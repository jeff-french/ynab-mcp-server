package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for the YNAB MCP server
type Config struct {
	YNABToken     string
	TransportMode string
	HTTPPort      int
	HTTPHost      string
	MCPAuthToken  string
	LogLevel      string
}

// Load reads configuration from multiple sources with precedence:
// CLI flags > environment variables > config file > defaults
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("transport_mode", "stdio")
	v.SetDefault("http_port", 8080)
	v.SetDefault("http_host", "0.0.0.0")
	v.SetDefault("log_level", "info")

	// Bind environment variables
	v.SetEnvPrefix("YNAB_MCP")
	v.AutomaticEnv()

	// Also check for YNAB_ACCESS_TOKEN without prefix (common convention)
	if token := os.Getenv("YNAB_ACCESS_TOKEN"); token != "" {
		v.Set("ynab_access_token", token)
	}
	if token := os.Getenv("MCP_AUTH_TOKEN"); token != "" {
		v.Set("mcp_auth_token", token)
	}

	// Load config file if specified or use default location
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Try default config location: ~/.config/ynab-mcp/config.json
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configDir := filepath.Join(homeDir, ".config", "ynab-mcp")
			configFile := filepath.Join(configDir, "config.json")

			// Create config directory if it doesn't exist
			if _, err := os.Stat(configDir); os.IsNotExist(err) {
				os.MkdirAll(configDir, 0755)
			}

			// Use config file if it exists
			if _, err := os.Stat(configFile); err == nil {
				v.SetConfigFile(configFile)
			}
		}
	}

	// Read config file (ignore error if file doesn't exist)
	if v.ConfigFileUsed() != "" || configPath != "" {
		if err := v.ReadInConfig(); err != nil {
			// Only error if a config file was explicitly specified
			if configPath != "" {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	}

	// Build config struct
	cfg := &Config{
		YNABToken:     v.GetString("ynab_access_token"),
		TransportMode: v.GetString("transport_mode"),
		HTTPPort:      v.GetInt("http_port"),
		HTTPHost:      v.GetString("http_host"),
		MCPAuthToken:  v.GetString("mcp_auth_token"),
		LogLevel:      v.GetString("log_level"),
	}

	// Validate required fields
	if cfg.YNABToken == "" {
		return nil, fmt.Errorf("YNAB access token is required (set YNAB_ACCESS_TOKEN env var or add to config file)")
	}

	return cfg, nil
}
