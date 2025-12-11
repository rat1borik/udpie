package main

import (
	"fmt"
	"os"

	"udpie/internal/config"
)

// loadConfig loads producer config with defaults
func loadConfig() *config.ProducerConfig {
	cfg, err := config.LoadProducerConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config, using defaults: %v\n", err)
		// Use defaults if config file not found
		cfg = &config.ProducerConfig{
			Signaller: config.ProducerSignallerConfig{
				URL: "http://localhost:8080",
			},
			STUN: config.STUNConfig{
				Servers:   []string{"stun.nextcloud.com:3478", "global.stun.twilio.com:3478", "stun.l.google.com:19302"},
				LocalPort: 50000,
				Timeout:   5,
			},
		}
	}
	return cfg
}
