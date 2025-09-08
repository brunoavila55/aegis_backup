package monitoring

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

// AlertRule defines when and how to send alerts
type AlertRule struct {
	Name          string        `json:"name"`
	Condition     string        `json:"condition"`
	Severity      AlertSeverity `json:"severity"`
	Channels      []string      `json:"channels"`
	Cooldown      time.Duration `json:"cooldown"`
	LastTriggered time.Time     `json:"last_triggered"`
	Enabled       bool          `json:"enabled"`
}

// Alert represents an active alert
type Alert struct {
	ID           string        `json:"id"`
	DeviceName   string        `json:"device_name"`
	RuleName     string        `json:"rule_name"`
	Severity     AlertSeverity `json:"severity"`
	Message      string        `json:"message"`
	Timestamp    time.Time     `json:"timestamp"`
	Channels     []string      `json:"channels"`
	Acknowledged bool          `json:"acknowledged"`
}

// AlertManager manages alert rules and active alerts
type AlertManager struct {
	rules        map[string]*AlertRule
	activeAlerts map[string]*Alert
	metrics      *MetricsCollector
	mu           sync.RWMutex
}

// NewAlertManager creates a new alert manager
func NewAlertManager(metrics *MetricsCollector) *AlertManager {
	am := &AlertManager{
		rules:        make(map[string]*AlertRule),
		activeAlerts: make(map[string]*Alert),
		metrics:      metrics,
	}

	// Add default alert rules
	am.addDefaultRules()

	return am
}

// addDefaultRules adds default alert rules
func (am *AlertManager) addDefaultRules() {
	defaultRules := []*AlertRule{
		{
			Name:      "consecutive_failures",
			Condition: "consecutive_failures >= 3",
			Severity:  SeverityCritical,
			Channels:  []string{"telegram"},
			Cooldown:  1 * time.Hour,
			Enabled:   true,
		},
		{
			Name:      "backup_size_change",
			Condition: "backup_size_change > 50%",
			Severity:  SeverityWarning,
			Channels:  []string{"telegram"},
			Cooldown:  24 * time.Hour,
			Enabled:   true,
		},
		{
			Name:      "backup_timeout",
			Condition: "backup_time > 5m",
			Severity:  SeverityWarning,
			Channels:  []string{"telegram"},
			Cooldown:  2 * time.Hour,
			Enabled:   true,
		},
		{
			Name:      "no_backup_24h",
			Condition: "last_backup_age > 24h",
			Severity:  SeverityCritical,
			Channels:  []string{"telegram"},
			Cooldown:  6 * time.Hour,
			Enabled:   true,
		},
	}

	for _, rule := range defaultRules {
		am.rules[rule.Name] = rule
	}
}

// CheckAlerts checks all alert rules for a device
func (am *AlertManager) CheckAlerts(deviceName string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	deviceMetrics, exists := am.metrics.GetDeviceMetrics(deviceName)
	if !exists {
		return
	}

	for _, rule := range am.rules {
		if !rule.Enabled {
			continue
		}

		// Check cooldown
		if time.Since(rule.LastTriggered) < rule.Cooldown {
			continue
		}

		// Check if condition is met
		if am.evaluateCondition(rule.Condition, deviceMetrics) {
			am.triggerAlert(deviceName, rule, deviceMetrics)
			rule.LastTriggered = time.Now()
		}
	}
}

// evaluateCondition evaluates an alert condition
func (am *AlertManager) evaluateCondition(condition string, metrics *DeviceMetrics) bool {
	condition = strings.ToLower(condition)

	switch {
	case strings.Contains(condition, "consecutive_failures >= 3"):
		return metrics.ConsecutiveFailures >= 3
	case strings.Contains(condition, "consecutive_failures >= 1"):
		return metrics.ConsecutiveFailures >= 1
	case strings.Contains(condition, "backup_time > 5m"):
		return metrics.AverageTime > 5*time.Minute
	case strings.Contains(condition, "last_backup_age > 24h"):
		return time.Since(metrics.LastBackup) > 24*time.Hour
	case strings.Contains(condition, "last_backup_age > 12h"):
		return time.Since(metrics.LastBackup) > 12*time.Hour
	case strings.Contains(condition, "status == critical"):
		return metrics.Status == StatusCritical
	case strings.Contains(condition, "status == warning"):
		return metrics.Status == StatusWarning
	}

	return false
}

// triggerAlert creates and triggers an alert
func (am *AlertManager) triggerAlert(deviceName string, rule *AlertRule, metrics *DeviceMetrics) {
	alertID := fmt.Sprintf("%s_%s_%d", deviceName, rule.Name, time.Now().Unix())

	message := am.generateAlertMessage(deviceName, rule, metrics)

	alert := &Alert{
		ID:           alertID,
		DeviceName:   deviceName,
		RuleName:     rule.Name,
		Severity:     rule.Severity,
		Message:      message,
		Timestamp:    time.Now(),
		Channels:     rule.Channels,
		Acknowledged: false,
	}

	am.activeAlerts[alertID] = alert

	log.Printf("ALERT [%s] %s: %s", rule.Severity, deviceName, message)
}

// generateAlertMessage generates a human-readable alert message
func (am *AlertManager) generateAlertMessage(deviceName string, rule *AlertRule, metrics *DeviceMetrics) string {
	switch rule.Name {
	case "consecutive_failures":
		return fmt.Sprintf("Device %s has failed %d consecutive backups. Last error: %s",
			deviceName, metrics.ConsecutiveFailures, metrics.LastError)
	case "backup_timeout":
		return fmt.Sprintf("Device %s backup is taking too long (%.1f minutes)",
			deviceName, metrics.AverageTime.Minutes())
	case "no_backup_24h":
		return fmt.Sprintf("Device %s has not been backed up for %.1f hours",
			deviceName, time.Since(metrics.LastBackup).Hours())
	case "backup_size_change":
		return fmt.Sprintf("Device %s backup size has changed significantly", deviceName)
	default:
		return fmt.Sprintf("Alert triggered for device %s: %s", deviceName, rule.Name)
	}
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts() map[string]*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make(map[string]*Alert)
	for id, alert := range am.activeAlerts {
		result[id] = alert
	}

	return result
}

// AcknowledgeAlert acknowledges an alert
func (am *AlertManager) AcknowledgeAlert(alertID string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.activeAlerts[alertID]
	if !exists {
		return false
	}

	alert.Acknowledged = true
	return true
}

// ClearAlert removes an alert
func (am *AlertManager) ClearAlert(alertID string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	_, exists := am.activeAlerts[alertID]
	if !exists {
		return false
	}

	delete(am.activeAlerts, alertID)
	return true
}

// AddRule adds a new alert rule
func (am *AlertManager) AddRule(rule *AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.rules[rule.Name] = rule
}

// GetRules returns all alert rules
func (am *AlertManager) GetRules() map[string]*AlertRule {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make(map[string]*AlertRule)
	for name, rule := range am.rules {
		result[name] = rule
	}

	return result
}

// UpdateRule updates an existing alert rule
func (am *AlertManager) UpdateRule(name string, rule *AlertRule) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	_, exists := am.rules[name]
	if !exists {
		return false
	}

	am.rules[name] = rule
	return true
}

// DeleteRule removes an alert rule
func (am *AlertManager) DeleteRule(name string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	_, exists := am.rules[name]
	if !exists {
		return false
	}

	delete(am.rules, name)
	return true
}
