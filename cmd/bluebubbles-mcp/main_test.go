package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadConfigDefaultsToReadonly(t *testing.T) {
	t.Setenv("BLUEBUBBLES_MODE", "")
	t.Setenv("BLUEBUBBLES_URL", "https://bluebubbles.example")
	t.Setenv("BLUEBUBBLES_PASSWORD", "password")
	t.Setenv("MCP_AUTH_TOKEN", "token")
	t.Setenv("MCP_LISTEN_ADDR", "")
	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != "readonly" {
		t.Fatalf("mode=%q", cfg.Mode)
	}
	if cfg.ListenAddr != "127.0.0.1:8080" {
		t.Fatalf("addr=%q", cfg.ListenAddr)
	}
}

func TestLoadConfigHeadersJSONAndAllowedOrigin(t *testing.T) {
	t.Setenv("BLUEBUBBLES_MODE", "readonly")
	t.Setenv("BLUEBUBBLES_URL", "http://bluebubbles.lan:1234")
	t.Setenv("BLUEBUBBLES_PASSWORD", "password")
	t.Setenv("MCP_AUTH_TOKEN", "token")
	t.Setenv("BLUEBUBBLES_HEADERS_JSON", `{"CF-Access-Client-Id":"id","CF-Access-Client-Secret":"secret"}`)
	t.Setenv("MCP_ALLOWED_HOSTS", "mcp.lan:8080,mcp.example.test")
	t.Setenv("MCP_ALLOWED_ORIGINS", "https://mcp.example.test")
	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BlueBubblesHeaders["CF-Access-Client-Id"] != "id" {
		t.Fatal("headers JSON not loaded")
	}
	if _, ok := cfg.AllowedOrigins["https://mcp.example.test"]; !ok {
		t.Fatal("origin not loaded")
	}
}

func TestLoadConfigRejectsUnknownMode(t *testing.T) {
	t.Setenv("BLUEBUBBLES_MODE", "unsafe")
	t.Setenv("MCP_AUTH_TOKEN", "token")
	if _, err := loadConfig(); err == nil {
		t.Fatal("expected mode error")
	}
}

func TestBearer(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(http.StatusNoContent) })
	cfg := Config{MCPAuthToken: "secret"}
	h := bearer(cfg, next)
	for _, tc := range []struct {
		name, auth string
		want       int
		call       bool
	}{
		{"missing", "", http.StatusUnauthorized, false},
		{"wrong", "Bearer nope", http.StatusUnauthorized, false},
		{"valid", "Bearer secret", http.StatusNoContent, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called = false
			r := httptest.NewRequest(http.MethodPost, "/mcp", nil)
			r.Header.Set("Authorization", tc.auth)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Errorf("%s: status=%d", tc.name, w.Code)
			}
			if called != tc.call {
				t.Errorf("%s: called=%v", tc.name, called)
			}
		})
	}
}

func TestHealthz(t *testing.T) {
	_ = os.Setenv("BLUEBUBBLES_MODE", "readonly")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	if req.URL.Path != "/healthz" {
		t.Fatal("bad request")
	}
}
