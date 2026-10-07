// Command embed builds the flattened SDK catalog (catalog.json.gz) from the data tree, the same composition relay's cmd/catalog-embed runs.
//
// Output is byte-reproducible: generatedAt comes from -generated-at instead of the clock, and the gzip header carries no name or mtime.
//
// Usage:
//
//	embed -generated-at <RFC3339> [-version <v>] [-o catalog.json.gz] [dir]
package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/wyolet/relay/app/catalogembed"
	"github.com/wyolet/relay/app/manifest"
)

func main() {
	var out, generatedAt, version string
	flag.StringVar(&out, "o", "catalog.json.gz", "output path (gzip)")
	flag.StringVar(&generatedAt, "generated-at", "", "generatedAt timestamp, RFC3339 (required)")
	flag.StringVar(&version, "version", "", "override the catalog version field (empty keeps the composed value)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: embed -generated-at <RFC3339> [-version <v>] [-o path] [dir]")
		flag.PrintDefaults()
	}
	flag.Parse()

	at, err := time.Parse(time.RFC3339, generatedAt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "-generated-at: %v\n", err)
		flag.Usage()
		os.Exit(2)
	}
	dir := flag.Arg(0)
	if dir == "" {
		dir = "./data"
	}

	docs, err := manifest.LoadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load %s: %v\n", dir, err)
		os.Exit(2)
	}
	cat, err := catalogembed.Compose(docs, at)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compose: %v\n", err)
		os.Exit(2)
	}
	if err := catalogembed.ValidateAdapters(cat); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if version != "" {
		cat.Version = version
	}
	data, err := catalogembed.MarshalJSON(cat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(2)
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
		os.Exit(2)
	}
	if _, err := gz.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
		os.Exit(2)
	}
	if err := gz.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
		os.Exit(2)
	}
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", out, err)
		os.Exit(2)
	}
}
