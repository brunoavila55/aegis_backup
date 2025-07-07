# 🛡️ Aegis Backup

**Aegis Backup** is a lightweight, concurrent, and configuration-driven backup tool for MikroTik routers.  
Built in Go, it connects to your devices over SSH and exports their configurations, saving them as timestamped text files.  

> Designed for network engineers, ISPs, and sysadmins who need automated, scriptable, no-nonsense config backups.

## ✨ Features

- 🔒 SSH-based export of MikroTik configs (`/export`)
- 📁 Configurable device list via `config.json`
- 🧵 Concurrent backups (goroutines + WaitGroup)
- 🗂️ Timestamped file naming
- 📦 Local file storage in customizable backup directory

## ⚙️ Usage

1. Add your MikroTik devices to `config.json`:
```json
{
  "backup_dir": "backups_mikrotik",
  "devices": [
    {
      "name": "POP-SantaClara",
      "address": "172.28.0.41",
      "username": "your_user",
      "password": "your_password"
    }
  ]
}
