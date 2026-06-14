package release

import "testing"

func TestAssetName(t *testing.T) {
	if got := AssetName("darwin", "arm64"); got != "cosmobar_darwin_arm64.tar.gz" {
		t.Errorf("asset = %q", got)
	}
	if got := AssetName("linux", "amd64"); got != "cosmobar_linux_amd64.tar.gz" {
		t.Errorf("asset = %q", got)
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v1.2.0", "v1.1.0", true},
		{"v1.2.0", "1.2.0", false},
		{"v1.2.0", "v1.2.0", false},
		{"v1.2.0", "dev", true}, // dev always upgradeable
		{"v1.0.0", "v1.2.0", false},
		{"v1.10.0", "v1.9.0", true},
	}
	for _, c := range cases {
		if got := IsNewer(c.latest, c.current); got != c.want {
			t.Errorf("IsNewer(%q,%q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestParseChecksums(t *testing.T) {
	body := "abc123  cosmobar_darwin_arm64.tar.gz\ndef456  cosmobar_linux_amd64.tar.gz\n"
	m := ParseChecksums(body)
	if m["cosmobar_darwin_arm64.tar.gz"] != "abc123" {
		t.Errorf("checksum = %q", m["cosmobar_darwin_arm64.tar.gz"])
	}
}
