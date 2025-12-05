package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// AppState represents the persistent application state
type AppState struct {
	URLs          []string   `json:"urls"`
	ActiveSession *Session   `json:"active_session,omitempty"`
}

// Session represents an active blocking session
type Session struct {
	EndTime   time.Time `json:"end_time"`
	Duration  string    `json:"duration"`
	StartTime time.Time `json:"start_time"`
}

var statePath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("cannot determine home directory: " + err.Error())
	}
	statePath = filepath.Join(home, ".config", "selfcontrol-tui", "state.json")
}

// Load reads the state from disk
func Load() (*AppState, error) {
	// Ensure config directory exists
	dir := filepath.Dir(statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Check if state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		// Return empty state
		return &AppState{
			URLs: []string{},
		}, nil
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}

	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// Save writes the state to disk
func Save(state *AppState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

// AddURL adds a URL to the state
func (s *AppState) AddURL(url string) {
	// Check for duplicates
	for _, u := range s.URLs {
		if u == url {
			return
		}
	}
	s.URLs = append(s.URLs, url)
}

// RemoveURLs removes URLs at the specified indices
func (s *AppState) RemoveURLs(indices []int) {
	// Create a map of indices to remove
	toRemove := make(map[int]bool)
	for _, idx := range indices {
		toRemove[idx] = true
	}

	// Build new slice without removed URLs
	newURLs := []string{}
	for i, url := range s.URLs {
		if !toRemove[i] {
			newURLs = append(newURLs, url)
		}
	}
	s.URLs = newURLs
}

// StartSession starts a new blocking session
func (s *AppState) StartSession(duration time.Duration, durationStr string) {
	s.ActiveSession = &Session{
		StartTime: time.Now(),
		EndTime:   time.Now().Add(duration),
		Duration:  durationStr,
	}
}

// EndSession ends the current blocking session
func (s *AppState) EndSession() {
	s.ActiveSession = nil
}

// IsSessionActive returns true if there is an active session
func (s *AppState) IsSessionActive() bool {
	if s.ActiveSession == nil {
		return false
	}
	return time.Now().Before(s.ActiveSession.EndTime)
}

// TimeRemaining returns the time remaining in the current session
func (s *AppState) TimeRemaining() time.Duration {
	if !s.IsSessionActive() {
		return 0
	}
	return time.Until(s.ActiveSession.EndTime)
}
