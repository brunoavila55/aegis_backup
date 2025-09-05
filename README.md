# 🛡️ Aegis Backup — v0.5

**Aegis Backup** is a lightweight, concurrent, configuration-driven backup tool for MikroTik routers.  
Written in Go, it connects to your devices over SSH, exports their configurations, and saves them as timestamped `.txt` files.  

> Built for network engineers, ISPs, and sysadmins who need fast, automated, no-nonsense config backups with intelligent notifications.

---

## 🚀 What's New in v0.5

✅ **Telegram Integration**: Automatic notifications and file sharing via Telegram bot  
✅ **Daily ZIP Archives**: Compress daily backups into organized ZIP files  
✅ **Smart Notifications**: Backup summaries, error alerts, and file delivery  
✅ **Automatic Cleanup**: Remove old ZIP files to save disk space  
✅ **Enhanced Monitoring**: Detailed logs for all operations including Telegram activities  

---

## 🚀 What's New in v0.4

✅ **Automated Scheduling**: Run backups automatically using cron expressions - daily, weekly, or any custom schedule  
✅ **Daemon Mode**: Run as a background service for continuous automated backups  
✅ **Timezone Support**: Configure backups to run in your local timezone  
✅ **Graceful Shutdown**: Proper signal handling for clean service stops  
✅ **Flexible Execution**: Choose between one-time runs or continuous daemon mode  

---

## ✨ Features

- **🤖 Telegram Integration**: Automated notifications and ZIP file delivery to your Telegram group
- **📦 Daily ZIP Archives**: Automatic compression of daily backups with cleanup
- **🔄 Automated Scheduling**: Set up recurring backups with cron expressions
- **⚡ Concurrent Processing**: Backup multiple devices simultaneously using worker pools
- **🌍 Timezone Support**: Schedule backups in your local timezone
- **🛠️ Flexible Configuration**: JSON-based configuration with CSV device lists
- **📁 Organized Storage**: Timestamped backups with device-specific naming
- **🔒 SSH Security**: Secure connections to your MikroTik devices
- **🚀 Daemon Mode**: Run as a background service for continuous operation
- **📊 Detailed Logging**: Comprehensive logging for monitoring and troubleshooting
- **🧹 Smart Cleanup**: Automatic removal of old archives to manage disk space

---

## 📋 Usage

### One-Time Backup (Default Mode)
```bash
# Run a single backup with default config
./aegis-backup

# Use custom config and devices files
./aegis-backup -config /path/to/config.json -devices /path/to/devices.csv
```

### Daemon Mode (Scheduled Backups)
```bash
# Run as daemon with automatic scheduling
./aegis-backup -daemon

# Run daemon with custom config
./aegis-backup -daemon -config /path/to/config.json -devices /path/to/devices.csv
```

### Command Line Options
- `-config`: Path to configuration file (default: `config.json`)
- `-devices`: Path to devices CSV file (default: `devices.csv`)
- `-daemon`: Run as daemon service with scheduler

---

## ⚙️ Configuration

### Main Configuration (`config.json`)
```json
{
    "backup_dir": "backups_mikrotik",
    "schedule": {
        "enabled": true,
        "cron": "0 2 * * *",
        "timezone": "America/Sao_Paulo"
    },
    "telegram": {
        "enabled": true,
        "bot_token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
        "chat_id": "-1001234567890",
        "send_zip": true,
        "send_logs": true
    }
}
```

#### Schedule Configuration Options:
- **`enabled`**: Enable/disable automatic scheduling (boolean)
- **`cron`**: Cron expression for backup timing (string)
- **`timezone`**: Timezone for scheduling (string, defaults to "UTC")

#### Telegram Configuration Options:
- **`enabled`**: Enable/disable Telegram notifications (boolean)
- **`bot_token`**: Your Telegram bot token (string)
- **`chat_id`**: Target chat/group ID (string)
- **`send_zip`**: Send daily ZIP archives (boolean)
- **`send_logs`**: Send backup completion summaries (boolean)

#### Common Cron Expressions:
- `"0 2 * * *"` - Daily at 2:00 AM
- `"0 */6 * * *"` - Every 6 hours
- `"0 0 * * 0"` - Weekly on Sunday at midnight
- `"0 1 1 * *"` - Monthly on the 1st at 1:00 AM
- `"0 9-17 * * 1-5"` - Every hour from 9 AM to 5 PM, Monday to Friday

