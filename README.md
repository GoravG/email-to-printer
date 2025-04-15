# Email to Printer

## Features

- Monitors an IMAP mailbox for new emails
- Automatically downloads and prints email attachments
- Supports file type filtering (e.g., only print PDFs)
- Configurable printer selection
- Interactive and scriptable installation/uninstallation
- Systemd service and timer support for scheduled operation
- Detailed logging to file and system journal
- Runs on Linux (including Raspberry Pi) and macOS
- Easy configuration via YAML file

## Prerequisites

For binary installation:
- CUPS printing system (for Linux/Unix systems)
- Access to an IMAP email account
- A configured printer virtual printer will also work for testing

For building from source (optional):
- Go 1.24 or higher

## Quick Install (Recommended)

You can use the provided `install.sh` script to automate installation, configuration, and service setup:

```sh
sudo bash install.sh
```

This script will:
- Install required dependencies (CUPS, etc.)
- Download the correct binary for your platform
- Guide you through configuration interactively
- Set up systemd service and timer for automatic operation

## Uninstall

To completely remove the service, configuration, and logs, use:

```sh
sudo bash uninstall.sh
```

This script will:
- Stop and disable the service and timer
- Remove installed binaries, configuration, and logs
- Optionally uninstall CUPS

## Manual Installation

### Option 1: Using Release Binaries (Recommended)

1. Download the latest release for your platform from the [Releases page](https://github.com/GoravG/email-to-printer/releases)
```sh
# For macOS
curl -LO https://github.com/GoravG/email-to-printer/releases/latest/download/email-printer-darwin-amd64.tar.gz
tar xzf email-printer-darwin-amd64.tar.gz

# For Linux
curl -LO https://github.com/GoravG/email-to-printer/releases/latest/download/email-printer-linux-amd64.tar.gz
tar xzf email-printer-linux-amd64.tar.gz
```

2. Create configuration file:
```sh
cp example.config.yaml config.yaml
```

3. Edit the configuration file with your settings:
```yaml
imap_server: "imap.gmail.com:993"
email: "your.email@gmail.com"
password: "your-password"
printer_name: "Your-Printer-Name"
debug: false
allowed_file_types:
  - ".pdf"
log_level: "info"
```

### Option 2: Building from Source

1. Ensure you have Go 1.24 or higher installed
2. Clone the repository:
```sh
git clone https://github.com/GoravG/email-to-printer.git
cd email-to-printer
```

3. Build the project:
```sh
go build -o bin/email-printer cmd/main.go
```

## Running as a Service

1. Copy the service file to systemd:
```sh
sudo cp email-to-printer.service /etc/systemd/system/
```

2. Reload systemd and enable the service:
```sh
sudo systemctl daemon-reload
sudo systemctl enable email-to-printer
sudo systemctl start email-to-printer
```

3. Check service status:
```sh
sudo systemctl status email-to-printer
```

## Configuration Options

| Option | Description |
|--------|------------|
| `imap_server` | IMAP server address with port |
| `email` | Email address for IMAP login |
| `password` | Password for IMAP login |
| `printer_name` | Name of the printer to use |
| `debug` | Enable debug logging |
| `allowed_file_types` | List of allowed file extensions |
| `log_level` | Logging level (debug/info/warn/error) |
| `attachment_retention` | How long to keep temporary files |
| `allowed_senders` | List of email addresses allowed to send prints |

## Logging

Logs are stored in:
- `/var/log/email-printer/` (when running as root)
- `~/.local/share/email-printer/logs/` (when running as user)

## Development

The project structure:
```
email-to-printer/
├── cmd/
│   └── main.go          # Application entry point
├── config/
│   └── config.go        # Configuration handling
├── internal/
│   ├── email/          # Email processing
│   └── printer/        # Printing functionality
└── utils/
    └── logger.go       # Logging utilities
```