// The latest tag marks, per provider and family, the model with the
// newest spec.releaseDate. It is stored in spec.tags so relay and the UI read
// it like any other tag, generated with -write-latest, and checked on every
// validate run. Models released the same day as the newest all carry it: size
// variants ship together and none of them is newer. A family of one gets no
// tag, since there is nothing to be newer than.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/wyolet/relay/app/catalogvalidate"
	"github.com/wyolet/relay/app/manifest"
)

const latestTag = "latest"

// reservedTags are computed, never stored: newest comes from spec.releaseDate
// in the UI.
var reservedTags = []string{"newest"}

func modelProvider(m *manifest.ModelDTO) string {
	if m.Metadata.Owner.Name != "" {
		return m.Metadata.Owner.Name
	}
	return m.Metadata.Owner.ID
}

// generateLatest returns the names of the models that should carry the latest
// tag. Undated models are skipped here; checkLatest reports them.
func generateLatest(docs []manifest.Document) map[string]bool {
	families := map[string][]*manifest.ModelDTO{}
	for _, d := range docs {
		if d.Model == nil {
			continue
		}
		key := modelProvider(d.Model) + "/" + d.Model.Spec.Family
		families[key] = append(families[key], d.Model)
	}
	out := map[string]bool{}
	for _, members := range families {
		if len(members) < 2 {
			continue
		}
		newest := ""
		for _, m := range members {
			newest = max(newest, m.Spec.ReleaseDate)
		}
		if newest == "" {
			continue
		}
		for _, m := range members {
			if m.Spec.ReleaseDate == newest {
				out[m.Metadata.Name] = true
			}
		}
	}
	return out
}

// checkLatest reports models without a full release date, stored tags that
// are computed elsewhere, and drift between the stored latest tags and the
// rule.
func checkLatest(docs []manifest.Document) []catalogvalidate.Issue {
	var out []catalogvalidate.Issue
	issue := func(name, field, msg string) {
		out = append(out, catalogvalidate.Issue{
			Severity: catalogvalidate.SeverityError,
			Kind:     catalogvalidate.KindInvariant,
			Source:   catalogvalidate.Ref{Kind: "Model", Name: name, Field: field},
			Message:  msg,
		})
	}
	want := generateLatest(docs)
	for _, d := range docs {
		m := d.Model
		if m == nil {
			continue
		}
		name := m.Metadata.Name
		if _, err := time.Parse(time.DateOnly, m.Spec.ReleaseDate); err != nil {
			issue(name, "spec.releaseDate", fmt.Sprintf("%q is not a YYYY-MM-DD date — latest is decided by it", m.Spec.ReleaseDate))
		}
		for _, t := range reservedTags {
			if slices.Contains(m.Spec.Tags, t) {
				issue(name, "spec.tags", fmt.Sprintf("%q is computed from spec.releaseDate and must not be stored", t))
			}
		}
		has := slices.Contains(m.Spec.Tags, latestTag)
		switch {
		case want[name] && !has:
			issue(name, "spec.tags", "newest in its family but not tagged latest — regenerate with -write-latest")
		case !want[name] && has:
			issue(name, "spec.tags", "tagged latest but not the newest in its family — regenerate with -write-latest")
		}
	}
	return out
}

// writeLatest adds or removes the latest tag in every Model document under
// dir so the stored tags match generateLatest. It edits lines in place so
// formatting and comments survive.
func writeLatest(dir string, docs []manifest.Document) (int, error) {
	want := generateLatest(docs)
	changed := 0
	err := filepath.WalkDir(dir, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			if e.Name() == "drafts" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".draft.yaml") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		next := setLatestTags(string(b), want)
		if next == string(b) {
			return nil
		}
		changed++
		return os.WriteFile(path, []byte(next), 0o644)
	})
	return changed, err
}

var (
	kindModelLine = regexp.MustCompile(`^kind:\s*Model\s*$`)
	metaNameLine  = regexp.MustCompile(`^(\s+)name:\s*['"]?([^'"\s]+)['"]?\s*$`)
)

