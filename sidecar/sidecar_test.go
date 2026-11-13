package sidecar

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"unicode/utf16"

	"github.com/nfisher/gstream/hash/murmur2"
)

const granularity = 0x40000

// encodeStored builds a bundle of Oodle stored (uncompressed) Kraken blocks: enough to exercise the decoder
// without an encoder.
func encodeStored(data []byte) []byte {
	blocks := (len(data) + granularity - 1) / granularity
	var payload bytes.Buffer
	sizes := make([]uint32, blocks)
	for i := range blocks {
		chunk := data[i*granularity : min(len(data), (i+1)*granularity)]
		payload.Write([]byte{0xcc, 0x06})
		payload.Write(chunk)
		sizes[i] = uint32(len(chunk) + 2)
	}
	head := bundleHead{
		UncompressedSize:             uint32(len(data)),
		TotalPayloadSize:             uint32(payload.Len()),
		HeadPayloadSize:              uint32(48 + 4*blocks),
		FirstFileEncode:              13,
		UncompressedSize2:            int64(len(data)),
		TotalPayloadSize2:            int64(payload.Len()),
		BlockCount:                   uint32(blocks),
		UncompressedBlockGranularity: granularity,
	}
	var out bytes.Buffer
	_ = binary.Write(&out, binary.LittleEndian, head)
	_ = binary.Write(&out, binary.LittleEndian, sizes)
	out.Write(payload.Bytes())
	return out.Bytes()
}

func TestDecodeBundle(t *testing.T) {
	data := bytes.Repeat([]byte("abcdefg"), granularity/3)
	raw := encodeStored(data)
	got, err := DecodeBundle(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("decoded %d bytes, want %d", len(got), len(data))
	}
	if _, err := DecodeBundle(raw[:len(raw)-1]); err == nil {
		t.Fatal("a truncated bundle decoded")
	}
	if _, err := DecodeBundle(append(raw, 0)); err == nil {
		t.Fatal("a bundle with trailing bytes decoded")
	}
}

func TestEntriesRoundTrip(t *testing.T) {
	in := []Entry{
		{Path: "data/b.datc64", Bundle: "Folders/0/data.datc64", Offset: 10, Size: 5, Confidence: Low, Method: "shape\tonly"},
		{Path: "data/a.datc64", Bundle: "Folders/0/data.datc64", Offset: 0, Size: 10, Confidence: High, Method: "identical"},
	}
	var buf bytes.Buffer
	if err := WriteEntries(&buf, in); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[1], "data/a.datc64\t") {
		t.Fatalf("not sorted by path:\n%s", buf.String())
	}
	got, err := ReadEntries(&buf)
	if err != nil {
		t.Fatal(err)
	}
	want := []Entry{in[1], in[0]}
	want[1].Method = "shape only"
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip:\n got %+v\nwant %+v", got, want)
	}
	if _, err := ReadEntries(strings.NewReader("path\tbundle\n")); err == nil {
		t.Fatal("a foreign header was accepted")
	}
}

func TestFS(t *testing.T) {
	blob := []byte("0123456789abcdef")
	lower := fstest.MapFS{"Bundles2/Folders/x.bundle.bin": {Data: encodeStored(blob)}}
	f := NewFS(lower, []Entry{
		{Path: "Data/Medium.datc64", Bundle: "Folders/x", Offset: 2, Size: 4, Confidence: Medium},
		{Path: "data/low.datc64", Bundle: "Folders/x", Offset: 8, Size: 4, Confidence: Low},
	}, Medium)
	got, err := fs.ReadFile(f, "data/medium.datc64")
	if err != nil || string(got) != "2345" {
		t.Fatalf("medium entry: %q %v", got, err)
	}
	if _, err := f.Open("data/low.datc64"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("an entry below the threshold was served: %v", err)
	}
	h, _ := f.Open("data/medium.datc64")
	if st, _ := h.Stat(); st.Size() != 4 || st.Name() != "medium.datc64" {
		t.Fatalf("stat: %+v", st)
	}
}

