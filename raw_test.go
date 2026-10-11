package main

import (
	"bufio"
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

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/fields"
	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
	st "github.com/Vilsol/go-pob-data/stat_translations"
	"github.com/Vilsol/go-pob-data/testdata/legacyraw"
)

const (
	staleSchema = "older dat schema than the struct (3.28+), re-encoding would drop its columns"
	// Neither its json nor its map asset decodes into the struct, so it has no tuple asset.
	undecodable = "older dat schema than the struct (3.28+), not even its map asset decodes"
)

var update = flag.Bool("update", false, "write tuple assets that are missing or do not match their source")

// Known mismatches between json.gz and the frozen map asset. Their tuple asset is built from the map
// asset instead, which is what consumers have been reading.
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

// 3.20 translations predate TranslationFile: a bare list of descriptors whose conditions cannot tell 0
// from absent, and no json. The current decoder already fails on them, so they have no tuple asset.
const staleTranslations = "3.20"

// 3.29 was re-extracted from patch 3.29.3.3, and its translations by the strict stat description parser,
// after its map assets were frozen, so its assets no longer decode to the rows of their map assets.
const reextracted = "3.29"

// rowCodecs decodes a table's tuple asset and its frozen map asset to the json of their rows. The tuple
// rows are projected onto the legacy row type, so columns added since the map assets froze are not compared.
type rowCodecs struct {
	tuple  func(asset []byte) ([]byte, error)
	legacy func(asset []byte) ([]byte, error)
}

// Legacy columns that are not compared. The schema renamed them, so their map assets missed their value
// (SkillGems' Str/Dex/Int from 3.24, Mods' BuffTemplate from 3.28), or the row type now carries them
// under the current schema names (DefaultMonsterStats, whose pre-3.28 Armour was AilmentThreshold, and
// the map difficulty tables); TestTupleMatchesJSON still checks them against the json.gz.
var correctedColumns = map[string][]string{
	"SkillGems":                {"Str", "Dex", "Int"},
	"Mods":                     {"BuffTemplate"},
	"DefaultMonsterStats":      {"Life", "AllyLife", "Damage2", "Armour"},
	"MonsterMapDifficulty":     {"StatsKey1", "StatsKey2", "StatsKey3", "StatsKey4", "Stat1Value", "Stat2Value"},
	"MonsterMapBossDifficulty": {"StatsKey1", "StatsKey2", "StatsKey3", "StatsKey4", "StatsKey5", "Stat1Value", "Stat2Value", "Stat3Value", "Stat4Value", "Stat5Value"},
}

func codecs[T, L any, PT interface {
	*T
	msgp.Decodable
}, PL interface {
	*L
	msgp.Decodable
}]() rowCodecs {
	return rowCodecs{tuple: asLegacy[L](rowsJSON[T, PT]), legacy: rowsJSON[L, PL]}
}

func asLegacy[L any](decode func(asset []byte) ([]byte, error)) func(asset []byte) ([]byte, error) {
	return func(asset []byte) ([]byte, error) {
		b, err := decode(asset)
		if err != nil {
			return nil, err
		}

		var rows []L
		if err := json.Unmarshal(b, &rows); err != nil {
			return nil, errors.Wrap(err, "failed to project rows")
		}

		b, err = json.Marshal(rows)
		return b, errors.Wrap(err, "failed to marshal rows")
	}
}

