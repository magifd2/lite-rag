// Package server implements the HTTP server for the lite-rag serve subcommand.
// It exposes a POST /api/ask endpoint (SSE streaming) and GET /api/status,
// and serves an embedded Web UI from the static/ directory.
package server

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"lite-rag/internal/config"
	"lite-rag/internal/database"
	"lite-rag/internal/llm"
)

// version is set by the serve command from the binary's build-time version var.
var version = "dev"

// Server holds the shared dependencies for all HTTP handlers.
type Server struct {
	db     *database.DB
	client *llm.Client
	cfg    *config.Config
	mu     sync.RWMutex // guards db access: RLock for ask, Lock for future writes
}

// New creates a Server. db must remain open for the lifetime of the server.
func New(db *database.DB, client *llm.Client, cfg *config.Config, v string) *Server {
	version = v
	return &Server{db: db, client: client, cfg: cfg}
}

// Start registers routes, serves static files, and blocks until SIGINT/SIGTERM.
// addr should be a host:port string, e.g. "127.0.0.1:8080".
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/ask", s.handleAsk)
	mux.HandleFunc("/api/status", s.handleStatus)

	// Static Web UI — strip the "static/" prefix from the embedded FS
	uiFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("server: embed sub: %w", err)
	}
	mux.Handle("/", http.FileServerFS(uiFS))

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server started", "addr", addr, "url", "http://"+addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-quit:
		slog.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}
