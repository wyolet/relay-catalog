package rules

import (
	"fmt"

	"github.com/wyolet/relay/app/catalogvalidate"
	"github.com/wyolet/relay/app/manifest"
)

// host-has-tier-policy: every Host should declare at least one tier
// Policy owned by it. Without a tier policy, the host has no
// rate-limit defaults for hostkeys to anchor on, and the catalog
// can't seed a working policy ladder (free/tier-1/tier-2/...).
//
// A few hosts intentionally have no formal tiers (Ollama-self has
// `ollama-unbounded`), so this is a warning, not an error.
func init() {
	All = append(All, catalogvalidate.Rule{
		Name:        "host-has-tier-policy",
		Description: "Host should declare at least one Policy owned by it (its tier ladder)",
		Severity:    catalogvalidate.SeverityWarning,
		Check:       checkHostHasTierPolicy,
	})
}

func checkHostHasTierPolicy(docs []manifest.Document) []catalogvalidate.Issue {
	tierCount := map[string]int{} // host name → count
	for _, d := range docs {
		if d.Policy == nil {
			continue
		}
		if d.Policy.Metadata.Owner.Kind != "host" {
			continue
		}
		hname := d.Policy.Metadata.Owner.Name
		if hname == "" {
			hname = d.Policy.Metadata.Owner.ID
		}
		if hname != "" {
			tierCount[hname]++
		}
	}

	var out []catalogvalidate.Issue
	for _, d := range docs {
		if d.Host == nil {
			continue
		}
		name := d.Host.Metadata.Name
		if tierCount[name] == 0 {
			out = append(out, catalogvalidate.Issue{
				Severity: catalogvalidate.SeverityWarning,
				Kind:     catalogvalidate.KindIncomplete,
				Source:   catalogvalidate.Ref{Kind: "Host", Name: name},
				Message:  fmt.Sprintf("host has no tier policies — hostkeys can't anchor a rate-limit ladder"),
			})
		}
	}
	return out
}
