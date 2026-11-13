package exelayout

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/Vilsol/go-pob-data/extractor"
)

// exeEnv names a local PathOfExile.exe of patch 3.30.0.1.2; the integration tests skip without it.
const exeEnv = "POB_EXE_3_30"

func analyzeEnv(t *testing.T) *Layout {
	t.Helper()
	path := os.Getenv(exeEnv)
	if path == "" {
		t.Skipf("set %s to a 3.30.0.1.2 PathOfExile.exe", exeEnv)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	l, err := Analyze(f)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestAnalyzeRowSizes(t *testing.T) {
	l := analyzeEnv(t)
	if len(l.Tables) != 794 {
		t.Errorf("found %d tables, want 794", len(l.Tables))
	}
	for name, want := range map[string]int{"Stats": 105, "AtlasNode": 262, "Mods": 692, "BaseItemTypes": 373, "ActiveSkills": 257} {
		got, ok := l.Table(name)
		if !ok {
			t.Errorf("%s not found", name)
			continue
		}
		if got.RowSize != want {
			t.Errorf("%s row is %d bytes, want %d", name, got.RowSize, want)
		}
		if !got.Walked || len(got.Pointers) == 0 {
			t.Errorf("%s fixup was not walked", name)
		}
	}
}

func TestSchemaDisagreements(t *testing.T) {
	l := analyzeEnv(t)
	raw, err := os.ReadFile("../data/3.30/schema.min.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema extractor.SchemaFile
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	tables := SchemaTables(&schema, extractor.PoE1)
	var got []string
	for _, ct := range l.Tables {
		if st, ok := tables[strings.ToLower(ct.Name)]; ok {
			for _, d := range Compare(ct, st) {
				got = append(got, d.String())
			}
		}
	}
	// None of these tables is exported. The three row sizes match the .datc64 files' own widths, so the
	// schema is behind; BetrayalUpgrades' byte 52 is a list of keys the schema reads as i32s.
	want := []string{
		"BetrayalUpgrades: byte 52: client reads list of 16-byte elements, schema has unnamed ([i32])",
		"DropPool: row is 54 bytes, schema has 32",
		"HideoutDoodads: row is 157 bytes, schema has 156",
		"MapDevices: row is 126 bytes, schema has 110",
	}
	if !slices.Equal(got, want) {
		t.Errorf("disagreements are\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
