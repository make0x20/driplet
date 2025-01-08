package config

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// Config is main config struct
type Config struct {
	Global    GlobalConfig              `mapstructure:"Global"`
	Endpoints map[string]EndpointConfig `mapstructure:"Endpoints"`
}

// GlobalConfig is the global config struct
type GlobalConfig struct {
	BindAddress string `mapstructure:"BindAddress"`
	Port        int    `mapstructure:"Port"`
	LogFile     string `mapstructure:"LogFile"`
	LogLevel    string `mapstructure:"LogLevel"`
}

// EndpointConfig is the endpoint config struct
type EndpointConfig struct {
	Name      string `mapstructure:"Name"`
	APISecret string `mapstructure:"APISecret"`
	JWTSecret string `mapstructure:"JWTSecret"`
}

// NewWithPath creates a new config from the given path.
func NewWithPath(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := getDefaultConfig()
		data, err := toml.Marshal(cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating default config: %w", err)
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return nil, fmt.Errorf("error writing default config: %w", err)
		}
	}
	return loadConfig(configPath)
}

// loadConfig loads the config from the given path.
func loadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Only set defaults for Global settings
	v.SetDefault("Global.BindAddress", "0.0.0.0")
	v.SetDefault("Global.Port", 4719)
	v.SetDefault("Global.LogFile", "")
	v.SetDefault("Global.LogLevel", "normal")

	v.SetConfigFile(configPath)
	v.SetConfigType("toml")
	v.AutomaticEnv()
	v.SetEnvPrefix("DRIPLET")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// getDefaultConfig returns the default config.
func getDefaultConfig() *Config {
	return &Config{
		Global: GlobalConfig{
			BindAddress: "0.0.0.0",
			Port:        4719,
			LogFile:     "",
			LogLevel:    "normal",
		},
		Endpoints: map[string]EndpointConfig{
			"default": {
				Name:      "default",
				APISecret: "change-this-api-secret",
				JWTSecret: "change-this-jwt-secret",
			},
		},
	}
}
