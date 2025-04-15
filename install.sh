#!/bin/bash

# Check if script is run as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run this script as root or with sudo"
    exit 1
fi

# Update and upgrade the system
apt update
apt upgrade -y
apt-get install -y cups
apt install cups-pdf
systemctl enable cups
systemctl start cups
cupsctl --remote-any
systemctl restart cups

# Verify CUPS installation
if command -v lpstat &> /dev/null; then
    echo "CUPS installed and running successfully"
else
    echo "Error: CUPS installation failed"
    exit 1
fi

arch=$(uname -m)
os=$(uname -s)

# Construct download URL based on architecture and OS
case "$os-$arch" in
  Linux-x86_64)
    binary_url="https://github.com/GoravG/email-to-printer/releases/download/latest/email-printer-linux-amd64.tar.gz"
    ;;
  Linux-armv7l|Linux-armv6l)
    binary_url="https://github.com/GoravG/email-to-printer/releases/download/latest/email-printer-linux-arm.tar.gz"
    ;;
  Linux-aarch64)
    binary_url="https://github.com/GoravG/email-to-printer/releases/download/latest/email-printer-linux-arm64.tar.gz"
    ;;
  Darwin-x86_64)
    binary_url="https://github.com/GoravG/email-to-printer/releases/download/latest/email-printer-darwin-amd64.tar.gz"
    ;;
  Darwin-aarch64)
    binary_url="https://github.com/GoravG/email-to-printer/releases/download/latest/email-printer-darwin-arm64.tar.gz"
    ;;
  *)
    echo "Unsupported OS/architecture: $os-$arch"
    exit 1
    ;;
esac

# Download the binary
echo "Downloading binary from $binary_url"
wget -q "$binary_url" -O email-printer.tar.gz

# Extract the binary
tar -xzf email-printer.tar.gz

# Set executable permissions on the binary
chmod +x email-printer

# Configuration file name
CONFIG_FILE="config.yaml"

# Prompt user for configuration details with validation
while true; do
    read -p "Enter IMAP server (e.g., imap.gmail.com:993): " imap_server
    if [[ -z "$imap_server" ]]; then
        echo "IMAP server cannot be empty. Please try again."
    else
        break
    fi
done

