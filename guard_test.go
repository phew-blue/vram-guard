package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func mockPSServer(models []psModel) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(psResponse{Models: models})
    }))
}

func TestCheckAllowlist_Allowed(t *testing.T) {
    cfg := &Config{Models: map[string]ModelConfig{"qwen2.5:7b": {VRAMMb: 4500}}}
    g := NewGuard(cfg)
    if !g.CheckAllowlist("qwen2.5:7b") {
        t.Error("expected allowed model to pass")
    }
}

func TestCheckAllowlist_Denied(t *testing.T) {
    cfg := &Config{Models: map[string]ModelConfig{}}
    g := NewGuard(cfg)
    if g.CheckAllowlist("malicious:latest") {
        t.Error("expected unknown model to fail")
    }
}

func TestCheckVRAM_Fits(t *testing.T) {
    srv := mockPSServer(nil) // nothing loaded
    defer srv.Close()

    cfg := &Config{
        OllamaURL: srv.URL,
        MaxVRAMMb: 8192,
        Models:    map[string]ModelConfig{"llama3.2:3b": {VRAMMb: 2000}},
    }
    g := NewGuard(cfg)

    fits, err := g.CheckVRAM("llama3.2:3b")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !fits {
        t.Error("expected fits=true with nothing loaded")
    }
}

func TestCheckVRAM_Overflow(t *testing.T) {
    // 7000 MB already loaded + 7500 MB for gemma3:12b = 14500 > 8192
    loaded := []psModel{{Name: "qwen2.5:7b", SizeVRAM: 7000 * 1024 * 1024}}
    srv := mockPSServer(loaded)
    defer srv.Close()

    cfg := &Config{
        OllamaURL: srv.URL,
        MaxVRAMMb: 8192,
        Models:    map[string]ModelConfig{"gemma3:12b": {VRAMMb: 7500}},
    }
    g := NewGuard(cfg)

    fits, err := g.CheckVRAM("gemma3:12b")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if fits {
        t.Error("expected fits=false with loaded model + requested exceeding budget")
    }
}

func TestCheckVRAM_OllamaUnreachable(t *testing.T) {
    // When /api/ps is unreachable, allow through (Ollama itself will fail)
    cfg := &Config{
        OllamaURL: "http://127.0.0.1:19999", // nothing listening
        MaxVRAMMb: 8192,
        Models:    map[string]ModelConfig{"llama3.2:3b": {VRAMMb: 2000}},
    }
    g := NewGuard(cfg)

    fits, err := g.CheckVRAM("llama3.2:3b")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !fits {
        t.Error("expected fits=true when ps unreachable (fail open)")
    }
}
