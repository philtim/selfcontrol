package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/phil/selfcontrol/internal/blocker"
	"github.com/phil/selfcontrol/internal/state"
	"github.com/phil/selfcontrol/internal/timer"
)

// View modes
type viewMode int

const (
	viewMain viewMode = iota
	viewAddURL
	viewDelete
	viewSelectDuration
)

// Model represents the UI state
type Model struct {
	state           *state.AppState
	mode            viewMode
	cursor          int
	textInput       textinput.Model
	deleteSelected  map[int]bool
	err             error
	quitting        bool
	lastTickTime    time.Time
	permissionError bool
}

// tickMsg is sent every second to update the timer
type tickMsg time.Time

// New creates a new UI model
func New() (*Model, error) {
	// Load state
	st, err := state.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	// Create text input for URL entry
	ti := textinput.New()
	ti.Placeholder = "example.com or *.example.*"
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	m := &Model{
		state:          st,
		mode:           viewMain,
		textInput:      ti,
		deleteSelected: make(map[int]bool),
		lastTickTime:   time.Now(),
	}

	// Check if session expired and clean up
	if st.ActiveSession != nil && !st.IsSessionActive() {
		// Session expired, unblock
		if err := blocker.Unblock(); err != nil {
			m.permissionError = true
			m.err = fmt.Errorf("session expired but failed to unblock: %w", err)
		}
		st.EndSession()
		state.Save(st)
	}

	return m, nil
}

// Init initializes the UI
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tickCmd(),
	)
}

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tickMsg:
		// Update timer
		m.lastTickTime = time.Time(msg)

		// Check if session expired
		if m.state.ActiveSession != nil && !m.state.IsSessionActive() {
			// Session expired, unblock
			if err := blocker.Unblock(); err != nil {
				m.permissionError = true
				m.err = fmt.Errorf("failed to unblock after timer expiry: %w", err)
			} else {
				m.state.EndSession()
				state.Save(m.state)
			}
		}

		return m, tickCmd()

	case tea.WindowSizeMsg:
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case viewMain:
		return m.handleMainKeys(msg)
	case viewAddURL:
		return m.handleAddURLKeys(msg)
	case viewDelete:
		return m.handleDeleteKeys(msg)
	case viewSelectDuration:
		return m.handleDurationKeys(msg)
	}
	return m, nil
}

// handleMainKeys processes keys in main view
func (m Model) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "a":
		// Enter add URL mode
		m.mode = viewAddURL
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, nil

	case "up", "k":
		// Navigate up in URL list
		if len(m.state.URLs) > 0 && m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		// Navigate down in URL list
		if len(m.state.URLs) > 0 && m.cursor < len(m.state.URLs)-1 {
			m.cursor++
		}
		return m, nil

	case "d":
		// Delete currently selected URL
		if len(m.state.URLs) > 0 && m.cursor < len(m.state.URLs) {
			m.state.RemoveURLs([]int{m.cursor})
			if err := state.Save(m.state); err != nil {
				m.err = err
			}
			// Adjust cursor if needed
			if m.cursor >= len(m.state.URLs) && len(m.state.URLs) > 0 {
				m.cursor = len(m.state.URLs) - 1
			}
			if len(m.state.URLs) == 0 {
				m.cursor = 0
			}
		}
		return m, nil

	case "s":
		// Start blocking session
		if len(m.state.URLs) > 0 && !m.state.IsSessionActive() {
			m.mode = viewSelectDuration
			m.cursor = 0
		}
		return m, nil
	}

	return m, nil
}

// handleAddURLKeys processes keys in add URL view
func (m Model) handleAddURLKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		url := strings.TrimSpace(m.textInput.Value())
		if url != "" {
			m.state.AddURL(url)
			if err := state.Save(m.state); err != nil {
				m.err = err
			}
			// Set cursor to the newly added URL (last item)
			m.cursor = len(m.state.URLs) - 1
		}
		m.mode = viewMain
		return m, nil

	case "esc":
		m.mode = viewMain
		return m, nil

	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

// handleDeleteKeys processes keys in delete view
func (m Model) handleDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = viewMain
		return m, nil

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.state.URLs)-1 {
			m.cursor++
		}
		return m, nil

	case " ":
		// Toggle selection
		m.deleteSelected[m.cursor] = !m.deleteSelected[m.cursor]
		return m, nil

	case "enter":
		// Delete selected URLs
		var toDelete []int
		for idx := range m.deleteSelected {
			if m.deleteSelected[idx] {
				toDelete = append(toDelete, idx)
			}
		}

		if len(toDelete) > 0 {
			m.state.RemoveURLs(toDelete)
			if err := state.Save(m.state); err != nil {
				m.err = err
			}
		}

		m.mode = viewMain
		m.cursor = 0
		return m, nil
	}

	return m, nil
}

