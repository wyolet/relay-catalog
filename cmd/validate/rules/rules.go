// Package rules declares catalog-specific curation conventions on top of
// relay's schema-generic graph linter. Each *.go file in this package
// exports one or more catalogvalidate.Rule values appended to the All
// slice via init() — keeping rule code, its constants, and its unit
// tests colocated.
//
// To add a rule:
//
//  1. Create rules/<rule_name>.go.
//  2. Implement a `func check<X>(docs []manifest.Document) []catalogvalidate.Issue`.
//  3. Append a Rule literal to All in that file's init().
//  4. Add rules/<rule_name>_test.go with table-driven coverage.
//
// Naming: rule Name field is kebab-case, stable, used by --skip flags
// and bug reports. Don't rename without a deprecation note.
package rules

import "github.com/wyolet/relay/app/catalogvalidate"

// All is the catalog repo's curation rule set. Init functions in
// sibling files append to this slice — keeps every rule discoverable in
// one place via `validate --list` and auto-generated docs.
var All []catalogvalidate.Rule
