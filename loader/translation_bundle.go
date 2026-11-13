package loader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"sync"

	"github.com/Vilsol/slox"
	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/raw"
)

// EncodeTranslationBundle is one language's stat description files, by name ("stat_descriptions"), as a
// file bundle of their tuple encodings.
func EncodeTranslationBundle(files map[string]*raw.TranslationFile) ([]byte, error) {
	encoded := make(map[string][]byte, len(files))
	for name, file := range files {
		b, err := file.MarshalMsg(nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode "+name)
		}
		encoded[name] = b
	}
	return EncodeFileBundle(encoded), nil
}

// TranslationBundleVariant is the variant of the translation bundles LoadTranslationBundle reads: the
// tuple layout of raw.TranslationFile.
func TranslationBundleVariant() string {
	return TupleVariant(raw.TranslationFile{})
}

// TranslationBundlePath is where language's bundle lives under its version directory, named by its layout
// and the hash of its EncodeTranslationBundle bytes: "stat_translations/en.t-c976d1d7-1a2b3c4d.msgpack.br".
func TranslationBundlePath(language string, encoded []byte) string {
	sum := sha256.Sum256(encoded)
	return AssetPath("stat_translations/"+language, TranslationBundleVariant()+"-"+hex.EncodeToString(sum[:4]))
}

var (
	translationBundles     = map[string]map[string][]byte{}
	translationBundlesLock sync.Mutex
)

// LoadTranslationBundle loads every stat description file language has in version, by name, each still
// tuple-encoded. A version whose manifest names no such bundle is ErrAssetNotFound. Bundles are kept
// decoded for the life of the process.
func LoadTranslationBundle(ctx context.Context, version string, language string, assetCache AssetCache) (map[string][]byte, error) {
	manifest, err := LoadManifest(ctx, version)
	if err != nil {
		return nil, err
	}

	path, ok := manifest.TranslationBundles[TranslationBundleVariant()][language]
	if !ok {
		return nil, errors.Wrap(ErrAssetNotFound, "no "+language+" translation bundle in "+version)
	}

	url := cdnBase + version + "/" + path

	translationBundlesLock.Lock()
	defer translationBundlesLock.Unlock()

	if files, ok := translationBundles[url]; ok {
		return files, nil
	}

	slox.From(ctx).Debug(
		"loading translation bundle",
		slog.String("version", version),
		slog.String("path", path),
	)

	b, fromCache, err := loadAsset(ctx, url, assetCache)
	if err != nil {
		return nil, err
	}

	files, err := decodeFileBundle(b)
	if err != nil && fromCache {
		slox.From(ctx).Warn(
			"discarding unreadable cached asset",
			slog.String("url", url),
			slog.Any("error", err),
		)

		b, err = fetchAsset(ctx, url, assetCache)
		if err != nil {
			return nil, err
		}

		files, err = decodeFileBundle(b)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to decode: "+url)
	}

	translationBundles[url] = files
	return files, nil
}

// LoadBundledTranslation is LoadTranslation read from language's translation bundle: one named stat
// description file ("stat_descriptions", or PoE2's "specific_skill_stat_descriptions/..."). A name the
// bundle does not have is ErrAssetNotFound.
func LoadBundledTranslation(ctx context.Context, version string, language string, name string, assetCache AssetCache) (*raw.TranslationFile, error) {
	files, err := LoadTranslationBundle(ctx, version, language, assetCache)
	if err != nil {
		return nil, err
	}

	b, ok := files[name]
	if !ok {
		return nil, errors.Wrap(ErrAssetNotFound, "no "+name+" in the "+language+" translation bundle of "+version)
	}

	var out raw.TranslationFile
	rest, err := out.UnmarshalMsg(b)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode "+name+" of the "+language+" translation bundle of "+version)
	}
	if len(rest) > 0 {
		return nil, errors.Errorf("%d trailing bytes after %s in the %s translation bundle of %s", len(rest), name, language, version)
	}

	return &out, nil
}
