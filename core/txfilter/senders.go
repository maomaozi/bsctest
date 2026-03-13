// Copyright 2024 BSC
// Bundle sender implementations

package txfilter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"
)

func sendBundle48WS(targetRawTx []byte, tokenAddr common.Address, nonce uint64, value *big.Int) common.Hash {
	wsMu.Lock()
	conn := ws48Conn
	wsMu.Unlock()

	if conn == nil {
		log.Error("48Club WS not connected")
		return common.Hash{}
	}

	data, _ := buyerABI.Pack("buyWithBNBAndBribeTo", tokenAddr, big.NewInt(0), big.NewInt(0), globalConfig.Club48BribeAddr, globalConfig.BribeAmountBNB)
	tx := types.NewTransaction(nonce, globalConfig.BuyerContract, value, globalConfig.GasLimit, globalConfig.GasPrice, data)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(56)), globalConfig.PrivateKey)
	myRawTx, _ := signedTx.MarshalBinary()

	tx0Hash := crypto.Keccak256(targetRawTx)
	tx2Hash := crypto.Keccak256(myRawTx)
	signature, _ := crypto.Sign(crypto.Keccak256(append(tx0Hash, tx2Hash...)), globalConfig.PrivateKey)

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "eth_sendBundle",
		"params": []interface{}{
			map[string]interface{}{
				"txs":      []string{fmt.Sprintf("0x%x", targetRawTx), fmt.Sprintf("0x%x", myRawTx)},
				"48spSign": fmt.Sprintf("0x%x", signature),
			},
		},
	}

	if err := conn.WriteJSON(payload); err != nil {
		log.Error("Failed to send 48Club bundle", "err", err)
		return common.Hash{}
	}

	log.Info("Sent bundle to 48Club")
	return signedTx.Hash()
}

func sendBundleBloxWS(targetRawTx []byte, tokenAddr common.Address, nonce uint64, value *big.Int) common.Hash {
	wsMu.Lock()
	conn := wsBloxConn
	wsMu.Unlock()

	if conn == nil {
		log.Error("BloxRoute WS not connected")
		return common.Hash{}
	}

	data, _ := buyerABI.Pack("buyWithBNBAndBribeTo", tokenAddr, big.NewInt(0), big.NewInt(0), globalConfig.BloxBribeAddr, globalConfig.BribeAmountBNB)
	tx := types.NewTransaction(nonce, globalConfig.BuyerContract, value, globalConfig.GasLimit, globalConfig.GasPrice, data)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(56)), globalConfig.PrivateKey)
	myRawTx, _ := signedTx.MarshalBinary()

	payload := map[string]interface{}{
		"id":     "1",
		"method": "blxr_submit_bundle",
		"params": map[string]interface{}{
			"transaction":              []string{fmt.Sprintf("%x", targetRawTx), fmt.Sprintf("%x", myRawTx)},
			"blockchain_network":       "BSC-Mainnet",
			"mev_builders":             map[string]string{"all": ""},
			"backrunme_reward_address": crypto.PubkeyToAddress(globalConfig.PrivateKey.PublicKey).Hex(),
		},
	}

	if err := conn.WriteJSON(payload); err != nil {
		log.Error("Failed to send BloxRoute bundle", "err", err)
		return common.Hash{}
	}

	log.Info("Sent bundle to BloxRoute")
	return signedTx.Hash()
}

func sendBundleBlockrazor(targetRawTx []byte, tokenAddr common.Address, nonce uint64, value *big.Int) common.Hash {
	data, _ := buyerABI.Pack("buyWithBNBAndBribeTo", tokenAddr, big.NewInt(0), big.NewInt(0), globalConfig.BlockrazorBribeAddr, globalConfig.BribeAmountBNB)
	tx := types.NewTransaction(nonce, globalConfig.BuyerContract, value, globalConfig.GasLimit, globalConfig.GasPrice, data)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(56)), globalConfig.PrivateKey)
	myRawTx, _ := signedTx.MarshalBinary()

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "eth_sendBundle",
		"params": []interface{}{
			map[string]interface{}{
				"txs": []string{fmt.Sprintf("0x%x", targetRawTx), fmt.Sprintf("0x%x", myRawTx)},
			},
		},
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(globalConfig.BlockrazorRPC, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Error("Blockrazor bundle failed", "err", err)
		return common.Hash{}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if _, ok := result["result"]; ok {
		log.Info("Blockrazor bundle submitted")
		return signedTx.Hash()
	}

	log.Error("Blockrazor bundle error", "result", result)
	return common.Hash{}
}

func sendBundleNodereal(targetRawTx []byte, tokenAddr common.Address, nonce uint64, value *big.Int) common.Hash {
	data, _ := buyerABI.Pack("buyWithBNBAndBribeTo", tokenAddr, big.NewInt(0), big.NewInt(0), globalConfig.NoderealBribeAddr, globalConfig.BribeAmountBNB)
	tx := types.NewTransaction(nonce, globalConfig.BuyerContract, value, globalConfig.GasLimit, globalConfig.GasPrice, data)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(56)), globalConfig.PrivateKey)
	myRawTx, _ := signedTx.MarshalBinary()

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "eth_sendBundle",
		"params": []interface{}{
			map[string]interface{}{
				"txs": []string{fmt.Sprintf("0x%x", targetRawTx), fmt.Sprintf("0x%x", myRawTx)},
			},
		},
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(globalConfig.NoderealRPC, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Error("Nodereal bundle failed", "err", err)
		return common.Hash{}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if _, ok := result["result"]; ok {
		log.Info("Nodereal bundle submitted")
		return signedTx.Hash()
	}

	log.Error("Nodereal bundle error", "result", result)
	return common.Hash{}
}

func maintainWSConnections() {
	go func() {
		for {
			wsMu.Lock()
			if ws48Conn == nil {
				conn, _, err := websocket.DefaultDialer.Dial(globalConfig.WS48Club, nil)
				if err == nil {
					ws48Conn = conn
					log.Info("48Club WS connected")
				}
			}
			wsMu.Unlock()
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for {
			wsMu.Lock()
			if wsBloxConn == nil {
				header := http.Header{}
				header.Add("Authorization", globalConfig.BloxRouteAuth)
				conn, _, err := websocket.DefaultDialer.Dial(globalConfig.WSBloxRoute, header)
				if err == nil {
					wsBloxConn = conn
					log.Info("BloxRoute WS connected")
				}
			}
			wsMu.Unlock()
			time.Sleep(5 * time.Second)
		}
	}()
}
