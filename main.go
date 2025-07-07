package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
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

// backupMikroTik connects to a device via SSH and performs a configuration export.
func backupMikroTik(device Device, backupDir string) error {
	log.Printf("Starting SSH backup for device: %s (%s)", device.Name, device.Address)

	sshConfig := &ssh.ClientConfig{
		User: device.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(device.Password),
		},
		// WARNING: In a production environment, avoid ignoring the host key.
		// Consider using a known_hosts file for better security.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	// Ensure the address has a port, defaulting to 22 if missing.
	// This is more robust than string checking, especially for IPv6.
	address := device.Address
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = net.JoinHostPort(address, "22")
	}

	// Connect to the SSH server.
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH connection failed for %s: %w", device.Name, err)
	}
	defer client.Close()

	log.Printf("Successfully connected to %s. Exporting configuration...", device.Name)

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

	configContent := stdoutBuf.Bytes()

	// Check if the output is empty.
	if len(configContent) == 0 {
		return fmt.Errorf("no configuration data returned from %s; check user permissions", device.Name)
	}

	// Generate a backup filename with a timestamp.
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("%s_%s.rsc", device.Name, timestamp) // Using .rsc extension for RouterOS scripts
	filePath := filepath.Join(backupDir, fileName)

	// Save the configuration to a file.
	if err := os.WriteFile(filePath, configContent, 0644); err != nil {
		return fmt.Errorf("failed to save backup file for %s: %w", device.Name, err)
	}

	log.Printf("Backup for %s completed successfully! Saved to: %s", device.Name, filePath)
	return nil
}

// worker represents a concurrent worker that processes backup jobs from a channel.
func worker(id int, wg *sync.WaitGroup, devices <-chan Device, backupDir string) {
	defer wg.Done()

	for d := range devices {
		log.Printf("Worker %d: processing device %s...", id, d.Name)
		if err := backupMikroTik(d, backupDir); err != nil {
			// Log errors but continue processing other devices.
			log.Printf("ERROR backing up %s (Worker %d): %v", d.Name, id, err)
		}
	}
	log.Printf("Worker %d finished.", id)
}

func main() {
	log.Println("Starting MikroTik Backup application...")

	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Fatal error loading configuration: %v", err)
	}

	// Create the backup directory with standard permissions (0755).
	if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
		log.Fatalf("Could not create backup directory '%s': %v", config.BackupDir, err)
	}

	log.Printf("Found %d devices to back up.", len(config.Devices))

	// Defines the number of concurrent backup jobs.
	// This can be tuned based on system resources and network capacity.
	const numWorkers = 5

	// Create a buffered channel to distribute devices to workers.
	devicesChan := make(chan Device, len(config.Devices))
	var wg sync.WaitGroup

	// Start worker goroutines.
	log.Printf("Starting %d workers...", numWorkers)
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, &wg, devicesChan, config.BackupDir)
	}

	// Distribute jobs to the workers.
	log.Println("Distributing devices to workers...")
	for _, device := range config.Devices {
		devicesChan <- device
	}

	// Close the channel to signal that no more jobs will be sent.
	close(devicesChan)

	// Wait for all workers to complete their jobs.
	log.Println("Waiting for all workers to finish...")
	wg.Wait()

	log.Println("Backup process completed.")
}
