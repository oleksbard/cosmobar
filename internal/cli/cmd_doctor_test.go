package cli

import (
	"strings"
	"testing"
)

func TestDoctorReport(t *testing.T) {
	rep := doctorReport()
	if !strings.Contains(rep, "git:") {
		t.Errorf("report should mention git: %q", rep)
	}
	if !strings.Contains(rep, "config:") {
		t.Errorf("report should mention config: %q", rep)
	}
	if !strings.Contains(rep, "color:") {
		t.Errorf("report should mention color: %q", rep)
	}
}
