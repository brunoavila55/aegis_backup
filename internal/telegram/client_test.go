package telegram

import (
	"testing"
	"time"
)

func TestFormatBackupSummary(t *testing.T) {
	tests := []struct {
		name        string
		deviceCount int
		duration    time.Duration
		zipFile     string
		timezone    string
		expected    string
	}{
		{
			name:        "valid summary",
			deviceCount: 3,
			duration:    2 * time.Minute,
			zipFile:     "/path/to/backup.zip",
			timezone:    "America/Sao_Paulo",
			expected:    "🛡️ <b>Aegis Backup Completed</b>",
		},
		{
			name:        "invalid timezone fallback",
			deviceCount: 1,
			duration:    time.Minute,
			zipFile:     "/path/to/backup.zip",
			timezone:    "Invalid/Timezone",
			expected:    "🛡️ <b>Aegis Backup Completed</b>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBackupSummary(tt.deviceCount, tt.duration, tt.zipFile, tt.timezone)

			if !contains(result, tt.expected) {
				t.Errorf("FormatBackupSummary() = %v, expected to contain %v", result, tt.expected)
			}

			if !contains(result, "Devices backed up: 3") && tt.deviceCount == 3 {
				t.Errorf("Expected device count in result")
			}
		})
	}
}

func TestFormatErrorMessage(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		err       error
		timezone  string
		expected  string
	}{
		{
			name:      "valid error message",
			operation: "Backup",
			err:       &testError{msg: "connection failed"},
			timezone:  "UTC",
			expected:  "❌ <b>Aegis Backup Error</b>",
		},
		{
			name:      "invalid timezone fallback",
			operation: "ZIP Creation",
			err:       &testError{msg: "file not found"},
			timezone:  "Invalid/Timezone",
			expected:  "❌ <b>Aegis Backup Error</b>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatErrorMessage(tt.operation, tt.err, tt.timezone)

			if !contains(result, tt.expected) {
				t.Errorf("FormatErrorMessage() = %v, expected to contain %v", result, tt.expected)
			}

			if !contains(result, tt.operation) {
				t.Errorf("Expected operation '%s' in result", tt.operation)
			}
		})
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && contains(s[1:], substr)
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
