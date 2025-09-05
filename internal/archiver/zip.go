package archiver

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ZipDailyBackups creates a ZIP file containing all backup files from a specific date
func ZipDailyBackups(backupDir string, date time.Time) (string, error) {
	// Format date as YYYY-MM-DD for file filtering
	dateStr := date.Format("2006-01-02")
	
	// Create ZIP filename with date
	zipFilename := fmt.Sprintf("backups_%s.zip", dateStr)
	zipPath := filepath.Join(backupDir, zipFilename)
	
	// Create ZIP file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create ZIP file: %w", err)
	}
	defer zipFile.Close()
	
	// Create ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	
	// Track files added to ZIP
	filesAdded := 0
	
	// Walk through backup directory
	err = filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and the ZIP file itself
		if info.IsDir() || strings.HasSuffix(path, ".zip") {
			return nil
		}
		
		// Check if file was created on the specified date
		if !isFileFromDate(info, date) {
			return nil
		}
		
		// Add file to ZIP
		if err := addFileToZip(zipWriter, path, backupDir); err != nil {
			return fmt.Errorf("failed to add file %s to ZIP: %w", path, err)
		}
		
		filesAdded++
		return nil
	})
	
	if err != nil {
		// Clean up ZIP file if there was an error
		os.Remove(zipPath)
		return "", fmt.Errorf("failed to walk backup directory: %w", err)
	}
	
	if filesAdded == 0 {
		// Remove empty ZIP file
		os.Remove(zipPath)
		return "", fmt.Errorf("no backup files found for date %s", dateStr)
	}
	
	return zipPath, nil
}

// isFileFromDate checks if a file was created on a specific date
func isFileFromDate(info os.FileInfo, targetDate time.Time) bool {
	fileDate := info.ModTime().Format("2006-01-02")
	targetDateStr := targetDate.Format("2006-01-02")
	return fileDate == targetDateStr
}

// addFileToZip adds a single file to the ZIP archive
func addFileToZip(zipWriter *zip.Writer, filePath, baseDir string) error {
	// Open source file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Get file info
	info, err := file.Stat()
	if err != nil {
		return err
	}
	
	// Create ZIP file header
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	
	// Set relative path in ZIP
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		relPath = filepath.Base(filePath)
	}
	header.Name = relPath
	
	// Set compression method
	header.Method = zip.Deflate
	
	// Create writer for this file in ZIP
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	
	// Copy file content to ZIP
	_, err = io.Copy(writer, file)
	return err
}

// CleanupOldZips removes ZIP files older than specified days
func CleanupOldZips(backupDir string, keepDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -keepDays)
	
	return filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Only process ZIP files
		if !strings.HasSuffix(strings.ToLower(path), ".zip") {
			return nil
		}
		
		// Check if file is older than cutoff date
		if info.ModTime().Before(cutoffDate) {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove old ZIP file %s: %w", path, err)
			}
		}
		
		return nil
	})
}