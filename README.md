# 🚫 SelfControl TUI

A production-grade Text User Interface (TUI) application for blocking distracting websites on Linux and macOS. Built with Go and Bubble Tea.

## Features

- ✅ **Cross-platform**: Runs on Linux (Arch Linux) and macOS
- ✅ **Safe /etc/hosts modification**: Uses markers to isolate changes
- ✅ **Wildcard support**: Block patterns like `*.linkedin.*`
- ✅ **Multiple durations**: 5min, 15min, 1h, 4h, 6h, 8h
- ✅ **Persistent state**: Sessions survive app restarts
- ✅ **Live countdown**: Real-time timer display
- ✅ **Multi-select delete**: Remove multiple URLs at once
- ✅ **Automatic unblocking**: Blocks removed when timer expires

## How It Works

### /etc/hosts Modification

The application modifies `/etc/hosts` to block websites by redirecting them to `127.0.0.1` (localhost). All modifications are isolated between markers:

```
# BEGIN SELFCONTROL-TUI
127.0.0.1 linkedin.com
127.0.0.1 www.linkedin.com
::1 linkedin.com
::1 www.linkedin.com
# END SELFCONTROL-TUI
```

**Safety guarantees:**
- Never modifies existing lines in `/etc/hosts`
- Only removes lines added by this application
- Blocks both IPv4 and IPv6

### Wildcard Matching

When you add a wildcard pattern like `*.linkedin.*`, the application expands it to common variations:

- `*.example.*` → `example.com`, `www.example.com`, `example.net`, `www.example.net`, etc.
- `*.example.com` → `example.com`, `www.example.com`, `m.example.com`, `mobile.example.com`, etc.

### Persistence & Timer Recovery

State is stored in `$HOME/.config/selfcontrol-tui/state.json` containing:
- List of blocked URLs
- Active session (if any) with end timestamp

When you reopen the TUI:
- Active sessions are restored with accurate countdown
- Expired sessions are automatically unblocked
- All URLs persist across restarts

### Background Daemon

The daemon (`selfcontrol-daemon`) runs in the background to automatically unblock websites when timers expire, even if the TUI is closed.

## Installation

### Prerequisites

- Go 1.21 or later
- Root/sudo access (required for `/etc/hosts` modification)

### Building

```bash
# Clone the repository
git clone https://github.com/phil/selfcontrol
cd selfcontrol

# Download dependencies
go mod download

# Build the main application
go build -o selfcontrol cmd/selfcontrol/main.go

# Build the daemon (optional but recommended)
go build -o selfcontrol-daemon cmd/selfcontrol-daemon/main.go

# Install to system (optional)
sudo cp selfcontrol /usr/local/bin/
sudo cp selfcontrol-daemon /usr/local/bin/
```

### Quick Build Script

```bash
# Build both binaries
go build -o selfcontrol ./cmd/selfcontrol
go build -o selfcontrol-daemon ./cmd/selfcontrol-daemon
```

## Usage

### Running the TUI

```bash
# Run with sudo (required for /etc/hosts modification)
sudo ./selfcontrol

# Or if installed system-wide
sudo selfcontrol
```

### Keyboard Controls

**Main View:**
- `a` - Add URL or pattern
- `d` - Delete URLs (multi-select mode)
- `s` - Start blocking session
- `q` - Quit

**Add URL View:**
- Type URL or pattern
- `Enter` - Add URL
- `Esc` - Cancel

**Delete Mode:**
- `↑`/`↓` or `j`/`k` - Navigate
- `Space` - Toggle selection
- `Enter` - Delete selected URLs
- `Esc` - Cancel

**Select Duration:**
- `↑`/`↓` or `j`/`k` - Navigate
- `Enter` - Start session with selected duration
- `Esc` - Cancel

### Setting Up the Background Daemon

The daemon ensures websites are automatically unblocked when timers expire, even if the TUI is closed.

#### Linux (systemd)

Create `/etc/systemd/system/selfcontrol-daemon.service`:

```ini
[Unit]
Description=SelfControl Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/selfcontrol-daemon
Restart=always
User=root

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable selfcontrol-daemon
sudo systemctl start selfcontrol-daemon
sudo systemctl status selfcontrol-daemon
```

#### macOS (launchd)

Create `/Library/LaunchDaemons/com.selfcontrol.daemon.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.selfcontrol.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/selfcontrol-daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>/var/log/selfcontrol-daemon.log</string>
    <key>StandardOutPath</key>
    <string>/var/log/selfcontrol-daemon.log</string>
</dict>
</plist>
```

Load the daemon:

```bash
sudo cp selfcontrol-daemon /usr/local/bin/
sudo launchctl load /Library/LaunchDaemons/com.selfcontrol.daemon.plist
sudo launchctl start com.selfcontrol.daemon
```

## Technical Details

### TUI Library