// utf16z is s as the null-terminated UTF-16 a table's variable section holds.
func utf16z(s string) []byte {
	var b []byte
	for _, u := range utf16.Encode([]rune(s)) {
		b = binary.LittleEndian.AppendUint16(b, u)
	}
	return append(b, 0, 0, 0, 0)
}

// table builds a .datc64: rows of an Id string then a u32 value, padded with extra zero bytes per row.
func table(ids []string, values []uint32, extra int) []byte {
	w := 12 + extra
	fixed := make([]byte, 0, w*len(ids))
	varData := bytes.Repeat([]byte{0xbb}, 8)
	for i, id := range ids {
		row := make([]byte, w)
		binary.LittleEndian.PutUint64(row, uint64(len(varData)))
		binary.LittleEndian.PutUint32(row[8:], values[i])
		fixed = append(fixed, row...)
		varData = append(varData, utf16z(id)...)
	}
	return append(append(binary.LittleEndian.AppendUint32(nil, uint32(len(ids))), fixed...), varData...)
}

// plain builds a .datc64 with no strings: rows of two u32s.
func plain(rows int, seed uint32) []byte {
	b := binary.LittleEndian.AppendUint32(nil, uint32(rows))
	for i := range rows {
		b = binary.LittleEndian.AppendUint32(b, seed+uint32(i))
		b = binary.LittleEndian.AppendUint32(b, uint32(i%3))
	}
	return append(b, bytes.Repeat([]byte{0xbb}, 8)...)
}

func text(s string) []byte {
	b := []byte{0xff, 0xfe}
	for _, u := range utf16.Encode([]rune(s)) {
		b = binary.LittleEndian.AppendUint16(b, u)
	}
	return b
}

func ids(prefix string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = prefix + string(rune('A'+i%26)) + string(rune('a'+i/26))
	}
	return out
}

func vals(n int) []uint32 {
	out := make([]uint32, n)
	for i := range out {
		out[i] = uint32(i * 7)
	}
	return out
}

// The reference bundle holds a text file, a keyed table and a plain one; the target grew the keyed table (more
// rows, 4 bytes appended to each), kept the plain table, edited the text, and added an unknown table.
func TestRebuild(t *testing.T) {
	refText := text("description\n\t1 some_stat\n" + strings.Repeat("line\n", 40))
	refMods := table(ids("Mod", 40), vals(40), 0)
	refPlain := plain(30, 1000)
	refBlob := append(append(append([]byte{}, refText...), refMods...), refPlain...)

	tgtText := text("description\n\t1 some_stat\n" + strings.Repeat("line\n", 41))
	tgtMods := table(append(ids("Mod", 40), "ModNew1", "ModNew2"), vals(42), 4)
	newTable := table([]string{"Zz1", "Zz2", "Zz3"}, []uint32{1, 2, 3}, 9)
	tgtBlob := append(append(append(append([]byte{}, tgtText...), tgtMods...), refPlain...), newTable...)

	const b = "Folders/0/data.datc64"
	refFiles := []RefFile{
		{Path: "metadata/statdescriptions/x.txt", Bundle: b, Offset: 0, Size: int64(len(refText))},
		{Path: "data/mods.datc64", Bundle: b, Offset: int64(len(refText)), Size: int64(len(refMods))},
		{Path: "data/french/mods.datc64", Bundle: b, Offset: int64(len(refText)), Size: int64(len(refMods))},
		{Path: "data/plain.datc64", Bundle: b, Offset: int64(len(refText) + len(refMods)), Size: int64(len(refPlain))},
	}
	src := func(blob []byte) BundleSource {
		return func(name string) ([]byte, error) {
			if name != b {
				return nil, fs.ErrNotExist
			}
			return blob, nil
		}
	}
	cols := func(table string) []Column {
		if table == "mods" {
			return []Column{{Type: "string", Offset: 0, Size: 8}, {Type: "u32", Offset: 8, Size: 4}}
		}
		return nil
	}
	entries, spans, err := Rebuild(src(tgtBlob), src(refBlob), refFiles, []string{b}, Options{Columns: cols})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Entry{}
	for _, e := range entries {
		got[e.Path] = e
	}
	if _, ok := got["data/french/mods.datc64"]; ok {
		t.Error("a localised copy was named")
	}
	mods := got["data/mods.datc64"]
	if mods.Offset != int64(len(tgtText)) || mods.Size != int64(len(tgtMods)) || mods.Confidence != Medium || !strings.Contains(mods.Method, "width=16/12 ids=1.00 width changed") {
		t.Errorf("mods: %+v", mods)
	}
	pl := got["data/plain.datc64"]
	if pl.Offset != int64(len(tgtText)+len(tgtMods)) || pl.Size != int64(len(refPlain)) || pl.Confidence != High {
		t.Errorf("plain: %+v", pl)
	}
	tx := got["metadata/statdescriptions/x.txt"]
	if tx.Offset != 0 || tx.Size != int64(len(tgtText)) || tx.Confidence != Medium {
		t.Errorf("text: %+v", tx)
	}
	if len(entries) != 3 {
		t.Errorf("%d entries, want 3: %+v", len(entries), entries)
	}
	if len(spans) != 1 || spans[0].Offset != int64(len(tgtBlob)-len(newTable)) || spans[0].Size != int64(len(newTable)) {
		t.Errorf("spans: %+v", spans)
	}
}

