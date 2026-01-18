package client

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/teams-cli/internal/config"
)

var (
	teamsClientInstance *Client
	clientInitError     error
)

// Initialize initializes the TeamsClient instance with the provided config path
func Initialize(ctx context.Context, configPath string) error {
	if teamsClientInstance != nil {
		return nil
	}
	if clientInitError != nil {
		return clientInitError
	}

	if configPath == "" {
		configPath = config.FindDefaultConfigFile()
	}

	if configPath == "" {
		clientInitError = fmt.Errorf("no configuration file found. Please create one using 'teams-cli config init'")
		return clientInitError
	}

	authConfig, senderConfig, cacheConfig, err := config.LoadConfigsFromFile(configPath)
	if err != nil {
		clientInitError = fmt.Errorf("failed to load config from file %s: %w", configPath, err)
		return clientInitError
	}

	teamsClientInstance, clientInitError = NewClient(ctx, authConfig, senderConfig, cacheConfig)
	return clientInitError
}

// GetInstance returns the existing TeamsClient instance or an error if not initialized
func GetInstance() (*Client, error) {
	if teamsClientInstance != nil {
		return teamsClientInstance, nil
	}
	if clientInitError != nil {
		return nil, clientInitError
	}
	return nil, fmt.Errorf("teams client not initialized")
}

// SetInstance overrides the current TeamsClient instance (useful for tests).
func SetInstance(instance *Client) {
	teamsClientInstance = instance
	clientInitError = nil
}

// ResetInstance clears the current TeamsClient instance and error.
func ResetInstance() {
	teamsClientInstance = nil
	clientInitError = nil
}
