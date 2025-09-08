// Package cli/types provides shared types for CLI commands
package cli

import "time"

// CommitInfo represents detailed commit information
type CommitInfo struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
}
