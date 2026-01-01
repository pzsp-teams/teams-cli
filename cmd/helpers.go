package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"github.com/pzsp-teams/cli/internal/client"
	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/templates"
	teams_config "github.com/pzsp-teams/lib/config"
)

var (
	teamsClientInstance *client.TeamsClient
	clientInitError     error
)

func getDecodeFunc(extension string) (file_readers.DecodeFunc, error) {
	switch extension {
	case "json":
		return file_readers.DecodeJSON, nil
	case "yaml", "yml":
		return file_readers.DecodeYAML, nil
	case "toml":
		return file_readers.DecodeTOML, nil
	case "csv":
		return file_readers.DecodeCSV, nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s (supported: json, yaml, yml, toml, csv)", extension)
	}
}

func GetOrCreateTeamsClient(ctx context.Context) (*client.TeamsClient, error) {
	if teamsClientInstance != nil {
		return teamsClientInstance, nil
	}
	if clientInitError != nil {
		return nil, clientInitError
	}

	authConfig := loadAuthConfig()
	senderConfig := newSenderConfig()
	cacheConfig := newCacheConfig()

	teamsClientInstance, clientInitError = client.NewTeamsClient(ctx, authConfig, senderConfig, cacheConfig)
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

func parseTemplateAndData(templatePath, dataPath string) (map[string]string, error) {
	templateFile, err := os.Open(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template file: %w", err)
	}

	dataFile, err := os.Open(dataPath)
	if err != nil {
		_ = templateFile.Close()
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}

	extension := strings.TrimPrefix(filepath.Ext(dataPath), ".")
	parser, err := getDecodeFunc(extension)
	if err != nil {
		_ = templateFile.Close()
		_ = dataFile.Close()
		return nil, err
	}

	messageParser, err := templates.NewMessageParser(templateFile, dataFile, parser)
	_ = templateFile.Close()
	_ = dataFile.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to create message parser: %w", err)
	}

	return messageParser.Parse()
}

func createMessagesFromString(message string, recipients []string) map[string]string {
	messages := make(map[string]string, len(recipients))
	for _, recipient := range recipients {
		messages[recipient] = message
	}
	return messages
}

func createMessagesFromFile(filePath string, recipients []string) (map[string]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read message file: %w", err)
	}

	message := string(content)
	messages := make(map[string]string, len(recipients))
	for _, recipient := range recipients {
		messages[recipient] = message
	}
	return messages, nil
}
