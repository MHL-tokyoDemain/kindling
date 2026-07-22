package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kindling/kindling/internal/parser"
	"github.com/kindling/kindling/pkg/types"
)

type Config struct {
	Port            int
	CredentialsFile string
	ProjectID       string
	MaxFileSize     int64
	Concurrency     int
	LogLevel        string
}

type Server struct {
	cfg        Config
	fw         firestoreWriter
	httpSrv    *http.Server
	mux        *http.ServeMux
	startTime  time.Time
	shutdownCh chan struct{}
}

func New(cfg Config, fw firestoreWriter) (*Server, error) {
	if cfg.Port <= 0 {
		cfg.Port = 9876
	}
	if cfg.MaxFileSize <= 0 {
		cfg.MaxFileSize = 1048576
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	setupLogging(cfg.LogLevel)

	mux := http.NewServeMux()
	s := &Server{
		cfg:        cfg,
		fw:         fw,
		mux:        mux,
		startTime:  time.Now(),
		shutdownCh: make(chan struct{}),
	}

	s.registerRoutes()

	s.httpSrv = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", cfg.Port),
		Handler:           s.corsMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s, nil
}

func setupLogging(level string) {
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slogLevel,
	})))
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.HealthHandler)
	s.mux.HandleFunc("POST /upload", HandleUpload(s.fw, parser.Parse, s.cfg))
	s.mux.HandleFunc("POST /shutdown", s.handleShutdown)
	s.mux.HandleFunc("POST /auth", s.handleAuth)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start() error {
	slog.Info("starting server", "addr", s.httpSrv.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("shutting down due to signal", "signal", sig.String())
	case <-s.shutdownCh:
		slog.Info("shutting down via /shutdown")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.Shutdown(ctx)
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("shutting down server")
	if s.fw != nil {
		if err := s.fw.Close(); err != nil {
			slog.Warn("error closing Firestore client", "error", err)
		}
	}
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	slog.Info("shutdown requested via /shutdown")

	WriteJSON(w, http.StatusOK, types.ShutdownResponse{
		Success: true,
		Status:  "shutting_down",
	})

	select {
	case s.shutdownCh <- struct{}{}:
	default:
	}
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusNotImplemented, map[string]string{"status": "not_implemented"})
}
