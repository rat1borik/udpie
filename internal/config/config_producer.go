package config

import (
	"github.com/spf13/viper"
)

const (
	defaultSTUNLocalPort = 50000
	defaultSTUNTimeout   = 5
)

type ProducerConfig struct {
	Signaller ProducerSignallerConfig `mapstructure:"signaller"`
	STUN      STUNConfig              `mapstructure:"stun"`
}

type ProducerSignallerConfig struct {
	URL string `mapstructure:"url"`
}

type STUNConfig struct {
	Servers   []string `mapstructure:"servers"`
	LocalPort int      `mapstructure:"local_port"`
	Timeout   int      `mapstructure:"timeout"` // seconds
}

func LoadProducerConfig() (*ProducerConfig, error) {
	viper.SetConfigType("toml")
	viper.SetConfigName("config.producer")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("signaller.url", "http://localhost:8080")
	viper.SetDefault("stun.servers", []string{
		"stun.nextcloud.com:3478",
		"global.stun.twilio.com:3478",
		"stun.l.google.com:19302",
	})
	viper.SetDefault("stun.local_port", defaultSTUNLocalPort)
	viper.SetDefault("stun.timeout", defaultSTUNTimeout)

	// Read environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// If config file not found, use defaults
		// This is OK, we'll use default values
		_ = err // explicitly ignore error
	}

	var config ProducerConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
