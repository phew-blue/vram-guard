package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ModelConfig struct {
	VRAMMb int `yaml:"vram_mb"`
}

type Config struct {
	OllamaURL string                 `yaml:"ollama_url"`
	MaxVRAMMb int                    `yaml:"max_vram_mb"`
	Models    map[string]ModelConfig `yaml:"models"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.OllamaURL == "" {
		return nil, fmt.Errorf("ollama_url is required")
	}
	if cfg.MaxVRAMMb == 0 {
		return nil, fmt.Errorf("max_vram_mb is required")
	}
	return &cfg, nil
}

func (c *Config) LookupModel(name string) (ModelConfig, bool) {
	m, ok := c.Models[name]
	return m, ok
}
