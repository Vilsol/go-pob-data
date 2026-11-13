package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/fields"
	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
	st "github.com/Vilsol/go-pob-data/stat_translations"
)

const (
	staleSchema = "older dat schema than the struct (3.28+), re-encoding would drop its columns"
	// Neither its json nor its map asset decodes into the struct, so it has no tuple asset.
	undecodable = "older dat schema than the struct (3.28+), not even its map asset decodes"
)

var update = flag.Bool("update", false, "write tuple assets that are missing or do not match their source")

// Known mismatches between json.gz and the tuple asset. A staleSchema tuple asset was built from the
// table's map asset, since deleted, so frozenTuples pins it instead.
var knownStale = map[string]string{
	"3.18/AtlasNodeDefinition": undecodable,
	"3.24/AtlasNodeDefinition": undecodable,
	"3.25/AtlasNode":           staleSchema,
	"3.25/AtlasNodeDefinition": undecodable,
	"3.26/AtlasNode":           staleSchema,
	"3.26/AtlasNodeDefinition": undecodable,
	"3.27/AtlasNode":           staleSchema,
	"3.27/AtlasNodeDefinition": undecodable,
}

// frozenTuples are the sha256 of the staleSchema tuple assets.
var frozenTuples = map[string]string{
	"3.25/AtlasNode": "a045267de61d1d6a45b7f95f7ccf63f1791178a0d6bfdbfd7a34c38a21f0fd2b",
	"3.26/AtlasNode": "0b58ce5060677f35701d8d2242f72acc6197d0bc863ef7362753bf597e3b1317",
	"3.27/AtlasNode": "0b58ce5060677f35701d8d2242f72acc6197d0bc863ef7362753bf597e3b1317",
}

func readGzip(t *testing.T, path string) []byte {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}

	b, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

// readAsset returns nil for an asset that does not exist.
func readAsset(t *testing.T, path string) []byte {
	t.Helper()

	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	if bytes.HasPrefix(b, []byte("version https://git-lfs")) {
		t.Skip("git lfs pointer, run git lfs pull")
	}

	return b
}

// isMapAsset tells a frozen map asset from a tuple (.t-) or .min (.min-) variant.
func isMapAsset(path string) bool {
	return !strings.Contains(strings.TrimSuffix(filepath.Base(path), ".msgpack.br"), ".")
}

// Raw tables publish as json.gz and tuple assets only: no consumer reads a map asset or a json.br.
func TestRawAssetsTupleOnly(t *testing.T) {
	assets, err := filepath.Glob("data/*/raw/*.msgpack.br")
	if err != nil {
		t.Fatal(err)
	}
	jsons, err := filepath.Glob("data/*/raw/*.json.br")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range assets {
		if isMapAsset(path) {
			t.Errorf("%s is a map asset, the extractor should only write tuple assets", path)
		}
	}
	for _, path := range jsons {
		t.Errorf("%s is a json.br, the extractor should only write json.gz", path)
	}
}

