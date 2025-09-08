package archiver

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestZipDailyBackups(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_backup")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test backup files
	testDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	// Create files for the test date
	file1 := filepath.Join(tempDir, "router1_2024-01-15_10-30-00.rsc")
	file2 := filepath.Join(tempDir, "router2_2024-01-15_11-00-00.rsc")

	// Create files for a different date (should not be included)
	file3 := filepath.Join(tempDir, "router3_2024-01-14_10-30-00.rsc")

	// Write test content
	testContent := []byte("# RouterOS configuration\n/interface ethernet\n")

	if err := os.WriteFile(file1, testContent, 0600); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	if err := os.WriteFile(file2, testContent, 0600); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}
	if err := os.WriteFile(file3, testContent, 0600); err != nil {
		t.Fatalf("Failed to create test file 3: %v", err)
	}

	// Set file modification times to match the test date
	os.Chtimes(file1, testDate, testDate)
	os.Chtimes(file2, testDate, testDate)
	// Set file3 to a different date
	differentDate := time.Date(2024, 1, 14, 10, 30, 0, 0, time.UTC)
	os.Chtimes(file3, differentDate, differentDate)

	// Test ZIP creation
	zipPath, err := ZipDailyBackups(tempDir, testDate)
	if err != nil {
		t.Fatalf("ZipDailyBackups() error = %v", err)
	}

	// Check if ZIP file was created
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Error("ZIP file was not created")
	}

	// Clean up
	os.Remove(zipPath)
}

func TestZipDailyBackupsNoFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_backup_empty")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test ZIP creation with no files
	testDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	_, err = ZipDailyBackups(tempDir, testDate)

	if err == nil {
		t.Error("Expected error when no files found, got nil")
	}

	if !contains(err.Error(), "no backup files found") {
		t.Errorf("Expected 'no backup files found' error, got: %v", err)
	}
}

func TestCleanupOldZips(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_cleanup")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test ZIP files with different ages
	oldFile := filepath.Join(tempDir, "old_backup.zip")
	newFile := filepath.Join(tempDir, "new_backup.zip")

	// Create old file (35 days ago)
	oldTime := time.Now().AddDate(0, 0, -35)
	if err := os.WriteFile(oldFile, []byte("old content"), 0600); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	os.Chtimes(oldFile, oldTime, oldTime)

	// Create new file (5 days ago)
	newTime := time.Now().AddDate(0, 0, -5)
	if err := os.WriteFile(newFile, []byte("new content"), 0600); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}
	os.Chtimes(newFile, newTime, newTime)

	// Test cleanup (keep 30 days)
	err = CleanupOldZips(tempDir, 30)
	if err != nil {
		t.Fatalf("CleanupOldZips() error = %v", err)
	}

	// Check that old file was removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should have been removed")
	}

	// Check that new file still exists
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New file should not have been removed")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && contains(s[1:], substr)
}
