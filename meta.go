package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/loader"
	st "github.com/Vilsol/go-pob-data/stat_translations"
)

type Metadata struct {
	Versions []string `json:"versions"`
}

// unlistedMarker in a version directory publishes that version's files and manifest but leaves it out of
// meta.json, for data that must not be offered as a version yet, such as a partial pre-launch extraction.
const unlistedMarker = ".unlisted"

func generateMeta() error {
	dir, err := os.ReadDir("data")
	if err != nil {
		return errors.Wrap(err, "failed reading data directory")
	}

	metadata := &Metadata{}
	for _, entry := range dir {
		if entry.IsDir() {
			if _, err := os.Stat(filepath.Join("data", entry.Name(), unlistedMarker)); err != nil {
				metadata.Versions = append(metadata.Versions, entry.Name())
			}

			manifest, err := manifestJSON(filepath.Join("data", entry.Name()))
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join("data", entry.Name(), "manifest.json"), manifest, 0o644); err != nil {
				return errors.Wrap(err, "failed writing manifest")
			}
		}
	}

	bs, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(err, "failed marshaling metadata")
	}

	return errors.Wrap(os.WriteFile("data/meta.json", bs, 0o644), "failed writing metadata")
}

// manifestJSON lists what versionDir publishes. Every table has a json.gz, whichever tuple layouts it also
// has; the translations are what the translation bundles hold.
func manifestJSON(versionDir string) ([]byte, error) {
	manifest := loader.Manifest{
		Tables:           []string{},
		StatTranslations: []string{},
		Languages:        []string{},
		Files:            []string{},
	}

	tables, err := filepath.Glob(filepath.Join(versionDir, "raw", "*.json.gz"))
	if err != nil {
		return nil, errors.Wrap(err, "failed listing tables")
	}
	for _, table := range tables {
		manifest.Tables = append(manifest.Tables, strings.TrimSuffix(filepath.Base(table), ".json.gz"))
	}

	tuples, err := filepath.Glob(filepath.Join(versionDir, "raw", "*.t-*.msgpack.br"))
	if err != nil {
		return nil, errors.Wrap(err, "failed listing tuple assets")
	}
	manifest.JSONOnly = len(tables) > 0 && len(tuples) == 0

	translationPaths, err := translationBundles(versionDir)
	if err != nil {
		return nil, err
	}
	translations, err := readTranslationBundles(versionDir)
	if err != nil {
		return nil, err
	}
	for _, language := range slices.Sorted(maps.Keys(translationPaths)) {
		if manifest.TranslationBundles == nil {
			manifest.TranslationBundles = map[string]map[string]string{loader.TranslationBundleVariant(): {}}
		}
		manifest.TranslationBundles[loader.TranslationBundleVariant()][language] = translationPaths[language][0]
		manifest.Languages = append(manifest.Languages, language)
	}
	manifest.StatTranslations = append(manifest.StatTranslations, slices.Sorted(maps.Keys(translations[st.PerFileLanguage]))...)

	filesDir := filepath.Join(versionDir, "files")
	err = filepath.WalkDir(filesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(filesDir, path)
		if err != nil {
			return errors.Wrap(err, "failed resolving file path")
		}
		manifest.Files = append(manifest.Files, strings.TrimSuffix(filepath.ToSlash(rel), ".br"))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "failed listing files")
	}

	slices.Sort(manifest.Files)

	bundles, err := filepath.Glob(filepath.Join(versionDir, loader.AssetPath("files", "*")))
	if err != nil {
		return nil, errors.Wrap(err, "failed listing file bundles")
	}
	for _, bundle := range bundles {
		name := filepath.Base(bundle)
		variant := strings.TrimSuffix(strings.TrimPrefix(name, "files."), ".msgpack.br")
		ext, _, _ := strings.Cut(variant, "-")
		if manifest.FileBundles == nil {
			manifest.FileBundles = map[string]string{}
		}
		manifest.FileBundles["."+ext] = name
	}

	derived, err := filepath.Glob(filepath.Join(versionDir, "derived", "*.msgpack.br"))
	if err != nil {
		return nil, errors.Wrap(err, "failed listing derived assets")
	}
	for _, asset := range derived {
		name, rest, _ := strings.Cut(filepath.Base(asset), ".")
		variant := rest[:strings.LastIndex(strings.TrimSuffix(rest, ".msgpack.br"), "-")]
		if manifest.Derived == nil {
			manifest.Derived = map[string]map[string]string{}
		}
		if manifest.Derived[name] == nil {
			manifest.Derived[name] = map[string]string{}
		}
		if previous, ok := manifest.Derived[name][variant]; ok {
			return nil, errors.Errorf("%s has two %s %s assets: %s and %s", versionDir, name, variant, previous, asset)
		}
		manifest.Derived[name][variant] = "derived/" + filepath.Base(asset)
	}

	manifest.Assets, err = manifestAssets(versionDir)
	if err != nil {
		return nil, err
	}

	b, err := json.MarshalIndent(manifest, "", "\t")
	return append(b, '\n'), errors.Wrap(err, "failed marshaling manifest")
}

var lfsOID = regexp.MustCompile(`(?m)^oid sha256:([0-9a-f]{64})$`)

// manifestAssets hashes every file under versionDir but the manifest and dotfiles such as unlistedMarker.
func manifestAssets(versionDir string) (map[string]loader.ManifestAsset, error) {
	assets := map[string]loader.ManifestAsset{}
	err := filepath.WalkDir(versionDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasPrefix(d.Name(), ".") || path == filepath.Join(versionDir, "manifest.json") {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "failed reading asset")
		}
		rel, err := filepath.Rel(versionDir, path)
		if err != nil {
			return errors.Wrap(err, "failed resolving asset path")
		}
		var asset loader.ManifestAsset
		if bytes.HasPrefix(b, []byte("version https://git-lfs")) {
			// An LFS pointer's oid is the content's sha256; a msgpack asset still needs its bytes for Records.
			oid := lfsOID.FindSubmatch(b)
			if oid == nil || strings.HasSuffix(path, ".msgpack.br") {
				return errors.Errorf("%s is a git lfs pointer, run git lfs pull", path)
			}
			assets[filepath.ToSlash(rel)] = loader.ManifestAsset{SHA256: string(oid[1])}
			return nil
		}
		sum := sha256.Sum256(b)
		asset.SHA256 = hex.EncodeToString(sum[:])
		if strings.HasSuffix(path, ".msgpack.br") {
			asset.Records = leadingCount(b)
		}
		assets[filepath.ToSlash(rel)] = asset
		return nil
	})
	return assets, errors.Wrap(err, "failed hashing assets")
}

// leadingCount is the int64 row count a tuple asset leads with, or the entry count of a bundle's map;
// nil for an asset that leads with neither, such as a single translation file.
func leadingCount(compressed []byte) *int64 {
	r := msgp.NewReader(brotli.NewReader(bytes.NewReader(compressed)))
	t, err := r.NextType()
	if err != nil {
		return nil
	}
	var n int64
	switch t {
	case msgp.IntType, msgp.UintType:
		n, err = r.ReadInt64()
	case msgp.MapType:
		var size uint32
		size, err = r.ReadMapHeader()
		n = int64(size)
	default:
		return nil
	}
	if err != nil {
		return nil
	}
	return &n
}

func IsNewHash(game extractor.Game, gameVersion string) bool {
	return extractor.IsNewHash(game, gameVersion)
}

func GetFileExtension(game extractor.Game, gameVersion string) string {
	return extractor.GetFileExtension(game, gameVersion)
}
