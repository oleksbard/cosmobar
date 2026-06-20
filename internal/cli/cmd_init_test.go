package cli

import "testing"

func TestBuildInitConfigCostAndBlockCost(t *testing.T) {
	cfg, customized := buildInitConfig("", "", "", "", "", "", "", "", "today,month", "off")
	if !customized {
		t.Error("expected customized=true when flags are set")
	}
	if len(cfg.Cost.Rollups) != 2 || cfg.Cost.Rollups[0] != "today" || cfg.Cost.Rollups[1] != "month" {
		t.Errorf("cost rollups = %v, want [today month]", cfg.Cost.Rollups)
	}
	if cfg.RateLimits.ShowBlockCost {
		t.Error("--block-cost off should set ShowBlockCost=false")
	}
}

func TestBuildInitConfigDefaultsWhenEmpty(t *testing.T) {
	cfg, customized := buildInitConfig("", "", "", "", "", "", "", "", "", "")
	if customized {
		t.Error("expected customized=false when no flags set")
	}
	if len(cfg.Cost.Rollups) != 1 || cfg.Cost.Rollups[0] != "today" {
		t.Errorf("default rollups = %v, want [today]", cfg.Cost.Rollups)
	}
}
