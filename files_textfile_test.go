package main

import (
	"errors"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

// 3.30 ships stat files as .csd where 3.29 had .txt; a text file is opened under either name
func TestOpenTextFileOtherExtension(t *testing.T) {
	bundles := fstest.MapFS{
		"metadata/statdescriptions/skillpopup_stat_filters.csd": {Data: []byte("csd")},
		"metadata/characters/character.ot":                      {Data: []byte("ot")},
	}
	read := func(name string) string {
		t.Helper()
		f, err := openTextFile(bundles, name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		defer f.Close()
		b, _ := io.ReadAll(f)
		return string(b)
	}
	if got := read("Metadata/StatDescriptions/skillpopup_stat_filters.txt"); got != "csd" {
		t.Fatalf("got %q", got)
	}
	if got := read("Metadata/Characters/Character.ot"); got != "ot" {
		t.Fatalf("got %q", got)
	}
	// neither extension, and a path outside StatDescriptions, stay not-found
	for _, name := range []string{"Metadata/StatDescriptions/missing.txt", "Metadata/Characters/Character.txt"} {
		if _, err := openTextFile(bundles, name); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("%s: err = %v, want not exist", name, err)
		}
	}
}