// setLatestTags rewrites the spec.tags of each Model document in a
// multi-document YAML file to carry latest exactly when want names it.
func setLatestTags(src string, want map[string]bool) string {
	docs := strings.SplitAfter(src, "\n---\n")
	for i, doc := range docs {
		docs[i] = setDocLatest(doc, want)
	}
	return strings.Join(docs, "")
}

func setDocLatest(doc string, want map[string]bool) string {
	lines := strings.SplitAfter(doc, "\n")
	isModel, name := false, ""
	inMeta := false
	spec := -1
	for i, l := range lines {
		t := strings.TrimRight(l, "\n")
		switch {
		case kindModelLine.MatchString(t):
			isModel = true
		case t == "metadata:":
			inMeta = true
		case t == "spec:":
			inMeta = false
			spec = i
		case inMeta && t != "" && !strings.HasPrefix(t, " "):
			inMeta = false
		}
		if inMeta && name == "" {
			if m := metaNameLine.FindStringSubmatch(t); m != nil && len(m[1]) <= 4 {
				name = m[2]
			}
		}
	}
	if !isModel || name == "" || spec < 0 || spec+1 >= len(lines) {
		return doc
	}
	indent := leadingSpace(lines[spec+1])
	// The spec's direct children, and where the tags block and its items sit.
	end := len(lines)
	for i := spec + 1; i < len(lines); i++ {
		t := strings.TrimRight(lines[i], "\n")
		if t != "" && !strings.HasPrefix(t, " ") {
			end = i
			break
		}
	}
	tagsAt, itemIndent := -1, ""
	for i := spec + 1; i < end; i++ {
		if strings.TrimRight(lines[i], "\n") == indent+"tags:" {
			tagsAt = i
		}
	}
	has := -1
	if tagsAt >= 0 {
		for i := tagsAt + 1; i < end; i++ {
			t := strings.TrimRight(lines[i], "\n")
			s := strings.TrimLeft(t, " ")
			if !strings.HasPrefix(s, "- ") || len(t)-len(s) < len(indent) {
				break
			}
			itemIndent = t[:len(t)-len(s)]
			if strings.TrimSpace(strings.TrimPrefix(s, "- ")) == latestTag {
				has = i
			}
		}
	}
	switch {
	case want[name] && has < 0 && tagsAt >= 0:
		if itemIndent == "" {
			itemIndent = seqIndent(lines, spec+1, end, indent)
		}
		lines = slices.Insert(lines, tagsAt+1, itemIndent+"- "+latestTag+"\n")
	case want[name] && has < 0:
		block := []string{indent + "tags:\n", seqIndent(lines, spec+1, end, indent) + "- " + latestTag + "\n"}
		at := end
		for at > spec+1 && strings.TrimSpace(lines[at-1]) == "" {
			at--
		}
		if at == len(lines) && !strings.HasSuffix(lines[at-1], "\n") {
			lines[at-1] += "\n"
		}
		lines = slices.Insert(lines, at, block...)
	case !want[name] && has >= 0:
		lines = slices.Delete(lines, has, has+1)
		// A tags key left with no items would parse as null; drop it too.
		if next := tagsAt + 1; next >= len(lines) || !strings.HasPrefix(strings.TrimLeft(lines[next], " "), "- ") || len(leadingSpace(lines[next])) < len(indent) {
			lines = slices.Delete(lines, tagsAt, tagsAt+1)
		}
	}
	return strings.Join(lines, "")
}

// seqIndent is the indent this document gives sequence items under a spec
// key, read from its snapshots list: some files indent items past the key,
// some align them with it.
func seqIndent(lines []string, from, to int, indent string) string {
	for i := from; i < to-1; i++ {
		if strings.TrimRight(lines[i], "\n") == indent+"snapshots:" {
			return leadingSpace(lines[i+1])
		}
	}
	return indent + "    "
}

func leadingSpace(l string) string {
	return l[:len(l)-len(strings.TrimLeft(l, " "))]
}
