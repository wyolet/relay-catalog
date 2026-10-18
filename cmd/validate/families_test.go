package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wyolet/relay/app/manifest"
)

func TestCheckFamilies(t *testing.T) {
	doc := func(name, family string) manifest.Document {
		return manifest.Document{Model: &manifest.ModelDTO{
			Metadata: manifest.WireMeta{Name: name},
			Spec:     manifest.ModelSpec{Family: family},
		}}
	}
	issues := checkFamilies([]manifest.Document{
		doc("ok", "gemini"),
		doc("versioned", "gemini-3-flash"),
		doc("none", ""),
	}, map[string]bool{"gemini": true})
	if len(issues) != 2 || issues[0].Source.Name != "versioned" || issues[1].Source.Name != "none" {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestLoadFamilies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "families.yaml")
	if err := os.WriteFile(path, []byte("families:\n    - name: gpt\n      aliases: [o]\n    - name: claude\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := loadFamilies(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got["gpt"] || !got["claude"] || got["o"] || len(got) != 2 {
		t.Fatalf("families = %v", got)
	}
}
