package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Config struct{ Mode, ListenAddr, MCPAuthToken string }

func loadConfig() (Config, error) {
	mode := os.Getenv("BLUEBUBBLES_MODE")
	if mode == "" {
		mode = "readonly"
	}
	if mode != "readonly" && mode != "interactive" {
		return Config{}, fmt.Errorf("BLUEBUBBLES_MODE must be readonly or interactive")
	}
	token := os.Getenv("MCP_AUTH_TOKEN")
	if token == "" {
		return Config{}, errors.New("MCP_AUTH_TOKEN is required")
	}
	addr := os.Getenv("MCP_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	return Config{Mode: mode, ListenAddr: addr, MCPAuthToken: token}, nil
}

type Health struct {
	Status string `json:"status"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("configuration invalid", "error", err)
		os.Exit(2)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	server := mcp.NewServer(&mcp.Implementation{Name: "bluebubbles-mcp", Version: "0.1.0"}, nil)
	readOnly, destructive := true, false
	mcp.AddTool(server, &mcp.Tool{Name: "ping", Description: "Return gateway availability and configured capability mode.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, DestructiveHint: &destructive, IdempotentHint: true, OpenWorldHint: &readOnly}}, pingTool(cfg))
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true, MaxRequestBodyBytes: 4 << 20})
	mux := http.NewServeMux()
	mux.Handle("/mcp", bearer(cfg.MCPAuthToken, mcpHandler))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Health{Status: "ok"})
	})
	httpServer := &http.Server{Addr: cfg.ListenAddr, Handler: requestLog(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()
	logger.Info("starting gateway", "addr", cfg.ListenAddr, "mode", cfg.Mode)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func pingTool(cfg Config) mcp.ToolHandlerFor[struct{}, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		b, _ := json.Marshal(map[string]string{"status": "ok", "mode": cfg.Mode})
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
	}
}

func bearer(expected string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, want := r.Header.Get("Authorization"), "Bearer "+expected
		if len(got) != len(want) || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
