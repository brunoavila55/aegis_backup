package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client represents a Telegram bot client
type Client struct {
	botToken string
	chatID   string
	baseURL  string
	client   *http.Client
}

// NewClient creates a new Telegram client
func NewClient(botToken, chatID string) *Client {
	return &Client{
		botToken: botToken,
		chatID:   chatID,
		baseURL:  fmt.Sprintf("https://api.telegram.org/bot%s", botToken),
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: true,
			},
		},
	}
}

// SendMessage sends a text message to the configured chat
func (c *Client) SendMessage(message string) error {
	url := fmt.Sprintf("%s/sendMessage", c.baseURL)

	payload := map[string]any{
		"chat_id":    c.chatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// SendDocument sends a file to the configured chat
func (c *Client) SendDocument(filePath, caption string) error {
	url := fmt.Sprintf("%s/sendDocument", c.baseURL)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add chat_id field
	if err := writer.WriteField("chat_id", c.chatID); err != nil {
		return fmt.Errorf("failed to write chat_id field: %w", err)
	}

	// Add caption field if provided
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return fmt.Errorf("failed to write caption field: %w", err)
		}
		if err := writer.WriteField("parse_mode", "HTML"); err != nil {
			return fmt.Errorf("failed to write parse_mode field: %w", err)
		}
	}

	// Add document field
	part, err := writer.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content to form
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// TestConnection tests the bot connection by sending a simple message
func (c *Client) TestConnection() error {
	url := fmt.Sprintf("%s/getMe", c.baseURL)

	resp, err := c.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to test connection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// FormatBackupSummary creates a formatted message for backup completion
func FormatBackupSummary(deviceCount int, duration time.Duration, zipFile string, timezone string) string {
	// Parse timezone
	location, err := time.LoadLocation(timezone)
	if err != nil {
		location = time.UTC
	}

	currentTime := time.Now().In(location)

	return fmt.Sprintf(`🛡️ <b>Aegis Backup Completed</b>

📊 <b>Summary:</b>
• Devices backed up: %d
• Duration: %v
• ZIP file: %s
• Date: %s (%s)

✅ All configurations have been successfully backed up and compressed.`,
		deviceCount,
		duration,
		filepath.Base(zipFile),
		currentTime.Format("2006-01-02 15:04:05"),
		timezone)
}

// FormatErrorMessage creates a formatted error message
func FormatErrorMessage(operation string, err error, timezone string) string {
	// Parse timezone
	location, parseErr := time.LoadLocation(timezone)
	if parseErr != nil {
		location = time.UTC
	}

	currentTime := time.Now().In(location)

	return fmt.Sprintf(`❌ <b>Aegis Backup Error</b>

🔧 <b>Operation:</b> %s
⚠️ <b>Error:</b> %s
📅 <b>Time:</b> %s (%s)

Please check the logs for more details.`,
		operation,
		err.Error(),
		currentTime.Format("2006-01-02 15:04:05"),
		timezone)
}
