package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"aegis_backup/internal/monitoring"
)

// Server represents the API server
type Server struct {
	metrics      *monitoring.MetricsCollector
	alertManager *monitoring.AlertManager
	port         int
}

// NewServer creates a new API server
func NewServer(metrics *monitoring.MetricsCollector, alertManager *monitoring.AlertManager, port int) *Server {
	return &Server{
		metrics:      metrics,
		alertManager: alertManager,
		port:         port,
	}
}

// Start starts the API server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Metrics endpoints
	mux.HandleFunc("/api/v1/metrics", s.handleMetrics)
	mux.HandleFunc("/api/v1/metrics/", s.handleDeviceMetrics)
	mux.HandleFunc("/api/v1/stats", s.handleStats)

	// Alerts endpoints
	mux.HandleFunc("/api/v1/alerts", s.handleAlerts)
	mux.HandleFunc("/api/v1/alerts/", s.handleAlertActions)

	// Dashboard endpoint
	mux.HandleFunc("/", s.handleDashboard)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("API server starting on port %d", s.port)
	return server.ListenAndServe()
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(time.Now()).String(),
	}

	json.NewEncoder(w).Encode(response)
}

// handleMetrics handles requests for all device metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := s.metrics.GetAllMetrics()
	json.NewEncoder(w).Encode(metrics)
}

// handleDeviceMetrics handles requests for specific device metrics
func (s *Server) handleDeviceMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract device name from URL path
	deviceName := r.URL.Path[len("/api/v1/metrics/"):]
	if deviceName == "" {
		http.Error(w, "Device name required", http.StatusBadRequest)
		return
	}

	metrics, exists := s.metrics.GetDeviceMetrics(deviceName)
	if !exists {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

// handleStats handles requests for overall statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := s.metrics.GetOverallStats()
	json.NewEncoder(w).Encode(stats)
}

// handleAlerts handles requests for active alerts
func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts := s.alertManager.GetActiveAlerts()
	json.NewEncoder(w).Encode(alerts)
}

