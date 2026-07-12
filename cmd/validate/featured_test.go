package main

import (
	"testing"

	"github.com/wyolet/relay/app/manifest"
)

func featuredDocs() []manifest.Document {
	return []manifest.Document{
		{Model: &manifest.ModelDTO{Metadata: manifest.WireMeta{
			Name:   "gpt-x",
			Owner:  manifest.WireOwner{Kind: "provider", Name: "openai"},
			Labels: map[string]string{"featured": "true"},
		}}},
		{Model: &manifest.ModelDTO{Metadata: manifest.WireMeta{
			Name:  "gpt-x-mini",
			Owner: manifest.WireOwner{Kind: "provider", Name: "openai"},
		}}},
		{Model: &manifest.ModelDTO{Metadata: manifest.WireMeta{
			Name:   "claude-y",
			Owner:  manifest.WireOwner{Kind: "provider", Name: "anthropic"},
			Labels: map[string]string{"featured": "true"},
		}}},
	}
}

func TestGenerateFeatured(t *testing.T) {
	m := generateFeatured(featuredDocs())
	if len(m.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d: %+v", len(m.Providers), m.Providers)
	}
	if m.Providers[0].Provider != "anthropic" || m.Providers[1].Provider != "openai" {
		t.Fatalf("providers not sorted: %+v", m.Providers)
	}
	if len(m.Providers[1].Models) != 1 || m.Providers[1].Models[0] != "gpt-x" {
		t.Fatalf("unexpected openai models: %v", m.Providers[1].Models)
	}
}

func TestCheckFeatured_InSync(t *testing.T) {
	docs := featuredDocs()
	if got := checkFeatured(docs, generateFeatured(docs)); len(got) != 0 {
		t.Fatalf("expected no issues, got: %v", got)
	}
}

func TestCheckFeatured_LabelNotInManifest(t *testing.T) {
	docs := featuredDocs()
	want := generateFeatured(docs)
	want.Providers = want.Providers[:1] // drop openai — gpt-x label now unindexed
	got := checkFeatured(docs, want)
	if len(got) != 1 || got[0].Source.Name != "gpt-x" {
		t.Fatalf("expected 1 issue on gpt-x, got: %v", got)
	}
}

func TestCheckFeatured_ManifestEntryNotLabeled(t *testing.T) {
	docs := featuredDocs()
	want := generateFeatured(docs)
	for i := range want.Providers {
		if want.Providers[i].Provider == "openai" {
			want.Providers[i].Models = append(want.Providers[i].Models, "gpt-x-mini")
		}
	}
	got := checkFeatured(docs, want)
	if len(got) != 1 || got[0].Source.Name != "gpt-x-mini" {
		t.Fatalf("expected 1 issue on gpt-x-mini, got: %v", got)
	}
}
