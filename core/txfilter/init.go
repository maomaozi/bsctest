// Copyright 2024 BSC
// TxFilter initialization for geth integration

package txfilter

import (
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
)

var (
	defaultConfigPath = "txfilter.json"
	isInitialized     bool
)

// InitFromConfigFile initializes the bundle handler from config file
// Call this during geth startup
func InitFromConfigFile(configPath string) error {
	if isInitialized {
		return nil
	}

	if configPath == "" {
		configPath = defaultConfigPath
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Warn("TxFilter config not found, bundle trading disabled", "path", configPath)
		return nil
	}

	config, err := LoadConfigFromFile(configPath)
	if err != nil {
		log.Error("Failed to load txfilter config", "err", err)
		return err
	}

	if err := InitHandler(config, config.HTTPRPC); err != nil {
		log.Error("Failed to init txfilter handler", "err", err)
		return err
	}

	isInitialized = true
	log.Info("TxFilter bundle trading initialized", "config", configPath)
	return nil
}

// GetDefaultConfigPath returns the default config path
func GetDefaultConfigPath(datadir string) string {
	if datadir == "" {
		return defaultConfigPath
	}
	return filepath.Join(datadir, defaultConfigPath)
}
