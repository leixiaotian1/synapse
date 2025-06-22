// internal/config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ProviderConfig struct {
	Name         string `yaml:"name"`
	APIKeyEnv    string `yaml:"api_key_env"`
	BaseURL      string `yaml:"base_url"`
	DefaultModel string `yaml:"default_model"`
	APIKey       string `yaml:"-"` // 不从文件读取，从 env 加载
}

type ServerConfig struct {
	MCP struct {
		Address string `yaml:"address"`
	} `yaml:"mcp"`
}

type Config struct {
	LogLevel       string                    `yaml:"log_level"`
	Providers      map[string]ProviderConfig `yaml:"providers"`
	ActiveProvider string                    `yaml:"active_provider"`
	Server         ServerConfig              `yaml:"server"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 加载 API keys
	for name, p := range cfg.Providers {
		p.APIKey = os.Getenv(p.APIKeyEnv)
		cfg.Providers[name] = p
	}

	return &cfg, nil
}
