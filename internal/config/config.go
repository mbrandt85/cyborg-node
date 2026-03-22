package config

import (
	"encoding/json"
	"os"
)

type GitToken struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	URLPattern string `json:"url_pattern"`
	Token      string `json:"token"`
}

type Config struct {
	GatewayURL  string     `json:"gateway_url"`
	NodeName    string     `json:"node_name"`
	NodeKey     string     `json:"node_key"`
	GitTokens   []GitToken `json:"git_tokens"`
	SearchPaths []string   `json:"search_paths"` // Paths to scan for git repos
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