func withoutColumns(t *testing.T, rowsJSON []byte, columns []string) []byte {
	t.Helper()

	var rows []map[string]any
	if err := json.Unmarshal(rowsJSON, &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		for _, column := range columns {
			delete(row, column)
		}
	}

	b, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

var tables = map[string]rowCodecs{
	"ActiveSkillType":                   codecs[raw.ActiveSkillType, legacyraw.ActiveSkillType](),
	"ActiveSkills":                      codecs[raw.ActiveSkill, legacyraw.ActiveSkill](),
	"AlternatePassiveAdditions":         codecs[raw.AlternatePassiveAddition, legacyraw.AlternatePassiveAddition](),
	"AlternatePassiveSkills":            codecs[raw.AlternatePassiveSkill, legacyraw.AlternatePassiveSkill](),
	"AlternateTreeVersions":             codecs[raw.AlternateTreeVersion, legacyraw.AlternateTreeVersion](),
	"ArmourTypes":                       codecs[raw.ArmourType, legacyraw.ArmourType](),
	"AtlasNode":                         codecs[raw.AtlasNode, legacyraw.AtlasNode](),
	"AtlasNodeDefinition":               codecs[raw.AtlasNodeDefinition, legacyraw.AtlasNodeDefinition](),
	"BaseItemTypes":                     codecs[raw.BaseItemType, legacyraw.BaseItemType](),
	"ComponentAttributeRequirements":    codecs[raw.ComponentAttributeRequirement, legacyraw.ComponentAttributeRequirement](),
	"ComponentCharges":                  codecs[raw.ComponentCharge, legacyraw.ComponentCharge](),
	"CostTypes":                         codecs[raw.CostType, legacyraw.CostType](),
	"CraftingBenchOptions":              codecs[raw.CraftingBenchOption, legacyraw.CraftingBenchOption](),
	"DefaultMonsterStats":               codecs[raw.DefaultMonsterStat, legacyraw.DefaultMonsterStat](),
	"Essences":                          codecs[raw.Essence, legacyraw.Essence](),
	"Flasks":                            codecs[raw.Flask, legacyraw.Flask](),
	"GrantedEffectQualityStats":         codecs[raw.GrantedEffectQualityStat, legacyraw.GrantedEffectQualityStat](),
	"GrantedEffectStatSets":             codecs[raw.GrantedEffectStatSet, legacyraw.GrantedEffectStatSet](),
	"GrantedEffectStatSetsPerLevel":     codecs[raw.GrantedEffectStatSetsPerLevel, legacyraw.GrantedEffectStatSetsPerLevel](),
	"GrantedEffects":                    codecs[raw.GrantedEffect, legacyraw.GrantedEffect](),
	"GrantedEffectsPerLevel":            codecs[raw.GrantedEffectsPerLevel, legacyraw.GrantedEffectsPerLevel](),
	"ItemClasses":                       codecs[raw.ItemClass, legacyraw.ItemClass](),
	"ItemExperiencePerLevel":            codecs[raw.ItemExperiencePerLevel, legacyraw.ItemExperiencePerLevel](),
	"Mods":                              codecs[raw.Mod, legacyraw.Mod](),
	"MonsterMapBossDifficulty":          codecs[raw.MonsterMapBossDifficulty, legacyraw.MonsterMapBossDifficulty](),
	"MonsterMapDifficulty":              codecs[raw.MonsterMapDifficulty, legacyraw.MonsterMapDifficulty](),
	"MonsterVarieties":                  codecs[raw.MonsterVariety, legacyraw.MonsterVariety](),
	"PantheonPanelLayout":               codecs[raw.PantheonPanelLayout, legacyraw.PantheonPanelLayout](),
	"PassiveSkills":                     codecs[raw.PassiveSkill, legacyraw.PassiveSkill](),
	"PassiveTreeExpansionJewels":        codecs[raw.PassiveTreeExpansionJewel, legacyraw.PassiveTreeExpansionJewel](),
	"PassiveTreeExpansionSkills":        codecs[raw.PassiveTreeExpansionSkill, legacyraw.PassiveTreeExpansionSkill](),
	"PassiveTreeExpansionSpecialSkills": codecs[raw.PassiveTreeExpansionSpecialSkill, legacyraw.PassiveTreeExpansionSpecialSkill](),
	"ShieldTypes":                       codecs[raw.ShieldType, legacyraw.ShieldType](),
	"SkillGems":                         codecs[raw.SkillGem, legacyraw.SkillGem](),
	"SkillTotemVariations":              codecs[raw.SkillTotemVariation, legacyraw.SkillTotemVariation](),
	"SkillTotems":                       codecs[raw.SkillTotem, legacyraw.SkillTotem](),
	"Stats":                             codecs[raw.Stat, legacyraw.Stat](),
	"Tags":                              codecs[raw.Tag, legacyraw.Tag](),
	"WeaponTypes":                       codecs[raw.WeaponType, legacyraw.WeaponType](),
}

var translationCodecs = rowCodecs{
	tuple:  fileJSON[raw.TranslationFile],
	legacy: fileJSON[legacyraw.TranslationFile],
}

// rowsJSON decodes an asset the way loader.LoadRaw does: a count, exactly that many rows, then EOF.
func rowsJSON[T any, PT interface {
	*T
	msgp.Decodable
}](asset []byte) ([]byte, error) {
	r := msgp.NewReader(brotli.NewReader(bytes.NewReader(asset)))
	n, err := r.ReadInt64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read row count")
	}

	rows := make([]T, n)
	for i := range rows {
		if err := PT(&rows[i]).DecodeMsg(r); err != nil {
			return nil, errors.Wrap(err, "failed to decode row")
		}
	}

	if _, err := r.R.ReadByte(); !errors.Is(err, io.EOF) {
		return nil, errors.New("trailing bytes after the last row")
	}

	b, err := json.Marshal(rows)
	return b, errors.Wrap(err, "failed to marshal rows")
}

