package extractor

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestGameGates(t *testing.T) {
	cases := []struct {
		game      Game
		version   string
		extension string
		tableDir  string
		cdn       string
		newHash   bool
	}{
		{PoE1, "3.21", "dat64", "Data/", "https://patch.poecdn.com/3.21/", false},
		{PoE1, "3.22", "dat64", "Data/", "https://patch.poecdn.com/3.22/", true},
		{PoE1, "3.24", "dat64", "Data/", "https://patch.poecdn.com/3.24/", true},
		{PoE1, "3.25", "datc64", "Data/", "https://patch.poecdn.com/3.25/", true},
		{PoE1, "3.29", "datc64", "Data/", "https://patch.poecdn.com/3.29/", true},
		// A minor below 22 reads as old hash and .dat64 for PoE1; PoE2 must not.
		{PoE2, "0.5", "datc64", "Data/Balance/", "https://patch-poe2.poecdn.com/0.5/", true},
		{PoE2, "0.1", "datc64", "Data/Balance/", "https://patch-poe2.poecdn.com/0.1/", true},
		{PoE2, "0.5.5.4.2", "datc64", "Data/Balance/", "https://patch-poe2.poecdn.com/0.5.5.4.2/", true},
	}
	for _, c := range cases {
		t.Run(string(c.game)+"/"+c.version, func(t *testing.T) {
			testza.AssertEqual(t, c.newHash, IsNewHash(c.game, c.version))
			testza.AssertEqual(t, c.extension, GetFileExtension(c.game, c.version))
			testza.AssertEqual(t, c.tableDir, c.game.TableDir())
			testza.AssertEqual(t, c.cdn, NewWebFS(c.game, c.version).BasePath)
		})
	}
}

func TestParseGame(t *testing.T) {
	for _, s := range []string{"poe1", "poe2"} {
		g, err := ParseGame(s)
		testza.AssertNoError(t, err)
		testza.AssertEqual(t, Game(s), g)
	}
	for _, s := range []string{"", "poe3", "PoE2", "3.29"} {
		_, err := ParseGame(s)
		testza.AssertNotNil(t, err, s)
	}
}

// Replies captured from both patch servers on 2026-10-15.
func TestParsePatchReply(t *testing.T) {
	cases := []struct {
		reply string
		patch string
	}{
		{
			"02" + zeros32 + "0028" + utf16le("https://patch-poe2.poecdn.com/0.5.5.4.2/") + "0028" + utf16le("https://patch-poe2.poecdn.com/0.5.5.4.2/"),
			"0.5.5.4.2",
		},
		{
			"02" + zeros32 + "0022" + utf16le("https://patch.poecdn.com/3.29.3.3/") + "0022" + utf16le("https://patch.poecdn.com/3.29.3.3/"),
			"3.29.3.3",
		},
	}
	for _, c := range cases {
		b, err := hex.DecodeString(c.reply)
		testza.AssertNoError(t, err)
		patch, err := parsePatchReply(b)
		testza.AssertNoError(t, err)
		testza.AssertEqual(t, c.patch, patch)
	}

	for _, bad := range []string{
		"02",
		"02" + zeros32 + "0000",
		"02" + zeros32 + "0028" + utf16le("https://patch"),
		"02" + zeros32 + "0001" + utf16le("/"),
	} {
		b, err := hex.DecodeString(bad)
		testza.AssertNoError(t, err)
		_, err = parsePatchReply(b)
		testza.AssertNotNil(t, err, bad)
	}
}

const zeros32 = "0000000000000000000000000000000000000000000000000000000000000000"

func utf16le(s string) string {
	out := make([]byte, 0, 2*len(s))
	for _, r := range s {
		out = append(out, byte(r), 0)
	}
	return hex.EncodeToString(out)
}

// dat-schema lists a table once per game it differs in (validFor 1 or 2) or once for both (3).
func TestLoadSchemaByGame(t *testing.T) {
	prevPath, prevFile, prevMap, prevSource := dataPath, schemaFile, tableMap, tableSource
	dataPath = t.TempDir()
	t.Cleanup(func() { dataPath, schemaFile, tableMap, tableSource = prevPath, prevFile, prevMap, prevSource })

	schema := `{"tables":[
		{"validFor":1,"name":"Split","columns":[{"name":"PoE1","type":"i32"}]},
		{"validFor":2,"name":"Split","columns":[{"name":"PoE2","type":"i32"}]},
		{"validFor":3,"name":"Shared","columns":[{"name":"Both","type":"i32"}]},
		{"validFor":1,"name":"OnlyPoE1","columns":[{"name":"A","type":"i32"}]},
		{"validFor":2,"name":"OnlyPoE2","columns":[{"name":"B","type":"i32"}]}
	]}`
	testza.AssertNoError(t, os.MkdirAll(filepath.Join(dataPath, "v"), 0o755))
	testza.AssertNoError(t, os.WriteFile(filepath.Join(dataPath, "v", "schema.min.json"), []byte(schema), 0o644))

	cases := []struct {
		game   Game
		split  string
		absent string
		only   string
	}{
		{PoE1, "PoE1", "OnlyPoE2", "OnlyPoE1"},
		{PoE2, "PoE2", "OnlyPoE1", "OnlyPoE2"},
	}
	for _, c := range cases {
		t.Run(string(c.game), func(t *testing.T) {
			LoadSchema(c.game, "v")
			testza.AssertEqual(t, c.split, *GetSchema("Split").Columns[0].Name)
			testza.AssertEqual(t, "Both", *GetSchema("Shared").Columns[0].Name)
			testza.AssertEqual(t, c.only, GetSchema(c.only).Name)
			testza.AssertEqual(t, "", GetSchema(c.absent).Name)
		})
	}
}

func TestLatestPatch(t *testing.T) {
	for _, game := range []Game{PoE1, PoE2} {
		patch, err := LatestPatch(game)
		testza.AssertNoError(t, err)
		testza.AssertTrue(t, patchVersion.MatchString(patch), patch)
	}
}

var patchVersion = regexp.MustCompile(`^\d+(\.\d+){3,}$`)
