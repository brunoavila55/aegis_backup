package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"aegis_backup/internal/config"
	"aegis_backup/internal/scheduler"
	"aegis_backup/internal/worker"
)

// The number of concurrent workers for processing backups.
const numWorkers = 5

func main() {
	// Define and parse the command-line flags
	configPath := flag.String("config", "config.json", "Path to the configuration file (e.g., config.json)")
	devicesPath := flag.String("devices", "devices.csv", "Path to the devices CSV file (e.g., devices.csv)")
	daemon := flag.Bool("daemon", false, "Run as daemon service with scheduler")
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

	// Load devices from the CSV file.
	devices, err := config.LoadDevicesFromCSV(*devicesPath)
	if err != nil {
		log.Fatalf("Error loading devices from CSV: %v", err)
	}
	conf.Devices = devices

	// Ensure the backup directory exists before starting the backup process.
	if err := os.MkdirAll(conf.BackupDir, 0755); err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	// Determine the execution mode
	if *daemon {
		runDaemonMode(conf)
	} else {
		runOnceMode(conf)
	}
}

// runOnceMode executes a single backup run and exits
func runOnceMode(conf *config.Config) {
	log.Println("Running backup once...")
	
	start := time.Now()
	
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
	
	duration := time.Since(start)
	log.Printf("Backup process completed successfully in %v", duration)
}

// runDaemonMode runs the application as a daemon service with scheduler
func runDaemonMode(conf *config.Config) {
	log.Println("Starting Aegis Backup daemon...")

	// Create and start the scheduler
	sched := scheduler.NewScheduler(conf)
	
	if err := sched.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// If scheduler is enabled, show next run time
	if conf.Schedule.Enabled && sched.IsRunning() {
		nextRun := sched.GetNextRun()
		if !nextRun.IsZero() {
			log.Printf("Next backup scheduled for: %s", nextRun.Format("2006-01-02 15:04:05 MST"))
		}
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Aegis Backup daemon is running. Press Ctrl+C to stop.")
	
	// Wait for shutdown signal
	<-sigChan
	
	log.Println("Shutdown signal received, stopping daemon...")
	sched.Stop()
	log.Println("Aegis Backup daemon stopped gracefully")
}