func TestParseIndex(t *testing.T) {
	const path = "data/mods.datc64"
	firstRep := uint64(0x1234)
	a := (firstRep ^ (firstRep >> 47)) * 0x5F7A0EA7E59B19BD
	seed := a ^ (a >> 47)

	spec := binary.LittleEndian.AppendUint32(nil, 1)
	spec = append(append(spec, path...), 0)
	pathBundle := encodeStored(spec)

	var d []byte
	le := binary.LittleEndian
	d = le.AppendUint32(d, 1)
	name := "Folders/0/data.datc64"
	d = le.AppendUint32(d, uint32(len(name)))
	d = append(d, name...)
	d = le.AppendUint32(d, 999)
	d = le.AppendUint32(d, 1)
	d = le.AppendUint64(d, murmur2.Hash([]byte(path), seed))
	d = le.AppendUint32(d, 0)
	d = le.AppendUint32(d, 64)
	d = le.AppendUint32(d, 128)
	d = le.AppendUint32(d, 1)
	d = le.AppendUint64(d, firstRep)
	d = le.AppendUint32(d, 0)
	d = le.AppendUint32(d, uint32(len(spec)))
	d = le.AppendUint32(d, uint32(len(spec)))
	d = append(d, pathBundle...)

	files, err := ParseIndex(encodeStored(d))
	if err != nil {
		t.Fatal(err)
	}
	want := []RefFile{{Path: path, Bundle: name, Offset: 64, Size: 128}}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("got %+v, want %+v", files, want)
	}
	if _, err := ParseIndex(encodeStored(d[:20])); err == nil {
		t.Fatal("a truncated index parsed")
	}
}

// Two changed UTF-16 files without byte order marks are cut where the second one's first line begins.
func TestSplitAtFirstLines(t *testing.T) {
	u16 := func(s string) []byte { return text(s)[2:] }
	ref := append(u16("\"Art/A\" 1\n\"Art/B\" 2\n"), u16("\"Shop/X\" 1\n")...)
	run := []refGroup{{off: 0, size: int64(len(u16("\"Art/A\" 1\n\"Art/B\" 2\n")))}, {off: int64(len(u16("\"Art/A\" 1\n\"Art/B\" 2\n"))), size: int64(len(u16("\"Shop/X\" 1\n")))}}
	first := u16("\"Art/A\" 1\n\"Art/B\" 2\n\"Art/C\" 3\n")
	b := append(append([]byte{}, first...), u16("\"Shop/X\" 1\n\"Shop/Y\" 2\n")...)
	got := matchTextRun(b, 0, len(b), ref, run)
	if len(got) != 2 || got[0].b != len(first) || got[1].a != len(first) || got[1].b != len(b) {
		t.Fatalf("cuts: %+v", got)
	}
}
