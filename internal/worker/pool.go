package worker

import (
	"log"
	"sync"

	"aegis_backup/internal/backup"
	"aegis_backup/internal/config"
	"aegis_backup/internal/monitoring"
)

// processJobs defines the work for a single worker goroutine.
// It receives devices from a channel and executes the backup for each one.
func processJobs(id int, wg *sync.WaitGroup, devices <-chan config.Device, backupDir string, metrics *monitoring.MetricsCollector, alertManager *monitoring.AlertManager) {
	// Signal to the WaitGroup that this worker is finished when the function returns.
	defer wg.Done()

	log.Printf("Worker %d started", id)

	// Range over the devices channel until it is closed.
	for device := range devices {
		log.Printf("Worker %d: processing device %s...", id, device.Name)

		if err := backup.Execute(device, backupDir, metrics, alertManager); err != nil {
			// Log the error but continue processing other devices.
			// This ensures that one failed backup doesn't stop the entire process.
			log.Printf("ERROR backing up %s (Worker %d): %v", device.Name, id, err)
		}
	}

	log.Printf("Worker %d finished.", id)
}

// StartPool initializes and starts a pool of workers to concurrently process device backups.
func StartPool(numWorkers int, wg *sync.WaitGroup, devices <-chan config.Device, backupDir string, metrics *monitoring.MetricsCollector, alertManager *monitoring.AlertManager) {
	// Start the specified number of worker goroutines.
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go processJobs(i, wg, devices, backupDir, metrics, alertManager)
	}
}