// handleAlertActions handles alert actions (acknowledge, clear)
func (s *Server) handleAlertActions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract alert ID from URL path
	alertID := r.URL.Path[len("/api/v1/alerts/"):]
	if alertID == "" {
		http.Error(w, "Alert ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		// Acknowledge alert
		success := s.alertManager.AcknowledgeAlert(alertID)
		if !success {
			http.Error(w, "Alert not found", http.StatusNotFound)
			return
		}

		response := map[string]interface{}{
			"status":   "acknowledged",
			"alert_id": alertID,
		}
		json.NewEncoder(w).Encode(response)

	case http.MethodDelete:
		// Clear alert
		success := s.alertManager.ClearAlert(alertID)
		if !success {
			http.Error(w, "Alert not found", http.StatusNotFound)
			return
		}

		response := map[string]interface{}{
			"status":   "cleared",
			"alert_id": alertID,
		}
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleDashboard serves a simple HTML dashboard
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	stats := s.metrics.GetOverallStats()
	alerts := s.alertManager.GetActiveAlerts()

	html := s.generateDashboardHTML(stats, alerts)
	w.Write([]byte(html))
}

// generateDashboardHTML generates a simple HTML dashboard
func (s *Server) generateDashboardHTML(stats monitoring.OverallStats, alerts map[string]*monitoring.Alert) string {
	criticalAlerts := 0
	warningAlerts := 0

	for _, alert := range alerts {
		if !alert.Acknowledged {
			switch alert.Severity {
			case monitoring.SeverityCritical:
				criticalAlerts++
			case monitoring.SeverityWarning:
				warningAlerts++
			}
		}
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Aegis Backup Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .stat-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stat-value { font-size: 2em; font-weight: bold; margin-bottom: 5px; }
        .stat-label { color: #666; }
        .healthy { color: #27ae60; }
        .warning { color: #f39c12; }
        .critical { color: #e74c3c; }
        .unknown { color: #95a5a6; }
        .devices-section { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .device-item { padding: 15px; margin: 10px 0; border-left: 4px solid; border-radius: 4px; display: flex; justify-content: space-between; align-items: center; }
        .device-healthy { border-left-color: #27ae60; background-color: #f0f9f0; }
        .device-warning { border-left-color: #f39c12; background-color: #fef9e7; }
        .device-critical { border-left-color: #e74c3c; background-color: #fdf2f2; }
        .device-unknown { border-left-color: #95a5a6; background-color: #f8f9fa; }
        .device-info { flex-grow: 1; }
        .device-name { font-weight: bold; font-size: 1.1em; margin-bottom: 5px; }
        .device-details { color: #666; font-size: 0.9em; }
        .device-status { font-weight: bold; padding: 4px 8px; border-radius: 4px; font-size: 0.8em; }
        .status-healthy { background-color: #27ae60; color: white; }
        .status-warning { background-color: #f39c12; color: white; }
        .status-critical { background-color: #e74c3c; color: white; }
        .status-unknown { background-color: #95a5a6; color: white; }
        .alerts-section { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .alert-item { padding: 10px; margin: 10px 0; border-left: 4px solid; border-radius: 4px; }
        .alert-critical { border-left-color: #e74c3c; background-color: #fdf2f2; }
        .alert-warning { border-left-color: #f39c12; background-color: #fef9e7; }
        .refresh-btn { background: #3498db; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #2980b9; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🛡️ Aegis Backup Dashboard</h1>
            <p>Real-time monitoring and alerting system</p>
        </div>
        
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Total Devices</div>
            </div>
            <div class="stat-card">
                <div class="stat-value healthy">%d</div>
                <div class="stat-label">Healthy</div>
            </div>
            <div class="stat-card">
                <div class="stat-value warning">%d</div>
                <div class="stat-label">Warning</div>
            </div>
            <div class="stat-card">
                <div class="stat-value critical">%d</div>
                <div class="stat-label">Critical</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Total Backups</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Total Failures</div>
            </div>
        </div>
        
        <div class="devices-section">
            <h2>📱 Device Status</h2>
            <button class="refresh-btn" onclick="location.reload()">Refresh</button>
            <div id="devices">
                %s
            </div>
        </div>
        
        <div class="alerts-section">
            <h2>🚨 Active Alerts (%d Critical, %d Warning)</h2>
            <div id="alerts">
                %s
            </div>
        </div>
        
        <script>
            // Auto-refresh every 30 seconds
            setTimeout(function() {
                location.reload();
            }, 30000);
        </script>
    </div>
</body>
</html>`,
		stats.TotalDevices,
		stats.HealthyDevices,
		stats.WarningDevices,
		stats.CriticalDevices,
		stats.TotalBackups,
		stats.TotalFailures,
		s.generateDevicesHTML(),
		criticalAlerts,
		warningAlerts,
		s.generateAlertsHTML(alerts),
	)
}

// generateDevicesHTML generates HTML for devices
func (s *Server) generateDevicesHTML() string {
	devices := s.metrics.GetAllMetrics()

	if len(devices) == 0 {
		return "<p>No devices configured</p>"
	}

	html := ""
	for _, device := range devices {
		// Determine CSS class based on status
		deviceClass := "device-unknown"
		statusClass := "status-unknown"
		statusText := "Unknown"

		switch device.Status {
		case monitoring.StatusHealthy:
			deviceClass = "device-healthy"
			statusClass = "status-healthy"
			statusText = "Healthy"
		case monitoring.StatusWarning:
			deviceClass = "device-warning"
			statusClass = "status-warning"
			statusText = "Warning"
		case monitoring.StatusCritical:
			deviceClass = "device-critical"
			statusClass = "status-critical"
			statusText = "Critical"
		}

		// Format last backup time
		lastBackup := "Never"
		if !device.LastBackup.IsZero() {
			lastBackup = device.LastBackup.Format("2006-01-02 15:04:05")
		}

		// Format average time
		avgTime := "N/A"
		if device.AverageTime > 0 {
			avgTime = device.AverageTime.String()
		}

		html += fmt.Sprintf(`
			<div class="device-item %s">
				<div class="device-info">
					<div class="device-name">%s</div>
					<div class="device-details">
						Last backup: %s | Avg time: %s | Success: %d | Failures: %d
						%s
					</div>
				</div>
				<div class="device-status %s">%s</div>
			</div>`,
			deviceClass,
			device.DeviceName,
			lastBackup,
			avgTime,
			device.SuccessCount,
			device.FailureCount,
			func() string {
				if device.LastError != "" {
					return fmt.Sprintf(" | Last error: %s", device.LastError)
				}
				return ""
			}(),
			statusClass,
			statusText,
		)
	}

	return html
}

// generateAlertsHTML generates HTML for alerts
func (s *Server) generateAlertsHTML(alerts map[string]*monitoring.Alert) string {
	if len(alerts) == 0 {
		return "<p>No active alerts</p>"
	}

	html := ""
	for _, alert := range alerts {
		if alert.Acknowledged {
			continue
		}

		class := "alert-warning"
		if alert.Severity == monitoring.SeverityCritical {
			class = "alert-critical"
		}

		html += fmt.Sprintf(`
			<div class="alert-item %s">
				<strong>%s</strong> - %s<br>
				<small>Device: %s | Time: %s</small>
			</div>`,
			class,
			alert.Severity,
			alert.Message,
			alert.DeviceName,
			alert.Timestamp.Format("2006-01-02 15:04:05"),
		)
	}

	return html
}
