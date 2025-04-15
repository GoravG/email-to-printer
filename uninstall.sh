#!/bin/bash

# Check if script is run as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run this script as root or with sudo"
    exit 1
fi

echo "Starting uninstallation of Email Printer service..."

# Stop and disable the timer
echo "Stopping and disabling email-printer timer..."
systemctl stop email-printer.timer
systemctl disable email-printer.timer

# Stop and disable the service if it exists
echo "Stopping and disabling email-printer service..."
systemctl stop email-printer.service 2>/dev/null || true

# Remove systemd service and timer files
echo "Removing systemd service and timer files..."
rm -f /etc/systemd/system/email-printer.service
rm -f /etc/systemd/system/email-printer.timer

# Reload systemd to apply changes
systemctl daemon-reload

# Remove executables and configuration
echo "Removing binaries and configuration files..."
rm -f /usr/local/bin/email-printer
rm -f /usr/local/bin/run.sh
rm -f /usr/local/bin/config.yaml

# Check if user wants to remove logs
read -p "Do you want to remove log files as well? (y/n): " remove_logs
if [[ "$remove_logs" =~ ^[Yy]$ ]]; then
    echo "Removing log directory and files..."
    rm -rf /var/log/email-printer
fi

# Ask about uninstalling CUPS
read -p "Do you want to uninstall CUPS and related packages? (y/n): " uninstall_cups
if [[ "$uninstall_cups" =~ ^[Yy]$ ]]; then
    echo "Uninstalling CUPS packages..."
    apt-get remove -y cups cups-pdf
    apt-get autoremove -y
    echo "CUPS has been uninstalled."
else
    echo "CUPS installation was kept as requested."
fi

echo "Email Printer service has been successfully uninstalled."
echo "Note: System updates and upgrades performed during installation have not been reverted."

# Final check for remaining files
if [ -f /usr/local/bin/email-printer ] || [ -f /usr/local/bin/config.yaml ] || [ -f /etc/systemd/system/email-printer.service ]; then
    echo "Warning: Some files could not be removed. Please check manually."
else
    echo "All Email Printer files have been successfully removed."
fi