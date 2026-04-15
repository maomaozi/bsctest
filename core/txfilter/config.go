// Copyright 2024 BSC
// Bundle trading configuration

package txfilter

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type BundleConfig struct {
	AddressesEnable  bool
	StartTime        string
	EndTime          string
	PrivateKey       *ecdsa.PrivateKey
	BuyAmountBNB     *big.Int
	BribeAmountBNB   *big.Int
	SellDelaySeconds float64
	GasLimit         uint64
	GasPrice         *big.Int

	BuyerContract    common.Address
	FourMemeContract common.Address

	Club48BribeAddr     common.Address
	BloxBribeAddr       common.Address
	BlockrazorBribeAddr common.Address
	NoderealBribeAddr   common.Address

	HTTPRPC       string
	BlockrazorRPC string
	NoderealRPC   string
	WS48Club      string
	WSBloxRoute   string
	BloxRouteAuth string

	TargetAddresses map[string]bool

	KeywordsEnable    bool
	KeywordsWsURL     string
	KeywordsTTLSeconds float64
}

type ConfigFile struct {
	AddressesEnable  bool     `json:"addresses_enable"`
	StartTime        string   `json:"start_time"`
	EndTime          string   `json:"end_time"`
	PrivateKey       string   `json:"private_key"`
	BuyAmountBNB     float64  `json:"buy_amount_bnb"`
	BribeAmountBNB   float64  `json:"bribe_amount_bnb"`
	SellDelaySeconds float64  `json:"sell_delay_seconds"`
	HTTPRPC          string   `json:"http_rpc"`
	TargetAddresses  []string `json:"target_addresses"`
	KeywordsEnable     bool    `json:"keywords_enable"`
	KeywordsWsURL      string  `json:"keywords_ws_url"`
	KeywordsTTLSeconds float64 `json:"keywords_ttl_seconds"`
}

func LoadConfigFromFile(configPath string) (*BundleConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg ConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := validateTimeRange(cfg.StartTime, cfg.EndTime); err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	targetAddrs := make(map[string]bool)
	for _, addr := range cfg.TargetAddresses {
		targetAddrs[common.HexToAddress(addr).Hex()] = true
	}

	return &BundleConfig{
		AddressesEnable:  cfg.AddressesEnable,
		StartTime:        cfg.StartTime,
		EndTime:          cfg.EndTime,
		PrivateKey:       privateKey,
		BuyAmountBNB:     toBNBWei(cfg.BuyAmountBNB),
		BribeAmountBNB:   toBNBWei(cfg.BribeAmountBNB),
		SellDelaySeconds: cfg.SellDelaySeconds,
		GasLimit:         600000,
		GasPrice:         big.NewInt(60000000),

		BuyerContract:    common.HexToAddress("0xd72c5102a98742cd47074f6c2fb55c88db7a31e7"),
		FourMemeContract: common.HexToAddress("0x5c952063c7fc8610FFDB798152D69F0B9550762b"),

		Club48BribeAddr:     common.HexToAddress("0x4848489f0b2BEdd788c696e2D79b6b69D7484848"),
		BloxBribeAddr:       common.HexToAddress("0x6374Ca2da5646C73Eb444aB99780495d61035f9b"),
		BlockrazorBribeAddr: common.HexToAddress("0x1266c6bE60392a8ff346e8D5EcCd3e69dD9c5F20"),
		NoderealBribeAddr:   common.HexToAddress("0xffffFFFfFFffffffffffffffFfFFFfffFFFfFFfE"),

		HTTPRPC:       cfg.HTTPRPC,
		BlockrazorRPC: "https://frankfurt.builder.blockrazor.io",
		NoderealRPC:   "https://bsc-mainnet-builder.nodereal.io",
		WS48Club:      "wss://puissant-builder.48.club/",
		WSBloxRoute:   "wss://api.blxrbdn.com/ws",
		BloxRouteAuth: "ZjBkN2QxZTQtZTVhMi00NGIyLTk2MzUtZGI0M2EyZjM5YWNmOmEyNGUzMjVhYzdkZTQ2NzQ0ODM5Njk5YTdhZWMzZWJk",

		TargetAddresses: targetAddrs,

		KeywordsEnable:    cfg.KeywordsEnable,
		KeywordsWsURL:     cfg.KeywordsWsURL,
		KeywordsTTLSeconds: cfg.KeywordsTTLSeconds,
	}, nil
}

func validateTimeRange(startTime, endTime string) error {
	if startTime == "" || endTime == "" {
		return nil
	}
	startH, startM, err := parseTimeString(startTime)
	if err != nil {
		return fmt.Errorf("invalid start_time: %v", err)
	}
	endH, endM, err := parseTimeString(endTime)
	if err != nil {
		return fmt.Errorf("invalid end_time: %v", err)
	}
	startMin := startH*60 + startM
	endMin := endH*60 + endM
	if startMin >= endMin {
		return fmt.Errorf("end_time must be greater than start_time")
	}
	return nil
}

func parseTimeString(timeStr string) (int, int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("format must be HH:MM")
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour")
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute")
	}
	return hour, minute, nil
}

func toBNBWei(bnb float64) *big.Int {
	wei := new(big.Float).Mul(big.NewFloat(bnb), big.NewFloat(1e18))
	result, _ := wei.Int(nil)
	return result
}