// handleDurationKeys processes keys in duration selection view
func (m Model) handleDurationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	durations := timer.PredefinedDurations()

	switch msg.String() {
	case "esc":
		m.mode = viewMain
		return m, nil

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(durations)-1 {
			m.cursor++
		}
		return m, nil

	case "enter":
		// Start blocking session
		selected := durations[m.cursor]
		m.state.StartSession(selected.Duration, selected.Label)

		// Apply blocking
		if err := blocker.Block(m.state.URLs); err != nil {
			m.permissionError = true
			m.err = fmt.Errorf("failed to apply blocking: %w", err)
			m.state.EndSession()
		}

		if err := state.Save(m.state); err != nil {
			m.err = err
		}

		m.mode = viewMain
		m.cursor = 0
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var s strings.Builder

	// Title bar
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#A3BB7D"))

	s.WriteString(titleStyle.Render("SelfControl"))
	s.WriteString("\n\n")

	// Show permission error if any
	if m.permissionError {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

		s.WriteString(errorStyle.Render("ERROR: Insufficient permissions!"))
		s.WriteString("\n")
		s.WriteString("Please run with sudo: sudo selfcontrol\n\n")

		if m.err != nil {
			s.WriteString(fmt.Sprintf("Details: %v\n\n", m.err))
		}

		s.WriteString("Press q to quit\n")
		return s.String()
	}

	// Show errors
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		s.WriteString("\n\n")
	}

	switch m.mode {
	case viewMain:
		s.WriteString(m.renderMainView())
	case viewAddURL:
		s.WriteString(m.renderAddURLView())
	case viewDelete:
		s.WriteString(m.renderDeleteView())
	case viewSelectDuration:
		s.WriteString(m.renderDurationView())
	}

	return s.String()
}

