// Copyright 2024 BSC
// Bundle transaction logic

package txfilter

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

func processTokenLifecycle(tokenAddr common.Address, symbol string, targetRawTx []byte) {
	nonceMu.Lock()
	isTradingActive = true
	nonceMu.Unlock()

	defer func() {
		nonceMu.Lock()
		isTradingActive = false
		nonceMu.Unlock()
	}()

	log.Info("Starting token lifecycle", "symbol", symbol)

	buyTxHashes := buyTokenStep(tokenAddr, targetRawTx)

	time.Sleep(time.Duration(globalConfig.SellDelaySeconds * float64(time.Second)))

	balance := waitForBuyConfirmation(buyTxHashes, tokenAddr)
	if balance == nil || balance.Cmp(big.NewInt(0)) == 0 {
		log.Error("Buy failed", "symbol", symbol)
		syncNonce()
		return
	}

	log.Info("Buy confirmed", "symbol", symbol, "balance", balance)

	if !sendApproveTx(tokenAddr, balance, symbol) {
		log.Error("Approve failed", "symbol", symbol)
		return
	}

	sellTokenStep(tokenAddr, balance, symbol)
	log.Info("Token lifecycle completed", "symbol", symbol)
}

func buyTokenStep(tokenAddr common.Address, targetRawTx []byte) []common.Hash {
	nonce := getNextNonce()
	value := new(big.Int).Add(globalConfig.BuyAmountBNB, globalConfig.BribeAmountBNB)

	results := make(chan common.Hash, 4)

	go func() {
		if hash := sendBundle48WS(targetRawTx, tokenAddr, nonce, value); hash != (common.Hash{}) {
			results <- hash
		}
	}()
	go func() {
		if hash := sendBundleBloxWS(targetRawTx, tokenAddr, nonce, value); hash != (common.Hash{}) {
			results <- hash
		}
	}()
	go func() {
		if hash := sendBundleBlockrazor(targetRawTx, tokenAddr, nonce, value); hash != (common.Hash{}) {
			results <- hash
		}
	}()
	go func() {
		if hash := sendBundleNodereal(targetRawTx, tokenAddr, nonce, value); hash != (common.Hash{}) {
			results <- hash
		}
	}()

	hashes := make([]common.Hash, 0, 4)
	timeout := time.After(2 * time.Second)
	for i := 0; i < 4; i++ {
		select {
		case h := <-results:
			hashes = append(hashes, h)
		case <-timeout:
			break
		}
	}
	return hashes
}
