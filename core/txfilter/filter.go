// Copyright 2024 BSC
// Transaction pre-filter for custom transaction processing

package txfilter

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

const (
	CreateFourMemeTokenSelector = "519ebb10"
	Creator3946                 = "0x757eba15a64468e6535532fcf093cef90e226f85"
	Creator5546                 = "0xc6496e138af13c0026e14ffbd32eae6764eab8b2"
	InitHash3946                = "0x3eb722ec5d79ddc2f52880ea62f1b7e7d95c66d4ae0dfe32f988ca9eca52b359"
	InitHash5546                = "0xa410105265035fe66e02e324fbc5f3d69c891340e6644808247a1cb49e4d5da0"
)

type TokenInfo struct {
	TokenAddress common.Address
	Name         string
	Symbol       string
	QuoteAddress common.Address
}

// TxFilter interface for modular transaction filtering
type TxFilter interface {
	Filter(tx *types.Transaction) bool
}

// FourMemeFilter detects FourMeme token creation transactions
type FourMemeFilter struct {
	fourMemeContract common.Address
	targetAddresses  map[string]bool
	addressesEnable  bool
	handler          func(*TokenInfo, *types.Transaction)
}

func NewFourMemeFilter(fourMemeContract common.Address, targetAddresses map[string]bool, addressesEnable bool, handler func(*TokenInfo, *types.Transaction)) *FourMemeFilter {
	return &FourMemeFilter{
		fourMemeContract: fourMemeContract,
		targetAddresses:  targetAddresses,
		addressesEnable:  addressesEnable,
		handler:          handler,
	}
}

// Filter checks if transaction is a FourMeme token creation
func (f *FourMemeFilter) Filter(tx *types.Transaction) bool {

	// Check to address
	if tx.To() == nil || tx.To().Hex() != f.fourMemeContract.Hex() {
		return false
	}

	data := tx.Data()
	if len(data) < 4 {
		return false
	}

	selector := hex.EncodeToString(data[:4])
	if selector != CreateFourMemeTokenSelector {
		return false
	}

	tokenInfo := predictTokenAddress(data)
	if tokenInfo == nil {
		return false
	}

	// Check from address
	from, err := types.Sender(types.NewEIP155Signer(big.NewInt(56)), tx)
	if err != nil {
		return false
	}

	log.Info("FourMeme token detected",
	"txHash", tx.Hash().Hex(),
	"token", tokenInfo.TokenAddress.Hex(),
	"symbol", tokenInfo.Symbol)

	inTargetList := f.addressesEnable && f.targetAddresses[from.Hex()]
	if inTargetList && !IsInTimeRange() {
		return false
	}

	keywordMatched := false
	matchedKw := ""
	if IsKeywordsEnabled() && tokenInfo.Name != "" {
		keywordMatched, matchedKw = MatchKeywords(tokenInfo.Name)
		if keywordMatched {
			log.Info("Token matched by keyword", "name", tokenInfo.Name, "symbol", tokenInfo.Symbol, "keyword", matchedKw)
		}
	}

	if !inTargetList && !keywordMatched {
		return false
	}

	if f.handler != nil {
		f.handler(tokenInfo, tx)
	}

	return true
}

func predictTokenAddress(inputData []byte) *TokenInfo {
	// Python: input_data_hex[202:] where input_data_hex includes "0x" prefix
	// "0x" (2 chars) + selector (8 chars) + data (192 chars) = 202 chars
	// In Go bytes: selector (4 bytes) + data (96 bytes) = 100 bytes
	// So we skip first 100 bytes to match Python's [202:]
	if len(inputData) < 100 {
		return nil
	}

	hexContent := inputData[100:]
	totalBytes := len(hexContent)

	if totalBytes < 320 {
		return nil
	}

	words := make([]*big.Int, 0, totalBytes/32)
	for i := 0; i < len(hexContent); i += 32 {
		if i+32 > len(hexContent) {
			break
		}
		word := new(big.Int).SetBytes(hexContent[i : i+32])
		words = append(words, word)
	}

	if len(words) < 10 || words[0].Uint64() != 32 {
		return nil
	}

	isOffset := func(v *big.Int) bool {
		val := v.Uint64()
		return val >= 64 && val%32 == 0 && val < uint64(totalBytes)
	}

	if !isOffset(words[3]) || !isOffset(words[4]) {
		return nil
	}

	salt := words[2]
	tokenID := words[9]

	var quoteAddr common.Address
	if len(words) > 8 {
		quoteAddr = common.BigToAddress(words[8])
	}

	creatorType := new(big.Int).Rsh(tokenID, 10).And(tokenID, big.NewInt(0x3f)).Uint64()

	var creator common.Address
	var initHash common.Hash

	if creatorType == 4 {
		creator = common.HexToAddress(Creator5546)
		initHash = common.HexToHash(InitHash5546)
	} else {
		creator = common.HexToAddress(Creator3946)
		initHash = common.HexToHash(InitHash3946)
	}

	saltBytes := common.LeftPadBytes(salt.Bytes(), 32)
	packedData := append([]byte{0xff}, creator.Bytes()...)
	packedData = append(packedData, saltBytes...)
	packedData = append(packedData, initHash.Bytes()...)

	hash := crypto.Keccak256(packedData)
	tokenAddress := common.BytesToAddress(hash[12:])

	decodeStr := func(offsetWord *big.Int) string {
		startIdx := int((32 + offsetWord.Uint64()) * 2)
		if startIdx+64 > len(hex.EncodeToString(hexContent)) {
			return ""
		}

		hexStr := hex.EncodeToString(hexContent)
		if startIdx+64 > len(hexStr) {
			return ""
		}

		lenHex := hexStr[startIdx : startIdx+64]
		strLen, _ := new(big.Int).SetString(lenHex, 16)
		if strLen == nil {
			return ""
		}

		contentStart := startIdx + 64
		contentEnd := contentStart + int(strLen.Uint64()*2)
		if contentEnd > len(hexStr) {
			return ""
		}

		strBytes, err := hex.DecodeString(hexStr[contentStart:contentEnd])
		if err != nil {
			return ""
		}
		return string(strBytes)
	}

	return &TokenInfo{
		TokenAddress: tokenAddress,
		Name:         decodeStr(words[3]),
		Symbol:       decodeStr(words[4]),
		QuoteAddress: quoteAddr,
	}
}