// renderMainView renders the main view
func (m Model) renderMainView() string {
	var s strings.Builder

	// Define colors - custom hex colors for each section
	urlsBorderColor := lipgloss.Color("#A3BB7D")    // Green for Blocked URLs
	sessionBorderColor := lipgloss.Color("#A69D88") // Tan for Session Status
	inactiveColor := lipgloss.Color("240")          // Gray
	highlightBg := lipgloss.Color("237")            // Dark gray for selection

	// Calculate width for proper alignment
	const tableWidth = 120
	const urlColumnWidth = tableWidth - 4 // Account for borders and padding

	// Blocked URLs Section
	urlsBorderStyle := lipgloss.NewStyle().Foreground(urlsBorderColor)
	urlsHeaderStyle := lipgloss.NewStyle().Foreground(urlsBorderColor).Bold(true)

	s.WriteString(urlsBorderStyle.Render("┌ Blocked URLs "))
	s.WriteString(urlsBorderStyle.Render(strings.Repeat("─", tableWidth-16)))
	s.WriteString(urlsBorderStyle.Render("┐"))
	s.WriteString("\n")

	// Table header
	s.WriteString(urlsBorderStyle.Render("│ "))
	s.WriteString(urlsHeaderStyle.Render(fmt.Sprintf("%-*s", urlColumnWidth, "URL / Pattern")))
	s.WriteString(urlsBorderStyle.Render(" │"))
	s.WriteString("\n")

	// Separator
	s.WriteString(urlsBorderStyle.Render("├"))
	s.WriteString(urlsBorderStyle.Render(strings.Repeat("─", tableWidth-2)))
	s.WriteString(urlsBorderStyle.Render("┤"))
	s.WriteString("\n")

	// URLs or empty message
	if len(m.state.URLs) == 0 {
		emptyMsg := lipgloss.NewStyle().Foreground(inactiveColor).Render("(no URLs added yet - press 'a' to add)")
		s.WriteString(urlsBorderStyle.Render("│ "))
		s.WriteString(fmt.Sprintf("%-*s", urlColumnWidth, emptyMsg))
		s.WriteString(urlsBorderStyle.Render(" │"))
		s.WriteString("\n")
	} else {
		for i, url := range m.state.URLs {
			// Truncate URL if too long
			displayURL := url
			if len(displayURL) > urlColumnWidth-6 {
				displayURL = displayURL[:urlColumnWidth-9] + "..."
			}

			// Add cursor indicator for selected item
			cursor := "  "
			if i == m.cursor {
				cursor = "▶ "
			}
			displayURL = cursor + displayURL

			lineStyle := lipgloss.NewStyle()
			if i == m.cursor {
				// Highlight selected item
				lineStyle = lineStyle.Background(highlightBg).Foreground(lipgloss.Color("117"))
			} else if i%2 == 0 {
				lineStyle = lineStyle.Background(lipgloss.Color("235"))
			}

			s.WriteString(urlsBorderStyle.Render("│ "))
			s.WriteString(lineStyle.Render(fmt.Sprintf("%-*s", urlColumnWidth, displayURL)))
			s.WriteString(urlsBorderStyle.Render(" │"))
			s.WriteString("\n")
		}
	}

	// Bottom border
	s.WriteString(urlsBorderStyle.Render("└"))
	s.WriteString(urlsBorderStyle.Render(strings.Repeat("─", tableWidth-2)))
	s.WriteString(urlsBorderStyle.Render("┘"))
	s.WriteString("\n\n")

	// Session Status Section
	sessionBorderStyle := lipgloss.NewStyle().Foreground(sessionBorderColor)

	s.WriteString(sessionBorderStyle.Render("┌ Session Status "))
	s.WriteString(sessionBorderStyle.Render(strings.Repeat("─", tableWidth-18)))
	s.WriteString(sessionBorderStyle.Render("┐"))
	s.WriteString("\n")

	if m.state.IsSessionActive() {
		remaining := m.state.TimeRemaining()
		elapsed := time.Since(m.state.ActiveSession.StartTime)

		// Status message
		statusMsg := fmt.Sprintf("🔒 ACTIVE  │  Time Remaining: %s  │  Elapsed: %s  │  Duration: %s",
			timer.FormatDuration(remaining),
			timer.FormatDuration(elapsed),
			m.state.ActiveSession.Duration)

		s.WriteString(sessionBorderStyle.Render("│ "))
		activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#C7AC75")).Bold(true)
		s.WriteString(activeStyle.Render(fmt.Sprintf("%-*s", urlColumnWidth, statusMsg)))
		s.WriteString(sessionBorderStyle.Render(" │"))
		s.WriteString("\n")
	} else {
		s.WriteString(sessionBorderStyle.Render("│ "))
		inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#C7AC75"))
		s.WriteString(inactiveStyle.Render(fmt.Sprintf("%-*s", urlColumnWidth, "No active session - press 's' to start blocking")))
		s.WriteString(sessionBorderStyle.Render(" │"))
		s.WriteString("\n")
	}

	s.WriteString(sessionBorderStyle.Render("└"))
	s.WriteString(sessionBorderStyle.Render(strings.Repeat("─", tableWidth-2)))
	s.WriteString(sessionBorderStyle.Render("┘"))
	s.WriteString("\n\n")

	// Command bar at the bottom
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	cmdKeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)

	commands := []string{}
	commands = append(commands, cmdKeyStyle.Render("a")+" "+cmdStyle.Render("Add"))

	if len(m.state.URLs) > 0 {
		commands = append(commands, cmdKeyStyle.Render("d")+" "+cmdStyle.Render("Delete"))
		commands = append(commands, cmdKeyStyle.Render("↑/↓")+" "+cmdStyle.Render("Navigate"))
	}

	if len(m.state.URLs) > 0 && !m.state.IsSessionActive() {
		commands = append(commands, cmdKeyStyle.Render("s")+" "+cmdStyle.Render("Start"))
	}

	commands = append(commands, cmdKeyStyle.Render("q")+" "+cmdStyle.Render("Quit"))

	s.WriteString(strings.Join(commands, " │ "))
	s.WriteString("\n")

	return s.String()
}