func fileJSON[T any, PT interface {
	*T
	msgp.Decodable
}](asset []byte) ([]byte, error) {
	var file T
	r := msgp.NewReader(brotli.NewReader(bytes.NewReader(asset)))
	if err := PT(&file).DecodeMsg(r); err != nil {
		return nil, errors.Wrap(err, "failed to decode file")
	}

	if _, err := r.R.ReadByte(); !errors.Is(err, io.EOF) {
		return nil, errors.New("trailing bytes after the file")
	}

	b, err := json.Marshal(file)
	return b, errors.Wrap(err, "failed to marshal file")
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

func encodeTranslation(t *testing.T, jsonData []byte) []byte {
	t.Helper()

	var file raw.TranslationFile
	if err := json.Unmarshal(jsonData, &file); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	bw := brotli.NewWriter(&buf)
	w := msgp.NewWriter(bw)
	if err := file.EncodeMsg(w); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := bw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

// The map assets are frozen: nothing re-encodes them any more, so only this list vouches for them.
func TestLegacyAssetsPinned(t *testing.T) {
	f, err := os.Open("testdata/legacy-assets.sha256")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var pinned []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sum, path, _ := strings.Cut(scanner.Text(), "  ")
		pinned = append(pinned, path)

		b := readAsset(t, filepath.Join("data", path))
		if b == nil {
			t.Errorf("%s is gone, but consumers of older versions still read it", path)
			continue
		}
		if got := sha256.Sum256(b); hex.EncodeToString(got[:]) != sum {
			t.Errorf("%s changed", path)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	onDisk, err := filepath.Glob("data/*/raw/*.msgpack.br")
	if err != nil {
		t.Fatal(err)
	}
	translationAssets, err := filepath.Glob("data/*/stat_translations/*/*.msgpack.br")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range append(onDisk, translationAssets...) {
		path = strings.TrimPrefix(filepath.ToSlash(path), "data/")
		if isMapAsset(path) && !slices.Contains(pinned, path) {
			t.Errorf("%s is a new map asset, the extractor should only write tuple assets", path)
		}
	}
}

// Every table has a tuple asset under the current layout of its row type, and it is a byte-identical
// re-encode of its json.gz (or, for knownStale, of its frozen map asset).
func TestTupleMatchesJSON(t *testing.T) {
	variants := map[string]string{}
	for _, file := range filesToExport {
		variants[filepath.Base(file.path)] = file.Variant
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
		want, ok := variants[name]
		if strings.Contains(path, "stat_translations") {
			want, ok = translationVariant, slices.Contains(st.Files, name)
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
		for _, file := range filesToExport {
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

				want := fromJSON
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

					rows, err := tables[name].legacy(readAsset(t, filepath.Join(dir, name+".msgpack.br")))
					if err != nil {
						t.Fatal(err)
					}
					if want, err = file.Encode(rows); err != nil {
						t.Fatal(err)
					}
				} else if jsonErr != nil {
					t.Fatal(jsonErr)
				}

				switch {
				case bytes.Equal(want, committed):
				case *update:
					if err := os.WriteFile(tuplePath, want, 0o755); err != nil {
						t.Fatal(err)
					}
				default:
					t.Errorf("%s is missing or stale: run with -update", tuplePath)
				}
			})
		}

		for _, lang := range langDirs(t, version) {
			for _, name := range st.Files {
				t.Run(version+"/"+lang+"/"+name, func(t *testing.T) {
					base := filepath.Join("data", version, "stat_translations", lang, name)
					tuplePath := loader.AssetPath(base, translationVariant)
					if version == staleTranslations {
						if readAsset(t, tuplePath) != nil {
							t.Errorf("%s exists, but its source is not a TranslationFile", tuplePath)
						}
						t.Skip("known stale: no TranslationFile to encode")
					}

					if _, err := os.Stat(base + ".json.gz"); os.IsNotExist(err) {
						t.Skip("not extracted for this version")
					}

					source := encodeTranslation(t, readGzip(t, base+".json.gz"))
					committed := readAsset(t, tuplePath)

					switch {
					case bytes.Equal(source, committed):
					case *update:
						if err := os.WriteFile(tuplePath, source, 0o755); err != nil {
							t.Fatal(err)
						}
					default:
						t.Errorf("%s is missing or stale: run with -update", tuplePath)
					}
				})
			}
		}
	}
}

func langDirs(t *testing.T, version string) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("data", version, "stat_translations"))
	if err != nil {
		t.Fatal(err)
	}

	langs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			langs = append(langs, entry.Name())
		}
	}
	return langs
}

