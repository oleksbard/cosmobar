package cli

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/oleksbard/cosmobar/internal/release"
)

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func httpGet(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "cosmobar-upgrade")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func latestRelease() (*ghRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", release.Owner, release.Repo)
	data, err := httpGet(url)
	if err != nil {
		return nil, err
	}
	var r ghRelease
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func assetURL(r *ghRelease, name string) string {
	for _, a := range r.Assets {
		if a.Name == name {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

func cmdUpgrade(args []string) int {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	checkOnly := fs.Bool("check", false, "report current vs latest and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	r, err := latestRelease()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: cannot fetch latest release:", err)
		return 1
	}
	fmt.Printf("current: %s   latest: %s\n", version, r.TagName)

	if !release.IsNewer(r.TagName, version) {
		fmt.Println("cosmobar: already up to date.")
		return 0
	}
	if *checkOnly {
		fmt.Println("cosmobar: an update is available. Run `cosmobar upgrade` to install it.")
		return 0
	}

	assetName := release.AssetName(runtime.GOOS, runtime.GOARCH)
	aURL := assetURL(r, assetName)
	cURL := assetURL(r, "checksums.txt")
	if aURL == "" || cURL == "" {
		fmt.Fprintf(os.Stderr, "cosmobar: release has no asset %q\n", assetName)
		return 1
	}

	archive, err := httpGet(aURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: download failed:", err)
		return 1
	}
	sumsBody, err := httpGet(cURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: checksums download failed:", err)
		return 1
	}
	sums := release.ParseChecksums(string(sumsBody))
	gotSum := sha256.Sum256(archive)
	if hex.EncodeToString(gotSum[:]) != sums[assetName] {
		fmt.Fprintln(os.Stderr, "cosmobar: checksum mismatch; aborting.")
		return 1
	}

	bin, err := extractBinary(archive)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: extract failed:", err)
		return 1
	}
	if err := replaceSelf(bin); err != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: replace failed:", err)
		return 1
	}
	fmt.Printf("cosmobar: upgraded to %s\n", r.TagName)
	return 0
}

// extractBinary pulls the "cosmobar" file out of a .tar.gz archive.
func extractBinary(archive []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytesReader(archive))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("cosmobar binary not found in archive")
		}
		if err != nil {
			return nil, err
		}
		if filepath.Base(hdr.Name) == "cosmobar" {
			return io.ReadAll(tr)
		}
	}
}

// replaceSelf atomically swaps the running executable with newBin.
func replaceSelf(newBin []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	dir := filepath.Dir(exe)
	tmp, err := os.CreateTemp(dir, ".cosmobar-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(newBin); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return err
	}
	return os.Rename(tmpName, exe)
}
