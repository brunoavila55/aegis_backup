package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Device represents a single device entry in the configuration file.
type Device struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config represents the structure of the main configuration file (e.g., config.json).
type Config struct {
	Devices   []Device `json:"devices"`
	BackupDir string   `json:"backup_dir"`
}

// LoadConfig reads and parses the JSON configuration from the given file path.
// It returns a Config struct populated with the data from the file.
func LoadConfig(path string) (*Config, error) {
	// Open the configuration file for reading.
	configFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	// Decode the JSON content directly from the file stream.
	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config JSON: %w", err)
	}

	// If no backup directory is specified in the config, set a default value.
	if config.BackupDir == "" {
		config.BackupDir = "./backups"
	}

	return &config, nil
}
