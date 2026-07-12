// featured.yaml is the one-place index of models carrying the
// metadata.labels.featured="true" label. The labels stay the source of
// truth (the seed loader and UI read them); the index is generated from
// them with -write-featured and checked for drift on every validate run.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/wyolet/relay/app/catalogvalidate"
	"github.com/wyolet/relay/app/manifest"
)

type featuredManifest struct {
	Description string             `yaml:"description,omitempty"`
	Providers   []featuredProvider `yaml:"providers"`
}

type featuredProvider struct {
	Provider string   `yaml:"provider"`
	Models   []string `yaml:"models"`
}

const featuredDescription = "Featured models — generated index of Model metadata.labels.featured. " +
	"Regenerate: go run ./cmd/validate -write-featured ./data. Not a seeded catalog kind; validate fails on drift."

// featuredPath resolves the manifest location: the repo root next to
// supported-hosts.yaml, one level above the data tree.
func featuredPath(dir string) string {
	return filepath.Join(dir, "..", "featured.yaml")
}

// generateFeatured collects featured-labeled models grouped by owner
// provider, sorted on both levels for a stable file.
func generateFeatured(docs []manifest.Document) featuredManifest {
	byProvider := map[string][]string{}
	for _, d := range docs {
		if d.Model == nil || d.Model.Metadata.Labels["featured"] != "true" {
			continue
		}
		p := d.Model.Metadata.Owner.Name
		if p == "" {
			p = d.Model.Metadata.Owner.ID
		}
		byProvider[p] = append(byProvider[p], d.Model.Metadata.Name)
	}

	m := featuredManifest{Description: featuredDescription}
	for p, models := range byProvider {
		sort.Strings(models)
		m.Providers = append(m.Providers, featuredProvider{Provider: p, Models: models})
	}
	sort.Slice(m.Providers, func(i, j int) bool { return m.Providers[i].Provider < m.Providers[j].Provider })
	return m
}

// checkFeatured reports drift between the checked-in manifest and the
// featured labels in docs, in both directions.
func checkFeatured(docs []manifest.Document, want featuredManifest) []catalogvalidate.Issue {
	got := generateFeatured(docs)

	key := func(provider, model string) string { return provider + "/" + model }
	gotSet := map[string]bool{}
	for _, p := range got.Providers {
		for _, m := range p.Models {
			gotSet[key(p.Provider, m)] = true
		}
	}
	wantSet := map[string]bool{}
	for _, p := range want.Providers {
		for _, m := range p.Models {
			wantSet[key(p.Provider, m)] = true
		}
	}

	var out []catalogvalidate.Issue
	for _, p := range got.Providers {
		for _, m := range p.Models {
			if !wantSet[key(p.Provider, m)] {
				out = append(out, catalogvalidate.Issue{
					Severity: catalogvalidate.SeverityError,
					Kind:     catalogvalidate.KindInvariant,
					Source:   catalogvalidate.Ref{Kind: "Model", Name: m, Field: "metadata.labels.featured"},
					Message:  fmt.Sprintf("labeled featured but absent from featured.yaml under provider %q — regenerate with -write-featured", p.Provider),
				})
			}
		}
	}
	for _, p := range want.Providers {
		for _, m := range p.Models {
			if !gotSet[key(p.Provider, m)] {
				out = append(out, catalogvalidate.Issue{
					Severity: catalogvalidate.SeverityError,
					Kind:     catalogvalidate.KindInvariant,
					Source:   catalogvalidate.Ref{Kind: "Model", Name: m},
					Message:  fmt.Sprintf("featured.yaml lists it under provider %q but no such model carries the featured label — fix the label or regenerate with -write-featured", p.Provider),
				})
			}
		}
	}
	return out
}

func loadFeatured(path string) (featuredManifest, bool, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return featuredManifest{}, false, nil
	}
	if err != nil {
		return featuredManifest{}, false, err
	}
	var m featuredManifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		return featuredManifest{}, false, fmt.Errorf("%s: %w", path, err)
	}
	return m, true, nil
}

func writeFeatured(path string, m featuredManifest) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(4)
	if err := enc.Encode(m); err != nil {
		return err
	}
	return enc.Close()
}
