package backup

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"aegis_backup/internal/config"
	"aegis_backup/internal/monitoring"

	"golang.org/x/crypto/ssh"
)

// Execute connects to a device via SSH and performs a configuration backup.
// It runs the "/export" command and saves the output to a file in the specified backup directory.
func Execute(device config.Device, backupDir string, metrics *monitoring.MetricsCollector, alertManager *monitoring.AlertManager) error {
	startTime := time.Now()
	log.Printf("Starting SSH backup for device: %s (%s)", device.Name, device.Address)

	// Configure the SSH client
	// Note: Using InsecureIgnoreHostKey for simplicity in controlled environments
	// For production with untrusted networks, consider implementing proper host key verification
	sshConfig := &ssh.ClientConfig{
		User: device.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(device.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	// Ensure the address has a port. If not, append the default SSH port "22".
	address := device.Address
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = net.JoinHostPort(address, "22")
	}

	// Dial the SSH server.
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		// Record failure metrics
		if metrics != nil {
			metrics.RecordBackupFailure(device.Name, err)
		}

		// Check for alerts
		if alertManager != nil {
			alertManager.CheckAlerts(device.Name)
		}

		return fmt.Errorf("failed to connect via SSH to %s: %w", device.Name, err)
	}
	defer client.Close()

	log.Printf("Successfully connected to %s. Exporting configuration...", device.Name)

	// Create a new SSH session.
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session for %s: %w", device.Name, err)
	}
	defer session.Close()

	// Capture the standard output of the remote command.
	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	// Execute the export command.
	if err := session.Run("/export"); err != nil {
		return fmt.Errorf("failed to run /export command on %s: %w", device.Name, err)
	}

	// Check if the command returned any data.
	configContent := stdoutBuf.Bytes()
	if len(configContent) == 0 {
		return fmt.Errorf("no configuration data returned from %s; check user permissions", device.Name)
	}

	// Create a unique filename for the backup file.
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("%s_%s.rsc", device.Name, timestamp)
	filePath := filepath.Join(backupDir, fileName)

	// Write the configuration data to the backup file with restricted permissions.
	if err := os.WriteFile(filePath, configContent, 0600); err != nil {
		return fmt.Errorf("failed to save backup file for %s: %w", device.Name, err)
	}

	duration := time.Since(startTime)
	fileInfo, _ := os.Stat(filePath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	// Record success metrics
	if metrics != nil {
		metrics.RecordBackupSuccess(device.Name, duration, fileSize)
	}

	// Check for alerts
	if alertManager != nil {
		alertManager.CheckAlerts(device.Name)
	}

	log.Printf("Backup for %s completed successfully! Saved to: %s (took %v, size: %d bytes)",
		device.Name, filePath, duration, fileSize)
	return nil
}
