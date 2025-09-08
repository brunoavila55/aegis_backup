package config

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Device represents a single device entry in the configuration file.
type Device struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ScheduleConfig represents the scheduling configuration
type ScheduleConfig struct {
	Enabled  bool   `json:"enabled"`  // Whether scheduling is enabled
	Cron     string `json:"cron"`     // Cron expression (e.g., "0 2 * * *" for daily at 2 AM)
	Timezone string `json:"timezone"` // Timezone (e.g., "America/Sao_Paulo", defaults to "UTC")
}

// TelegramConfig represents the Telegram bot configuration
type TelegramConfig struct {
	Enabled  bool   `json:"enabled"`   // Whether Telegram notifications are enabled
	BotToken string `json:"bot_token"` // Telegram bot token
	ChatID   string `json:"chat_id"`   // Chat ID (group or channel)
	SendZip  bool   `json:"send_zip"`  // Whether to send daily backup ZIP files
	SendLogs bool   `json:"send_logs"` // Whether to send backup completion notifications
}

// BackupRetention represents the backup retention policy
type BackupRetention struct {
	Enabled  bool `json:"enabled"`
	KeepDays int  `json:"keep_days"`
}

// Config represents the structure of the main configuration file (e.g., config.json).
type Config struct {
	Devices         []Device
	BackupDir       string          `json:"backup_dir"`
	Schedule        ScheduleConfig  `json:"schedule"`
	Telegram        TelegramConfig  `json:"telegram"`
	BackupRetention BackupRetention `json:"backup_retention"`
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

	// Set default timezone if not specified
	if config.Schedule.Timezone == "" {
		config.Schedule.Timezone = "UTC"
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// LoadDevicesFromCSV reads devices from a CSV file.
func LoadDevicesFromCSV(path string) ([]Device, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open devices file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Read header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header from devices file: %w", err)
	}

	var devices []Device
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read from devices file: %w", err)
		}

		if len(record) < 4 {
			return nil, fmt.Errorf("invalid record in devices file: %v", record)
		}

		devices = append(devices, Device{
			Name:     record[0],
			Address:  record[1],
			Username: record[2],
			Password: record[3],
		})
	}

	return devices, nil
}

// validateConfig validates the configuration structure
func validateConfig(config *Config) error {
	// Validate backup directory
	if config.BackupDir == "" {
		return fmt.Errorf("backup directory cannot be empty")
	}

	// Validate schedule configuration
	if config.Schedule.Enabled {
		if config.Schedule.Cron == "" {
			return fmt.Errorf("cron expression is required when schedule is enabled")
		}

		// Validate timezone
		if _, err := time.LoadLocation(config.Schedule.Timezone); err != nil {
			return fmt.Errorf("invalid timezone '%s': %w", config.Schedule.Timezone, err)
		}
	}

	// Validate Telegram configuration
	if config.Telegram.Enabled {
		if config.Telegram.BotToken == "" {
			return fmt.Errorf("telegram bot token is required when telegram is enabled")
		}
		if config.Telegram.ChatID == "" {
			return fmt.Errorf("telegram chat ID is required when telegram is enabled")
		}
	}

	// Validate backup retention
	if config.BackupRetention.Enabled {
		if config.BackupRetention.KeepDays <= 0 {
			return fmt.Errorf("keep_days must be greater than 0 when backup retention is enabled")
		}
		if config.BackupRetention.KeepDays > 365 {
			return fmt.Errorf("keep_days cannot exceed 365 days")
		}
	}

	return nil
}
