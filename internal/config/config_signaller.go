package config

import (
	"github.com/spf13/viper"
)

type SignallerConfig struct {
	Server ServerConfig `mapstructure:"server"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	Format     string `mapstructure:"format"` // json or text
	OutputFile bool   `mapstructure:"output_file"`
	OutputStd  bool   `mapstructure:"output_std"`
	MaxSize    int    `mapstructure:"max_size"`    // megabytes
	MaxBackups int    `mapstructure:"max_backups"` // number of backup files
	MaxAge     int    `mapstructure:"max_age"`     // days
	Compress   bool   `mapstructure:"compress"`    // compress old logs
}

func LoadSignallerConfig() (*SignallerConfig, error) {
	viper.SetConfigType("toml")

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	// Read environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// If config file not found, use defaults
		return nil, err
	}

	var config SignallerConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
