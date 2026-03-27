package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func buildMux(cfg *Config, guard *Guard) http.Handler {
	target, err := url.Parse(cfg.OllamaURL)
	if err != nil {
		log.Fatalf("parse ollama_url: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /api/pull", func(w http.ResponseWriter, r *http.Request) {
		model, err := extractModel(r)
		if err != nil || !guard.CheckAllowlist(model) {
			http.Error(w, "model not allowed", http.StatusForbidden)
			return
		}
		proxy.ServeHTTP(w, r)
	})

	for _, ep := range []string{"/api/generate", "/api/chat", "/api/embed"} {
		ep := ep
		mux.HandleFunc("POST "+ep, func(w http.ResponseWriter, r *http.Request) {
			model, err := extractModel(r)
			if err != nil || !guard.CheckAllowlist(model) {
				http.Error(w, "model not allowed", http.StatusForbidden)
				return
			}
			fits, _ := guard.CheckVRAM(model)
			if !fits {
				http.Error(w, "insufficient VRAM", http.StatusServiceUnavailable)
				return
			}
			proxy.ServeHTTP(w, r)
		})
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	return mux
}

// extractModel reads the "model" field from the JSON request body and restores
// the body so it can be forwarded to Ollama unchanged.
func extractModel(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var payload struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	return strings.TrimSpace(payload.Model), nil
}
