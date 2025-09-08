package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"aegis_backup/internal/archiver"
	"aegis_backup/internal/config"
	"aegis_backup/internal/monitoring"
	"aegis_backup/internal/telegram"
	"aegis_backup/internal/worker"

	"github.com/robfig/cron/v3"
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

// Scheduler manages the backup scheduling functionality
type Scheduler struct {
	cron         *cron.Cron
	config       *config.Config
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
	mu           sync.RWMutex
	telegram     *telegram.Client
	metrics      *monitoring.MetricsCollector
	alertManager *monitoring.AlertManager
}

// NewScheduler creates a new scheduler instance
func NewScheduler(conf *config.Config, metrics *monitoring.MetricsCollector, alertManager *monitoring.AlertManager) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	// Parse timezone
	location, err := time.LoadLocation(conf.Schedule.Timezone)
	if err != nil {
		log.Printf("Warning: Invalid timezone '%s', using UTC", conf.Schedule.Timezone)
		location = time.UTC
	}

	scheduler := &Scheduler{
		cron:         cron.New(cron.WithLocation(location)),
		config:       conf,
		ctx:          ctx,
		cancel:       cancel,
		metrics:      metrics,
		alertManager: alertManager,
	}

	// Initialize Telegram client if enabled
	if conf.Telegram.Enabled && conf.Telegram.BotToken != "" && conf.Telegram.ChatID != "" {
		scheduler.telegram = telegram.NewClient(conf.Telegram.BotToken, conf.Telegram.ChatID)

		// Test connection
		if err := scheduler.telegram.TestConnection(); err != nil {
			log.Printf("Warning: Failed to connect to Telegram: %v", err)
			scheduler.telegram = nil
		} else {
			log.Println("Telegram client initialized successfully")
		}
	}

	return scheduler
}

// Start begins the scheduled backup service
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	if !s.config.Schedule.Enabled {
		log.Println("Scheduler is disabled in configuration")
		return nil
	}

	if s.config.Schedule.Cron == "" {
		log.Println("No cron expression provided, scheduler not started")
		return nil
	}

	// Add the backup job to the cron scheduler
	_, err := s.cron.AddFunc(s.config.Schedule.Cron, s.runBackup)
	if err != nil {
		return err
	}

	s.cron.Start()
	s.running = true

	log.Printf("Scheduler started with cron expression: %s (timezone: %s)",
		s.config.Schedule.Cron, s.config.Schedule.Timezone)

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.cron.Stop()
	s.cancel()
	s.running = false
	log.Println("Scheduler stopped")
}

// IsRunning returns whether the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// runBackup executes the backup process with compression and Telegram notification
func (s *Scheduler) runBackup() {
	log.Println("Starting scheduled backup...")

	start := time.Now()

	// Create a buffered channel to distribute devices to the worker pool
	devicesChan := make(chan config.Device, len(s.config.Devices))
	var wg sync.WaitGroup

	// Calculate optimal number of workers
	numWorkers := calculateWorkers(len(s.config.Devices))
	log.Printf("Starting scheduled backup with %d workers for %d devices", numWorkers, len(s.config.Devices))

	// Start the worker pool
	worker.StartPool(numWorkers, &wg, devicesChan, s.config.BackupDir, s.metrics, s.alertManager)

	// Add all devices from the configuration to the channel for processing
	for _, device := range s.config.Devices {
		devicesChan <- device
	}
	close(devicesChan) // Close the channel to signal that no more devices will be sent

	// Wait for all workers to finish their jobs
	wg.Wait()

	duration := time.Since(start)
	log.Printf("Backup process completed in %v", duration)

	// Post-backup operations
	s.postBackupOperations(start, duration)
}

// postBackupOperations handles compression and Telegram notifications
func (s *Scheduler) postBackupOperations(backupTime time.Time, duration time.Duration) {
	var zipPath string
	var zipErr error

	// Create ZIP file if Telegram is enabled and send_zip is true
	if s.config.Telegram.Enabled && s.config.Telegram.SendZip {
		log.Println("Creating daily backup ZIP file...")

		zipPath, zipErr = archiver.ZipDailyBackups(s.config.BackupDir, backupTime)
		if zipErr != nil {
			log.Printf("Failed to create ZIP file: %v", zipErr)
			s.sendErrorNotification("ZIP Creation", zipErr)
			// Continue execution even if ZIP creation fails
			zipPath = ""
		} else {
			log.Printf("ZIP file created: %s", zipPath)
		}
	}

	// Send Telegram notifications
	if s.telegram != nil {
		if s.config.Telegram.SendLogs {
			s.sendBackupSummary(len(s.config.Devices), duration, zipPath)
		}

		if s.config.Telegram.SendZip && zipPath != "" && zipErr == nil {
			s.sendZipFile(zipPath)
		}
	}

	// Cleanup old ZIP files
	if s.config.BackupRetention.Enabled {
		if err := archiver.CleanupOldZips(s.config.BackupDir, s.config.BackupRetention.KeepDays); err != nil {
			log.Printf("Warning: Failed to cleanup old ZIP files: %v", err)
		}
	}

	log.Println("Scheduled backup completed successfully")
}

// sendBackupSummary sends a backup completion summary to Telegram
func (s *Scheduler) sendBackupSummary(deviceCount int, duration time.Duration, zipFile string) {
	if s.telegram == nil {
		return
	}

	message := telegram.FormatBackupSummary(deviceCount, duration, zipFile, s.config.Schedule.Timezone)

	if err := s.telegram.SendMessage(message); err != nil {
		log.Printf("Failed to send backup summary to Telegram: %v", err)
	} else {
		log.Println("Backup summary sent to Telegram")
	}
}

// sendZipFile sends the ZIP file to Telegram
func (s *Scheduler) sendZipFile(zipPath string) {
	if s.telegram == nil || zipPath == "" {
		return
	}

	log.Println("Sending ZIP file to Telegram...")

	caption := "📦 Daily backup archive"

	if err := s.telegram.SendDocument(zipPath, caption); err != nil {
		log.Printf("Failed to send ZIP file to Telegram: %v", err)
		s.sendErrorNotification("ZIP File Upload", err)
	} else {
		log.Println("ZIP file sent to Telegram successfully")
	}
}

// sendErrorNotification sends an error notification to Telegram
func (s *Scheduler) sendErrorNotification(operation string, err error) {
	if s.telegram == nil {
		return
	}

	message := telegram.FormatErrorMessage(operation, err, s.config.Schedule.Timezone)

	if sendErr := s.telegram.SendMessage(message); sendErr != nil {
		log.Printf("Failed to send error notification to Telegram: %v", sendErr)
	}
}

// GetNextRun returns the next scheduled run time
func (s *Scheduler) GetNextRun() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return time.Time{}
	}

	entries := s.cron.Entries()
	if len(entries) > 0 {
		return entries[0].Next
	}

	return time.Time{}
}

// Wait blocks until the scheduler is stopped
func (s *Scheduler) Wait() {
	<-s.ctx.Done()
}
