package exelayout

import (
	"encoding/hex"
	"slices"
	"strings"
	"testing"

	"golang.org/x/arch/x86/x86asm"

	"github.com/Vilsol/go-pob-data/extractor"
)

// code decodes hex machine code, one instruction per string, loaded at 0x1000.
func code(t *testing.T, lines ...string) []ins {
	t.Helper()
	b, err := hex.DecodeString(strings.ReplaceAll(strings.Join(lines, ""), " ", ""))
	if err != nil {
		t.Fatal(err)
	}
	insns := decodeAt(b, 0x1000)
	if len(insns) != len(lines) {
		t.Fatalf("decoded %d instructions from %d lines", len(insns), len(lines))
	}
	for _, in := range insns {
		if in.bad {
			t.Fatalf("undecodable byte at %#x", in.pc)
		}
		t.Log(x86asm.IntelSyntax(in.Inst, uint64(in.pc), nil))
	}
	return insns
}

func TestLoaderPattern(t *testing.T) {
	insns := code(t,
		"8b 41 10",                      // mov eax, [rcx+0x10]   ; the row count
		"48 6b d0 69",                   // imul rdx, rax, 0x69
		"49 b9 bb bb bb bb bb bb bb bb", // mov r9, 0xBBBBBBBBBBBBBBBB
		"4c 39 0c 1a",                   // cmp [rdx+rbx], r9
		"75 08",                         // jne
		"49 89 d0",                      // mov r8, rdx
		"e8 00 01 00 00",                // call fixup
	)
	size, fixup, ok := loaderPattern(insns)
	if !ok || size != 105 || fixup != 0x1000+0x1f+0x100 {
		t.Fatalf("got size %d fixup %#x ok %v, want 105 %#x", size, fixup, ok, 0x1000+0x1f+0x100)
	}
}

func TestLoaderPatternShiftedCount(t *testing.T) {
	insns := code(t,
		"8b 41 10",                      // mov eax, [rcx+0x10]
		"48 c1 e0 04",                   // shl rax, 4
		"49 b9 bb bb bb bb bb bb bb bb", // mov r9, 0xBBBBBBBBBBBBBBBB
		"4c 39 0c 18",                   // cmp [rax+rbx], r9
	)
	size, fixup, ok := loaderPattern(insns)
	if !ok || size != 16 || fixup != 0 {
		t.Fatalf("got size %d fixup %#x ok %v, want 16 and no fixup", size, fixup, ok)
	}
}

func TestLoaderPatternNeedsMarkerCheck(t *testing.T) {
	insns := code(t,
		"8b 41 10",                      // mov eax, [rcx+0x10]
		"48 6b d0 69",                   // imul rdx, rax, 0x69
		"49 b9 bb bb bb bb bb bb bb bb", // mov r9, 0xBBBBBBBBBBBBBBBB
		"4c 39 ca",                      // cmp rdx, r9   ; not a memory compare
	)
	if _, _, ok := loaderPattern(insns); ok {
		t.Fatal("matched without a compare against memory")
	}
}

func TestHelperScale(t *testing.T) {
	insns := code(t,
		"48 6b c2 69", // imul rax, rdx, 0x69
		"48 03 41 08", // add rax, [rcx+8]
		"c3",          // ret
	)
	if got := helperScale(insns); got != 105 {
		t.Fatalf("got %d, want 105", got)
	}
	insns = code(t,
		"48 8b 01",    // mov rax, [rcx]
		"48 8d 04 d0", // lea rax, [rax+rdx*8]
		"c3",          // ret
	)
	if got := helperScale(insns); got != 8 {
		t.Fatalf("got %d, want 8", got)
	}
}

