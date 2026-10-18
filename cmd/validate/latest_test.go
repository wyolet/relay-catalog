package main

import (
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/wyolet/relay/app/manifest"
)

func model(provider, family, name, released string, tags ...string) manifest.Document {
	return manifest.Document{Model: &manifest.ModelDTO{
		Metadata: manifest.WireMeta{
			Name:  name,
			Owner: manifest.WireOwner{Kind: "provider", Name: provider},
		},
		Spec: manifest.ModelSpec{Family: family, ReleaseDate: released, Tags: tags},
	}}
}

func latestNames(docs []manifest.Document) []string {
	var out []string
	for n := range generateLatest(docs) {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func TestGenerateLatest(t *testing.T) {
	docs := []manifest.Document{
		model("acme", "rocket", "rocket-1", "2025-01-10"),
		model("acme", "rocket", "rocket-2-small", "2026-03-01"),
		model("acme", "rocket", "rocket-2-large", "2026-03-01"),
		// Same family label under another provider is another family.
		model("other", "rocket", "rocket-x", "2027-01-01"),
		model("other", "rocket", "rocket-y", "2024-01-01"),
		// A family of one has nothing to be newer than.
		model("acme", "solo", "solo-1", "2026-05-01"),
	}
	want := []string{"rocket-2-large", "rocket-2-small", "rocket-x"}
	if got := latestNames(docs); !slices.Equal(got, want) {
		t.Fatalf("latest = %v, want %v", got, want)
	}
}

func TestCheckLatest(t *testing.T) {
	docs := []manifest.Document{
		model("acme", "rocket", "rocket-1", "2025-01-10", "latest"),
		model("acme", "rocket", "rocket-2", "2026-03-01"),
		model("acme", "rocket", "rocket-3", "2026-03"),
		model("acme", "solo", "solo-1", "", "newest"),
	}
	var got []string
	for _, is := range checkLatest(docs) {
		got = append(got, is.Source.Name+" "+is.Source.Field)
	}
	sort.Strings(got)
	want := []string{
		"rocket-1 spec.tags",        // tagged but not newest
		"rocket-2 spec.tags",        // newest but untagged
		"rocket-3 spec.releaseDate", // month only
		"solo-1 spec.releaseDate",   // undated
		"solo-1 spec.tags",          // newest is never stored
	}
	if !slices.Equal(got, want) {
		t.Fatalf("issues = %v\nwant %v", got, want)
	}
}

const latestFixture = `kind: Model
metadata:
    name: a
    labels:
        family: f
spec:
    snapshots:
        - name: a
    pointer: a
---
kind: Model
metadata:
  name: b
spec:
  tags:
  - open-weights
  snapshots:
  - name: b
  pointer: b
---
kind: Model
metadata:
    name: c
spec:
    # kept
    tags:
        - latest
    pointer: c
`

func TestSetLatestTags(t *testing.T) {
	got := setLatestTags(latestFixture, map[string]bool{"a": true, "b": true})
	for _, want := range []string{
		"    pointer: a\n    tags:\n        - latest\n---\n",
		"  tags:\n  - latest\n  - open-weights\n",
		"    # kept\n    pointer: c\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if again := setLatestTags(got, map[string]bool{"a": true, "b": true}); again != got {
		t.Errorf("not idempotent:\n%s", again)
	}
	back := setLatestTags(got, map[string]bool{"c": true})
	if strings.Count(back, "- latest") != 1 || !strings.HasSuffix(back, "    pointer: c\n    tags:\n        - latest\n") {
		t.Errorf("moving latest to c:\n%s", back)
	}
}
