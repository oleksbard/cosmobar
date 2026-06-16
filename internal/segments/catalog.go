package segments

// Meta describes a segment for discovery and setup UIs (e.g. the guided
// installer). It is the machine-readable catalog of "all possible nodes".
type Meta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DefaultOn   bool   `json:"default_on"`
	RequiresGit bool   `json:"requires_git"`
	ProMaxOnly  bool   `json:"pro_max_only"`
	Role        string `json:"role"`
}

// catalog lists every segment in display order with a human description.
// Keep it in sync with the registered segments — TestCatalogMatchesRegistry
// fails if a segment is registered without a catalog entry (or vice versa).
var catalog = []Meta{
	{Name: "dir", Description: "Current directory name", DefaultOn: true, Role: "identity"},
	{Name: "git", Description: "Branch + ahead/behind + staged/modified/untracked counts", DefaultOn: true, RequiresGit: true, Role: "vcs"},
	{Name: "model", Description: "Active model name", DefaultOn: true, Role: "identity"},
	{Name: "context", Description: "Context-window usage gauge", DefaultOn: true, Role: "gauge"},
	{Name: "cost", Description: "Session cost in USD", DefaultOn: true, Role: "metric"},
	{Name: "tokens", Description: "Cumulative session tokens (input + output)", Role: "usage"},
	{Name: "clock", Description: "Current time", DefaultOn: true, Role: "ambient"},
	{Name: "rate_limits", Description: "5-hour and 7-day usage limits", ProMaxOnly: true, Role: "gauge"},
	{Name: "duration", Description: "Session wall-clock duration", Role: "metric"},
	{Name: "lines", Description: "Lines added/removed vs last commit", RequiresGit: true, Role: "metric"},
	{Name: "output_style", Description: "Active output style name", Role: "ambient"},
	{Name: "git_stash", Description: "Git stash count", RequiresGit: true, Role: "vcs"},
	{Name: "effort", Description: "Reasoning effort level", Role: "ambient"},
}

// Catalog returns every segment's metadata in display order.
func Catalog() []Meta { return catalog }

// DefaultOrder returns the names of the default-on segments, in display order.
func DefaultOrder() []string {
	var out []string
	for _, m := range catalog {
		if m.DefaultOn {
			out = append(out, m.Name)
		}
	}
	return out
}

// MetaByName returns the catalog entry for a segment name.
func MetaByName(name string) (Meta, bool) {
	for _, m := range catalog {
		if m.Name == name {
			return m, true
		}
	}
	return Meta{}, false
}
