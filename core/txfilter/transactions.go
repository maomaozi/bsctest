// Copyright 2024 BSC
// Transaction confirmation and execution logic

package txfilter

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum"
)

func waitForBuyConfirmation(txHashes []common.Hash, tokenAddr common.Address) *big.Int {
	results := make(chan *big.Int, len(txHashes))
	done := make(chan struct{})

	for _, hash := range txHashes {
		go func(h common.Hash) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			receipt, err := waitForReceipt(ctx, h)
			if err != nil || receipt.Status != 1 {
				select {
				case results <- nil:
				case <-done:
				}
				return
			}

			balance := getTokenBalance(tokenAddr)
			select {
			case results <- balance:
			case <-done:
			}
		}(hash)
	}

	for i := 0; i < len(txHashes); i++ {
		balance := <-results
		if balance != nil && balance.Cmp(big.NewInt(0)) > 0 {
			close(done)
			return balance
		}
	}
	return nil
}

func waitForReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			receipt, err := globalClient.TransactionReceipt(context.Background(), hash)
			if err == nil {
				return receipt, nil
			}
		}
	}
}

func getTokenBalance(tokenAddr common.Address) *big.Int {
	data, _ := erc20ABI.Pack("balanceOf", crypto.PubkeyToAddress(globalConfig.PrivateKey.PublicKey))

	msg := ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}

	result, err := globalClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		return big.NewInt(0)
	}

	balance := new(big.Int).SetBytes(result)
	return balance
}

func sendApproveTx(tokenAddr common.Address, amount *big.Int, symbol string) bool {
	for attempt := 1; attempt <= 3; attempt++ {
		nonce := getNextNonce()
		log.Info("Sending Approve", "symbol", symbol, "attempt", attempt)

		data, _ := erc20ABI.Pack("approve", globalConfig.FourMemeContract, amount)
		tx := types.NewTransaction(nonce, tokenAddr, big.NewInt(0), globalConfig.GasLimit, globalConfig.GasPrice, data)
		signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(56)), globalConfig.PrivateKey)

		err := globalClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Error("Approve send failed", "symbol", symbol, "attempt", attempt, "err", err)
			syncNonce()
			continue
		}

		log.Info("Approve sent, waiting for confirmation", "symbol", symbol, "hash", signedTx.Hash().Hex())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		receipt, err := waitForReceipt(ctx, signedTx.Hash())
		cancel()

		if err == nil && receipt.Status == 1 {
			log.Info("Approve confirmed", "symbol", symbol)
			return true
		}

		log.Error("Approve failed", "symbol", symbol, "attempt", attempt)
		syncNonce()
	}

	log.Error("All Approve attempts failed", "symbol", symbol)
	return false
}

func sellTokenStep(tokenAddr common.Address, amount *big.Int, symbol string) bool {
	for attempt := 1; attempt <= 3; attempt++ {
		nonce := getNextNonce()
		log.Info("Sending Sell", "symbol", symbol, "attempt", attempt)

		data, _ := fourmemeABI.Pack("sellToken", tokenAddr, amount, big.NewInt(0))
		tx := types.NewTransaction(nonce, globalConfig.FourMemeContract, big.NewInt(0), globalConfig.GasLimit, globalConfig.GasPrice, data)
		signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(56)), globalConfig.PrivateKey)

		err := globalClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Error("Sell send failed", "symbol", symbol, "attempt", attempt, "err", err)
			syncNonce()
			time.Sleep(1 * time.Second)
			continue
		}

		log.Info("Sell sent, waiting for confirmation", "symbol", symbol, "hash", signedTx.Hash().Hex())

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		receipt, err := waitForReceipt(ctx, signedTx.Hash())
		cancel()

		if err == nil && receipt.Status == 1 {
			log.Info("Sell confirmed", "symbol", symbol)
			return true
		}

		log.Error("Sell failed", "symbol", symbol, "attempt", attempt)
		syncNonce()
		time.Sleep(1 * time.Second)
	}

	log.Error("All Sell attempts failed", "symbol", symbol)
	return false
}

var isTradingActive bool

func getNextNonce() uint64 {
	nonceMu.Lock()
	defer nonceMu.Unlock()
	current := globalNonce
	globalNonce++
	return current
}

func syncNonce() {
	nonceMu.Lock()
	defer nonceMu.Unlock()
	nonce, err := globalClient.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(globalConfig.PrivateKey.PublicKey))
	if err == nil {
		globalNonce = nonce
		log.Info("Nonce manually synced", "nonce", nonce)
	} else {
		log.Error("Nonce manual sync failed", "err", err)
	}
}

func asyncSyncNonce() {
	nonceMu.Lock()
	if isTradingActive {
		nonceMu.Unlock()
		return
	}
	nonceMu.Unlock()

	nonce, err := globalClient.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(globalConfig.PrivateKey.PublicKey))

	nonceMu.Lock()
	defer nonceMu.Unlock()
	if isTradingActive {
		return
	}
	if err == nil {
		globalNonce = nonce
	}
}

func startNonceSync() {
	go func() {
		for {
			time.Sleep(3 * time.Second)
			asyncSyncNonce()
		}
	}()
}
