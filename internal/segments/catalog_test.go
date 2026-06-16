package segments

import (
	"reflect"
	"testing"

	"github.com/oleksbard/cosmobar/internal/config"
)

func TestCatalogMatchesRegistry(t *testing.T) {
	// every catalog entry must be a registered segment
	for _, m := range Catalog() {
		if _, ok := Get(m.Name); !ok {
			t.Errorf("catalog lists %q but it is not registered", m.Name)
		}
	}
	// every registered segment must appear in the catalog
	inCatalog := map[string]bool{}
	for _, m := range Catalog() {
		inCatalog[m.Name] = true
	}
	for name := range registry {
		if !inCatalog[name] {
			t.Errorf("registered segment %q is missing from the catalog", name)
		}
	}
}

func TestDefaultOrderMatchesConfigDefault(t *testing.T) {
	if got, want := DefaultOrder(), config.Default().Order; !reflect.DeepEqual(got, want) {
		t.Errorf("DefaultOrder() = %v, config default order = %v", got, want)
	}
}

func TestEverySegmentHasValidRole(t *testing.T) {
	valid := map[string]bool{"identity": true, "vcs": true, "metric": true, "gauge": true, "ambient": true, "usage": true}
	for _, m := range Catalog() {
		if !valid[m.Role] {
			t.Errorf("%s has invalid role %q", m.Name, m.Role)
		}
	}
}

func TestTokensUsesUsageRole(t *testing.T) {
	m, ok := MetaByName("tokens")
	if !ok || m.Role != "usage" {
		t.Errorf("tokens role = %q ok=%v, want usage", m.Role, ok)
	}
}

func TestMetaByName(t *testing.T) {
	m, ok := MetaByName("git")
	if !ok || m.Role != "vcs" {
		t.Errorf("git role = %q ok=%v", m.Role, ok)
	}
	if _, ok := MetaByName("nope"); ok {
		t.Error("unknown name should not resolve")
	}
}
