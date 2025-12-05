package blocker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	hostsFile  = "/etc/hosts"
	beginMarker = "# BEGIN SELFCONTROL-TUI"
	endMarker   = "# END SELFCONTROL-TUI"
)

// Block adds blocking rules to /etc/hosts
func Block(urls []string) error {
	// First, ensure we're not already blocking
	if err := Unblock(); err != nil {
		return fmt.Errorf("failed to clear existing blocks: %w", err)
	}

	// Read current hosts file
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}

	// Expand wildcards to actual hostnames
	hosts := expandWildcards(urls)

	// Build blocking rules
	var blockingRules strings.Builder
	blockingRules.WriteString("\n")
	blockingRules.WriteString(beginMarker)
	blockingRules.WriteString("\n")

	for _, host := range hosts {
		// Block both with and without www
		blockingRules.WriteString(fmt.Sprintf("127.0.0.1 %s\n", host))

		// Also block IPv6
		blockingRules.WriteString(fmt.Sprintf("::1 %s\n", host))
	}

	blockingRules.WriteString(endMarker)
	blockingRules.WriteString("\n")

	// Append to hosts file
	newContent := string(content) + blockingRules.String()

	if err := os.WriteFile(hostsFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write hosts file (are you running with sudo?): %w", err)
	}

	return nil
}

// Unblock removes blocking rules from /etc/hosts
func Unblock() error {
	// Read current hosts file
	file, err := os.Open(hostsFile)
	if err != nil {
		return fmt.Errorf("failed to open hosts file: %w", err)
	}
	defer file.Close()

	var newContent strings.Builder
	scanner := bufio.NewScanner(file)
	inBlockSection := false

	for scanner.Scan() {
		line := scanner.Text()

		// Check if we're entering our block section
		if strings.TrimSpace(line) == beginMarker {
			inBlockSection = true
			continue
		}

		// Check if we're leaving our block section
		if strings.TrimSpace(line) == endMarker {
			inBlockSection = false
			continue
		}

		// Only write lines that are not in our block section
		if !inBlockSection {
			newContent.WriteString(line)
			newContent.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}

	// Write back the modified content
	if err := os.WriteFile(hostsFile, []byte(newContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write hosts file (are you running with sudo?): %w", err)
	}

	return nil
}

// expandWildcards converts wildcard patterns to actual hostnames
func expandWildcards(urls []string) []string {
	var result []string

	for _, url := range urls {
		// Remove protocol if present
		url = strings.TrimPrefix(url, "http://")
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimSpace(url)

		if strings.Contains(url, "*") {
			// Handle wildcard patterns
			// For patterns like *.linkedin.*, *.linkedin.com, linkedin.*
			// Generate common variations
			base := strings.ReplaceAll(url, "*", "")
			base = strings.Trim(base, ".")

			if base == "" {
				continue
			}

			// Generate common patterns
			variations := []string{
				base,
				"www." + base,
			}

			// If it's a pattern like *.linkedin.*, also add common subdomains
			if strings.HasPrefix(url, "*") {
				commonSubdomains := []string{"m", "mobile", "app", "api", "mail", "login", "account"}
				for _, sub := range commonSubdomains {
					variations = append(variations, sub+"."+base)
				}
			}

			// If the base looks like it might have a TLD pattern, add common TLDs
			if strings.HasSuffix(url, ".*") {
				baseDomain := strings.TrimSuffix(base, ".")
				commonTLDs := []string{"com", "net", "org", "io"}
				for _, tld := range commonTLDs {
					variations = append(variations, baseDomain+"."+tld)
					variations = append(variations, "www."+baseDomain+"."+tld)
				}
			}

			result = append(result, variations...)
		} else {
			// No wildcard, use as-is
			result = append(result, url)

			// Also add www variant if not present
			if !strings.HasPrefix(url, "www.") {
				result = append(result, "www."+url)
			}
		}
	}

	return result
}

// IsBlocked checks if our blocking rules are currently in place
func IsBlocked() (bool, error) {
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return false, fmt.Errorf("failed to read hosts file: %w", err)
	}

	return strings.Contains(string(content), beginMarker), nil
}
