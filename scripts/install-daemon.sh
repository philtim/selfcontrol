#!/bin/bash

# Installation script for SelfControl daemon
# Run with: sudo ./scripts/install-daemon.sh

set -e

if [ "$EUID" -ne 0 ]; then
    echo "Please run with sudo"
    exit 1
fi

echo "Installing SelfControl daemon..."

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "Detected Linux - installing systemd service"

    # Copy binary
    cp selfcontrol-daemon /usr/local/bin/
    chmod +x /usr/local/bin/selfcontrol-daemon

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

    # Copy binary
    cp selfcontrol-daemon /usr/local/bin/
    chmod +x /usr/local/bin/selfcontrol-daemon

    # Copy plist file
    cp scripts/com.selfcontrol.daemon.plist /Library/LaunchDaemons/

    # Load service
    launchctl load /Library/LaunchDaemons/com.selfcontrol.daemon.plist

    echo "✓ Service installed and started"
    echo "Check status with: sudo launchctl list | grep selfcontrol"
    echo "View logs with: tail -f /var/log/selfcontrol-daemon.log"

else
    echo "Unsupported OS: $OSTYPE"
    exit 1
fi

echo ""
echo "Daemon installation complete!"
