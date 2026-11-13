package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/Vilsol/go-pob-data/extractor"
)

type ggpkEntry struct {
	name     string
	data     []byte
	children []*ggpkEntry
}

// buildGGPK writes a GGPK record stream: the GGPK header, then every FILE/PDIR after its children so
// each PDIR knows its children's offsets.
func buildGGPK(t *testing.T, version int32, root *ggpkEntry) []byte {
	t.Helper()

	var body bytes.Buffer
	const headerSize = 28

	var write func(e *ggpkEntry) int64
	write = func(e *ggpkEntry) int64 {
		type child struct {
			hash   uint32
			offset int64
		}
		children := make([]child, 0, len(e.children))
		for _, c := range e.children {
			children = append(children, child{murmur32UTF16(c.name), write(c)})
		}

		name := utf16.Encode([]rune(e.name + "\x00"))
		var rec bytes.Buffer
		le := func(v any) { _ = binary.Write(&rec, binary.LittleEndian, v) }
		if e.children == nil {
			le(uint32(len(name)))
			rec.Write(make([]byte, 32))
			le(name)
			rec.Write(e.data)
		} else {
			le(uint32(len(name)))
			le(uint32(len(children)))
			rec.Write(make([]byte, 32))
			le(name)
			for _, c := range children {
				le(c.hash)
				le(c.offset)
			}
		}

		tag := "FILE"
		if e.children != nil {
			tag = "PDIR"
		}
		offset := headerSize + int64(body.Len())
		_ = binary.Write(&body, binary.LittleEndian, uint32(8+rec.Len()))
		body.WriteString(tag)
		body.Write(rec.Bytes())
		return offset
	}
	rootOffset := write(root)

	var out bytes.Buffer
	le := func(v any) { _ = binary.Write(&out, binary.LittleEndian, v) }
	le(uint32(headerSize))
	out.WriteString("GGPK")
	le(version)
	le(rootOffset)
	le(int64(0))
	out.Write(body.Bytes())
	return out.Bytes()
}

// murmur32UTF16 is pogo's PDIR child-name hash (ggpk.murmur32_utf16, unexported).
func murmur32UTF16(name string) uint32 {
	codepoints := utf16.Encode([]rune(strings.ToLower(name)))

	m := uint32(0x5bd1e995)
	h := uint32(len(codepoints) * 2)
	for i := 0; i+1 < len(codepoints); i += 2 {
		k := uint32(codepoints[i]) + uint32(codepoints[i+1])<<16
		k *= m
		k ^= k >> 24
		k *= m
		h *= m
		h ^= k
	}
	if len(codepoints)%2 == 1 {
		h ^= uint32(codepoints[len(codepoints)-1])
		h *= m
	}
	h ^= h >> 13
	h *= m
	h ^= h >> 15
	return h
}

func bundledTree(index string) *ggpkEntry {
	return &ggpkEntry{children: []*ggpkEntry{
		{name: "Bundles2", children: []*ggpkEntry{
			{name: "Readme.txt", data: []byte("readme")},
			{name: "_.index.bin", data: []byte(index)},
		}},
	}}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func readIndex(t *testing.T, source fs.FS) string {
	t.Helper()
	b, err := fs.ReadFile(source, bundleIndex)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestOpenSource(t *testing.T) {
	ggpkV3 := buildGGPK(t, 3, bundledTree("from ggpk"))

	t.Run("directory with bundles", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "Bundles2", "_.index.bin"), []byte("from dir"))
		// A directory with both keeps reading Bundles2/, as before.
		writeFile(t, filepath.Join(dir, ggpkName), ggpkV3)
		source, err := openSource(extractor.PoE1, dir)
		if err != nil {
			t.Fatal(err)
		}
		if got := readIndex(t, source); got != "from dir" {
			t.Fatalf("read %q", got)
		}
	})

	t.Run("ggpk file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "Content.ggpk")
		writeFile(t, path, ggpkV3)
		source, err := openSource(extractor.PoE1, path)
		if err != nil {
			t.Fatal(err)
		}
		if got := readIndex(t, source); got != "from ggpk" {
			t.Fatalf("read %q", got)
		}
	})

	t.Run("directory with ggpk", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, ggpkName), ggpkV3)
		source, err := openSource(extractor.PoE1, dir)
		if err != nil {
			t.Fatal(err)
		}
		if got := readIndex(t, source); got != "from ggpk" {
			t.Fatalf("read %q", got)
		}
	})

	t.Run("cdn", func(t *testing.T) {
		source, err := openSource(extractor.PoE1, "cdn:3.29.3.3")
		if err != nil {
			t.Fatal(err)
		}
		web, ok := source.(*extractor.WebFS)
		if !ok || web.BasePath != "https://patch.poecdn.com/3.29.3.3/" {
			t.Fatalf("got %#v", source)
		}
	})

	refusals := []struct {
		name string
		want string
		data []byte
	}{
		{"pre-bundle ggpk", "predates bundles", buildGGPK(t, 2, bundledTree("old"))},
		{"ggpk without bundles", "has no Bundles2/_.index.bin", buildGGPK(t, 3, &ggpkEntry{children: []*ggpkEntry{{name: "Data", children: []*ggpkEntry{}}}})},
		{"not a ggpk", "could not read ggpk", []byte(strings.Repeat("x", 64))},
	}
	for _, r := range refusals {
		t.Run(r.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "Content.ggpk")
			writeFile(t, path, r.data)
			_, err := openSource(extractor.PoE1, path)
			if err == nil || !strings.Contains(err.Error(), r.want) {
				t.Fatalf("want error containing %q, got %v", r.want, err)
			}
		})
	}

	t.Run("directory with neither", func(t *testing.T) {
		_, err := openSource(extractor.PoE1, t.TempDir())
		if err == nil || !strings.Contains(err.Error(), "could not find") {
			t.Fatalf("got %v", err)
		}
	})
}

// GGPK_TEST=<path to a Content.ggpk> reads the bundle index and a table through the bundle loader.
func TestRealGGPK(t *testing.T) {
	path := os.Getenv("GGPK_TEST")
	if path == "" {
		t.Skip("GGPK_TEST is not set")
	}
	const version = "3.30"

	source, err := openSource(extractor.PoE1, path)
	if err != nil {
		t.Fatal(err)
	}
	if index := readIndex(t, source); len(index) == 0 {
		t.Fatal("empty bundle index")
	}

	bundles, err := extractor.GetBundleLoader(source, IsNewHash(extractor.PoE1, version))
	if err != nil {
		t.Fatal(err)
	}
	f, err := bundles.Open(PackedFile{path: "Data/BaseItemTypes"}.FilePath(extractor.PoE1, version))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 4 || binary.LittleEndian.Uint32(data) == 0 || !bytes.Contains(data, bytes.Repeat([]byte{0xbb}, 8)) {
		t.Fatalf("BaseItemTypes does not look like a table (%d bytes)", len(data))
	}
	t.Logf("BaseItemTypes: %d rows, %d bytes", binary.LittleEndian.Uint32(data), len(data))
}
