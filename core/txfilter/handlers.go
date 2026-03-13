// Copyright 2024 BSC
// Transaction filter handlers

package txfilter

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"
)

var (
	globalConfig     *BundleConfig
	globalClient     *ethclient.Client
	globalNonce      uint64
	nonceMu          sync.Mutex
	ws48Conn         *websocket.Conn
	wsBloxConn       *websocket.Conn
	wsMu             sync.Mutex
	buyerABI         abi.ABI
	erc20ABI         abi.ABI
	fourmemeABI      abi.ABI
	globalFilter     *FourMemeFilter
	httpClient       *http.Client
)

func InitHandler(config *BundleConfig, rpcURL string) error {
	globalConfig = config

	// Delay RPC connection in background
	go func() {
		for i := 0; i < 30; i++ {
			time.Sleep(2 * time.Second)
			client, err := ethclient.Dial(rpcURL)
			if err != nil {
				log.Debug("Waiting for RPC to be ready", "attempt", i+1, "err", err)
				continue
			}
			globalClient = client

			nonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(config.PrivateKey.PublicKey))
			if err != nil {
				log.Debug("Failed to get nonce", "err", err)
				continue
			}
			globalNonce = nonce
			log.Info("TxFilter RPC connected", "nonce", nonce)

			startNonceSync()
			go maintainWSConnections()
			break
		}
	}()

	httpClient = &http.Client{Timeout: 10 * time.Second}

	buyerABI, _ = abi.JSON(bytes.NewReader([]byte(buyerABIJSON)))
	erc20ABI, _ = abi.JSON(bytes.NewReader([]byte(erc20ABIJSON)))
	fourmemeABI, _ = abi.JSON(bytes.NewReader([]byte(fourmemeABIJSON)))

	globalFilter = NewFourMemeFilter(config.FourMemeContract, config.TargetAddresses, FourMemeHandler)

	return nil
}

// FourMemeHandler handles FourMeme token creation events
func FourMemeHandler(info *TokenInfo, tx *types.Transaction) {
	log.Info("FourMeme token start to handle",
		"txHash", tx.Hash().Hex(),
		"token", info.TokenAddress.Hex(),
		"symbol", info.Symbol)

	targetRawTx, err := tx.MarshalBinary()
	if err != nil {
		log.Error("Failed to marshal tx", "err", err)
		return
	}

	go processTokenLifecycle(info.TokenAddress, info.Symbol, targetRawTx)
}

// GetFilter returns the initialized filter for use in geth integration
func GetFilter() *FourMemeFilter {
	return globalFilter
}
