package main

import (
	"os"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoadConfig(t *testing.T) {
	path := writeTempConfig(t, `
ollama_url: http://localhost:11435
max_vram_mb: 8192
models:
  llama3.2:3b:
    vram_mb: 2000
`)
	cfg, err := LoadConfig(path)
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
	path := writeTempConfig(t, "max_vram_mb: 8192\n")
	_, err := LoadConfig(path)
	if err == nil {
		t.Error("expected error for missing ollama_url")
	}
}

func TestLoadConfig_MissingMaxVRAM(t *testing.T) {
	path := writeTempConfig(t, "ollama_url: http://localhost:11435\n")
	_, err := LoadConfig(path)
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
