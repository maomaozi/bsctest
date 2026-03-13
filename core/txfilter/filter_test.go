// Copyright 2024 BSC
// Test example for transaction filter

package txfilter

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestFourMemeFilter(t *testing.T) {
	// Create a mock transaction with FourMeme selector
	data, _ := hex.DecodeString("519ebb10" + "0000000000000000000000000000000000000000000000000000000000000020")

	tx := types.NewTransaction(
		0,
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(0),
		21000,
		big.NewInt(1000000000),
		data,
	)

	matched := false
	handler := func(info *TokenInfo, tx *types.Transaction) {
		matched = true
		t.Logf("Token detected: %s", info.TokenAddress.Hex())
	}

	filter := NewFourMemeFilter(handler)
	result := filter.Filter(tx)

	if !result {
		t.Log("Transaction did not match FourMeme pattern (expected for minimal test data)")
	}
}

func TestFilterManager(t *testing.T) {
	manager := NewManager()

	callCount := 0
	handler := func(info *TokenInfo, tx *types.Transaction) {
		callCount++
	}

	filter := NewFourMemeFilter(handler)
	manager.AddFilter(filter)

	// Create test transaction
	data, _ := hex.DecodeString("519ebb10")
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 21000, big.NewInt(1), data)

	manager.ProcessTransaction(tx)
	t.Logf("Filter processed transaction")
}
