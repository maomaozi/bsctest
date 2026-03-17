package txfilter

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"
)

const (
	kwWsPingInterval   = 20 * time.Second
	kwWsPongWait       = 30 * time.Second
	kwWsReconnectDelay = 3 * time.Second
	kwMsgChanSize      = 4096
	kwMaxStaleSeconds  = 5
	kwTTL              = 3 * time.Second
	kwCleanInterval    = 1 * time.Second
)

var (
	activeKeywords   = make(map[string]time.Time)
	keywordsMu       sync.RWMutex
	keywordsStopOnce sync.Once
	keywordsStopCh   chan struct{}
)

type keywordsPushMessage struct {
	Type        string   `json:"type"`
	Keywords    []string `json:"keywords"`
	PublishTime int64    `json:"publishTime"`
}

func startKeywordsClient(wsURL string) {
	keywordsStopCh = make(chan struct{})
	log.Info("Keywords WS client starting", "url", wsURL)
	for {
		select {
		case <-keywordsStopCh:
			log.Info("Keywords WS client stopped")
			return
		default:
		}
		runKeywordsConnection(wsURL)
		log.Info("Keywords WS reconnecting", "delay", kwWsReconnectDelay)
		select {
		case <-keywordsStopCh:
			return
		case <-time.After(kwWsReconnectDelay):
		}
	}
}

func runKeywordsConnection(wsURL string) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Error("Keywords WS connect failed", "err", err)
		return
	}
	defer conn.Close()

	log.Info("Keywords WS connected", "url", wsURL)

	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(kwWsPongWait))
		return nil
	})
	conn.SetReadDeadline(time.Now().Add(kwWsPongWait))

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(kwWsPingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(
					websocket.PingMessage, nil,
					time.Now().Add(5*time.Second),
				); err != nil {
					log.Error("Keywords WS ping failed", "err", err)
					return
				}
			case <-done:
				return
			case <-keywordsStopCh:
				return
			}
		}
	}()

	msgCh := make(chan []byte, kwMsgChanSize)
	go consumeKeywordMessages(msgCh)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			close(done)
			close(msgCh)
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Info("Keywords WS server closed normally")
			} else {
				log.Warn("Keywords WS read error", "err", err)
			}
			return
		}
		select {
		case msgCh <- msg:
		default:
			log.Warn("Keywords message channel full, dropping")
		}
	}
}

func consumeKeywordMessages(msgCh <-chan []byte) {
	for raw := range msgCh {
		var probe struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &probe); err != nil {
			continue
		}

		switch probe.Type {
		case "ready":
			log.Info("Keywords WS server ready")
		case "keywords":
			var m keywordsPushMessage
			if err := json.Unmarshal(raw, &m); err != nil {
				log.Error("Keywords message parse failed", "err", err)
				continue
			}
			age := time.Now().Unix() - m.PublishTime
			if age > kwMaxStaleSeconds || age < -kwMaxStaleSeconds {
				log.Info("Keywords message stale, dropping", "keywords", m.Keywords, "age", age)
				continue
			}
			storeKeywords(m.Keywords)
		}
	}
}

func storeKeywords(keywords []string) {
	now := time.Now()
	keywordsMu.Lock()
	for _, kw := range keywords {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw != "" {
			activeKeywords[kw] = now
		}
	}
	keywordsMu.Unlock()
	log.Info("Keywords stored", "keywords", keywords, "active_count", len(activeKeywords))
}

func cleanExpiredKeywords() {
	ticker := time.NewTicker(kwCleanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			keywordsMu.Lock()
			for kw, t := range activeKeywords {
				if now.Sub(t) > kwTTL {
					delete(activeKeywords, kw)
				}
			}
			keywordsMu.Unlock()
		case <-keywordsStopCh:
			return
		}
	}
}

// MatchKeywords checks if tokenName contains any active keyword (case-insensitive).
func MatchKeywords(tokenName string) (bool, string) {
	lower := strings.ToLower(tokenName)
	keywordsMu.RLock()
	defer keywordsMu.RUnlock()
	for kw := range activeKeywords {
		if strings.Contains(lower, kw) {
			return true, kw
		}
	}
	return false, ""
}

// IsKeywordsEnabled checks if keywords filtering is enabled in current config.
func IsKeywordsEnabled() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return currentConfig != nil && currentConfig.KeywordsEnable
}
