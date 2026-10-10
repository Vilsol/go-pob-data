package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Every version's manifest lists exactly what its directory publishes.
func TestManifestsCurrent(t *testing.T) {
	dirs, err := filepath.Glob("data/*/raw")
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range dirs {
		versionDir := filepath.Dir(dir)
		t.Run(filepath.Base(versionDir), func(t *testing.T) {
			want, err := manifestJSON(versionDir)
			if err != nil {
				t.Fatal(err)
			}

			path := filepath.Join(versionDir, "manifest.json")
			got, err := os.ReadFile(path)
			switch {
			case err == nil && bytes.Equal(got, want):
			case *update:
				if err := os.WriteFile(path, want, 0o644); err != nil {
					t.Fatal(err)
				}
			default:
				t.Errorf("%s is missing or stale: run with -update", path)
			}
		})
	}
}
