// families.yaml is the allowlist of Model spec.family values. A family is the
// base model lineage, so an import that writes a versioned or product-line
// value (gemini-3-flash, gemma3, gpt-mini) fails here instead of splitting a
// lineage into many families.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/wyolet/relay/app/catalogvalidate"
	"github.com/wyolet/relay/app/manifest"
)

type familiesManifest struct {
	Families []struct {
		Name    string   `yaml:"name"`
		Aliases []string `yaml:"aliases,omitempty"`
	} `yaml:"families"`
}

func familiesPath(dir string) string {
	return filepath.Join(dir, "..", "families.yaml")
}

func loadFamilies(path string) (map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m familiesManifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	out := make(map[string]bool, len(m.Families))
	for _, f := range m.Families {
		out[f.Name] = true
	}
	return out, nil
}

// checkFamilies reports every model whose spec.family is missing or not in
// the allowlist.
func checkFamilies(docs []manifest.Document, allowed map[string]bool) []catalogvalidate.Issue {
	var out []catalogvalidate.Issue
	for _, d := range docs {
		if d.Model == nil || allowed[d.Model.Spec.Family] {
			continue
		}
		msg := fmt.Sprintf("%q is not in families.yaml — use the base model's family, or add a new lineage there", d.Model.Spec.Family)
		if d.Model.Spec.Family == "" {
			msg = "missing — every model names its base model lineage from families.yaml"
		}
		out = append(out, catalogvalidate.Issue{
			Severity: catalogvalidate.SeverityError,
			Kind:     catalogvalidate.KindInvariant,
			Source:   catalogvalidate.Ref{Kind: "Model", Name: d.Model.Metadata.Name, Field: "spec.family"},
			Message:  msg,
		})
	}
	return out
}
