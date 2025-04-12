#!/bin/bash

# Function to trim log file
trim_log() {
    local log_file="$1"
    local max_lines=500
    if [ -f "$log_file" ]; then
        tail -n $max_lines "$log_file" > "$log_file.tmp" && mv "$log_file.tmp" "$log_file"
    fi
}

cd "$HOME" || exit 1
./email-printer >> email-to-printer.log 2>&1

# Trim log file after adding new entries
trim_log "email-to-printer.log"
