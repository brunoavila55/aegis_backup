package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Device represents a single MikroTik device in the config file.
type Device struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config represents the structure of the config.json file.
type Config struct {
	Devices   []Device `json:"devices"`
	BackupDir string   `json:"backup_dir"`
}

// loadConfig reads the configuration from a JSON file.
func loadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	// Default to "./backups" if no backup directory is specified.
	if config.BackupDir == "" {
		config.BackupDir = "./backups"
	}

	return &config, nil
}

// backupMikroTik connects to a device via SSH and performs a backup.
func backupMikroTik(device Device, backupDir string) error {
	log.Printf("Starting SSH backup for device: %s (%s)", device.Name, device.Address)

	// Set up SSH client configuration.
	sshConfig := &ssh.ClientConfig{
		User: device.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(device.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: Ignores host key verification. Fine for internal networks, but consider using a known_hosts file for production.
		Timeout:         15 * time.Second,
	}

	// Ensure SSH port (22) is included if not specified in the address.
	address := device.Address
	if !strings.Contains(address, ":") {
		address = address + ":22"
	}

	// Connect to the SSH server.
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH connection failed for %s: %w", device.Name, err)
	}
	defer client.Close()

	log.Printf("Successfully connected to %s via SSH. Exporting configuration...", device.Name)

	// Create a new SSH session.
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session for %s: %w", device.Name, err)
	}
	defer session.Close()

	// Capture the command output.
	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	// Run the export command to retrieve the configuration.
	if err := session.Run("/export"); err != nil {
		return fmt.Errorf("failed to run /export command on %s: %w", device.Name, err)
	}

	configContent := stdoutBuf.String()

	// Check if the output is empty.
	if configContent == "" {
		return fmt.Errorf("no configuration data returned from /export for %s. Check user permissions and RouterOS version", device.Name)
	}

	// Generate backup filename with timestamp.
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("%s_%s.txt", device.Name, timestamp)
	filePath := filepath.Join(backupDir, fileName)

	// Save the configuration to a file.
	if err := os.WriteFile(filePath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to save backup file for %s: %w", device.Name, err)
	}

	log.Printf("Backup for %s completed successfully! Saved to: %s", device.Name, filePath)
	return nil
}

func main() {
	log.Println("Starting MikroTik Backup Application (via SSH)...")

	// Load configuration from file.
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Fatal error loading config: %v", err)
	}

	// Create backup directory if it doesn't exist.
	if err := os.MkdirAll(config.BackupDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create backup directory '%s': %v", config.BackupDir, err)
	}

	log.Printf("Found %d devices to back up.", len(config.Devices))

	// Run backups concurrently for each device.
	var wg sync.WaitGroup
	for _, device := range config.Devices {
		wg.Add(1)
		go func(d Device) {
			defer wg.Done()
			if err := backupMikroTik(d, config.BackupDir); err != nil {
				log.Printf("Backup error for %s: %v", d.Name, err)
			}
		}(device)
	}

	// Wait for all backups to complete.
	wg.Wait()

	log.Println("Backup process completed.")
}
