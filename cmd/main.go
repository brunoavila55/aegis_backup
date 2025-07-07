package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"aegis_backup/internal/config"
	"aegis_backup/internal/worker"
)

// The number of concurrent workers for processing backups.
const numWorkers = 5

func main() {
	// Define and parse the command-line flag for the config file path.
	configPath := flag.String("config", "config.json", "Path to the configuration file (e.g., config.json)")
	flag.Parse()

	// If the config flag was not set, check for an environment variable as a fallback.
	if *configPath == "config.json" {
		if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
			*configPath = envPath
		}
	}

	log.Printf("Using configuration file: %s", *configPath)

	// Load the application configuration from the specified file.
	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Ensure the backup directory exists before starting the backup process.
	if err := os.MkdirAll(conf.BackupDir, 0755); err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	// Create a buffered channel to distribute devices to the worker pool.
	devicesChan := make(chan config.Device, len(conf.Devices))
	var wg sync.WaitGroup

	// Start the worker pool.
	worker.StartPool(numWorkers, &wg, devicesChan, conf.BackupDir)

	// Add all devices from the configuration to the channel for processing.
	for _, device := range conf.Devices {
		devicesChan <- device
	}
	close(devicesChan) // Close the channel to signal that no more devices will be sent.

	// Wait for all workers to finish their jobs.
	wg.Wait()
	log.Println("Backup process completed successfully.")
}
