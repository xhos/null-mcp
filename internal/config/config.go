package config

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type Config struct {
	NullCoreURL string
	APIKey      string

	UserID        uuid.UUID
	ListenAddress string
	BaseURL       string

	LogFormat string // "json" | "text"
	LogLevel  log.Level
}

// parseAddress handles "8080", ":8080", or "127.0.0.1:8080".
func parseAddress(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ":") {
		return s
	}
	return ":" + s
}

func Load() Config {
	nullCoreURL := os.Getenv("NULL_CORE_URL")
	if nullCoreURL == "" {
		panic("NULL_CORE_URL environment variable is required")
	}

	apiKey := os.Getenv("NULL_API_KEY")
	if apiKey == "" {
		panic("NULL_API_KEY environment variable is required")
	}

	userID := os.Getenv("NULL_MCP_USER_ID")
	if userID == "" {
		panic("NULL_MCP_USER_ID environment variable is required")
	}

	baseURL := strings.TrimRight(os.Getenv("NULL_MCP_BASE_URL"), "/")
	if baseURL == "" {
		panic("NULL_MCP_BASE_URL environment variable is required")
	}

	listenAddress := os.Getenv("NULL_MCP_LISTEN_ADDRESS")
	if listenAddress == "" {
		listenAddress = "127.0.0.1:55553"
	}

	logLevel, err := log.ParseLevel(os.Getenv("NULL_MCP_LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}

	logFormat := strings.ToLower(strings.TrimSpace(os.Getenv("NULL_MCP_LOG_FORMAT")))
	if logFormat != "json" {
		logFormat = "text"
	}

	return Config{
		NullCoreURL:   nullCoreURL,
		APIKey:        apiKey,
		UserID:        uuid.MustParse(userID),
		ListenAddress: parseAddress(listenAddress),
		BaseURL:       baseURL,
		LogLevel:      logLevel,
		LogFormat:     logFormat,
	}
}
