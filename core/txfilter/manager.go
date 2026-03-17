// Copyright 2024 BSC
// Transaction filter manager for coordinating multiple filters

package txfilter

import (
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
)

// Manager coordinates multiple transaction filters
type Manager struct {
	filters      []TxFilter
	mu           sync.RWMutex
	processedTxs map[string]bool
}

func NewManager() *Manager {
	return &Manager{
		filters:      make([]TxFilter, 0),
		processedTxs: make(map[string]bool),
	}
}

// AddFilter registers a new filter
func (m *Manager) AddFilter(filter TxFilter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filters = append(m.filters, filter)
}

// ProcessTransaction runs all filters on a transaction
// Returns true if any filter matched
func (m *Manager) ProcessTransaction(tx *types.Transaction) bool {
	if !IsFilterEnabled() {
		return false
	}

	txHash := tx.Hash().Hex()

	m.mu.Lock()
	if m.processedTxs[txHash] {
		m.mu.Unlock()
		return false
	}
	m.processedTxs[txHash] = true
	if len(m.processedTxs) > 10000 {
		m.processedTxs = make(map[string]bool)
		m.processedTxs[txHash] = true
	}
	m.mu.Unlock()

	matched := false
	for _, filter := range m.filters {
		if filter.Filter(tx) {
			matched = true
		}
	}
	return matched
}

// ProcessTransactions processes a batch of transactions
func (m *Manager) ProcessTransactions(txs []*types.Transaction) {
	for _, tx := range txs {
		m.ProcessTransaction(tx)
	}
}
