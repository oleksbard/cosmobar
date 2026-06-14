package segments

// Meta describes a segment for discovery and setup UIs (e.g. the guided
// installer). It is the machine-readable catalog of "all possible nodes".
type Meta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DefaultOn   bool   `json:"default_on"`
	RequiresGit bool   `json:"requires_git"`
	ProMaxOnly  bool   `json:"pro_max_only"`
}

// catalog lists every segment in display order with a human description.
// Keep it in sync with the registered segments — TestCatalogMatchesRegistry
// fails if a segment is registered without a catalog entry (or vice versa).
var catalog = []Meta{
	{Name: "dir", Description: "Current directory name", DefaultOn: true},
	{Name: "git", Description: "Branch + ahead/behind + staged/modified/untracked counts", DefaultOn: true, RequiresGit: true},
	{Name: "model", Description: "Active model name", DefaultOn: true},
	{Name: "context", Description: "Context-window usage gauge", DefaultOn: true},
	{Name: "cost", Description: "Session cost in USD", DefaultOn: true},
	{Name: "clock", Description: "Current time", DefaultOn: true},
	{Name: "rate_limits", Description: "5-hour and 7-day usage limits", ProMaxOnly: true},
	{Name: "duration", Description: "Session wall-clock duration"},
	{Name: "lines", Description: "Lines added/removed this session"},
	{Name: "output_style", Description: "Active output style name"},
	{Name: "git_stash", Description: "Git stash count", RequiresGit: true},
	{Name: "effort", Description: "Reasoning effort level"},
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
