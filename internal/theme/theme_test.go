package theme

import "testing"

func TestGetKnown(t *testing.T) {
	p, ok := Get("coral")
	if !ok {
		t.Fatal("coral should exist")
	}
	if p.Accent == (RGB{}) {
		t.Error("coral accent should be non-zero")
	}
}

func TestGetUnknown(t *testing.T) {
	if _, ok := Get("does-not-exist"); ok {
		t.Error("unknown theme should not be found")
	}
}

func TestNamesSortedAndComplete(t *testing.T) {
	names := Names()
	if len(names) != 4 {
		t.Fatalf("expected 4 themes, got %v", names)
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("Names() not sorted: %v", names)
		}
	}
}

func TestPaletteHasExtendedColors(t *testing.T) {
	for _, name := range Names() {
		p, _ := Get(name)
		if p.Secondary == (RGB{}) || p.Tertiary == (RGB{}) {
			t.Errorf("%s: Secondary/Tertiary must be set", name)
		}
		if p.Dark == (RGB{}) || p.Light == (RGB{}) {
			t.Errorf("%s: Dark/Light contrast inks must be set", name)
		}
	}
}