// renderAddURLView renders the add URL view
func (m Model) renderAddURLView() string {
	var s strings.Builder

	borderColor := lipgloss.Color("142")
	headerColor := lipgloss.Color("184")

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	headerStyle := lipgloss.NewStyle().Foreground(headerColor).Bold(true)

	const tableWidth = 120
	const contentWidth = tableWidth - 4

	// Title
	s.WriteString(borderStyle.Render("┌ Add URL / Pattern "))
	s.WriteString(borderStyle.Render(strings.Repeat("─", tableWidth-21)))
	s.WriteString(borderStyle.Render("┐"))
	s.WriteString("\n")

	// Input field
	s.WriteString(borderStyle.Render("│"))
	s.WriteString("\n")
	s.WriteString(borderStyle.Render("│ "))
	s.WriteString(headerStyle.Render("URL or pattern: "))
	s.WriteString(m.textInput.View())
	padding := contentWidth - 16 - len(m.textInput.Value())
	if padding < 0 {
		padding = 0
	}
	s.WriteString(strings.Repeat(" ", padding))
	s.WriteString(borderStyle.Render(" │"))
	s.WriteString("\n")
	s.WriteString(borderStyle.Render("│"))
	s.WriteString(strings.Repeat(" ", contentWidth))
	s.WriteString(borderStyle.Render(" │"))
	s.WriteString("\n")

	// Examples section
	s.WriteString(borderStyle.Render("│ "))
	exampleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	s.WriteString(exampleStyle.Render("Examples:"))
	s.WriteString(strings.Repeat(" ", contentWidth-9))
	s.WriteString(borderStyle.Render(" │"))
	s.WriteString("\n")

	examples := []string{
		"  linkedin.com          - Block specific domain",
		"  *.linkedin.*          - Block all LinkedIn domains",
		"  *.reddit.com          - Block all Reddit subdomains",
		"  www.example.com       - Block specific URL",
	}

	for _, example := range examples {
		s.WriteString(borderStyle.Render("│ "))
		s.WriteString(exampleStyle.Render(fmt.Sprintf("%-*s", contentWidth, example)))
		s.WriteString(borderStyle.Render(" │"))
		s.WriteString("\n")
	}

	// Bottom border
	s.WriteString(borderStyle.Render("└"))
	s.WriteString(borderStyle.Render(strings.Repeat("─", tableWidth-2)))
	s.WriteString(borderStyle.Render("┘"))
	s.WriteString("\n\n")

	// Command bar
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	cmdKeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)

	s.WriteString(cmdKeyStyle.Render("Enter") + " " + cmdStyle.Render("Add"))
	s.WriteString(" │ ")
	s.WriteString(cmdKeyStyle.Render("Esc") + " " + cmdStyle.Render("Cancel"))
	s.WriteString("\n")

	return s.String()
}

// renderDeleteView renders the delete mode view
func (m Model) renderDeleteView() string {
	var s strings.Builder

	borderColor := lipgloss.Color("142")
	headerColor := lipgloss.Color("184")
	highlightBg := lipgloss.Color("237")
	selectedBg := lipgloss.Color("235")

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	headerStyle := lipgloss.NewStyle().Foreground(headerColor).Bold(true)

	const tableWidth = 120

	// Title
	s.WriteString(borderStyle.Render("┌ Delete URLs "))
	s.WriteString(borderStyle.Render(strings.Repeat("─", tableWidth-15)))
	s.WriteString(borderStyle.Render("┐"))
	s.WriteString("\n")

	// Table header
	s.WriteString(borderStyle.Render("│ "))
	s.WriteString(headerStyle.Render(fmt.Sprintf("%-5s", "")))
	s.WriteString(borderStyle.Render("│ "))
	s.WriteString(headerStyle.Render(fmt.Sprintf("%-5s", "Sel")))
	s.WriteString(borderStyle.Render("│ "))
	s.WriteString(headerStyle.Render(fmt.Sprintf("%-103s", "URL / Pattern")))
	s.WriteString(borderStyle.Render(" │"))
	s.WriteString("\n")

	// Separator
	s.WriteString(borderStyle.Render("├─────┼─────┼"))
	s.WriteString(borderStyle.Render(strings.Repeat("─", 103)))
	s.WriteString(borderStyle.Render("─┤"))
	s.WriteString("\n")

	// URLs
	for i, url := range m.state.URLs {
		displayURL := url
		if len(displayURL) > 100 {
			displayURL = displayURL[:97] + "..."
		}

		cursor := "  "
		checkbox := "[ ]"
		if m.deleteSelected[i] {
			checkbox = "[✓]"
		}

		lineStyle := lipgloss.NewStyle()
		if i == m.cursor {
			cursor = "▶ "
			lineStyle = lineStyle.Background(highlightBg).Foreground(lipgloss.Color("117"))
		} else if i%2 == 0 {
			lineStyle = lineStyle.Background(selectedBg)
		}

		s.WriteString(borderStyle.Render("│ "))
		s.WriteString(lineStyle.Render(fmt.Sprintf("%-5s", cursor)))
		s.WriteString(borderStyle.Render("│ "))
		s.WriteString(lineStyle.Render(fmt.Sprintf("%-5s", checkbox)))
		s.WriteString(borderStyle.Render("│ "))
		s.WriteString(lineStyle.Render(fmt.Sprintf("%-103s", displayURL)))
		s.WriteString(borderStyle.Render(" │"))
		s.WriteString("\n")
	}

	// Bottom border
	s.WriteString(borderStyle.Render("└"))
	s.WriteString(borderStyle.Render(strings.Repeat("─", tableWidth-2)))
	s.WriteString(borderStyle.Render("┘"))
	s.WriteString("\n\n")

	// Command bar
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	cmdKeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)

	s.WriteString(cmdKeyStyle.Render("Space") + " " + cmdStyle.Render("Select"))
	s.WriteString(" │ ")
	s.WriteString(cmdKeyStyle.Render("Enter") + " " + cmdStyle.Render("Delete"))
	s.WriteString(" │ ")
	s.WriteString(cmdKeyStyle.Render("j,↓") + " " + cmdStyle.Render("Down"))
	s.WriteString(" │ ")
	s.WriteString(cmdKeyStyle.Render("k,↑") + " " + cmdStyle.Render("Up"))
	s.WriteString(" │ ")
	s.WriteString(cmdKeyStyle.Render("Esc") + " " + cmdStyle.Render("Cancel"))
	s.WriteString("\n")

	return s.String()
}

