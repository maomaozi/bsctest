// Copyright 2024 BSC

package txfilter

import (
	"time"

	"github.com/ethereum/go-ethereum/log"
)

func startConfigReloader(path string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		config, err := LoadConfigFromFile(path)
		if err != nil {
			log.Error("Failed to reload txfilter config", "err", err)
			continue
		}

		configMu.Lock()
		currentConfig = config
		globalConfig = config
		globalFilter = NewFourMemeFilter(config.FourMemeContract, config.TargetAddresses, config.AddressesEnable, FourMemeHandler)
		configMu.Unlock()

		log.Info("TxFilter config reloaded",
			"addresses_enable", config.AddressesEnable,
			"start_time", config.StartTime,
			"end_time", config.EndTime,
			"sell_delay_seconds", config.SellDelaySeconds,
			"target_addresses", len(config.TargetAddresses),
			"keywords_enable", config.KeywordsEnable)
	}
}

func IsFilterEnabled() bool {
	configMu.RLock()
	defer configMu.RUnlock()

	if currentConfig == nil {
		return false
	}

	return currentConfig.AddressesEnable || currentConfig.KeywordsEnable
}

func IsInTimeRange() bool {
	configMu.RLock()
	defer configMu.RUnlock()

	if currentConfig == nil || currentConfig.StartTime == "" || currentConfig.EndTime == "" {
		return true
	}

	return isInTimeRange(currentConfig.StartTime, currentConfig.EndTime)
}

func isInTimeRange(startTime, endTime string) bool {
	now := time.Now()
	currentMin := now.Hour()*60 + now.Minute()

	startH, startM, _ := parseTimeString(startTime)
	endH, endM, _ := parseTimeString(endTime)

	startMin := startH*60 + startM
	endMin := endH*60 + endM

	return currentMin >= startMin && currentMin < endMin
}
