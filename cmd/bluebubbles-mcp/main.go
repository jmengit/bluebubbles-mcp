package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmengit/bluebubbles-mcp/internal/config"
	"github.com/jmengit/bluebubbles-mcp/internal/mcpserver"
)

type Config = config.Config

func loadConfig() (config.Config, error) {
	return config.Load()
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
	mcpHandler := mcpserver.New(cfg).Handler()
	mux := http.NewServeMux()
	mux.Handle("/mcp", bearer(cfg, mcpHandler))
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

func bearer(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if host := r.Host; len(cfg.AllowedHosts) > 0 {
			if _, ok := cfg.AllowedHosts[host]; !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		if origin := r.Header.Get("Origin"); origin != "" && len(cfg.AllowedOrigins) > 0 {
			if _, ok := cfg.AllowedOrigins[origin]; !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		got, want := r.Header.Get("Authorization"), "Bearer "+cfg.MCPAuthToken
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
