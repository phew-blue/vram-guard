package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	cfgPath := os.Getenv("VRAM_GUARD_CONFIG")
	if cfgPath == "" {
		cfgPath = "/etc/vram-guard/config.yaml"
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	guard := NewGuard(cfg)
	mux := buildMux(cfg, guard)

	addr := ":11434"
	log.Printf("vram-guard listening on %s, proxying to %s", addr, cfg.OllamaURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}
