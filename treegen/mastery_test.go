package treegen

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"unicode/utf16"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/extractor"
)

// datc64 builds a .datc64 table: fixed-width rows, then the variable data its strings and arrays point into.
type datc64 struct {
	fixed bytes.Buffer
	data  bytes.Buffer
	rows  int
}

func newDatc64() *datc64 {
	d := &datc64{}
	d.data.Write(bytes.Repeat([]byte{0xbb}, 8))
	return d
}

func (d *datc64) u64(v uint64) { _ = binary.Write(&d.fixed, binary.LittleEndian, v) }

func (d *datc64) str(s string) {
	d.u64(uint64(d.data.Len()))
	for _, c := range utf16.Encode([]rune(s)) {
		_ = binary.Write(&d.data, binary.LittleEndian, c)
	}
	d.data.Write(make([]byte, 4))
}

func (d *datc64) row(i uint64) { d.u64(i); d.u64(0) }

func (d *datc64) nullRow() { d.fixed.Write(bytes.Repeat([]byte{0xfe}, 16)) }

func (d *datc64) rowArray(rows []int) {
	d.u64(uint64(len(rows)))
	d.u64(uint64(d.data.Len()))
	for _, r := range rows {
		_ = binary.Write(&d.data, binary.LittleEndian, [2]uint64{uint64(r), 0})
	}
}

func (d *datc64) bool(b bool) {
	if b {
		d.fixed.WriteByte(1)
	} else {
		d.fixed.WriteByte(0)
	}
}

func (d *datc64) bytes() []byte {
	var out bytes.Buffer
	_ = binary.Write(&out, binary.LittleEndian, uint32(d.rows))
	out.Write(d.fixed.Bytes())
	out.Write(d.data.Bytes())
	return out.Bytes()
}

func column(name, typ string, array bool, ref string) map[string]any {
	c := map[string]any{"name": name, "type": typ, "array": array}
	if name == "" {
		c["name"] = nil
	}
	if ref != "" {
		c["references"] = map[string]any{"table": ref}
	}
	return c
}

// masterySchema is dat-schema's PassiveSkillMasteryGroups up to 3.29 or, with art, PoE2's layout that 3.30
// uses: the three art strings moved to a PassiveSkillTreeMasteryArt row.
func masterySchema(art bool) map[string]any {
	groups := []any{column("Id", "string", false, ""), column("MasteryEffects", "foreignrow", true, "PassiveSkillMasteryEffects")}
	if !art {
		groups = append(groups, column("InactiveIcon", "string", false, ""), column("ActiveIcon", "string", false, ""), column("ActiveEffectImage", "string", false, ""))
	}
	groups = append(groups, column("", "bool", false, ""), column("SoundEffect", "foreignrow", false, "SoundEffects"), column("MasteryCountStat", "foreignrow", false, "Stats"))
	tables := []any{map[string]any{"name": "PassiveSkillMasteryGroups", "validFor": 1, "columns": groups}}
	if art {
		groups = append(groups, column("Art", "foreignrow", false, "PassiveSkillTreeMasteryArt"))
		tables[0].(map[string]any)["columns"] = groups
		tables = append(tables, map[string]any{"name": "PassiveSkillTreeMasteryArt", "validFor": 1, "columns": []any{
			column("Id", "string", false, ""), column("InactiveIcon", "string", false, ""),
			column("ActiveIcon", "string", false, ""), column("ActiveEffectImage", "string", false, ""),
		}})
	}
	return map[string]any{"version": 1, "tables": tables}
}

// withSchema runs f with gameVersion's schema read from a data/ directory holding only schema.
func withSchema(t *testing.T, gameVersion string, schema map[string]any, f func()) {
	t.Helper()
	dir := t.TempDir()
	b, err := json.Marshal(schema)
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, os.MkdirAll(filepath.Join(dir, "data", gameVersion), 0o755))
	testza.AssertNoError(t, os.WriteFile(filepath.Join(dir, "data", gameVersion, "schema.min.json"), b, 0o644))
	wd, err := os.Getwd()
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, os.Chdir(dir))
	defer func() { testza.AssertNoError(t, os.Chdir(wd)) }()
	extractor.LoadParser(extractor.PoE1, gameVersion)
	f()
}

func masteryGroups(t *testing.T, loader fs.FS, gameVersion string, art bool) []masteryGroup {
	t.Helper()
	var groups []masteryGroup
	withSchema(t, gameVersion, masterySchema(art), func() {
		l := &tableLoader{loader: loader, game: extractor.PoE1, gameVersion: gameVersion}
		loadMasteryGroups(l, &groups)
		testza.AssertNoError(t, l.err)
	})
	return groups
}

func TestMasteryArt(t *testing.T) {
	want := []masteryGroup{
		{InactiveIcon: "a/inactive.dds", ActiveIcon: "a/active.dds", ActiveEffectImage: "a/pattern", MasteryEffects: []int{3}},
		{InactiveIcon: "b/inactive.dds", ActiveIcon: "b/active.dds", ActiveEffectImage: "b/pattern", MasteryEffects: []int{4, 5}},
	}

	inline := newDatc64()
	for _, g := range want {
		inline.rows++
		inline.str("Group")
		inline.rowArray(g.MasteryEffects)
		inline.str(g.InactiveIcon)
		inline.str(g.ActiveIcon)
		inline.str(g.ActiveEffectImage)
		inline.bool(true)
		inline.nullRow()
		inline.row(7)
	}
	got := masteryGroups(t, fstest.MapFS{"data/passiveskillmasterygroups.datc64": {Data: inline.bytes()}}, "3.29", false)
	testza.AssertEqual(t, want, got)

	// 3.30: the groups name their art rows in reverse, so a group must take its own row, not its index.
	groups, art := newDatc64(), newDatc64()
	for i, g := range want {
		groups.rows++
		groups.str("Group")
		groups.rowArray(g.MasteryEffects)
		groups.bool(true)
		groups.nullRow()
		groups.row(7)
		groups.row(uint64(len(want) - 1 - i))
	}
	for i := range want {
		g := want[len(want)-1-i]
		art.rows++
		art.str("Art")
		art.str(g.InactiveIcon)
		art.str(g.ActiveIcon)
		art.str(g.ActiveEffectImage)
	}
	got = masteryGroups(t, fstest.MapFS{
		"data/passiveskillmasterygroups.datc64":  {Data: groups.bytes()},
		"data/passiveskilltreemasteryart.datc64": {Data: art.bytes()},
	}, "3.30", true)
	testza.AssertEqual(t, want, got)
}

// TestRecoveredMasteryArt reads 3.30's PassiveSkillMasteryGroups and PassiveSkillTreeMasteryArt under
// GOPOB_330_SCRATCH, a directory holding game files at their game paths.
func TestRecoveredMasteryArt(t *testing.T) {
	dir := os.Getenv("GOPOB_330_SCRATCH")
	if dir == "" {
		t.Skip("GOPOB_330_SCRATCH is not set")
	}
	groups := masteryGroups(t, os.DirFS(dir), "3.30", true)
	testza.AssertGreater(t, len(groups), 0)
	for _, g := range groups {
		testza.AssertTrue(t, filepath.Ext(g.InactiveIcon) == ".dds" && filepath.Ext(g.ActiveIcon) == ".dds" && g.ActiveEffectImage != "", g)
	}
	t.Logf("%d mastery groups, first %+v", len(groups), groups[0])
}
