#!/bin/bash

# Set up log directory
LOG_DIR="/var/log/email-printer"
LOG_FILE="${LOG_DIR}/email-to-printer.log"

# Function to trim log file
trim_log() {
    local log_file="$1"
    local max_lines=500
    if [ -f "$log_file" ]; then
        tail -n $max_lines "$log_file" > "$log_file.tmp" && mv "$log_file.tmp" "$log_file"
    fi
}

# Create log directory with proper permissions if it doesn't exist
if [ ! -d "$LOG_DIR" ]; then
    sudo mkdir -p "$LOG_DIR"
    sudo chown $USER:$USER "$LOG_DIR"
    sudo chmod 755 "$LOG_DIR"
fi

# Create log file if it doesn't exist
if [ ! -f "$LOG_FILE" ]; then
    sudo touch "$LOG_FILE"
    sudo chown $USER:$USER "$LOG_FILE"
    sudo chmod 644 "$LOG_FILE"
fi

# Run the email printer and log output
./email-printer >> "$LOG_FILE" 2>&1

# Trim log file after adding new entries
trim_log "$LOG_FILE"