// renderDurationView renders the duration selection view
func (m Model) renderDurationView() string {
	var s strings.Builder

	borderColor := lipgloss.Color("142")
	headerColor := lipgloss.Color("184")
	highlightBg := lipgloss.Color("237")
	selectedBg := lipgloss.Color("235")

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	headerStyle := lipgloss.NewStyle().Foreground(headerColor).Bold(true)

	const tableWidth = 120

	// Title
	s.WriteString(borderStyle.Render("┌ Select Blocking Duration "))
	s.WriteString(borderStyle.Render(strings.Repeat("─", tableWidth-28)))
	s.WriteString(borderStyle.Render("┐"))
	s.WriteString("\n")

	// Table header
	s.WriteString(borderStyle.Render("│ "))
	s.WriteString(headerStyle.Render(fmt.Sprintf("%-5s", "")))
	s.WriteString(borderStyle.Render("│ "))
	s.WriteString(headerStyle.Render(fmt.Sprintf("%-30s", "Duration")))
	s.WriteString(borderStyle.Render("│ "))
	s.WriteString(headerStyle.Render(fmt.Sprintf("%-75s", "Description")))
	s.WriteString(borderStyle.Render(" │"))
	s.WriteString("\n")

	// Separator
	s.WriteString(borderStyle.Render("├─────┼"))
	s.WriteString(borderStyle.Render(strings.Repeat("─", 30)))
	s.WriteString(borderStyle.Render("┼"))
	s.WriteString(borderStyle.Render(strings.Repeat("─", 75)))
	s.WriteString(borderStyle.Render("─┤"))
	s.WriteString("\n")

	// Durations
	durations := timer.PredefinedDurations()
	descriptions := map[string]string{
		"5 minutes":  "Quick focus session",
		"15 minutes": "Short break blocker",
		"1 hour":     "Standard work session",
		"4 hours":    "Deep work block",
		"6 hours":    "Extended focus period",
		"8 hours":    "Full work day",
	}

	for i, dur := range durations {
		cursor := "  "
		lineStyle := lipgloss.NewStyle()

		if i == m.cursor {
			cursor = "▶ "
			lineStyle = lineStyle.Background(highlightBg).Foreground(lipgloss.Color("117"))
		} else if i%2 == 0 {
			lineStyle = lineStyle.Background(selectedBg)
		}

		description := descriptions[dur.Label]
		if description == "" {
			description = "Custom duration"
		}

		s.WriteString(borderStyle.Render("│ "))
		s.WriteString(lineStyle.Render(fmt.Sprintf("%-5s", cursor)))
		s.WriteString(borderStyle.Render("│ "))
		s.WriteString(lineStyle.Render(fmt.Sprintf("%-30s", dur.Label)))
		s.WriteString(borderStyle.Render("│ "))
		s.WriteString(lineStyle.Render(fmt.Sprintf("%-75s", description)))
		s.WriteString(borderStyle.Render(" │"))
		s.WriteString("\n")
	}

	// Bottom border
	s.WriteString(borderStyle.Render("└"))
	s.WriteString(borderStyle.Render(strings.Repeat("─", tableWidth-2)))
	s.WriteString(borderStyle.Render("┘"))
	s.WriteString("\n\n")

	// Command bar
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	cmdKeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)

	s.WriteString(cmdKeyStyle.Render("Enter") + " " + cmdStyle.Render("Start"))
	s.WriteString(" │ ")
	s.WriteString(cmdKeyStyle.Render("j,↓") + " " + cmdStyle.Render("Down"))
	s.WriteString(" │ ")
	s.WriteString(cmdKeyStyle.Render("k,↑") + " " + cmdStyle.Render("Up"))
	s.WriteString(" │ ")
	s.WriteString(cmdKeyStyle.Render("Esc") + " " + cmdStyle.Render("Cancel"))
	s.WriteString("\n")

	return s.String()
}
