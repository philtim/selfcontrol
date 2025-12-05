package main

import (
	"fmt"
	"os"
	"time"

	"github.com/phil/selfcontrol/internal/blocker"
	"github.com/phil/selfcontrol/internal/state"
)

func main() {
	// This daemon runs in the background and checks for expired sessions
	// It should be run with root privileges

	if os.Geteuid() != 0 {
		fmt.Println("Error: This daemon must be run as root to modify /etc/hosts")
		os.Exit(1)
	}

	fmt.Println("SelfControl Daemon started - monitoring for expired sessions...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Load current state
		st, err := state.Load()
		if err != nil {
			fmt.Printf("Error loading state: %v\n", err)
			continue
		}

		// Check if session expired
		if st.ActiveSession != nil && !st.IsSessionActive() {
			fmt.Printf("Session expired at %s, unblocking...\n", st.ActiveSession.EndTime)

			// Unblock
			if err := blocker.Unblock(); err != nil {
				fmt.Printf("Error unblocking: %v\n", err)
				continue
			}

			// End session
			st.EndSession()
			if err := state.Save(st); err != nil {
				fmt.Printf("Error saving state: %v\n", err)
			}

			fmt.Println("Successfully unblocked websites")
		}
	}
}
