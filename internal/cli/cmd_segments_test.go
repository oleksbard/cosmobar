package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSegmentsListIncludesKnownSegments(t *testing.T) {
	out := segmentsList()
	for _, n := range []string{"dir", "git", "context", "rate_limits", "effort"} {
		if !strings.Contains(out, n) {
			t.Errorf("segments list missing %q:\n%s", n, out)
		}
	}
}

func TestSegmentsJSONIsValid(t *testing.T) {
	// segmentsJSON helper round-trips the catalog through JSON.
	var data []map[string]any
	if err := json.Unmarshal([]byte(segmentsJSON()), &data); err != nil {
		t.Fatalf("segments --json not valid JSON: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected at least one segment")
	}
	if _, ok := data[0]["name"]; !ok {
		t.Error("each segment should have a name field")
	}
}
