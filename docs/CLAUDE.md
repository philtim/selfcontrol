# SelfControl TUI

A production-grade Text User Interface (TUI) application for blocking distracting websites on Linux and macOS. Built with Go and Bubble Tea.

## Overview

SelfControl TUI helps you stay focused by blocking websites for a specified duration. Once a blocking session starts, the websites remain blocked until the timer expires - even if you close the application.

**Key Features:**
- 🚫 Block websites using `/etc/hosts` modification
- ⏱️ Predefined durations: 30sec (testing), 5min, 15min, 1h, 4h, 6h, 8h
- 🔄 Persistent sessions across app restarts
- 🎯 Simple keyboard-driven interface
- 🌐 Wildcard pattern support (`*.linkedin.*`)
- 🤖 Background daemon for automatic unblocking
- 💻 Cross-platform: Linux (Arch) and macOS

## Quick Start

### Installation

```bash
# Clone or navigate to the project directory
cd selfcontrol

# Build, install, and start daemon (one command does it all)
make update

# Or do it step-by-step:
make build              # Build both binaries
make install            # Install to /usr/local/bin
make install-daemon     # Install and start daemon
```

### Updating to Latest Version

After pulling the latest code changes:

```bash
# This command will:
# - Build the latest binaries
# - Stop the running daemon
# - Install updated binaries
# - Restart the daemon
make update
```

### First Use

```bash
# Run the TUI (requires sudo for /etc/hosts access)
sudo ./selfcontrol

# Or if installed system-wide
sudo selfcontrol
```

**Basic workflow:**
1. Press `a` to add a URL (e.g., `*.linkedin.*`)
2. Use `↑`/`↓` arrows to navigate between URLs
3. Press `s` to start a blocking session
4. Select duration and press Enter
5. Sites are now blocked - close the TUI with `q` if desired
6. Blocking continues until timer expires!

## Features

### 1. URL Management

**Add URLs:**
- Press `a` to add a new URL or pattern
- Examples:
  - `linkedin.com` - Block specific domain
  - `*.linkedin.*` - Block all LinkedIn domains
  - `*.reddit.com` - Block all Reddit subdomains
  - `www.example.com` - Block specific URL

**Navigate & Delete:**
- Use `↑`/`↓` arrow keys (or `j`/`k`) to select URLs
- Press `d` to instantly delete the selected URL
- No confirmation needed - each URL deleted individually

### 2. Blocking Sessions

**Start a Session:**
- Press `s` to start blocking
- Choose from predefined durations:
  - 30 seconds - Quick test (for debugging)
  - 5 minutes - Quick focus session
  - 15 minutes - Short break blocker
  - 1 hour - Standard work session
  - 4 hours - Deep work block
  - 6 hours - Extended focus period
  - 8 hours - Full work day

**Live Timer:**
- Real-time countdown displayed in Session Status
- Shows time remaining, elapsed time, and total duration
- Timer continues even if TUI is closed

### 3. Wildcard Patterns

Wildcard patterns automatically expand to block multiple variations:

**`*.linkedin.*`** expands to:
- `linkedin.com`, `www.linkedin.com`
- `linkedin.net`, `www.linkedin.net`
- `linkedin.org`, `linkedin.io`, etc.

**`*.example.com`** expands to:
- `example.com`, `www.example.com`
- `m.example.com`, `mobile.example.com`
- `api.example.com`, `app.example.com`, etc.

### 4. Persistence

**State stored in:** `/tmp/selfcontrol-tui-state.json`

**Why /tmp?**
- Accessible by both TUI (run with sudo) and daemon
- Works consistently across all users and system services
- Cleared on reboot (which would expire timers anyway)

**What persists:**
- All blocked URLs
- Active session with end timestamp
- Timer continues accurately across:
  - App restarts
  - System sleep/wake
  - Note: Cleared on reboot (timers would expire anyway)

## How It Works

### /etc/hosts Modification

The application blocks websites by adding entries to `/etc/hosts`:

```
# BEGIN SELFCONTROL-TUI
127.0.0.1 linkedin.com
127.0.0.1 www.linkedin.com
::1 linkedin.com
::1 www.linkedin.com
# END SELFCONTROL-TUI
```

**Safety guarantees:**
- ✅ Never modifies existing hosts file entries
- ✅ Only removes lines added by this app
- ✅ Uses clear markers for isolation
- ✅ Blocks both IPv4 (127.0.0.1) and IPv6 (::1)

### Background Daemon

**What it does:**
- Runs silently in the background
- Checks every 10 seconds for expired sessions
- Automatically removes blocks when timer ends
- Works even when TUI is closed

**Why you need it:**
Without the daemon, websites stay blocked forever (until you manually reopen the TUI). With the daemon, blocking automatically ends when the timer expires.

**Installation:**
```bash
make install-daemon
```

