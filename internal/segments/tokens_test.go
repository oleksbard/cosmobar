package segments

import (
	"testing"

	"github.com/oleksbard/cosmobar/internal/session"
)

func TestHumanTokens(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{847, "847"},
		{1000, "1k"},
		{1500, "1.5k"},
		{9400, "9.4k"},
		{99_500, "99.5k"},
		{100_000, "100k"},
		{412_300, "412k"},
		{1_240_000, "1.2M"},
		{12_000_000, "12M"},
		{999, "999"},
		{99_950, "100k"},
		{999_999, "999k"},
	}
	for _, c := range cases {
		if got := humanTokens(c.n); got != c.want {
			t.Errorf("humanTokens(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestTokensSegmentHiddenWhenNil(t *testing.T) {
	if seg, ok := (tokensSeg{}).Render(&Context{}); ok {
		t.Errorf("expected hidden, got %+v", seg)
	}
}

func TestTokensSegmentRenders(t *testing.T) {
	ctx := &Context{SessionTokens: &session.TokenUsage{Input: 400_000, Output: 12_000}}
	seg, ok := (tokensSeg{}).Render(ctx)
	if !ok {
		t.Fatal("expected visible segment")
	}
	if seg.Text != "412k tok" {
		t.Errorf("Text = %q, want %q", seg.Text, "412k tok")
	}
}