// Every tuple asset decodes to exactly the rows its frozen map asset decodes to with the map decoder
// consumers used until now (testdata/legacyraw), compared as json of the decoded values.
func TestTupleMatchesLegacy(t *testing.T) {
	legacyAssets, err := filepath.Glob("data/*/raw/*.msgpack.br")
	if err != nil {
		t.Fatal(err)
	}
	translationAssets, err := filepath.Glob("data/*/stat_translations/*/*.msgpack.br")
	if err != nil {
		t.Fatal(err)
	}

	variants := map[string]string{}
	for _, file := range filesToExport {
		variants[filepath.Base(file.path)] = file.Variant
	}

	for _, legacyPath := range append(legacyAssets, translationAssets...) {
		if !isMapAsset(legacyPath) {
			continue
		}

		t.Run(strings.TrimPrefix(legacyPath, "data/"), func(t *testing.T) {
			base := strings.TrimSuffix(legacyPath, ".msgpack.br")
			name := filepath.Base(base)

			version := strings.Split(filepath.ToSlash(legacyPath), "/")[1]

			if version == reextracted {
				t.Skip("re-extracted after its map asset was frozen")
			}

			codec, variant := tables[name], variants[name]
			if strings.Contains(legacyPath, "stat_translations") {
				if version == staleTranslations {
					t.Skip("known stale: no TranslationFile to decode")
				}
				codec, variant = translationCodecs, loader.TupleVariant(raw.TranslationFile{})
			} else if knownStale[version+"/"+name] == undecodable {
				t.Skip("known stale: " + undecodable)
			}
			if codec.tuple == nil {
				t.Fatalf("no codecs for %s", name)
			}

			tuple := readAsset(t, loader.AssetPath(base, variant))
			if tuple == nil {
				t.Fatalf("no tuple asset for %s", legacyPath)
			}

			want, err := codec.legacy(readAsset(t, legacyPath))
			if err != nil {
				t.Fatal(err)
			}
			got, err := codec.tuple(tuple)
			if err != nil {
				t.Fatal(err)
			}

			if columns, ok := correctedColumns[name]; ok {
				got, want = withoutColumns(t, got, columns), withoutColumns(t, want, columns)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("tuple asset decodes to different rows than %s", legacyPath)
			}
		})
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
