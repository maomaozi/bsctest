// Copyright 2024 BSC
// Bundle trading configuration

package txfilter

import (
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type BundleConfig struct {
	PrivateKey       *ecdsa.PrivateKey
	BuyAmountBNB     *big.Int
	BribeAmountBNB   *big.Int
	SellDelaySeconds int
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
}

type ConfigFile struct {
	PrivateKey       string   `json:"private_key"`
	BuyAmountBNB     float64  `json:"buy_amount_bnb"`
	BribeAmountBNB   float64  `json:"bribe_amount_bnb"`
	SellDelaySeconds int      `json:"sell_delay_seconds"`
	HTTPRPC          string   `json:"http_rpc"`
	TargetAddresses  []string `json:"target_addresses"`
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

	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	targetAddrs := make(map[string]bool)
	for _, addr := range cfg.TargetAddresses {
		targetAddrs[common.HexToAddress(addr).Hex()] = true
	}

	return &BundleConfig{
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
	}, nil
}

func toBNBWei(bnb float64) *big.Int {
	wei := new(big.Float).Mul(big.NewFloat(bnb), big.NewFloat(1e18))
	result, _ := wei.Int(nil)
	return result
}
