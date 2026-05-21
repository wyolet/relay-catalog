package rules

import (
	"fmt"

	"github.com/wyolet/relay/app/catalogvalidate"
	"github.com/wyolet/relay/app/manifest"
)

// pricing-target-must-bind-to-host: every Pricing.targetModels entry
// must be served by the host that owns the Pricing. Pricing is a
// (Host, Model) tuple — pricing a model that doesn't actually bind to
// the host produces phantom price rows that never apply at request time.
func init() {
	All = append(All, catalogvalidate.Rule{
		Name:        "pricing-target-must-bind-to-host",
		Description: "Pricing.targetModels must each declare a HostBinding to the Pricing's owner host",
		Severity:    catalogvalidate.SeverityError,
		Check:       checkPricingTargetHostBinding,
	})
}

func checkPricingTargetHostBinding(docs []manifest.Document) []catalogvalidate.Issue {
	models := map[string]*manifest.ModelDTO{}
	for _, d := range docs {
		if d.Model != nil {
			models[d.Model.Metadata.Name] = d.Model
		}
	}

	var out []catalogvalidate.Issue
	for _, d := range docs {
		if d.Pricing == nil {
			continue
		}
		host := d.Pricing.Metadata.Owner.Name
		if host == "" {
			host = d.Pricing.Metadata.Owner.ID
		}
		if d.Pricing.Metadata.Owner.Kind != "host" || host == "" {
			// Non-host-owned pricing or missing owner is a different
			// kind of bug — handled by the per-entity Validate() and
			// owner refs in ValidateGraph. Skip here.
			continue
		}
		for i, mname := range d.Pricing.Spec.TargetModels {
			m, ok := models[mname]
			if !ok {
				// ValidateGraph already reports the ref-missing.
				continue
			}
			if !hasBindingToHost(m, host) {
				out = append(out, catalogvalidate.Issue{
					Severity: catalogvalidate.SeverityError,
					Kind:     catalogvalidate.KindInvariant,
					Source: catalogvalidate.Ref{
						Kind:  "Pricing",
						Name:  d.Pricing.Metadata.Name,
						Field: fmt.Sprintf("spec.targetModels[%d]", i),
					},
					Target:  catalogvalidate.Ref{Kind: "Model", Name: mname},
					Message: fmt.Sprintf("model %q has no host binding to %q — pricing row will never apply", mname, host),
				})
			}
		}
	}
	return out
}

func hasBindingToHost(m *manifest.ModelDTO, host string) bool {
	for _, hb := range m.Spec.Hosts {
		if hb.Host == host {
			return true
		}
	}
	return false
}
