package monitoring

import (
	"sync"
	"time"
)

// DeviceMetrics holds metrics for a single device
type DeviceMetrics struct {
	DeviceName          string        `json:"device_name"`
	LastBackup          time.Time     `json:"last_backup"`
	LastSuccess         time.Time     `json:"last_success"`
	LastFailure         time.Time     `json:"last_failure"`
	BackupSize          int64         `json:"backup_size"`
	SuccessCount        int           `json:"success_count"`
	FailureCount        int           `json:"failure_count"`
	AverageTime         time.Duration `json:"average_time"`
	LastError           string        `json:"last_error"`
	Status              DeviceStatus  `json:"status"`
	ConsecutiveFailures int           `json:"consecutive_failures"`
	mu                  sync.RWMutex
}

// DeviceStatus represents the current status of a device
type DeviceStatus string

const (
	StatusHealthy  DeviceStatus = "healthy"
	StatusWarning  DeviceStatus = "warning"
	StatusCritical DeviceStatus = "critical"
	StatusUnknown  DeviceStatus = "unknown"
)

// MetricsCollector manages metrics for all devices
type MetricsCollector struct {
	devices map[string]*DeviceMetrics
	mu      sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		devices: make(map[string]*DeviceMetrics),
	}
}

// GetOrCreateDevice gets existing device metrics or creates new ones
func (mc *MetricsCollector) GetOrCreateDevice(deviceName string) *DeviceMetrics {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if device, exists := mc.devices[deviceName]; exists {
		return device
	}

	device := &DeviceMetrics{
		DeviceName: deviceName,
		Status:     StatusUnknown,
	}
	mc.devices[deviceName] = device
	return device
}

// RecordBackupSuccess records a successful backup
func (mc *MetricsCollector) RecordBackupSuccess(deviceName string, duration time.Duration, size int64) {
	device := mc.GetOrCreateDevice(deviceName)

	device.mu.Lock()
	defer device.mu.Unlock()

	device.LastBackup = time.Now()
	device.LastSuccess = time.Now()
	device.BackupSize = size
	device.SuccessCount++
	device.ConsecutiveFailures = 0

	// Update average time
	if device.AverageTime == 0 {
		device.AverageTime = duration
	} else {
		device.AverageTime = (device.AverageTime + duration) / 2
	}

	// Update status
	device.Status = StatusHealthy
}

// RecordBackupFailure records a failed backup
func (mc *MetricsCollector) RecordBackupFailure(deviceName string, err error) {
	device := mc.GetOrCreateDevice(deviceName)

	device.mu.Lock()
	defer device.mu.Unlock()

	device.LastBackup = time.Now()
	device.LastFailure = time.Now()
	device.FailureCount++
	device.ConsecutiveFailures++
	device.LastError = err.Error()

	// Update status based on consecutive failures
	if device.ConsecutiveFailures >= 3 {
		device.Status = StatusCritical
	} else if device.ConsecutiveFailures >= 1 {
		device.Status = StatusWarning
	}
}

// GetDeviceMetrics returns metrics for a specific device
func (mc *MetricsCollector) GetDeviceMetrics(deviceName string) (*DeviceMetrics, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	device, exists := mc.devices[deviceName]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	device.mu.RLock()
	defer device.mu.RUnlock()

	return &DeviceMetrics{
		DeviceName:          device.DeviceName,
		LastBackup:          device.LastBackup,
		LastSuccess:         device.LastSuccess,
		LastFailure:         device.LastFailure,
		BackupSize:          device.BackupSize,
		SuccessCount:        device.SuccessCount,
		FailureCount:        device.FailureCount,
		AverageTime:         device.AverageTime,
		LastError:           device.LastError,
		Status:              device.Status,
		ConsecutiveFailures: device.ConsecutiveFailures,
	}, true
}

// GetAllMetrics returns metrics for all devices
func (mc *MetricsCollector) GetAllMetrics() map[string]*DeviceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*DeviceMetrics)
	for name, device := range mc.devices {
		device.mu.RLock()
		result[name] = &DeviceMetrics{
			DeviceName:          device.DeviceName,
			LastBackup:          device.LastBackup,
			LastSuccess:         device.LastSuccess,
			LastFailure:         device.LastFailure,
			BackupSize:          device.BackupSize,
			SuccessCount:        device.SuccessCount,
			FailureCount:        device.FailureCount,
			AverageTime:         device.AverageTime,
			LastError:           device.LastError,
			Status:              device.Status,
			ConsecutiveFailures: device.ConsecutiveFailures,
		}
		device.mu.RUnlock()
	}

	return result
}

// GetOverallStats returns overall statistics
func (mc *MetricsCollector) GetOverallStats() OverallStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := OverallStats{
		TotalDevices:      len(mc.devices),
		HealthyDevices:    0,
		WarningDevices:    0,
		CriticalDevices:   0,
		UnknownDevices:    0,
		TotalBackups:      0,
		TotalFailures:     0,
		AverageBackupTime: 0,
	}

	var totalTime time.Duration
	deviceCount := 0

	for _, device := range mc.devices {
		device.mu.RLock()

		switch device.Status {
		case StatusHealthy:
			stats.HealthyDevices++
		case StatusWarning:
			stats.WarningDevices++
		case StatusCritical:
			stats.CriticalDevices++
		case StatusUnknown:
			stats.UnknownDevices++
		}

		stats.TotalBackups += device.SuccessCount
		stats.TotalFailures += device.FailureCount

		if device.AverageTime > 0 {
			totalTime += device.AverageTime
			deviceCount++
		}

		device.mu.RUnlock()
	}

	if deviceCount > 0 {
		stats.AverageBackupTime = totalTime / time.Duration(deviceCount)
	}

	return stats
}

// OverallStats represents overall system statistics
type OverallStats struct {
	TotalDevices      int           `json:"total_devices"`
	HealthyDevices    int           `json:"healthy_devices"`
	WarningDevices    int           `json:"warning_devices"`
	CriticalDevices   int           `json:"critical_devices"`
	UnknownDevices    int           `json:"unknown_devices"`
	TotalBackups      int           `json:"total_backups"`
	TotalFailures     int           `json:"total_failures"`
	AverageBackupTime time.Duration `json:"average_backup_time"`
}
