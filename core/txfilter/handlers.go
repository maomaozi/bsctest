// Copyright 2024 BSC
// Transaction filter handlers

package txfilter

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// FourMemeHandler handles FourMeme token creation events
func FourMemeHandler(info *TokenInfo, tx *types.Transaction) {
	log.Info("FourMeme token creation detected, ready to handle",
		"txHash", tx.Hash().Hex(),
		"token", info.TokenAddress.Hex(),
		"name", info.Name,
		"symbol", info.Symbol,
		"quote", info.QuoteAddress.Hex())
}
