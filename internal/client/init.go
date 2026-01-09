package client

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"

	teams_config "github.com/pzsp-teams/lib/config"
)

var (
	teamsClientInstance *TeamsClient
	clientInitError     error
)

// GetOrCreateInstance initializes or returns the existing TeamsClient instance
// Single-threaded application, no need for sync.Once
func GetOrCreateInstance(ctx context.Context) (*TeamsClient, error) {
	if teamsClientInstance != nil {
		return teamsClientInstance, nil
	}
	if clientInitError != nil {
		return nil, clientInitError
	}

	authConfig := loadAuthConfig()
	senderConfig := newSenderConfig()
	cacheConfig := newCacheConfig()

	teamsClientInstance, clientInitError = NewTeamsClient(ctx, authConfig, senderConfig, cacheConfig)
	return teamsClientInstance, clientInitError
}

func newSenderConfig() *teams_config.SenderConfig {
	return &teams_config.SenderConfig{
		MaxRetries:     3,
		NextRetryDelay: 2,
		Timeout:        10,
	}
}

func newCacheConfig() *teams_config.CacheConfig {
	return &teams_config.CacheConfig{
		Mode:     teams_config.CacheAsync,
		Provider: teams_config.CacheProviderJSONFile,
		Path:     nil,
	}
}

func loadAuthConfig() *teams_config.AuthConfig {
	_ = godotenv.Load()

	cfg := &teams_config.AuthConfig{
		ClientID:   getEnv("CLIENT_ID", ""),
		Tenant:     getEnv("TENANT_ID", ""),
		Email:      getEnv("EMAIL", ""),
		Scopes:     strings.Split(getEnv("SCOPES", "https://graph.microsoft.com/.default"), ","),
		AuthMethod: getAuthMethod(),
	}

	validateAuthConfig(cfg)
	return cfg
}

func validateAuthConfig(cfg *teams_config.AuthConfig) {
	if cfg.ClientID == "" {
		fmt.Fprintf(os.Stderr, "Error: Missing CLIENT_ID environment variable\n")
		os.Exit(1)
	}
	if cfg.Tenant == "" {
		fmt.Fprintf(os.Stderr, "Error: Missing TENANT_ID environment variable\n")
		os.Exit(1)
	}
	if cfg.Email == "" {
		fmt.Fprintf(os.Stderr, "Error: Missing EMAIL environment variable\n")
		os.Exit(1)
	}
	if cfg.AuthMethod != "DEVICE_CODE" && cfg.AuthMethod != "INTERACTIVE" {
		fmt.Fprintf(os.Stderr, "Error: AUTH_METHOD must be either DEVICE_CODE or INTERACTIVE\n")
		os.Exit(1)
	}
}

func getAuthMethod() teams_config.Method {
	switch getEnv("AUTH_METHOD", "DEVICE_CODE") {
	case "INTERACTIVE":
		return teams_config.Interactive
	default:
		return teams_config.DeviceCode
	}
}

func getEnv(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}
