# 🛡️ Aegis Backup — v0.3

**Aegis Backup** is a lightweight, concurrent, configuration-driven backup tool for MikroTik routers.  
Written in Go, it connects to your devices over SSH, exports their configurations, and saves them as timestamped `.txt` files.  

> Built for network engineers, ISPs, and sysadmins who need fast, automated, no-nonsense config backups.

---

## 🚀 What’s New in v0.3

✅ Flexible Configuration: Set the config file path via the `-config` flag or a `CONFIG_PATH` environment variable.  
✅ Smarter Defaults: If `backup_dir` is not set in your config, it now defaults automatically to `./backups`. 
✅ Connection Timeout: Added a 15-second timeout for SSH connections to prevent hangs with unresponsive devices.  
✅ Better Error Messages: Now provides clearer feedback on permission issues if a backup fails to export data.  
✅ Project Restructuring: The codebase has been organized into cmd and internal packages for better maintenance and scalability. 

---

## ✨ Features

- 🔒 SSH-based MikroTik config export (`/export`)
- 📁 Configurable device list via `config.json`
- 🧵 Concurrent backups using a worker pool
- ⏳ Connection timeout to avoid hanging on offline devices
- 🗂️ Timestamped `.rsc` file naming convention
- 📦 Local storage in a configurable backup directory
- 👀 Clear log messages and permission checks

---

## 📦 Installation

### Prerequisites

- Go 1.18+
- SSH access to devices with permission to run `/export`

### Clone the repository

```bash
git clone https://github.com/brunoavila55/aegis_backup.git
cd aegis_backup
```

### Build

```bash
go build -o aegis_backup main.go
```

---

## ⚙️ Configuration

Create or edit the `config.json` file in the project root with the following structure:

```json
{
  "backup_dir": "./backups",
  "devices": [
    {
      "name": "BR_RS_POA_3F_2C_RB01",
      "address": "172.28.0.41",
      "username": "your_user",
      "password": "your_password"
    },
    {
      "name": "BR_RS_SGB_1F_1C_RB03",
      "address": "192.168.88.1:22",
      "username": "admin",
      "password": "admin123"
    }
  ]
}
```

- `backup_dir`: Directory where backups will be saved.
- `devices`: List of devices with name, IP (with or without port), username, and password.

---

## 🧪 Running

```bash
./aegis_backup
```

📋 Example log output:
```
2025/07/07 22:00:00 Starting MikroTik Backup application...
2025/07/07 22:00:00 Found 3 devices to back up.
2025/07/07 22:00:00 Starting 5 workers...
2025/07/07 22:00:01 Worker 1: processing device BR_RS_POA_3F_2C_RB01...
2025/07/07 22:00:10 Backup for BR_RS_POA_3F_2C_RB01 completed successfully! Saved to: ./backups/BR_RS_POA_3F_2C_RB01_2025-07-07_22-00-10.rsc
```

---

## 📜 Notes

- This project currently ignores host key verification for simplicity (not recommended for production).  
- Perfect for running via cron/scheduler for automation.  
- Number of workers is currently fixed at 5 (`const numWorkers`), but you can easily adjust it in the code.

---

## 📄 License

MIT — Free to use, improve, and share.