#### Supported Timezones:
Use standard timezone names like:
- `"UTC"` (default)
- `"America/Sao_Paulo"`
- `"America/New_York"`
- `"Europe/London"`
- `"Asia/Tokyo"`

### Device Configuration (`devices.csv`)
```csv
name,address,username,password
Router-01,192.168.1.1,admin,password123
Router-02,192.168.1.2,admin,password456
```

---

## 🤖 Telegram Setup

### 1. Create a Telegram Bot
1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Send `/newbot` and follow the instructions
3. Save the bot token (e.g., `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

### 2. Get Chat ID
**For a Group:**
1. Add your bot to the group
2. Send a message in the group
3. Visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
4. Look for `"chat":{"id":-1001234567890}` in the response

**For a Private Chat:**
1. Start a conversation with your bot
2. Send any message
3. Visit the same URL as above
4. Look for `"chat":{"id":123456789}` (positive number for private chats)

### 3. Configure Permissions
- Ensure the bot can send messages to the target chat
- For groups, the bot needs permission to send files

### 4. Test Configuration
The application will test the Telegram connection on startup and log any issues.

---

## 📦 ZIP Archive Features

- **Daily Compression**: All backups from the same day are compressed into a single ZIP file
- **Organized Naming**: ZIP files are named `backups_YYYY-MM-DD.zip`
- **Automatic Cleanup**: Old ZIP files are removed after 30 days
- **Telegram Delivery**: ZIP files are automatically sent to your configured Telegram chat
- **Error Handling**: Failed compressions are logged and reported via Telegram

---

## 🔧 Installation

### Download Binary
Download the latest release from the releases page and extract it to your preferred directory.

### Build from Source
```bash
git clone https://github.com/yourusername/aegis-backup.git
cd aegis-backup
go build -o aegis-backup ./cmd
```

---

## 🐳 Running as a System Service

### Linux (systemd)
Create a service file at `/etc/systemd/system/aegis-backup.service`:

```ini
[Unit]
Description=Aegis Backup Service
After=network.target

[Service]
Type=simple
User=aegis
WorkingDirectory=/opt/aegis-backup
ExecStart=/opt/aegis-backup/aegis-backup -daemon
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl enable aegis-backup
sudo systemctl start aegis-backup
sudo systemctl status aegis-backup
```

### Windows (Service)
Use tools like NSSM (Non-Sucking Service Manager) to run as a Windows service:

```cmd
nssm install AegisBackup "C:\path\to\aegis-backup.exe" "-daemon"
nssm start AegisBackup
```

---

## 📊 Monitoring

### Logs
The application provides detailed logging for monitoring:
- Scheduler start/stop events
- Backup execution times and results
- ZIP file creation and cleanup
- Telegram connection status and message delivery
- Error reporting and warnings

### Telegram Notifications
When enabled, you'll receive:
- **Backup Summaries**: Device count, duration, and ZIP file info
- **Error Alerts**: Detailed error messages with timestamps
- **ZIP Files**: Daily backup archives delivered directly to your chat

### Checking Service Status
```bash
# Check if daemon is running
ps aux | grep aegis-backup

# View logs (if using systemd)
journalctl -u aegis-backup -f

# Check next scheduled backup
# (This information is logged when the daemon starts)
```

---

## 🛠️ Troubleshooting

### Common Issues

**Scheduler not starting:**
- Verify `schedule.enabled` is set to `true`
- Check that the cron expression is valid
- Ensure timezone is correctly specified

**Telegram not working:**
- Verify bot token and chat ID are correct
- Check that the bot has permission to send messages to the target chat
- Test the bot manually by messaging it first
- Review logs for connection errors

**ZIP files not created:**
- Ensure there are backup files from the current day
- Check disk space and write permissions
- Review logs for compression errors

**Files not sent to Telegram:**
- Verify `send_zip` is enabled in configuration
- Check file size limits (Telegram has a 50MB limit for bots)
- Ensure stable internet connection

**Permission errors:**
- Ensure the user running the service has write access to the backup directory
- Check SSH credentials and network connectivity to devices

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.