while true; do
    read -p "Enter email address: " email
    if [[ "$email" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        break
    else
        echo "Invalid email format. Please try again."
    fi
done

while true; do
    read -s -p "Enter password: " password
    echo  # Add a newline after the password prompt
    if [[ -z "$password" ]]; then
        echo "Password cannot be empty. Please try again."
    else
        break
    fi
done

# Show available printers and let user select
echo "Available printers:"
lpstat -p | cut -d' ' -f2 | while read -r printer; do
    echo "- $printer"
done

while true; do
    read -p "Enter printer name from the list above: " printer_name
    if lpstat -p | cut -d' ' -f2 | grep -q "^${printer_name}$"; then
        break
    else
        echo "Invalid printer name. Please select from the list above."
    fi
done

while true; do
    read -p "Enable debug mode? (true/false): " debug
    if [[ "$debug" == "true" || "$debug" == "false" ]]; then
        break
    else
        echo "Please enter 'true' or 'false'"
    fi
done

while true; do
    read -p "Enter allowed file types (comma-separated, e.g., .pdf,.txt): " allowed_file_types
    if [[ -z "$allowed_file_types" ]]; then
        echo "At least one file type must be specified. Please try again."
    else
        break
    fi
done

while true; do
    read -p "Enter allowed senders (comma-separated, e.g., email1,email2): " allowed_senders
    if [[ -z "$allowed_senders" ]]; then
        echo "At least one sender must be specified. Please try again."
    else
        break
    fi
done

# Create config.yaml file with all values quoted
cat > "$CONFIG_FILE" <<EOL
imap_server: "${imap_server}"
email: "${email}"
password: "${password}"
printer_name: "${printer_name}"
debug: "${debug}"
allowed_file_types:
$(IFS=','; for type in $allowed_file_types; do echo "  - \"$type\""; done)
allowed_senders:
$(IFS=','; for sender in $allowed_senders; do echo "  - \"$sender\""; done)
EOL

echo "Configuration file created: $CONFIG_FILE"

# Move the binary to /usr/local/bin
mv email-printer /usr/local/bin/
echo "Binary moved to /usr/local/bin"

# Move config.yaml to /usr/local/bin
mv "$CONFIG_FILE" /usr/local/bin/
echo "config.yaml moved to /usr/local/bin"

# Determine the correct user for service without using sudo
if [ -z "${SUDO_USER}" ]; then
    # Script was run with 'su' or directly as root
    echo "Available users:"
    getent passwd | grep /home | cut -d: -f1 | while read -r user; do
        echo "- $user"
    done
    
    while true; do
        read -p "Enter the user to run the service as: " service_user
        if id "$service_user" >/dev/null 2>&1; then
            break
        else
            echo "Invalid user. Please select from the list above."
        fi
    done
else
    # Script was run with sudo
    service_user="${SUDO_USER}"
    echo "Service will run as user: $service_user"
fi

# Show available groups for selected user
echo "Available groups for $service_user:"
groups $service_user | tr ' ' '\n' | sort | uniq | while read -r group; do
    if [[ -n "$group" ]]; then
        echo "- $group"
    fi
done

while true; do
    read -p "Enter the group to run the service as (default: $service_user): " service_group
    if [ -z "$service_group" ]; then
        service_group=$service_user
        break
    elif getent group "$service_group" >/dev/null 2>&1; then
        if groups $service_user | grep -q "\b${service_group}\b"; then
            break
        else
            echo "User $service_user is not a member of group $service_group"
        fi
    else
        echo "Invalid group. Please select from the list above."
    fi
done

# Set up log directory and file
LOG_DIR="/var/log/email-printer"
LOG_FILE="${LOG_DIR}/email-to-printer.log"

# Create log directory with proper permissions
mkdir -p "$LOG_DIR"
chown $service_user:$service_group "$LOG_DIR"
chmod 755 "$LOG_DIR"

# Create log file
touch "$LOG_FILE"
chown $service_user:$service_group "$LOG_FILE"
chmod 644 "$LOG_FILE"

# Create simplified run.sh
cat > run.sh <<EOL
#!/bin/bash

# Set up log paths
LOG_FILE="/var/log/email-printer/email-to-printer.log"

# Function to trim log file
trim_log() {
    local max_lines=500
    if [ -f "\$LOG_FILE" ]; then
        tail -n \$max_lines "\$LOG_FILE" > "\$LOG_FILE.tmp" && mv "\$LOG_FILE.tmp" "\$LOG_FILE"
    fi
}

# Run the email printer and log output
/usr/local/bin/email-printer >> "\$LOG_FILE" 2>&1

# Trim log file after adding new entries
trim_log
EOL

mv run.sh /usr/local/bin/
chmod +x /usr/local/bin/run.sh

# Create the service file with the dynamic user
cat > email-printer.service <<EOL
[Unit]
Description=Email to Printer Service
After=network.target

[Service]
WorkingDirectory=/usr/local/bin
ExecStart=/usr/local/bin/run.sh
Restart=on-failure
User=$service_user
Group=$service_group

[Install]
WantedBy=multi-user.target
EOL

# Create the timer file
cat > email-printer.timer <<EOL
[Unit]
Description=Run Email to Printer every 5 minutes

[Timer]
OnCalendar=*:0/5
Persistent=false

[Install]
WantedBy=timers.target
EOL

# Copy the service and timer files to the systemd directory
mv email-printer.service /etc/systemd/system/
mv email-printer.timer /etc/systemd/system/

# Enable the timer
systemctl enable email-printer.timer

# Start the timer
systemctl start email-printer.timer

echo "Email to Printer service and timer installed and started."