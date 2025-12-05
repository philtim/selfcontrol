package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/phil/selfcontrol/internal/ui"
)

func main() {
	// Check if running with appropriate permissions
	// Note: On most systems, modifying /etc/hosts requires root
	if os.Geteuid() != 0 {
		fmt.Println("⚠️  Warning: Not running as root.")
		fmt.Println("You may need to run with sudo to modify /etc/hosts:")
		fmt.Println("  sudo selfcontrol")
		fmt.Println()
		fmt.Println("Continuing anyway... (errors will be shown if permissions are insufficient)")
		fmt.Println()
	}

	// Create UI model
	m, err := ui.New()
	if err != nil {
		fmt.Printf("Error initializing: %v\n", err)
		os.Exit(1)
	}

	// Create Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
