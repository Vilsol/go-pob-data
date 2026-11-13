package exelayout

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "3.30.0.1.2", cacheName)
	l := &Layout{Patch: "3.30.0.1.2", Version: formatVersion, Tables: []Table{
		{Name: "Stats", RowSize: 105, Walked: true, Pointers: []Pointer{{Offset: 0, Kind: KindString}}},
	}}
	if err := writeCached(path, l); err != nil {
		t.Fatal(err)
	}
	got, err := readCached(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, l) {
		t.Fatalf("got %+v, want %+v", got, l)
	}

	l.Version = formatVersion - 1
	if err := writeCached(path, l); err != nil {
		t.Fatal(err)
	}
	if _, err := readCached(path); err == nil {
		t.Fatal("a layout from an older analysis was reused")
	}
}
