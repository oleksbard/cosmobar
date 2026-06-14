package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type marketplaceManifest struct {
	Name  string `json:"name"`
	Owner struct {
		Name string `json:"name"`
	} `json:"owner"`
	Plugins []struct {
		Name        string `json:"name"`
		Source      string `json:"source"`
		Description string `json:"description"`
	} `json:"plugins"`
}

type pluginManifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

func TestMarketplaceManifestValid(t *testing.T) {
	data, err := os.ReadFile(".claude-plugin/marketplace.json")
	if err != nil {
		t.Fatalf("read marketplace.json: %v", err)
	}
	var m marketplaceManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("marketplace.json is not valid JSON: %v", err)
	}
	if m.Name == "" {
		t.Error("marketplace.json: name must be set")
	}
	if m.Owner.Name == "" {
		t.Error("marketplace.json: owner.name must be set")
	}
	if len(m.Plugins) == 0 {
		t.Fatal("marketplace.json: plugins must be non-empty")
	}
	for i, p := range m.Plugins {
		if p.Name == "" || p.Source == "" {
			t.Errorf("marketplace.json: plugins[%d] needs name+source, got %+v", i, p)
		}
	}
}

func TestPluginManifestValid(t *testing.T) {
	data, err := os.ReadFile(".claude-plugin/plugin.json")
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	var p pluginManifest
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("plugin.json is not valid JSON: %v", err)
	}
	if p.Name == "" {
		t.Error("plugin.json: name must be set")
	}
	if p.Version == "" {
		t.Error("plugin.json: version must be set")
	}
	if p.Description == "" {
		t.Error("plugin.json: description must be set")
	}
}

func TestInstallSkillHasFrontmatter(t *testing.T) {
	data, err := os.ReadFile("skills/install/SKILL.md")
	if err != nil {
		t.Fatalf("read skills/install/SKILL.md: %v", err)
	}
	s := string(data)
	if !strings.HasPrefix(s, "---") {
		t.Error("SKILL.md must start with YAML frontmatter (---)")
	}
	if !strings.Contains(s, "name:") || !strings.Contains(s, "description:") {
		t.Error("SKILL.md frontmatter must include name and description")
	}
	for _, want := range []string{"install.sh", "cosmobar init", "/cosmobar"} {
		if !strings.Contains(s, want) {
			t.Errorf("SKILL.md should reference %q", want)
		}
	}
}

func TestReadmeDocumentsPluginInstall(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	s := string(data)
	for _, want := range []string{
		"/plugin marketplace add oleksbard/cosmobar",
		"/plugin install cosmobar@cosmobar",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("README.md should document %q", want)
		}
	}
}
