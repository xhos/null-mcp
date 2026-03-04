package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/server"

	"github.com/xhos/null-mcp/internal/config"
	"github.com/xhos/null-mcp/internal/gen/null/v1/nullv1connect"
	"github.com/xhos/null-mcp/internal/tools"
	"github.com/xhos/null-mcp/internal/version"
)

func main() {
	cfg := config.Load()

	logger := log.NewWithOptions(os.Stdout, log.Options{
		Prefix: "null-mcp",
		Level:  cfg.LogLevel,
	})

	logger.Debug("debug is enabled")
	logger.Debug(version.FullVersion())
	logger.Info("starting null-mcp...", "version", version.Version)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	opts := connect.WithInterceptors(authInterceptor(cfg.APIKey))

	h := tools.New(
		cfg.UserID.String(),
		nullv1connect.NewAccountServiceClient(httpClient, cfg.NullCoreURL, opts),
		nullv1connect.NewTransactionServiceClient(httpClient, cfg.NullCoreURL, opts),
		nullv1connect.NewCategoryServiceClient(httpClient, cfg.NullCoreURL, opts),
		nullv1connect.NewDashboardServiceClient(httpClient, cfg.NullCoreURL, opts),
		logger,
	)

	s := server.NewMCPServer("null-mcp", version.Version)
	h.Register(s)

	logger.Info("serving SSE", "addr", cfg.ListenAddress, "base_url", cfg.BaseURL)
	sse := server.NewSSEServer(s,
		server.WithBaseURL(cfg.BaseURL),
		server.WithKeepAlive(true),
		server.WithKeepAliveInterval(30*time.Second),
	)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debug("incoming request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr, "ua", r.UserAgent())
			sse.ServeHTTP(w, r)
		})
		errCh <- http.ListenAndServe(cfg.ListenAddress, handler)
	}()

	select {
	case <-sigCh:
		logger.Info("shutting down")
		os.Exit(0)
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	}
}

func authInterceptor(apiKey string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("X-Internal-Key", apiKey)
			return next(ctx, req)
		}
	}
}
