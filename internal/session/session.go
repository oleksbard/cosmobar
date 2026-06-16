// Package session parses the JSON Claude Code sends to a statusLine command.
package session

import (
	"encoding/json"
	"io"
)

type Session struct {
	SessionID      string        `json:"session_id"`
	TranscriptPath string        `json:"transcript_path"`
	Cwd            string        `json:"cwd"`
	Model          Model         `json:"model"`
	Workspace      Workspace     `json:"workspace"`
	OutputStyle    OutputStyle   `json:"output_style"`
	Cost           Cost          `json:"cost"`
	ContextWindow  ContextWindow `json:"context_window"`
	RateLimits     *RateLimits   `json:"rate_limits"`
	Effort         *Effort       `json:"effort"`
}

type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type Workspace struct {
	CurrentDir  string `json:"current_dir"`
	ProjectDir  string `json:"project_dir"`
	GitWorktree string `json:"git_worktree"`
	Repo        *Repo  `json:"repo"`
}

type Repo struct {
	Host  string `json:"host"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type OutputStyle struct {
	Name string `json:"name"`
}

type Cost struct {
	TotalCostUSD      float64 `json:"total_cost_usd"`
	TotalDurationMS   int64   `json:"total_duration_ms"`
	TotalLinesAdded   int     `json:"total_lines_added"`
	TotalLinesRemoved int     `json:"total_lines_removed"`
}

type ContextWindow struct {
	UsedPercentage    *float64 `json:"used_percentage"`
	ContextWindowSize int      `json:"context_window_size"`
}

type RateLimits struct {
	FiveHour *RateWindow `json:"five_hour"`
	SevenDay *RateWindow `json:"seven_day"`
}

type RateWindow struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

type Effort struct {
	Level string `json:"level"`
}

// TokenUsage is a cumulative token count for a session.
type TokenUsage struct {
	Input  int
	Output int
}

// Total returns input + output tokens.
func (t TokenUsage) Total() int { return t.Input + t.Output }

// Dir returns the best available working directory.
func (s *Session) Dir() string {
	if s.Workspace.CurrentDir != "" {
		return s.Workspace.CurrentDir
	}
	return s.Cwd
}

// Parse reads one JSON object from r. Unknown fields are ignored; absent
// fields keep zero values (or nil for pointer fields).
func Parse(r io.Reader) (*Session, error) {
	var s Session
	if err := json.NewDecoder(r).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}
