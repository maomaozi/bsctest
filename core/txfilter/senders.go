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
		wsMu.Lock()
		if ws48Conn != nil {
			ws48Conn.Close()
			ws48Conn = nil
		}
		wsMu.Unlock()
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
		wsMu.Lock()
		if wsBloxConn != nil {
			wsBloxConn.Close()
			wsBloxConn = nil
		}
		wsMu.Unlock()
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
	// 48Club WS: 连接 + 接收循环 + 断线自动重连（对齐 Python _connect_loop_48）
	go func() {
		for {
			log.Info("Connecting to 48Club WS...")
			conn, _, err := websocket.DefaultDialer.Dial(globalConfig.WS48Club, nil)
			if err != nil {
				log.Warn("48Club WS connect failed", "err", err)
				time.Sleep(1 * time.Second)
				continue
			}
			conn.SetPingHandler(nil) // 使用默认 pong 响应
			wsMu.Lock()
			ws48Conn = conn
			wsMu.Unlock()
			log.Info("48Club WS connected")

			// 心跳 goroutine：每5秒发送 ping
			stopPing := make(chan struct{})
			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						wsMu.Lock()
						c := ws48Conn
						wsMu.Unlock()
						if c != nil {
							if err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(3*time.Second)); err != nil {
								log.Warn("48Club WS ping failed", "err", err)
							}
						}
					case <-stopPing:
						return
					}
				}
			}()

			// 接收循环：持续读取响应，检测断线
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Warn("48Club WS read error, reconnecting...", "err", err)
					break
				}
				log.Info("48Club Resp", "msg", string(msg))
			}

			close(stopPing)
			wsMu.Lock()
			ws48Conn = nil
			conn.Close()
			wsMu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()

	// BloxRoute WS: 连接 + 接收循环 + 断线自动重连（对齐 Python _connect_loop_bx）
	go func() {
		for {
			log.Info("Connecting to BloxRoute WS...")
			header := http.Header{}
			header.Add("Authorization", globalConfig.BloxRouteAuth)
			conn, _, err := websocket.DefaultDialer.Dial(globalConfig.WSBloxRoute, header)
			if err != nil {
				log.Warn("BloxRoute WS connect failed", "err", err)
				time.Sleep(1 * time.Second)
				continue
			}
			conn.SetPingHandler(nil)
			wsMu.Lock()
			wsBloxConn = conn
			wsMu.Unlock()
			log.Info("BloxRoute WS connected")

			stopPing := make(chan struct{})
			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						wsMu.Lock()
						c := wsBloxConn
						wsMu.Unlock()
						if c != nil {
							if err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(3*time.Second)); err != nil {
								log.Warn("BloxRoute WS ping failed", "err", err)
							}
						}
					case <-stopPing:
						return
					}
				}
			}()

			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Warn("BloxRoute WS read error, reconnecting...", "err", err)
					break
				}
				log.Info("BloxRoute Resp", "msg", string(msg))
			}

			close(stopPing)
			wsMu.Lock()
			wsBloxConn = nil
			conn.Close()
			wsMu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()

	// HTTP RPC 心跳：每20秒向 Blockrazor 和 Nodereal 发送心跳（对齐 Python session_heartbeat）
	go httpHeartbeat(globalConfig.BlockrazorRPC, "blockrazor", 20*time.Second)
	go httpHeartbeat(globalConfig.NoderealRPC, "nodereal", 20*time.Second)
}

func httpHeartbeat(url string, name string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		var payload []byte
		if name == "blockrazor" {
			payload, _ = json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "eth_sendBundle",
				"params":  []interface{}{map[string]interface{}{"txs": []string{}, "noMerge": true}},
			})
		} else {
			payload, _ = json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "eth_chainId",
				"params":  []interface{}{},
			})
		}

		resp, err := httpClient.Post(url, "application/json", bytes.NewReader(payload))
		if err != nil {
			log.Warn("Heartbeat failed", "name", name, "err", err)
			continue
		}
		if resp.StatusCode != 200 {
			log.Warn("Heartbeat non-200", "name", name, "status", resp.StatusCode)
		}
		resp.Body.Close()
	}
}
