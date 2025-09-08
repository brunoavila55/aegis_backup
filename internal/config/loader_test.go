package config

import (
	"os"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				BackupDir: "/tmp/backups",
				Schedule: ScheduleConfig{
					Enabled:  true,
					Cron:     "0 2 * * *",
					Timezone: "UTC",
				},
				Telegram: TelegramConfig{
					Enabled:  true,
					BotToken: "test_token",
					ChatID:   "test_chat",
				},
				BackupRetention: BackupRetention{
					Enabled:  true,
					KeepDays: 30,
				},
			},
			wantErr: false,
		},
		{
			name: "empty backup dir",
			config: Config{
				BackupDir: "",
			},
			wantErr: true,
		},
		{
			name: "schedule enabled without cron",
			config: Config{
				BackupDir: "/tmp/backups",
				Schedule: ScheduleConfig{
					Enabled: true,
					Cron:    "",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid timezone",
			config: Config{
				BackupDir: "/tmp/backups",
				Schedule: ScheduleConfig{
					Enabled:  true,
					Cron:     "0 2 * * *",
					Timezone: "Invalid/Timezone",
				},
			},
			wantErr: true,
		},
		{
			name: "telegram enabled without token",
			config: Config{
				BackupDir: "/tmp/backups",
				Telegram: TelegramConfig{
					Enabled:  true,
					BotToken: "",
					ChatID:   "test_chat",
				},
			},
			wantErr: true,
		},
		{
			name: "telegram enabled without chat ID",
			config: Config{
				BackupDir: "/tmp/backups",
				Telegram: TelegramConfig{
					Enabled:  true,
					BotToken: "test_token",
					ChatID:   "",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid keep days",
			config: Config{
				BackupDir: "/tmp/backups",
				BackupRetention: BackupRetention{
					Enabled:  true,
					KeepDays: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "keep days too high",
			config: Config{
				BackupDir: "/tmp/backups",
				BackupRetention: BackupRetention{
					Enabled:  true,
					KeepDays: 400,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadDevicesFromCSV(t *testing.T) {
	// Create a temporary CSV file
	tempFile, err := os.CreateTemp("", "test_devices.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test data
	csvContent := `name,address,username,password
router1,192.168.1.1,admin,password123
router2,192.168.1.2,admin,password456
router3,192.168.1.3,admin,password789`

	if _, err := tempFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tempFile.Close()

	// Test loading devices
	devices, err := LoadDevicesFromCSV(tempFile.Name())
	if err != nil {
		t.Fatalf("LoadDevicesFromCSV() error = %v", err)
	}

	if len(devices) != 3 {
		t.Errorf("Expected 3 devices, got %d", len(devices))
	}

	// Check first device
	if devices[0].Name != "router1" {
		t.Errorf("Expected name 'router1', got '%s'", devices[0].Name)
	}
	if devices[0].Address != "192.168.1.1" {
		t.Errorf("Expected address '192.168.1.1', got '%s'", devices[0].Address)
	}
	if devices[0].Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", devices[0].Username)
	}
	if devices[0].Password != "password123" {
		t.Errorf("Expected password 'password123', got '%s'", devices[0].Password)
	}
}

func TestLoadDevicesFromCSVEmpty(t *testing.T) {
	// Create a temporary CSV file with only header
	tempFile, err := os.CreateTemp("", "test_empty_devices.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write only header
	if _, err := tempFile.WriteString("name,address,username,password\n"); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tempFile.Close()

	// Test loading devices
	devices, err := LoadDevicesFromCSV(tempFile.Name())
	if err != nil {
		t.Fatalf("LoadDevicesFromCSV() error = %v", err)
	}

	if len(devices) != 0 {
		t.Errorf("Expected 0 devices, got %d", len(devices))
	}
}
