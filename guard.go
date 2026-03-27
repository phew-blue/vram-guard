package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type psModel struct {
    Name     string `json:"name"`
    SizeVRAM int64  `json:"size_vram"`
}

type psResponse struct {
    Models []psModel `json:"models"`
}

type Guard struct {
    cfg        *Config
    httpClient *http.Client
}

func NewGuard(cfg *Config) *Guard {
    return &Guard{cfg: cfg, httpClient: &http.Client{}}
}

// CheckAllowlist returns true if the model is in the configured allowlist.
func (g *Guard) CheckAllowlist(model string) bool {
    _, ok := g.cfg.LookupModel(model)
    return ok
}

// CheckVRAM returns true if loading the model would not exceed the VRAM budget.
// Fails open: if /api/ps is unreachable, allows the request through.
func (g *Guard) CheckVRAM(model string) (bool, error) {
    modelCfg, ok := g.cfg.LookupModel(model)
    if !ok {
        return false, fmt.Errorf("model %q not in config", model)
    }

    resp, err := g.httpClient.Get(g.cfg.OllamaURL + "/api/ps")
    if err != nil {
        return true, nil // fail open
    }
    defer resp.Body.Close()

    var ps psResponse
    if err := json.NewDecoder(resp.Body).Decode(&ps); err != nil {
        return true, nil // fail open on parse error
    }

    var loadedMB int64
    for _, m := range ps.Models {
        loadedMB += m.SizeVRAM / (1024 * 1024)
    }

    return int(loadedMB)+modelCfg.VRAMMb <= g.cfg.MaxVRAMMb, nil
}
