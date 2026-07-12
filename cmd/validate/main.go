// validate is the catalog repo's CI / local-check entry point. It runs:
//
//  1. catalogvalidate.ValidateGraph — relay-side, schema-generic graph
//     linter (refs, owners, snapshots, dupes, orphan warnings).
//  2. catalogvalidate.RunRules with rules.All — this repo's curation
//     conventions (e.g. tag completeness, pricing-target invariants).
//
// Flags:
//
//	--strict          promote warnings to errors before exit
//	--skip <name>     suppress a named rule (repeatable)
//	--list            print rule registry (name, severity, description) and exit
//
// Exit codes:
//
//	0  no errors (warnings may be present unless --strict)
//	1  at least one error
//	2  internal failure
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wyolet/relay/app/catalogvalidate"
	"github.com/wyolet/relay/app/manifest"

	"github.com/wyolet/relay-catalog/cmd/validate/rules"
)

type stringSlice []string

func (s *stringSlice) String() string     { return fmt.Sprintf("%v", *s) }
func (s *stringSlice) Set(v string) error { *s = append(*s, v); return nil }

func main() {
	var strict, list, regenFeatured bool
	var skip stringSlice
	flag.BoolVar(&strict, "strict", false, "promote warnings to errors before exit")
	flag.BoolVar(&list, "list", false, "print rule registry and exit")
	flag.BoolVar(&regenFeatured, "write-featured", false, "regenerate featured.yaml from Model featured labels and exit")
	flag.Var(&skip, "skip", "suppress a named rule (repeatable)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: validate [--strict] [--list] [--write-featured] [--skip <name>...] <dir>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if list {
		printRules()
		return
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	dir := flag.Arg(0)

	docs, err := manifest.LoadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load %s: %v\n", dir, err)
		os.Exit(2)
	}

	if regenFeatured {
		path := featuredPath(dir)
		if err := writeFeatured(path, generateFeatured(docs)); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
			os.Exit(2)
		}
		fmt.Println("wrote", path)
		return
	}

	skipSet := make(map[string]bool, len(skip))
	for _, n := range skip {
		skipSet[n] = true
	}

	issues := catalogvalidate.ValidateGraph(docs)
	issues = append(issues, catalogvalidate.RunRules(rules.All, docs, skipSet)...)

	if want, ok, err := loadFeatured(featuredPath(dir)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	} else if ok {
		issues = append(issues, checkFeatured(docs, want)...)
	}

	if strict {
		issues = catalogvalidate.Promote(issues)
	}

	fmt.Print(catalogvalidate.Format(issues))
	if catalogvalidate.HasErrors(issues) {
		os.Exit(1)
	}
}

// printRules emits the rule registry sorted by Name. CI doesn't use
// this; humans do (`validate --list`).
func printRules() {
	for _, r := range catalogvalidate.ListRules(rules.All) {
		fmt.Printf("%-40s %-8s %s\n", r.Name, r.Severity, r.Description)
	}
}