**Bubble Tea** - Chosen for its:
- Robust terminal handling
- Clean, composable architecture (Elm-inspired)
- Active development and community
- Excellent documentation
- Built-in utilities (textinput, etc.)

### Project Structure

```
selfcontrol/
├── cmd/
│   ├── selfcontrol/          # Main TUI application
│   │   └── main.go
│   └── selfcontrol-daemon/   # Background daemon
│       └── main.go
├── internal/
│   ├── blocker/              # /etc/hosts manipulation
│   │   └── blocker.go
│   ├── state/                # Persistence logic
│   │   └── state.go
│   ├── timer/                # Timer utilities
│   │   └── timer.go
│   └── ui/                   # Bubble Tea UI
│       └── ui.go
├── go.mod
└── README.md
```

### State Storage

Location: `$HOME/.config/selfcontrol-tui/state.json`

Example:
```json
{
  "urls": [
    "linkedin.com",
    "*.reddit.*",
    "twitter.com"
  ],
  "active_session": {
    "end_time": "2025-12-05T15:30:00Z",
    "duration": "1 hour",
    "start_time": "2025-12-05T14:30:00Z"
  }
}
```

### Wildcard Implementation

The blocker expands wildcards intelligently:

1. **`*.domain.*`** - Expands to multiple TLDs:
   - `domain.com`, `www.domain.com`
   - `domain.net`, `www.domain.net`
   - `domain.org`, `domain.io`, etc.

2. **`*.domain.com`** - Expands to common subdomains:
   - `domain.com`, `www.domain.com`
   - `m.domain.com`, `mobile.domain.com`
   - `api.domain.com`, `app.domain.com`, etc.

3. **`domain.com`** - Direct match:
   - `domain.com`, `www.domain.com`

### Timer Mechanism

1. **During active session**:
   - TUI updates countdown every second
   - State persisted with `end_time` timestamp

2. **When TUI is closed**:
   - Blocking remains active in `/etc/hosts`
   - Daemon checks every 10 seconds for expired sessions

3. **When TUI reopens**:
   - Loads state from disk
   - Calculates remaining time from `end_time`
   - Auto-unblocks if expired

4. **Automatic unblocking**:
   - Daemon removes markers from `/etc/hosts`
   - State cleared from disk
   - No manual intervention needed

## Examples

### Block LinkedIn

```
1. Run: sudo selfcontrol
2. Press 'a'
3. Type: *.linkedin.*
4. Press Enter
5. Press 's'
6. Select duration (e.g., "1 hour")
7. Press Enter
```

Now `linkedin.com`, `www.linkedin.com`, and all subdomains are blocked for 1 hour.

### Block Multiple Sites

```
1. Add: *.linkedin.*
2. Add: *.reddit.*
3. Add: twitter.com
4. Start session
```

### Check Active Session

If you started a session and closed the TUI, simply reopen it:

```bash
sudo selfcontrol
```

You'll see the remaining time and active blocking status.

## Troubleshooting

### Permission Denied

**Error**: `failed to write hosts file (are you running with sudo?)`

**Solution**: Run with sudo:
```bash
sudo selfcontrol
```

### Daemon Not Unblocking

**Check daemon status:**

Linux:
```bash
sudo systemctl status selfcontrol-daemon
journalctl -u selfcontrol-daemon -f
```

macOS:
```bash
sudo launchctl list | grep selfcontrol
tail -f /var/log/selfcontrol-daemon.log
```

### Manual Unblock

If you need to manually remove blocking:

```bash
# Open hosts file
sudo nano /etc/hosts

# Remove lines between:
# BEGIN SELFCONTROL-TUI
# ... (remove these lines)
# END SELFCONTROL-TUI

# Save and exit
```

Or use the blocker directly:
```bash
# In Go code
go run -exec sudo ./test-unblock.go
```

### State Reset

To completely reset the application:

```bash
# Remove state file
rm ~/.config/selfcontrol-tui/state.json

# Manually clean hosts file (if needed)
sudo nano /etc/hosts  # Remove SelfControl section
```

## Development

### Running Tests

```bash
go test ./...
```

### Code Structure

- **`internal/state`**: JSON persistence, session management
- **`internal/blocker`**: Safe `/etc/hosts` manipulation, wildcard expansion
- **`internal/timer`**: Duration formatting, predefined durations
- **`internal/ui`**: Bubble Tea models, views, and update logic

### Adding New Features

1. **New duration**: Edit `timer.PredefinedDurations()`
2. **New wildcard pattern**: Edit `blocker.expandWildcards()`
3. **New view**: Add mode to `ui.viewMode` and implement handlers

## Security Considerations

- **Root required**: Modifying `/etc/hosts` requires root privileges
- **Safe isolation**: Only modifies lines between markers
- **No network access**: Purely local file manipulation
- **Transparent operation**: All changes visible in `/etc/hosts`

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Submit a pull request

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

Inspired by the original [SelfControl](https://selfcontrolapp.com/) for macOS.
