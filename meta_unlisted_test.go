package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// a version directory holding the unlisted marker keeps its manifest but is left out of meta.json
func TestGenerateMetaSkipsUnlisted(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	for _, dir := range []string{"data/3.29/raw", "data/3.30/raw"} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join("data", "3.30", unlistedMarker), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := generateMeta(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile("data/meta.json")
	if err != nil {
		t.Fatal(err)
	}
	var meta Metadata
	if err := json.Unmarshal(b, &meta); err != nil {
		t.Fatal(err)
	}
	if len(meta.Versions) != 1 || meta.Versions[0] != "3.29" {
		t.Fatalf("versions = %v, want [3.29]", meta.Versions)
	}
	for _, v := range []string{"3.29", "3.30"} {
		if _, err := os.Stat(filepath.Join("data", v, "manifest.json")); err != nil {
			t.Fatalf("%s has no manifest: %v", v, err)
		}
	}
}
