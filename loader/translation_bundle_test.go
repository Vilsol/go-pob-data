package loader

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/andybalholm/brotli"

	"github.com/Vilsol/go-pob-data/raw"
	"github.com/Vilsol/go-pob-data/testdata"
)

func TestTranslationBundleRoundTrip(t *testing.T) {
	files := map[string]*raw.TranslationFile{
		"stat_descriptions": {
			Descriptors: []*raw.StatTranslation{{Identifier: "a", IDs: []string{"a"}}},
			Includes:    []string{},
		},
		"specific_skill_stat_descriptions/x": {
			Descriptors: []*raw.StatTranslation{},
			Includes:    []string{"Data/StatDescriptions/stat_descriptions.csd"},
		},
	}
	encoded, err := EncodeTranslationBundle(files)
	testza.AssertNoError(t, err)

	path := TranslationBundlePath("de", encoded)
	testza.AssertTrue(t, strings.HasPrefix(path, "stat_translations/de."+TupleVariant(raw.TranslationFile{})+"-"), path)
	other, err := EncodeTranslationBundle(map[string]*raw.TranslationFile{"stat_descriptions": files["stat_descriptions"]})
	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, path, TranslationBundlePath("de", other))

	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	_, err = w.Write(encoded)
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, w.Close())

	got, err := decodeFileBundle(buf.Bytes())
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, len(files), len(got))
	for name, file := range files {
		want, err := file.MarshalMsg(nil)
		testza.AssertNoError(t, err)
		testza.AssertTrue(t, bytes.Equal(want, got[name]), name)
	}
}

// A bundled translation is the file LoadTranslation reads one at a time, and every language has one.
func TestBundledTranslationMatchesPerFile(t *testing.T) {
	cdn := testdata.ServeCommittedData(t)
	ctx := context.Background()

	for _, version := range []string{"0.5", "3.18", testdata.LatestDataVersion()} {
		for _, name := range []string{"stat_descriptions", "passive_skill_stat_descriptions"} {
			want, err := LoadTranslation(ctx, version, "en", name, FakeCache{})
			testza.AssertNoError(t, err)

			got, err := LoadBundledTranslation(ctx, version, "en", name, FakeCache{})
			testza.AssertNoError(t, err)
			wantBytes, err := want.MarshalMsg(nil)
			testza.AssertNoError(t, err)
			gotBytes, err := got.MarshalMsg(nil)
			testza.AssertNoError(t, err)
			testza.AssertTrue(t, bytes.Equal(wantBytes, gotBytes), version+" "+name)
		}

		manifest, err := LoadManifest(ctx, version)
		testza.AssertNoError(t, err)
		testza.AssertGreater(t, len(manifest.Languages), 1)
		for _, lang := range manifest.Languages {
			bundle, err := LoadTranslationBundle(ctx, version, lang, FakeCache{})
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, len(manifest.StatTranslations), len(bundle), version+" "+lang)

			file, err := LoadBundledTranslation(ctx, version, lang, "stat_descriptions", FakeCache{})
			testza.AssertNoError(t, err)
			testza.AssertGreater(t, len(file.Descriptors), 100, version+" "+lang)
		}
	}

	testza.AssertEqual(t, []string(nil), cdn.Misses())

	_, err := LoadBundledTranslation(ctx, "0.5", "en", "no_such_stat_descriptions", FakeCache{})
	testza.AssertErrorIs(t, err, ErrAssetNotFound)

	_, err = LoadBundledTranslation(ctx, "3.20", "en", "stat_descriptions", FakeCache{})
	testza.AssertErrorIs(t, err, ErrAssetNotFound, "3.20 has no TranslationFile to bundle")
}
