package session

import (
	"bufio"
	"encoding/json"
	"os"
)

// transcriptEntry is the minimal shape read from each JSONL line.
type transcriptEntry struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Message   struct {
		ID    string `json:"id"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// SessionTokens reads a Claude Code transcript (JSONL) and returns the
// cumulative input+output token usage across all assistant turns. Duplicate
// turns (same message id + request id) are counted once; subagent (sidechain)
// turns are included. Malformed lines are skipped. A read error after the file
// opens returns the usage gathered so far with no error (best effort); only a
// failure to open the file is returned as an error.
func SessionTokens(path string) (TokenUsage, error) {
	f, err := os.Open(path)
	if err != nil {
		return TokenUsage{}, err
	}
	defer f.Close()

	var usage TokenUsage
	seen := map[string]bool{}
	r := bufio.NewReader(f)
	for {
		line, readErr := r.ReadBytes('\n')
		if len(line) > 0 {
			var e transcriptEntry
			if json.Unmarshal(line, &e) == nil && e.Type == "assistant" {
				key := e.Message.ID + ":" + e.RequestID
				if e.Message.ID == "" || !seen[key] {
					if e.Message.ID != "" {
						seen[key] = true
					}
					usage.Input += e.Message.Usage.InputTokens
					usage.Output += e.Message.Usage.OutputTokens
				}
			}
		}
		if readErr != nil {
			break // io.EOF (final partial line already handled) or a read error
		}
	}
	return usage, nil
}
