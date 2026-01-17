package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
	lib "github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/teams-cli/internal/file_readers"
	"gopkg.in/yaml.v3"
)

// defaultScopes contains the default OAuth scopes for Microsoft Teams
var defaultScopes = []string{
	"openid",
	"profile",
	"User.Read",
	"Team.ReadBasic.All",
	"Channel.ReadBasic.All",
	"ChannelSettings.ReadWrite.All",
	"Channel.Create",
	"Channel.Delete.All",
	"ChannelMessage.Send",
	"Team.Create",
	"TeamSettings.ReadWrite.All",
	"Group.ReadWrite.All",
}

type config struct {
	Auth   authSection   `json:"auth" yaml:"auth" toml:"auth"`
	Sender senderSection `json:"sender" yaml:"sender" toml:"sender"`
	Cache  cacheSection  `json:"cache" yaml:"cache" toml:"cache"`
}

type authSection struct {
	ClientID   string `json:"client_id" yaml:"client_id" toml:"client_id"`
	TenantID   string `json:"tenant_id" yaml:"tenant_id" toml:"tenant_id"`
	Email      string `json:"email" yaml:"email" toml:"email"`
	AuthMethod string `json:"auth_method" yaml:"auth_method" toml:"auth_method"`
}

type senderSection struct {
	MaxRetries     int `json:"max_retries" yaml:"max_retries" toml:"max_retries"`
	NextRetryDelay int `json:"next_retry_delay" yaml:"next_retry_delay" toml:"next_retry_delay"`
	Timeout        int `json:"timeout" yaml:"timeout" toml:"timeout"`
}

type cacheSection struct {
	Mode string `json:"mode" yaml:"mode" toml:"mode"`
	Path string `json:"path,omitempty" yaml:"path,omitempty" toml:"path,omitempty"`
}

var configFormats = []string{
	"toml",
	"json",
	"yaml",
	"yml",
}

// GetDefaultConfigPath returns the default configuration file path for the given format
func GetDefaultConfigPath(format string) string {
	configDir := filepath.Join(xdg.ConfigHome, "teams-cli")
	filename := fmt.Sprintf("config.%s", format)
	return filepath.Join(configDir, filename)
}

// FindDefaultConfigFile looks for a config file in the current directory
// Returns empty string if no config file is found
func FindDefaultConfigFile() string {
	for _, format := range configFormats {
		path := GetDefaultConfigPath(format)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func loadConfigFromFile(configFilePath string) (*config, error) {
	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	extension := strings.TrimPrefix(filepath.Ext(configFilePath), ".")
	decoder, err := file_readers.GetDecoderByExtension(extension)
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	var cfg config
	if err := decoder(file, &cfg); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	_ = file.Close()
	return &cfg, nil
}

// LoadConfigsFromFile loads all configuration from a file and returns the individual configs
func LoadConfigsFromFile(configFilePath string) (*lib.AuthConfig, *lib.SenderConfig, *lib.CacheConfig, error) {
	cfg, err := loadConfigFromFile(configFilePath)
	if err != nil {
		return nil, nil, nil, err
	}

	authConfig := buildAuthConfig(cfg)
	senderConfig := buildSenderConfig(cfg)
	cacheConfig := buildCacheConfig(cfg)

	if err := validateAuthConfig(authConfig); err != nil {
		return nil, nil, nil, err
	}

	return authConfig, senderConfig, cacheConfig, nil
}

// CreateDefaultConfig creates a default configuration with empty values
func CreateDefaultConfig(format string) ([]byte, error) {
	cfg := &config{
		Auth: authSection{
			ClientID:   "",
			TenantID:   "",
			Email:      "",
			AuthMethod: "interactive",
		},
		Sender: senderSection{
			MaxRetries:     3,
			NextRetryDelay: 2,
			Timeout:        10,
		},
		Cache: cacheSection{
			Mode: "async",
			Path: "",
		},
	}

	switch format {
	case "json":
		return json.MarshalIndent(cfg, "", "  ")
	case "yaml", "yml":
		return yaml.Marshal(cfg)
	case "toml":
		var buf strings.Builder
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(cfg); err != nil {
			return nil, err
		}
		return []byte(buf.String()), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func buildSenderConfig(cfg *config) *lib.SenderConfig {
	maxRetries := cfg.Sender.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}
	nextRetryDelay := cfg.Sender.NextRetryDelay
	if nextRetryDelay == 0 {
		nextRetryDelay = 2
	}
	timeout := cfg.Sender.Timeout
	if timeout == 0 {
		timeout = 10
	}

	return &lib.SenderConfig{
		MaxRetries:     maxRetries,
		NextRetryDelay: nextRetryDelay,
		Timeout:        timeout,
	}
}

func buildCacheConfig(cfg *config) *lib.CacheConfig {
	mode := lib.CacheAsync
	if cfg.Cache.Mode == "sync" {
		mode = lib.CacheSync
	}

	provider := lib.CacheProviderJSONFile

	var path *string
	if cfg.Cache.Path != "" {
		path = &cfg.Cache.Path
	}

	return &lib.CacheConfig{
		Mode:     mode,
		Provider: provider,
		Path:     path,
	}
}

func buildAuthConfig(cfg *config) *lib.AuthConfig {
	return &lib.AuthConfig{
		ClientID:   cfg.Auth.ClientID,
		Tenant:     cfg.Auth.TenantID,
		Email:      cfg.Auth.Email,
		Scopes:     defaultScopes,
		AuthMethod: parseAuthMethod(cfg.Auth.AuthMethod),
	}
}

func validateAuthConfig(cfg *lib.AuthConfig) error {
	if cfg.ClientID == "" {
		return fmt.Errorf("error: missing client_id")
	}
	if cfg.Tenant == "" {
		return fmt.Errorf("error: missing tenant_id")
	}
	if cfg.Email == "" {
		return fmt.Errorf("error: missing email")
	}
	if cfg.AuthMethod != lib.DeviceCode && cfg.AuthMethod != lib.Interactive {
		return fmt.Errorf("error: auth_method must be either device_code or interactive")
	}
	return nil
}

func parseAuthMethod(method string) lib.Method {
	switch method {
	case "interactive":
		return lib.Interactive
	case "device_code":
		return lib.DeviceCode
	default:
		return lib.Interactive
	}
}
