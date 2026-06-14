package segments

import (
	"fmt"
	"strings"

	"github.com/oleksbard/cosmobar/internal/render"
)

type gitSeg struct{}

func (gitSeg) Name() string { return "git" }

// maxBranchWidth caps long branch names; longer names get a middle ellipsis.
const maxBranchWidth = 16

func (gitSeg) Render(ctx *Context) (Segment, bool) {
	st := ctx.Git
	if !st.InRepo {
		return Segment{}, false
	}
	branch := st.Branch
	if branch == "" {
		branch = "(detached)"
	}
	branch = render.Truncate(branch, maxBranchWidth)
	var flags []string
	if st.Ahead > 0 {
		flags = append(flags, fmt.Sprintf("↑%d", st.Ahead))
	}
	if st.Behind > 0 {
		flags = append(flags, fmt.Sprintf("↓%d", st.Behind))
	}
	if st.Staged > 0 {
		flags = append(flags, fmt.Sprintf("+%d", st.Staged))
	}
	if st.Modified > 0 {
		flags = append(flags, fmt.Sprintf("~%d", st.Modified))
	}
	if st.Untracked > 0 {
		flags = append(flags, fmt.Sprintf("?%d", st.Untracked))
	}
	text := branch
	if len(flags) > 0 {
		text += " " + strings.Join(flags, " ")
	}
	return Segment{Name: "git", Text: text, State: render.StateNone, Prio: 90}, true
}

func init() { register(gitSeg{}) }
