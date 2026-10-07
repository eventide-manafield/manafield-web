package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eventide-manafield/manafield-web/internal/coreclient"
	shell "github.com/eventide-manafield/manafield-web/internal/web"
)

var version = "0.0.1-dev"

func main() {
	coreURL := envOr("MANAFIELD_CORE_URL", "http://127.0.0.1:8080")
	addr := listenAddr()

	client, err := coreclient.New(coreURL, &http.Client{Timeout: 3 * time.Second})
	if err != nil {
		slog.Error("invalid Core URL", "url", coreURL, "error", err)
		os.Exit(1)
	}

	basePath := envOr("MANAFIELD_WEB_BASE_PATH", "/")

	handler, err := shell.New(client, version, basePath)
	if err != nil {
		slog.Error("failed to initialize Web Shell", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("Manafield Web is listening", "addr", addr, "core", coreURL, "version", version)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Web Shell stopped", "error", err)
		os.Exit(1)
	}
}

func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func listenAddr() string {
	if addr := strings.TrimSpace(os.Getenv("MANAFIELD_WEB_ADDR")); addr != "" {
		return addr
	}

	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		if strings.HasPrefix(port, ":") {
			return port
		}
		return ":" + port
	}

	return ":8080"
}
