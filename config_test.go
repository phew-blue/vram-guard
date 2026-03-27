package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `
ollama_url: http://localhost:11435
max_vram_mb: 8192
models:
  llama3.2:3b:
    vram_mb: 2000
`
	f, _ := os.CreateTemp("", "config*.yaml")
	f.WriteString(content)
	f.Close()
	defer os.Remove(f.Name())

	cfg, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.OllamaURL != "http://localhost:11435" {
		t.Errorf("OllamaURL = %q", cfg.OllamaURL)
	}
	if cfg.MaxVRAMMb != 8192 {
		t.Errorf("MaxVRAMMb = %d", cfg.MaxVRAMMb)
	}
}

func TestLoadConfig_MissingOllamaURL(t *testing.T) {
	f, _ := os.CreateTemp("", "config*.yaml")
	f.WriteString("max_vram_mb: 8192\n")
	f.Close()
	defer os.Remove(f.Name())

	_, err := LoadConfig(f.Name())
	if err == nil {
		t.Error("expected error for missing ollama_url")
	}
}

func TestLoadConfig_MissingMaxVRAM(t *testing.T) {
	f, _ := os.CreateTemp("", "config*.yaml")
	f.WriteString("ollama_url: http://localhost:11435\n")
	f.Close()
	defer os.Remove(f.Name())

	_, err := LoadConfig(f.Name())
	if err == nil {
		t.Error("expected error for missing max_vram_mb")
	}
}

func TestLookupModel_Found(t *testing.T) {
	cfg := &Config{
		Models: map[string]ModelConfig{
			"qwen2.5:7b": {VRAMMb: 4500},
		},
	}
	m, ok := cfg.LookupModel("qwen2.5:7b")
	if !ok {
		t.Fatal("expected model to be found")
	}
	if m.VRAMMb != 4500 {
		t.Errorf("VRAMMb = %d", m.VRAMMb)
	}
}

func TestLookupModel_NotFound(t *testing.T) {
	cfg := &Config{Models: map[string]ModelConfig{}}
	_, ok := cfg.LookupModel("unknown:latest")
	if ok {
		t.Error("expected model to not be found")
	}
}