// Every table has a tuple asset under the current layout of its row type, and it is a byte-identical
// re-encode of its json.gz (or, for knownStale, of its frozen map asset).
func TestTupleMatchesJSON(t *testing.T) {
	variants := map[extractor.Game]map[string]string{}
	for _, game := range []extractor.Game{extractor.PoE1, extractor.PoE2} {
		variants[game] = map[string]string{}
		for _, file := range tablesToExport(game) {
			if file.Encode != nil {
				variants[game][filepath.Base(file.path)] = file.Variant
			}
		}
	}
	translationVariant := loader.TupleVariant(raw.TranslationFile{})

	rawAssets, err := filepath.Glob("data/*/raw/*.t-*.msgpack.br")
	if err != nil {
		t.Fatal(err)
	}
	translationAssets, err := filepath.Glob("data/*/stat_translations/*/*.t-*.msgpack.br")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range append(rawAssets, translationAssets...) {
		name, variant, _ := strings.Cut(strings.TrimSuffix(filepath.Base(path), ".msgpack.br"), ".")
		want, ok := variants[extractor.GameOf(strings.Split(filepath.ToSlash(path), "/")[1])][name]
		if strings.Contains(path, "stat_translations") {
			want, ok = translationVariant, slices.Contains(st.Files, name) || slices.Contains(st.PoE2Files, name)
		}
		switch {
		case !ok:
			t.Errorf("%s is not a known asset", path)
		case variant != want:
			// Consumers of an older library version still read the layout they were built with.
			t.Logf("%s is an older layout than %s, kept for older consumers", path, want)
		}
	}

	dirs, err := filepath.Glob("data/*/raw")
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range dirs {
		version := filepath.Base(filepath.Dir(dir))
		for _, file := range tablesToExport(extractor.GameOf(version)) {
			if file.Encode == nil {
				continue
			}
			name := filepath.Base(file.path)
			key := version + "/" + name
			t.Run(key, func(t *testing.T) {
				jsonPath := filepath.Join(dir, name+".json.gz")
				if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
					t.Skip("not extracted for this version")
				}

				tuplePath := filepath.Join("data", version, loader.AssetPath("raw/"+name, file.Variant))
				committed := readAsset(t, tuplePath)
				fromJSON, jsonErr := file.Encode(readGzip(t, jsonPath))

				if reason, stale := knownStale[key]; stale {
					if jsonErr == nil && bytes.Equal(fromJSON, committed) {
						t.Errorf("%s is no longer stale, remove it from knownStale", key)
					}
					if reason == undecodable {
						if committed != nil {
							t.Errorf("%s exists, but %s", tuplePath, reason)
						}
						t.Skip("known stale: " + reason)
					}

					if sum := sha256.Sum256(committed); hex.EncodeToString(sum[:]) != frozenTuples[key] {
						t.Errorf("%s changed, and its map asset source is gone", tuplePath)
					}
					t.Skip("known stale: " + reason)
				} else if jsonErr != nil {
					t.Fatal(jsonErr)
				}

				switch {
				case bytes.Equal(fromJSON, committed):
				case *update:
					if err := os.WriteFile(tuplePath, fromJSON, 0o755); err != nil {
						t.Fatal(err)
					}
				default:
					t.Errorf("%s is missing or stale: run with -update", tuplePath)
				}
			})
		}
	}
}

// fields names a list's assets and their row types on its own, so it must agree with what is extracted.
func TestFieldsKnowEveryAsset(t *testing.T) {
	for _, file := range filesToExport {
		name := filepath.Base(file.path)
		rowType, ok := fields.RowType("raw/" + name)
		if !ok {
			t.Errorf("fields does not know raw/%s", name)
			continue
		}
		if variant := loader.TupleVariant(reflect.New(rowType).Interface()); variant != file.Variant {
			t.Errorf("raw/%s: fields decodes it in %s, the extractor writes %s", name, variant, file.Variant)
		}
	}

	for _, name := range st.Files {
		if _, ok := fields.RowType("stat_translations/en/" + name); !ok {
			t.Errorf("fields does not know stat_translations/<lang>/%s", name)
		}
	}
}

// CannotGrantToMinion moved between unnamed columns across schema versions; every version carries it.
func TestStatsCannotGrantToMinion(t *testing.T) {
	jsons, err := filepath.Glob("data/*/raw/Stats.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range jsons {
		t.Run(path, func(t *testing.T) {
			// raw.Stat is a PoE1 layout.
			if extractor.GameOf(filepath.Base(filepath.Dir(filepath.Dir(path)))) == extractor.PoE2 {
				t.Skip("poe2")
			}
			var stats []raw.Stat
			if err := json.Unmarshal(readGzip(t, path), &stats); err != nil {
				t.Fatal(err)
			}
			flagged := 0
			for _, stat := range stats {
				if stat.CannotGrantToMinion {
					flagged++
				}
			}
			if flagged < 300 || flagged > 600 {
				t.Errorf("%d stats cannot be granted to minions", flagged)
			}
		})
	}
}
