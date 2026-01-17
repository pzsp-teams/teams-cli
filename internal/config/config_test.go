package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	lib "github.com/pzsp-teams/lib/config"
)

func TestGetDefaultConfigPath(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "json format",
			format: "json",
			want:   filepath.Join(xdg.ConfigHome, "teams-cli", "config.json"),
		},
		{
			name:   "yaml format",
			format: "yaml",
			want:   filepath.Join(xdg.ConfigHome, "teams-cli", "config.yaml"),
		},
		{
			name:   "toml format",
			format: "toml",
			want:   filepath.Join(xdg.ConfigHome, "teams-cli", "config.toml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDefaultConfigPath(tt.format)
			if got != tt.want {
				t.Errorf("GetDefaultConfigPath(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "json format",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "yaml format",
			format:  "yaml",
			wantErr: false,
		},
		{
			name:    "yml format",
			format:  "yml",
			wantErr: false,
		},
		{
			name:    "toml format",
			format:  "toml",
			wantErr: false,
		},
		{
			name:    "unsupported format",
			format:  "xml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateDefaultConfig(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDefaultConfig(%q) error = %v, wantErr %v", tt.format, err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 {
				t.Errorf("CreateDefaultConfig(%q) returned empty config", tt.format)
			}
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	t.Run("load json config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		configData := `{
			"auth": {
				"client_id": "test-client",
				"tenant_id": "test-tenant",
				"email": "test@example.com",
				"auth_method": "interactive"
			},
			"sender": {
				"max_retries": 5,
				"next_retry_delay": 3,
				"timeout": 15
			},
			"cache": {
				"mode": "sync",
				"path": "/tmp/cache"
			}
		}`

		if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("loadConfigFromFile() error = %v", err)
		}

		if cfg.Auth.ClientID != "test-client" {
			t.Errorf("cfg.Auth.ClientID = %q, want %q", cfg.Auth.ClientID, "test-client")
		}
		if cfg.Auth.TenantID != "test-tenant" {
			t.Errorf("cfg.Auth.TenantID = %q, want %q", cfg.Auth.TenantID, "test-tenant")
		}
		if cfg.Auth.Email != "test@example.com" {
			t.Errorf("cfg.Auth.Email = %q, want %q", cfg.Auth.Email, "test@example.com")
		}
		if cfg.Sender.MaxRetries != 5 {
			t.Errorf("cfg.Sender.MaxRetries = %d, want %d", cfg.Sender.MaxRetries, 5)
		}
	})

	t.Run("load yaml config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		configData := `auth:
  client_id: test-client
  tenant_id: test-tenant
  email: test@example.com
  auth_method: device_code
sender:
  max_retries: 5
  next_retry_delay: 3
  timeout: 15
cache:
  mode: async
  path: /tmp/cache
`

		if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("loadConfigFromFile() error = %v", err)
		}

		if cfg.Auth.ClientID != "test-client" {
			t.Errorf("cfg.Auth.ClientID = %q, want %q", cfg.Auth.ClientID, "test-client")
		}
		if cfg.Auth.AuthMethod != "device_code" {
			t.Errorf("cfg.Auth.AuthMethod = %q, want %q", cfg.Auth.AuthMethod, "device_code")
		}
	})

	t.Run("load toml config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")

		configData := `[auth]
client_id = "test-client"
tenant_id = "test-tenant"
email = "test@example.com"
auth_method = "interactive"

[sender]
max_retries = 5
next_retry_delay = 3
timeout = 15

[cache]
mode = "sync"
path = "/tmp/cache"
`

		if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigFromFile(configPath)
		if err != nil {
			t.Fatalf("loadConfigFromFile() error = %v", err)
		}

		if cfg.Auth.ClientID != "test-client" {
			t.Errorf("cfg.Auth.ClientID = %q, want %q", cfg.Auth.ClientID, "test-client")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := loadConfigFromFile("/nonexistent/config.json")
		if err == nil {
			t.Error("loadConfigFromFile() expected error for nonexistent file")
		}
	})
}

func TestLoadConfigsFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := `{
		"auth": {
			"client_id": "test-client",
			"tenant_id": "test-tenant",
			"email": "test@example.com",
			"auth_method": "interactive"
		},
		"sender": {
			"max_retries": 5,
			"next_retry_delay": 3,
			"timeout": 15
		},
		"cache": {
			"mode": "sync",
			"path": "/tmp/cache"
		}
	}`

	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatal(err)
	}

	authCfg, senderCfg, cacheCfg, err := LoadConfigsFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigsFromFile() error = %v", err)
	}

	if authCfg.ClientID != "test-client" {
		t.Errorf("authCfg.ClientID = %q, want %q", authCfg.ClientID, "test-client")
	}
	if authCfg.Tenant != "test-tenant" {
		t.Errorf("authCfg.Tenant = %q, want %q", authCfg.Tenant, "test-tenant")
	}
	if authCfg.Email != "test@example.com" {
		t.Errorf("authCfg.Email = %q, want %q", authCfg.Email, "test@example.com")
	}
	if authCfg.AuthMethod != lib.Interactive {
		t.Errorf("authCfg.AuthMethod = %v, want %v", authCfg.AuthMethod, lib.Interactive)
	}
	if len(authCfg.Scopes) == 0 {
		t.Error("authCfg.Scopes should not be empty")
	}

	if senderCfg.MaxRetries != 5 {
		t.Errorf("senderCfg.MaxRetries = %d, want %d", senderCfg.MaxRetries, 5)
	}
	if senderCfg.NextRetryDelay != 3 {
		t.Errorf("senderCfg.NextRetryDelay = %d, want %d", senderCfg.NextRetryDelay, 3)
	}
	if senderCfg.Timeout != 15 {
		t.Errorf("senderCfg.Timeout = %d, want %d", senderCfg.Timeout, 15)
	}

	if cacheCfg.Mode != lib.CacheSync {
		t.Errorf("cacheCfg.Mode = %v, want %v", cacheCfg.Mode, lib.CacheSync)
	}
	if cacheCfg.Path == nil || *cacheCfg.Path != "/tmp/cache" {
		t.Errorf("cacheCfg.Path = %v, want %q", cacheCfg.Path, "/tmp/cache")
	}
}

func TestLoadConfigsFromFile_MissingClientID(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := `{
		"auth": {
			"tenant_id": "test-tenant",
			"email": "test@example.com",
			"auth_method": "interactive"
		},
		"sender": {},
		"cache": {}
	}`

	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, _, err := LoadConfigsFromFile(configPath)
	if err == nil {
		t.Error("LoadConfigsFromFile() expected error for missing client_id")
	}
}

func TestLoadConfigsFromFile_MissingTenantID(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := `{
		"auth": {
			"client_id": "test-client",
			"email": "test@example.com",
			"auth_method": "interactive"
		},
		"sender": {},
		"cache": {}
	}`

	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, _, err := LoadConfigsFromFile(configPath)
	if err == nil {
		t.Error("LoadConfigsFromFile() expected error for missing tenant_id")
	}
}

func TestLoadConfigsFromFile_MissingEmail(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := `{
		"auth": {
			"client_id": "test-client",
			"tenant_id": "test-tenant",
			"auth_method": "interactive"
		},
		"sender": {},
		"cache": {}
	}`

	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, _, err := LoadConfigsFromFile(configPath)
	if err == nil {
		t.Error("LoadConfigsFromFile() expected error for missing email")
	}
}

func TestBuildSenderConfig(t *testing.T) {
	t.Run("with custom values", func(t *testing.T) {
		cfg := &config{
			Sender: senderSection{
				MaxRetries:     5,
				NextRetryDelay: 3,
				Timeout:        15,
			},
		}

		senderCfg := buildSenderConfig(cfg)
		if senderCfg.MaxRetries != 5 {
			t.Errorf("MaxRetries = %d, want %d", senderCfg.MaxRetries, 5)
		}
		if senderCfg.NextRetryDelay != 3 {
			t.Errorf("NextRetryDelay = %d, want %d", senderCfg.NextRetryDelay, 3)
		}
		if senderCfg.Timeout != 15 {
			t.Errorf("Timeout = %d, want %d", senderCfg.Timeout, 15)
		}
	})

	t.Run("with default values", func(t *testing.T) {
		cfg := &config{
			Sender: senderSection{},
		}

		senderCfg := buildSenderConfig(cfg)
		if senderCfg.MaxRetries != 3 {
			t.Errorf("MaxRetries = %d, want default %d", senderCfg.MaxRetries, 3)
		}
		if senderCfg.NextRetryDelay != 2 {
			t.Errorf("NextRetryDelay = %d, want default %d", senderCfg.NextRetryDelay, 2)
		}
		if senderCfg.Timeout != 10 {
			t.Errorf("Timeout = %d, want default %d", senderCfg.Timeout, 10)
		}
	})
}

func TestBuildCacheConfig(t *testing.T) {
	t.Run("async mode with path", func(t *testing.T) {
		cfg := &config{
			Cache: cacheSection{
				Mode: "async",
				Path: "/tmp/cache",
			},
		}

		cacheCfg := buildCacheConfig(cfg)
		if cacheCfg.Mode != lib.CacheAsync {
			t.Errorf("Mode = %v, want %v", cacheCfg.Mode, lib.CacheAsync)
		}
		if cacheCfg.Provider != lib.CacheProviderJSONFile {
			t.Errorf("Provider = %v, want %v", cacheCfg.Provider, lib.CacheProviderJSONFile)
		}
		if cacheCfg.Path == nil || *cacheCfg.Path != "/tmp/cache" {
			t.Errorf("Path = %v, want %q", cacheCfg.Path, "/tmp/cache")
		}
	})

	t.Run("sync mode without path", func(t *testing.T) {
		cfg := &config{
			Cache: cacheSection{
				Mode: "sync",
			},
		}

		cacheCfg := buildCacheConfig(cfg)
		if cacheCfg.Mode != lib.CacheSync {
			t.Errorf("Mode = %v, want %v", cacheCfg.Mode, lib.CacheSync)
		}
		if cacheCfg.Path != nil {
			t.Errorf("Path = %v, want nil", cacheCfg.Path)
		}
	})

	t.Run("defaults to async mode", func(t *testing.T) {
		cfg := &config{
			Cache: cacheSection{},
		}

		cacheCfg := buildCacheConfig(cfg)
		if cacheCfg.Mode != lib.CacheAsync {
			t.Errorf("Mode = %v, want default %v", cacheCfg.Mode, lib.CacheAsync)
		}
	})
}

func TestBuildAuthConfig(t *testing.T) {
	cfg := &config{
		Auth: authSection{
			ClientID:   "test-client",
			TenantID:   "test-tenant",
			Email:      "test@example.com",
			AuthMethod: "interactive",
		},
	}

	authCfg := buildAuthConfig(cfg)
	if authCfg.ClientID != "test-client" {
		t.Errorf("ClientID = %q, want %q", authCfg.ClientID, "test-client")
	}
	if authCfg.Tenant != "test-tenant" {
		t.Errorf("Tenant = %q, want %q", authCfg.Tenant, "test-tenant")
	}
	if authCfg.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", authCfg.Email, "test@example.com")
	}
	if authCfg.AuthMethod != lib.Interactive {
		t.Errorf("AuthMethod = %v, want %v", authCfg.AuthMethod, lib.Interactive)
	}
	if len(authCfg.Scopes) == 0 {
		t.Error("Scopes should not be empty")
	}
}

func TestValidateAuthConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &lib.AuthConfig{
			ClientID:   "test-client",
			Tenant:     "test-tenant",
			Email:      "test@example.com",
			AuthMethod: lib.Interactive,
		}

		err := validateAuthConfig(cfg)
		if err != nil {
			t.Errorf("validateAuthConfig() unexpected error = %v", err)
		}
	})

	t.Run("missing client_id", func(t *testing.T) {
		cfg := &lib.AuthConfig{
			Tenant:     "test-tenant",
			Email:      "test@example.com",
			AuthMethod: lib.Interactive,
		}

		err := validateAuthConfig(cfg)
		if err == nil {
			t.Error("validateAuthConfig() expected error for missing client_id")
		}
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		cfg := &lib.AuthConfig{
			ClientID:   "test-client",
			Email:      "test@example.com",
			AuthMethod: lib.Interactive,
		}

		err := validateAuthConfig(cfg)
		if err == nil {
			t.Error("validateAuthConfig() expected error for missing tenant_id")
		}
	})

	t.Run("missing email", func(t *testing.T) {
		cfg := &lib.AuthConfig{
			ClientID:   "test-client",
			Tenant:     "test-tenant",
			AuthMethod: lib.Interactive,
		}

		err := validateAuthConfig(cfg)
		if err == nil {
			t.Error("validateAuthConfig() expected error for missing email")
		}
	})
}

func TestParseAuthMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   lib.Method
	}{
		{
			name:   "interactive",
			method: "interactive",
			want:   lib.Interactive,
		},
		{
			name:   "device_code",
			method: "device_code",
			want:   lib.DeviceCode,
		},
		{
			name:   "unknown defaults to interactive",
			method: "unknown",
			want:   lib.Interactive,
		},
		{
			name:   "empty defaults to interactive",
			method: "",
			want:   lib.Interactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAuthMethod(tt.method)
			if got != tt.want {
				t.Errorf("parseAuthMethod(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}
