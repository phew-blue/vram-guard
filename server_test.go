package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestMux creates a test mux with a single mock backend that handles both
// /api/ps (for the guard) and all other requests (as the Ollama proxy target).
func newTestMux(psModels []psModel, maxVRAMMb int, models map[string]ModelConfig) (http.Handler, func()) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/ps" {
			json.NewEncoder(w).Encode(psResponse{Models: psModels})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	cfg := &Config{
		OllamaURL: backend.URL,
		MaxVRAMMb: maxVRAMMb,
		Models:    models,
	}
	guard := NewGuard(cfg)
	mux := buildMux(cfg, guard)
	return mux, backend.Close
}

var defaultModels = map[string]ModelConfig{
	"llama3.2:3b": {VRAMMb: 2000},
}

func TestHealth(t *testing.T) {
	mux, cleanup := newTestMux(nil, 8192, defaultModels)
	defer cleanup()

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestGenerate_AllowlistReject(t *testing.T) {
	mux, cleanup := newTestMux(nil, 8192, defaultModels)
	defer cleanup()

	body := `{"model":"unknown:latest","prompt":"hi"}`
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body)))

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestGenerate_VRAMReject(t *testing.T) {
	// 7500 MB already loaded; llama3.2:3b needs 2000 MB; total 9500 > 8192
	loaded := []psModel{{Name: "gemma3:12b", SizeVRAM: 7500 * 1024 * 1024}}
	mux, cleanup := newTestMux(loaded, 8192, defaultModels)
	defer cleanup()

	body := `{"model":"llama3.2:3b","prompt":"hi"}`
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body)))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestGenerate_Allowed(t *testing.T) {
	mux, cleanup := newTestMux(nil, 8192, defaultModels)
	defer cleanup()

	body := `{"model":"llama3.2:3b","prompt":"hi"}`
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(body)))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestChat_AllowlistReject(t *testing.T) {
	mux, cleanup := newTestMux(nil, 8192, defaultModels)
	defer cleanup()

	body := `{"model":"bad:model","messages":[]}`
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(body)))

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestPull_AllowlistReject(t *testing.T) {
	mux, cleanup := newTestMux(nil, 8192, defaultModels)
	defer cleanup()

	body := `{"model":"malicious:latest"}`
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/pull", strings.NewReader(body)))

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestPull_Allowed(t *testing.T) {
	mux, cleanup := newTestMux(nil, 8192, defaultModels)
	defer cleanup()

	body := `{"model":"llama3.2:3b"}`
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/pull", strings.NewReader(body)))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestPassthrough_Tags(t *testing.T) {
	mux, cleanup := newTestMux(nil, 8192, defaultModels)
	defer cleanup()

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tags", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