func TestPointers(t *testing.T) {
	insns := code(t,
		"48 89 cb",    // mov rbx, rcx          ; the loop starts after this
		"48 8b 03",    // mov rax, [rbx]        ; string at 0: read and rewritten
		"48 01 d0",    // add rax, rdx
		"48 89 03",    // mov [rbx], rax
		"8b 43 08",    // mov eax, [rbx+8]      ; row at 8: 4-byte read, 8-byte write
		"48 89 43 08", // mov [rbx+8], rax
		"48 8b 43 10", // mov rax, [rbx+0x10]   ; key at 16: index read and rewritten, handle written
		"48 89 43 10", // mov [rbx+0x10], rax
		"48 89 4b 18", // mov [rbx+0x18], rcx
		"48 8b 43 20", // mov rax, [rbx+0x20]   ; list at 32: count read, offset read and rewritten
		"48 8b 4b 28", // mov rcx, [rbx+0x28]
		"48 01 d1",    // add rcx, rdx
		"48 89 4b 28", // mov [rbx+0x28], rcx
		"48 83 c3 30", // add rbx, 0x30         ; next row
		"48 39 fb",    // cmp rbx, rdi
		"75 d4",       // jne
		"c3",          // ret
	)
	got, ok := newFixups(nil, nil).pointers(insns, 0x30)
	if !ok {
		t.Fatal("row loop not found")
	}
	want := []Pointer{
		{Offset: 0, Kind: KindString},
		{Offset: 8, Kind: KindRow},
		{Offset: 16, Kind: KindKey},
		{Offset: 32, Kind: KindList},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestClassifyCell(t *testing.T) {
	for _, tc := range []struct {
		want    Kind
		c, next cell
	}{
		{KindKey, cell{read8: true, write8: true}, cell{write8: true}},
		{KindKey, cell{read4: true, write8: true}, cell{write8: true}},
		{KindList, cell{read8: true}, cell{read8: true, write8: true}},
		{KindString, cell{read8: true, write8: true}, cell{}},
		{KindRow, cell{read4: true, write8: true}, cell{}},
		{"", cell{read8: true}, cell{}},
	} {
		if got, _ := classifyCell(tc.c, &tc.next); got != tc.want {
			t.Errorf("%+v then %+v: got %q, want %q", tc.c, tc.next, got, tc.want)
		}
	}
}

func TestRowNames(t *testing.T) {
	wide := func(s string) []byte {
		out := make([]byte, 0, 2*len(s))
		for _, c := range []byte(s) {
			out = append(out, c, 0)
		}
		return out
	}
	d := slices.Concat([]byte{0, 0}, wide("Stats_Row"), []byte{0, 0}, wide("-ab_Row"), []byte{0, 0})
	got := rowNamesIn(d, 0x100)
	if len(got) != 2 {
		t.Fatalf("got %d strings, want 2", len(got))
	}
	names := make([]string, 0, len(got[0]))
	for _, c := range got[0] {
		names = append(names, c.name)
	}
	if !slices.Equal(names, []string{"Stats", "tats", "ats", "ts", "s"}) || got[0][0].rva != 0x102 {
		t.Errorf("got %v at %#x", names, got[0][0].rva)
	}
	if got[1][0].name != "ab" {
		t.Errorf("got %q, want ab", got[1][0].name)
	}
}

func TestTableFromPath(t *testing.T) {
	for in, want := range map[string]string{
		"Data/Stats.dat":           "Stats",
		"Data/Balance/Mods.dat":    "Mods",
		"Data/Stats.datc64":        "",
		"Art/Stats.dat":            "",
		"Data/.dat":                "",
		"Data/StatDescriptions.da": "",
	} {
		if got := tableFromPath(in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}

func TestCompare(t *testing.T) {
	name := func(s string) *string { return &s }
	st := extractor.Table{Name: "T", Columns: []extractor.Column{
		{Name: name("Id"), Type: extractor.TypeString},
		{Name: name("Stat"), Type: extractor.TypeForeignRow, References: &extractor.References{Table: "Stats"}},
		{Name: name("Values"), Type: extractor.TypeI32, Array: true},
		{Name: name("Text"), Type: extractor.TypeString},
		{Type: extractor.TypeI32, Interval: true},
	}}
	if w := RowSize(st); w != 56 {
		t.Fatalf("schema row is %d, want 56", w)
	}

	agree := Table{Name: "T", RowSize: 56, Walked: true, Pointers: []Pointer{
		{Offset: 0, Kind: KindString},
		{Offset: 8, Kind: KindKey, Target: "Stats", Proven: true},
		{Offset: 24, Kind: KindList, Element: 4},
		{Offset: 40, Kind: KindString},
	}}
	if d := Compare(agree, st); len(d) != 0 {
		t.Fatalf("got %v", d)
	}

	if d := Compare(Table{Name: "T", RowSize: 60}, st); len(d) != 1 || d[0].Message != "row is 60 bytes, schema has 56" {
		t.Fatalf("got %v", d)
	}

	differ := Table{Name: "T", RowSize: 56, Walked: true, Pointers: []Pointer{
		{Offset: 8, Kind: KindKey, Target: "Mods", Proven: true},
		{Offset: 24, Kind: KindList, Element: 16},
		{Offset: 44, Kind: KindRow},
	}}
	got := make([]string, 0, 4)
	for _, d := range Compare(differ, st) {
		got = append(got, d.String())
	}
	want := []string{
		"T: byte 8: client resolves Stat (foreignrow) against Mods, schema references Stats",
		"T: byte 24: client reads list (element 16), schema has Values ([i32])",
		"T: byte 44: client reads row, schema has Text (string) from byte 40",
		"T: byte 0: schema has Id (string), client fixes up nothing there",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
