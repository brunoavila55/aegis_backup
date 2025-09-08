package monitoring

import (
	"testing"
	"time"
)

func TestMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector()

	// Test device creation
	device := mc.GetOrCreateDevice("test-device")
	if device.DeviceName != "test-device" {
		t.Errorf("Expected device name 'test-device', got '%s'", device.DeviceName)
	}

	// Test success recording
	mc.RecordBackupSuccess("test-device", 2*time.Minute, 1024)
	metrics, exists := mc.GetDeviceMetrics("test-device")
	if !exists {
		t.Fatal("Device metrics should exist")
	}

	if metrics.SuccessCount != 1 {
		t.Errorf("Expected success count 1, got %d", metrics.SuccessCount)
	}

	if metrics.Status != StatusHealthy {
		t.Errorf("Expected status 'healthy', got '%s'", metrics.Status)
	}

	if metrics.BackupSize != 1024 {
		t.Errorf("Expected backup size 1024, got %d", metrics.BackupSize)
	}

	// Test failure recording
	mc.RecordBackupFailure("test-device", &testError{msg: "connection failed"})
	metrics, exists = mc.GetDeviceMetrics("test-device")
	if !exists {
		t.Fatal("Device metrics should exist")
	}

	if metrics.FailureCount != 1 {
		t.Errorf("Expected failure count 1, got %d", metrics.FailureCount)
	}

	if metrics.ConsecutiveFailures != 1 {
		t.Errorf("Expected consecutive failures 1, got %d", metrics.ConsecutiveFailures)
	}

	if metrics.Status != StatusWarning {
		t.Errorf("Expected status 'warning', got '%s'", metrics.Status)
	}

	// Test multiple consecutive failures
	mc.RecordBackupFailure("test-device", &testError{msg: "connection failed again"})
	mc.RecordBackupFailure("test-device", &testError{msg: "connection failed again"})
	mc.RecordBackupFailure("test-device", &testError{msg: "connection failed again"})

	metrics, exists = mc.GetDeviceMetrics("test-device")
	if !exists {
		t.Fatal("Device metrics should exist")
	}

	if metrics.ConsecutiveFailures != 4 {
		t.Errorf("Expected consecutive failures 4, got %d", metrics.ConsecutiveFailures)
	}

	if metrics.Status != StatusCritical {
		t.Errorf("Expected status 'critical', got '%s'", metrics.Status)
	}

	// Test overall stats
	stats := mc.GetOverallStats()
	if stats.TotalDevices != 1 {
		t.Errorf("Expected total devices 1, got %d", stats.TotalDevices)
	}

	if stats.CriticalDevices != 1 {
		t.Errorf("Expected critical devices 1, got %d", stats.CriticalDevices)
	}
}

func TestAlertManager(t *testing.T) {
	mc := NewMetricsCollector()
	am := NewAlertManager(mc)

	// Test default rules
	rules := am.GetRules()
	if len(rules) == 0 {
		t.Error("Expected default alert rules to be present")
	}

	// Test alert triggering
	mc.RecordBackupFailure("test-device", &testError{msg: "connection failed"})
	mc.RecordBackupFailure("test-device", &testError{msg: "connection failed"})
	mc.RecordBackupFailure("test-device", &testError{msg: "connection failed"})

	am.CheckAlerts("test-device")

	alerts := am.GetActiveAlerts()
	if len(alerts) == 0 {
		t.Error("Expected alerts to be triggered")
	}

	// Test alert acknowledgment
	var alertID string
	for id := range alerts {
		alertID = id
		break
	}

	success := am.AcknowledgeAlert(alertID)
	if !success {
		t.Error("Expected alert acknowledgment to succeed")
	}

	alert, exists := alerts[alertID]
	if !exists {
		t.Fatal("Alert should exist")
	}

	if !alert.Acknowledged {
		t.Error("Expected alert to be acknowledged")
	}

	// Test alert clearing
	success = am.ClearAlert(alertID)
	if !success {
		t.Error("Expected alert clearing to succeed")
	}

	alerts = am.GetActiveAlerts()
	if len(alerts) != 0 {
		t.Error("Expected alerts to be cleared")
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