**Check status:**
```bash
# Linux
sudo systemctl status selfcontrol-daemon

# macOS
sudo launchctl list | grep selfcontrol
```

## Keyboard Controls

### Main View
- `a` - Add URL
- `d` - Delete selected URL
- `↑`/`↓` or `j`/`k` - Navigate through URLs
- `s` - Start blocking session (when URLs exist)
- `q` - Quit

### Add URL View
- Type URL or pattern
- `Enter` - Add URL and return to main view
- `Esc` - Cancel and return to main view

### Select Duration View
- `↑`/`↓` or `j`/`k` - Navigate durations
- `Enter` - Start session with selected duration
- `Esc` - Cancel and return to main view

## Building from Source

### Prerequisites
- Go 1.21 or later
- Root/sudo access (required for `/etc/hosts` modification)

### Build Commands

```bash
# Update everything (build, install, restart daemon)
make update

# Build both binaries
make build

# Install binaries to /usr/local/bin
make install

# Install background daemon service
make install-daemon

# Run TUI (with sudo)
make run

# Run daemon in foreground (for testing)
make daemon

# Clean build artifacts
make clean
```

### Manual Build

```bash
# Build TUI
go build -o selfcontrol ./cmd/selfcontrol

# Build daemon
go build -o selfcontrol-daemon ./cmd/selfcontrol-daemon

# Run TUI
sudo ./selfcontrol

# Run daemon
sudo ./selfcontrol-daemon
```

## Project Structure

```
selfcontrol/
├── cmd/
│   ├── selfcontrol/          # Main TUI application
│   │   └── main.go
│   └── selfcontrol-daemon/   # Background daemon
│       └── main.go
│
├── internal/
│   ├── ui/                   # Bubble Tea TUI
│   │   └── ui.go            # All UI rendering and logic
│   ├── state/               # State persistence
│   │   └── state.go         # JSON state management
│   ├── blocker/             # Hosts file manipulation
│   │   └── blocker.go       # Safe /etc/hosts editing
│   └── timer/               # Timer utilities
│       └── timer.go         # Duration formatting
│
├── scripts/                 # Installation scripts
│   ├── install-daemon.sh
│   ├── selfcontrol-daemon.service    # systemd (Linux)
│   └── com.selfcontrol.daemon.plist  # launchd (macOS)
│
├── go.mod                   # Go module definition
├── Makefile                 # Build automation
└── CLAUDE.md               # This file
```

## Technical Details

### TUI Framework: Bubble Tea

**Why Bubble Tea?**
- Clean Model-View-Update architecture (Elm-inspired)
- Robust cross-platform terminal handling
- Composable components (textinput, styling)
- Type-safe Go implementation
- Active development and community

### Color Scheme

The UI uses carefully chosen colors for visual hierarchy:

