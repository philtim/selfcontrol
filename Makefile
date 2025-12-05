.PHONY: build install clean test run daemon install-daemon uninstall-daemon

# Build both binaries
build:
	@echo "Building selfcontrol..."
	go build -o selfcontrol ./cmd/selfcontrol
	@echo "Building selfcontrol-daemon..."
	go build -o selfcontrol-daemon ./cmd/selfcontrol-daemon
	@echo "✓ Build complete"

# Install binaries to /usr/local/bin
install: build
	@echo "Installing binaries..."
	sudo cp selfcontrol /usr/local/bin/
	sudo cp selfcontrol-daemon /usr/local/bin/
	sudo chmod +x /usr/local/bin/selfcontrol
	sudo chmod +x /usr/local/bin/selfcontrol-daemon
	@echo "✓ Binaries installed to /usr/local/bin"

# Install daemon service
install-daemon: install
	@echo "Installing daemon service..."
	sudo ./scripts/install-daemon.sh
	@echo "✓ Daemon installed"

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
	sudo ./selfcontrol

# Run the daemon in foreground (for testing)
daemon: build
	sudo ./selfcontrol-daemon

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f selfcontrol selfcontrol-daemon
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
	@echo "  make build           - Build both binaries"
	@echo "  make install         - Install binaries to /usr/local/bin"
	@echo "  make install-daemon  - Install and start background daemon"
	@echo "  make uninstall-daemon- Stop and remove background daemon"
	@echo "  make run             - Build and run TUI (with sudo)"
	@echo "  make daemon          - Build and run daemon in foreground"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make test            - Run tests"
	@echo "  make deps            - Download and tidy dependencies"
	@echo "  make fmt             - Format code"
	@echo "  make lint            - Run linter"
	@echo ""
