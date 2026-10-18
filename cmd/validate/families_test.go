package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wyolet/relay/app/manifest"
)

func modelDoc(name, family, line string) manifest.Document {
	return manifest.Document{Model: &manifest.ModelDTO{
		Metadata: manifest.WireMeta{Name: name},
		Spec:     manifest.ModelSpec{Family: family, ProductLine: line},
	}}
}

func TestCheckFamilies(t *testing.T) {
	issues := checkFamilies([]manifest.Document{
		modelDoc("ok", "gemini", ""),
		modelDoc("versioned", "gemini-3-flash", ""),
		modelDoc("none", "", ""),
	}, map[string]map[string]bool{"gemini": nil})
	if len(issues) != 2 || issues[0].Source.Name != "versioned" || issues[1].Source.Name != "none" {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestCheckProductLines(t *testing.T) {
	allowed := map[string]map[string]bool{
		"claude": {"opus": true},
		"gpt":    {"o": true},
		"falcon": {},
	}
	issues := checkProductLines([]manifest.Document{
		modelDoc("ok", "claude", "opus"),
		modelDoc("empty", "gpt", ""),
		modelDoc("other-family", "gpt", "opus"),
		modelDoc("line-alias", "gpt", "o-mini"),
		modelDoc("no-lines", "falcon", "falcon"),
		modelDoc("unknown-family", "gemini-3", "flash"),
	}, allowed)
	want := []string{"other-family", "line-alias", "no-lines"}
	if len(issues) != len(want) {
		t.Fatalf("issues = %+v", issues)
	}
	for i, name := range want {
		if issues[i].Source.Name != name || issues[i].Source.Field != "spec.productLine" {
			t.Fatalf("issues[%d] = %+v, want %s", i, issues[i], name)
		}
	}
}

func TestLoadFamilies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "families.yaml")
	body := "families:\n" +
		"    - name: gpt\n" +
		"      aliases: [o]\n" +
		"      lines:\n" +
		"        - name: o\n" +
		"          aliases: [o-mini]\n" +
		"        - name: codex\n" +
		"    - name: claude\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := loadFamilies(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["o"]; ok || len(got) != 2 {
		t.Fatalf("families = %v", got)
	}
	if gpt := got["gpt"]; !gpt["o"] || !gpt["codex"] || gpt["o-mini"] || len(gpt) != 2 {
		t.Fatalf("gpt lines = %v", gpt)
	}
	if claude, ok := got["claude"]; !ok || len(claude) != 0 {
		t.Fatalf("claude lines = %v", claude)
	}
}
