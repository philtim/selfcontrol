#!/bin/bash

# Installation script for SelfControl daemon
# Run with: sudo ./scripts/install-daemon.sh
# Handles both fresh installs and updates

set -e

if [ "$EUID" -ne 0 ]; then
    echo "Please run with sudo"
    exit 1
fi

echo "Installing SelfControl daemon..."

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "Detected Linux - installing systemd service"

    # Binary should already be installed by 'make install'
    # Just verify it exists
    if [ ! -f /usr/local/bin/selfcontrol-daemon ]; then
        echo "Error: selfcontrol-daemon not found in /usr/local/bin"
        echo "Please run 'make install' first"
        exit 1
    fi

    # Stop service if already running
    systemctl stop selfcontrol-daemon 2>/dev/null || true

    # Copy service file
    cp scripts/selfcontrol-daemon.service /etc/systemd/system/

    # Enable and start service
    systemctl daemon-reload
    systemctl enable selfcontrol-daemon
    systemctl start selfcontrol-daemon

    echo "✓ Service installed and started"
    echo "Check status with: sudo systemctl status selfcontrol-daemon"
    echo "View logs with: sudo journalctl -u selfcontrol-daemon -f"

elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Detected macOS - installing launchd service"

    # Binary should already be installed by 'make install'
    # Just verify it exists
    if [ ! -f /usr/local/bin/selfcontrol-daemon ]; then
        echo "Error: selfcontrol-daemon not found in /usr/local/bin"
        echo "Please run 'make install' first"
        exit 1
    fi

    # Unload service if already running
    launchctl unload /Library/LaunchDaemons/com.selfcontrol.daemon.plist 2>/dev/null || true

    # Copy plist file
    cp scripts/com.selfcontrol.daemon.plist /Library/LaunchDaemons/

    # Load and start service
    launchctl load /Library/LaunchDaemons/com.selfcontrol.daemon.plist

    # Give it a moment to start
    sleep 1

    echo "✓ Service installed and started"
    echo "Check status with: sudo launchctl list | grep selfcontrol"
    echo "View logs with: tail -f /var/log/selfcontrol-daemon.log"

else
    echo "Unsupported OS: $OSTYPE"
    exit 1
fi

echo ""
echo "Daemon installation complete!"