| Element | Color | Usage |
|---------|-------|-------|
| Title | `#A3BB7D` | App title |
| Blocked URLs Border | `#A3BB7D` | Section border & headers |
| Session Status Border | `#A69D88` | Section border |
| Session Status Text | `#C7AC75` | Active/inactive status |
| Command Keys | Cyan (#117) | Keyboard shortcuts |
| Selected Item | Cyan + Gray BG | Currently selected URL |
| Alternating Rows | Dark Gray | Every other row |

### State Management

**File:** `/tmp/selfcontrol-tui-state.json`

**Example:**
```json
{
  "urls": [
    "*.linkedin.*",
    "*.reddit.*",
    "twitter.com"
  ],
  "active_session": {
    "start_time": "2025-12-05T14:30:00Z",
    "end_time": "2025-12-05T18:30:00Z",
    "duration": "4 hours"
  }
}
```

**Key design:**
- Stores absolute `end_time` (not relative duration) for accurate timers across system sleep/clock changes
- Located in `/tmp` for universal access by both TUI and daemon
- Cleared on system reboot (which would expire active sessions anyway)

### Platform Compatibility

**Works on:**
- ✅ Linux (tested on Arch Linux)
- ✅ macOS (10.12+)

**Platform differences:**
| Feature | Linux | macOS |
|---------|-------|-------|
| Daemon system | systemd | launchd |
| Log location | journalctl | `/var/log/*.log` |
| Service file | `.service` | `.plist` |

No platform-specific code in the application itself - all handled by standard library.

## Troubleshooting

### Permission Denied

**Error:** `failed to write hosts file (are you running with sudo?)`

**Solution:** Run with sudo:
```bash
sudo selfcontrol
```

### Sites Not Blocking

1. **Check daemon is running:**
   ```bash
   # Linux
   sudo systemctl status selfcontrol-daemon

   # macOS
   sudo launchctl list | grep selfcontrol
   ```

2. **Verify /etc/hosts entries:**
   ```bash
   sudo cat /etc/hosts | grep SELFCONTROL
   ```

3. **Clear browser cache:**
   Some browsers cache DNS lookups. Try:
   - Hard refresh (Ctrl+Shift+R)
   - Clear browser cache
   - Restart browser

### Timer Not Auto-Unblocking

**Install the daemon:**
```bash
make install-daemon
```

Or manually unblock by reopening the TUI:
```bash
sudo selfcontrol  # Will check for expired sessions on startup
```

### Manual Unblock (Emergency)

If you need to manually remove blocking:

**Option 1: Edit hosts file**
```bash
sudo nano /etc/hosts
# Delete lines between SELFCONTROL-TUI markers
# Save and exit
```

**Option 2: Reset state**
```bash
rm /tmp/selfcontrol-tui-state.json
sudo nano /etc/hosts  # Clean up manually
```

## Common Workflows

### Daily Focus Session

```bash
# Morning routine
sudo selfcontrol

# Add your distraction sites
# Press 'a' for each:
# - *.linkedin.*
# - *.reddit.*
# - *.twitter.*

# Start 4-hour session
# Press 's', select "4 hours", press Enter

# Close TUI
# Press 'q'

# Work distraction-free!
# Sites automatically unblock after 4 hours
```

### Quick Break Block

```bash
sudo selfcontrol
# Press 's'
# Select "15 minutes"
# Press Enter and 'q'
```

### Emergency Edit

Need to temporarily access a blocked site?

```bash
# Stop daemon (Linux)
sudo systemctl stop selfcontrol-daemon

# Edit hosts file
sudo nano /etc/hosts
# Remove the site you need

# Access the site

# Restart daemon when done
sudo systemctl start selfcontrol-daemon
```

## Development

### Running Tests

```bash
go test ./...
```

### Adding Features

**Example: Add new duration**

Edit `internal/timer/timer.go`:
```go
func PredefinedDurations() []Duration {
    return []Duration{
        {Label: "5 minutes", Duration: 5 * time.Minute},
        {Label: "2 hours", Duration: 2 * time.Hour},  // Add this
        // ... existing durations
    }
}
```

**Example: Add new wildcard expansion**

Edit `internal/blocker/blocker.go`, function `expandWildcards()`.

### Code Style

- **Packages:** lowercase, single word
- **Exported:** PascalCase
- **Unexported:** camelCase
- **Errors:** Always wrap with context using `fmt.Errorf("context: %w", err)`
- **Comments:** All exported functions documented

## Performance

- **Binary size:** ~5 MB (TUI), ~3 MB (daemon)
- **Memory usage:** ~15 MB (TUI), ~10 MB (daemon)
- **CPU usage:** ~0-1% (TUI), ~0% (daemon)
- **Startup time:** <100ms (instant)
- **State file size:** <1 KB

## Security & Privacy

**What this app does:**
- ✅ Modifies `/etc/hosts` locally
- ✅ Stores URLs and timestamps locally
- ✅ No network access
- ✅ No telemetry or tracking
- ✅ Open source - audit the code

**What this app does NOT protect against:**
- ❌ Determined users editing `/etc/hosts` manually
- ❌ Using IP addresses instead of domain names
- ❌ VPN/proxy bypass
- ❌ Using a different browser/device

**This is a self-control tool**, not a security solution. The goal is to add friction to habitual distractions, not to create an impenetrable block.

## Uninstallation

### Remove binaries
```bash
sudo rm /usr/local/bin/selfcontrol
sudo rm /usr/local/bin/selfcontrol-daemon
```

### Remove daemon service

**Linux:**
```bash
sudo systemctl stop selfcontrol-daemon
sudo systemctl disable selfcontrol-daemon
sudo rm /etc/systemd/system/selfcontrol-daemon.service
sudo systemctl daemon-reload
```

**macOS:**
```bash
sudo launchctl unload /Library/LaunchDaemons/com.selfcontrol.daemon.plist
sudo rm /Library/LaunchDaemons/com.selfcontrol.daemon.plist
```

### Remove state
```bash
rm /tmp/selfcontrol-tui-state.json
```

### Clean /etc/hosts
```bash
sudo nano /etc/hosts
# Remove SELFCONTROL-TUI section if present
```

## Credits

**Built with:**
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

**Inspired by:**
- [SelfControl for macOS](https://selfcontrolapp.com/)

## License

MIT License - See LICENSE file for details.

## Version

**Current Version:** 1.0
**Platform:** Linux, macOS
**Go Version:** 1.21+

---

**Need help?** Check the troubleshooting section above or review the code in the `internal/` directory.

**Found a bug?** Feel free to modify the code - it's designed to be readable and maintainable.

**Want to contribute?** The codebase is clean and well-structured. Common additions:
- New durations (edit `internal/timer/timer.go`)
- Custom wildcard patterns (edit `internal/blocker/blocker.go`)
- UI improvements (edit `internal/ui/ui.go`)
