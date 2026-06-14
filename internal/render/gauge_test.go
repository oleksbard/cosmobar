package render

import "testing"

func TestGauge(t *testing.T) {
	txt, st := Gauge(42, 8, 70, 90, false)
	if txt != "▓▓▓░░░░░ 42%" {
		t.Errorf("gauge = %q", txt)
	}
	if st != StateOK {
		t.Errorf("state = %v, want OK", st)
	}
}

func TestGaugeClampAndStates(t *testing.T) {
	if txt, _ := Gauge(0, 8, 70, 90, false); txt != "░░░░░░░░ 0%" {
		t.Errorf("0%% gauge = %q", txt)
	}
	if txt, st := Gauge(150, 8, 70, 90, false); txt != "▓▓▓▓▓▓▓▓ 100%" || st != StateCrit {
		t.Errorf("clamp failed: %q %v", txt, st)
	}
	if _, st := Gauge(75, 8, 70, 90, false); st != StateWarn {
		t.Errorf("75%% should be warn")
	}
}

func TestGaugeASCII(t *testing.T) {
	if txt, _ := Gauge(50, 4, 70, 90, true); txt != "##-- 50%" {
		t.Errorf("ascii gauge = %q", txt)
	}
}
