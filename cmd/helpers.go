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
	teams "github.com/pzsp-teams/lib"
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

	teamsClientInstance, clientInitError = client.NewTeamsClient(ctx, authConfig, senderConfig)
	return teamsClientInstance, clientInitError
}

func newSenderConfig() *teams.SenderConfig {
	return &teams.SenderConfig{
		MaxRetries:     3,
		NextRetryDelay: 2,
		Timeout:        10,
	}
}

func loadAuthConfig() *teams.AuthConfig {
	_ = godotenv.Load() // Silently ignore if .env doesn't exist

	cfg := &teams.AuthConfig{
		ClientID:   getEnv("CLIENT_ID", ""),
		Tenant:     getEnv("TENANT_ID", ""),
		Email:      getEnv("EMAIL", ""),
		Scopes:     strings.Split(getEnv("SCOPES", "https://graph.microsoft.com/.default"), ","),
		AuthMethod: getEnv("AUTH_METHOD", "DEVICE_CODE"),
	}

	validateAuthConfig(cfg)
	return cfg
}

func validateAuthConfig(cfg *teams.AuthConfig) {
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
