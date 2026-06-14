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
