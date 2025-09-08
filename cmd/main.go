package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"aegis_backup/internal/api"
	"aegis_backup/internal/archiver"
	"aegis_backup/internal/config"
	"aegis_backup/internal/monitoring"
	"aegis_backup/internal/scheduler"
	"aegis_backup/internal/telegram"
	"aegis_backup/internal/worker"
)

// calculateWorkers calculates the optimal number of workers based on device count
func calculateWorkers(deviceCount int) int {
	if deviceCount <= 0 {
		return 1
	}
	if deviceCount <= 5 {
		return deviceCount
	}
	if deviceCount <= 10 {
		return 5
	}
	// For more than 10 devices, use 8 workers max to avoid overwhelming the system
	return 8
}

func main() {
	// Define and parse the command-line flags
	configPath := flag.String("config", "config.json", "Path to the configuration file (e.g., config.json)")
	devicesPath := flag.String("devices", "devices.csv", "Path to the devices CSV file (e.g., devices.csv)")
	daemon := flag.Bool("daemon", false, "Run as daemon service with scheduler")
	apiPort := flag.Int("api-port", 8080, "Port for the monitoring API server")
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

	// Validate devices configuration
	if len(devices) == 0 {
		log.Fatalf("No devices found in CSV file: %s", *devicesPath)
	}

	// Validate each device
	for i, device := range devices {
		if device.Name == "" {
			log.Fatalf("Device %d: name is required", i+1)
		}
		if device.Address == "" {
			log.Fatalf("Device %d (%s): address is required", i+1, device.Name)
		}
		if device.Username == "" {
			log.Fatalf("Device %d (%s): username is required", i+1, device.Name)
		}
		if device.Password == "" {
			log.Fatalf("Device %d (%s): password is required", i+1, device.Name)
		}
	}

	log.Printf("Loaded %d devices for backup", len(devices))
	conf.Devices = devices

	// Ensure the backup directory exists before starting the backup process.
	if err := os.MkdirAll(conf.BackupDir, 0700); err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	// Initialize monitoring system
	metrics := monitoring.NewMetricsCollector()
	alertManager := monitoring.NewAlertManager(metrics)

	// Start API server in background
	go func() {
		apiServer := api.NewServer(metrics, alertManager, *apiPort)
		if err := apiServer.Start(); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	log.Printf("Monitoring API available at http://localhost:%d", *apiPort)

	// Determine the execution mode
	if *daemon {
		runDaemonMode(conf, metrics, alertManager)
	} else {
		runOnceMode(conf, metrics, alertManager)
	}
}

// runOnceMode executes a single backup run and exits
func runOnceMode(conf *config.Config, metrics *monitoring.MetricsCollector, alertManager *monitoring.AlertManager) {
	log.Println("Running backup once...")

	start := time.Now()

	// Create a buffered channel to distribute devices to the worker pool.
	devicesChan := make(chan config.Device, len(conf.Devices))
	var wg sync.WaitGroup

	// Calculate optimal number of workers
	numWorkers := calculateWorkers(len(conf.Devices))
	log.Printf("Starting backup with %d workers for %d devices", numWorkers, len(conf.Devices))

	// Start the worker pool.
	worker.StartPool(numWorkers, &wg, devicesChan, conf.BackupDir, metrics, alertManager)

	// Add all devices from the configuration to the channel for processing.
	for _, device := range conf.Devices {
		devicesChan <- device
	}
	close(devicesChan) // Close the channel to signal that no more devices will be sent.

	// Wait for all workers to finish their jobs.
	wg.Wait()

	duration := time.Since(start)
	log.Printf("Backup process completed successfully in %v", duration)

	// Post-backup operations
	postBackupOperations(conf, start, duration)
}

// runDaemonMode runs the application as a daemon service with scheduler
func runDaemonMode(conf *config.Config, metrics *monitoring.MetricsCollector, alertManager *monitoring.AlertManager) {
	log.Println("Starting Aegis Backup daemon...")

	// Create and start the scheduler
	sched := scheduler.NewScheduler(conf, metrics, alertManager)

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

// postBackupOperations handles compression and Telegram notifications
func postBackupOperations(conf *config.Config, backupTime time.Time, duration time.Duration) {
	var zipPath string
	var zipErr error

	// Create ZIP file if Telegram is enabled and send_zip is true
	if conf.Telegram.Enabled && conf.Telegram.SendZip {
		log.Println("Creating daily backup ZIP file...")

		zipPath, zipErr = archiver.ZipDailyBackups(conf.BackupDir, backupTime)
		if zipErr != nil {
			log.Printf("Failed to create ZIP file: %v", zipErr)
			sendErrorNotification(conf, "ZIP Creation", zipErr)
			// Continue execution even if ZIP creation fails
			zipPath = ""
		} else {
			log.Printf("ZIP file created: %s", zipPath)
		}
	}

	// Send Telegram notifications
	if conf.Telegram.Enabled {
		if conf.Telegram.SendLogs {
			sendBackupSummary(conf, len(conf.Devices), duration, zipPath)
		}

		if conf.Telegram.SendZip && zipPath != "" && zipErr == nil {
			sendZipFile(conf, zipPath)
		}
	}

	// Cleanup old ZIP files
	if conf.BackupRetention.Enabled {
		if err := archiver.CleanupOldZips(conf.BackupDir, conf.BackupRetention.KeepDays); err != nil {
			log.Printf("Warning: Failed to cleanup old ZIP files: %v", err)
		}
	}

	log.Println("Backup completed successfully")
}

// sendBackupSummary sends a backup completion summary to Telegram
func sendBackupSummary(conf *config.Config, deviceCount int, duration time.Duration, zipFile string) {
	tgClient := telegram.NewClient(conf.Telegram.BotToken, conf.Telegram.ChatID)
	message := telegram.FormatBackupSummary(deviceCount, duration, zipFile, conf.Schedule.Timezone)

	if err := tgClient.SendMessage(message); err != nil {
		log.Printf("Failed to send backup summary to Telegram: %v", err)
	} else {
		log.Println("Backup summary sent to Telegram")
	}
}

// sendZipFile sends the ZIP file to Telegram
func sendZipFile(conf *config.Config, zipPath string) {
	if zipPath == "" {
		return
	}

	tgClient := telegram.NewClient(conf.Telegram.BotToken, conf.Telegram.ChatID)
	log.Println("Sending ZIP file to Telegram...")

	caption := "📦 Daily backup archive"

	if err := tgClient.SendDocument(zipPath, caption); err != nil {
		log.Printf("Failed to send ZIP file to Telegram: %v", err)
		sendErrorNotification(conf, "ZIP File Upload", err)
	} else {
		log.Println("ZIP file sent to Telegram successfully")
	}
}

// sendErrorNotification sends an error notification to Telegram
func sendErrorNotification(conf *config.Config, operation string, err error) {
	tgClient := telegram.NewClient(conf.Telegram.BotToken, conf.Telegram.ChatID)
	message := telegram.FormatErrorMessage(operation, err, conf.Schedule.Timezone)

	if sendErr := tgClient.SendMessage(message); sendErr != nil {
		log.Printf("Failed to send error notification to Telegram: %v", sendErr)
	}
}
