package main

import (
	"bytes"
	"encoding/binary"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/Vilsol/go-pob-data/extractor"
)

// storedBundle is a one-block bundle whose block is an Oodle stored (uncompressed) Kraken chunk.
func storedBundle(data []byte) []byte {
	var b bytes.Buffer
	le := binary.LittleEndian
	_ = binary.Write(&b, le, []uint32{uint32(len(data)), uint32(len(data) + 2), 52, 13, 0})
	_ = binary.Write(&b, le, []int64{int64(len(data)), int64(len(data) + 2)})
	_ = binary.Write(&b, le, []uint32{1, 0x40000, 0, 0, 0, 0, uint32(len(data) + 2)})
	b.Write([]byte{0xcc, 0x06})
	b.Write(data)
	return b.Bytes()
}

func TestSidecarLoader(t *testing.T) {
	ggpkPath := filepath.Join(t.TempDir(), "Content.ggpk")
	writeFile(t, ggpkPath, buildGGPK(t, 3, &ggpkEntry{children: []*ggpkEntry{
		{name: "Bundles2", children: []*ggpkEntry{
			{name: "Folders", children: []*ggpkEntry{{name: "x.bundle.bin", data: storedBundle([]byte("hello world"))}}},
		}},
	}}))
	sidecarPath := filepath.Join(t.TempDir(), "sidecar.tsv")
	writeFile(t, sidecarPath, []byte("path\tbundle\toffset\tsize\tconfidence\tmethod\n"+
		"data/low.datc64\tFolders/x\t0\t5\tlow\tguess\n"+
		"data/world.datc64\tFolders/x\t6\t5\tmedium\tids\n"))

	t.Setenv(sidecarEnv, sidecarPath)
	source, err := openSource(extractor.PoE1, ggpkPath)
	if err != nil {
		t.Fatalf("a ggpk without an index was refused with a sidecar named: %v", err)
	}
	loader, err := bundleLoader(source, extractor.PoE1, "3.30")
	if err != nil {
		t.Fatal(err)
	}
	if b, err := fs.ReadFile(loader, "data/world.datc64"); err != nil || string(b) != "world" {
		t.Fatalf("medium entry: %q %v", b, err)
	}
	if _, err := fs.ReadFile(loader, "data/low.datc64"); err == nil {
		t.Fatal("a low entry was served at the default threshold")
	}

	t.Setenv(sidecarMinEnv, "low")
	loader, err = bundleLoader(source, extractor.PoE1, "3.30")
	if err != nil {
		t.Fatal(err)
	}
	if b, err := fs.ReadFile(loader, "data/low.datc64"); err != nil || string(b) != "hello" {
		t.Fatalf("low entry at threshold low: %q %v", b, err)
	}

	t.Setenv(sidecarMinEnv, "certain")
	if _, err := bundleLoader(source, extractor.PoE1, "3.30"); err == nil {
		t.Fatal("an unknown threshold was accepted")
	}
}
