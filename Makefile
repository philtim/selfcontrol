.PHONY: build install clean test run daemon install-daemon uninstall-daemon update stop-daemon

# Build both binaries to dist/ folder
build:
	@echo "Building selfcontrol..."
	@mkdir -p dist
	go build -o dist/selfcontrol ./cmd/selfcontrol
	@echo "Building selfcontrol-daemon..."
	go build -o dist/selfcontrol-daemon ./cmd/selfcontrol-daemon
	@echo "✓ Build complete"
	@echo "Binaries available in: dist/"

# Stop daemon if running (cross-platform)
stop-daemon:
	@echo "Stopping daemon if running..."
	@if [ "$(shell uname)" = "Linux" ]; then \
		sudo systemctl stop selfcontrol-daemon 2>/dev/null || true; \
	elif [ "$(shell uname)" = "Darwin" ]; then \
		sudo launchctl unload /Library/LaunchDaemons/com.selfcontrol.daemon.plist 2>/dev/null || true; \
	fi

# Install binaries to /usr/local/bin (stops daemon first if running)
install: build stop-daemon
	@echo "Installing binaries..."
	sudo cp dist/selfcontrol /usr/local/bin/
	sudo cp dist/selfcontrol-daemon /usr/local/bin/
	sudo chmod +x /usr/local/bin/selfcontrol
	sudo chmod +x /usr/local/bin/selfcontrol-daemon
	@echo "✓ Binaries installed to /usr/local/bin"

# Install daemon service (restarts if already installed)
install-daemon: install
	@echo "Installing daemon service..."
	sudo ./scripts/install-daemon.sh
	@echo "✓ Daemon installed and started"

# Update everything (build, install, restart daemon)
update: install-daemon
	@echo ""
	@echo "✓ Update complete!"
	@echo ""
	@echo "To verify daemon is running:"
	@if [ "$(shell uname)" = "Linux" ]; then \
		echo "  sudo systemctl status selfcontrol-daemon"; \
		echo "  sudo journalctl -u selfcontrol-daemon -f"; \
	elif [ "$(shell uname)" = "Darwin" ]; then \
		echo "  sudo launchctl list | grep selfcontrol"; \
		echo "  tail -f /var/log/selfcontrol-daemon.log"; \
	fi

# Uninstall daemon service
uninstall-daemon:
	@echo "Uninstalling daemon service..."
	@if [ "$(shell uname)" = "Linux" ]; then \
		sudo systemctl stop selfcontrol-daemon || true; \
		sudo systemctl disable selfcontrol-daemon || true; \
		sudo rm -f /etc/systemd/system/selfcontrol-daemon.service; \
		sudo systemctl daemon-reload; \
	elif [ "$(shell uname)" = "Darwin" ]; then \
		sudo launchctl unload /Library/LaunchDaemons/com.selfcontrol.daemon.plist || true; \
		sudo rm -f /Library/LaunchDaemons/com.selfcontrol.daemon.plist; \
	fi
	@echo "✓ Daemon uninstalled"

# Run the TUI (with sudo)
run: build
	sudo ./dist/selfcontrol

# Run the daemon in foreground (for testing)
daemon: build
	sudo ./dist/selfcontrol-daemon

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf dist
	@echo "✓ Clean complete"

# Run tests
test:
	go test -v ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

# Show help
help:
	@echo "SelfControl TUI - Makefile commands"
	@echo ""
	@echo "Installation & Updates:"
	@echo "  make update          - Build, install, and restart daemon (recommended)"
	@echo "  make install         - Install binaries to /usr/local/bin"
	@echo "  make install-daemon  - Install and start background daemon"
	@echo "  make uninstall-daemon- Stop and remove background daemon"
	@echo ""
	@echo "Development:"
	@echo "  make build           - Build both binaries to dist/"
	@echo "  make run             - Build and run TUI (with sudo)"
	@echo "  make daemon          - Build and run daemon in foreground"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make test            - Run tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  make deps            - Download and tidy dependencies"
	@echo "  make fmt             - Format code"
	@echo "  make lint            - Run linter"
	@echo ""
