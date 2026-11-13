package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/fields"
	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
	st "github.com/Vilsol/go-pob-data/stat_translations"
)

// writeTranslations publishes translations, by language directory then name, under versionDir: one bundle
// per language, and the files go-pob reads one at a time (st.PerFileRoots) as their own assets. Every other
// per-file translation asset under versionDir is removed.
func writeTranslations(versionDir string, translations map[string]map[string]*raw.TranslationFile) error {
	for _, lang := range slices.Sorted(maps.Keys(translations)) {
		if err := writeTranslationBundle(versionDir, lang, translations[lang]); err != nil {
			return err
		}
	}

	keep := perFileTranslations(versionDir, translations[st.PerFileLanguage])
	for _, name := range keep {
		path := filepath.Join(versionDir, "stat_translations", st.PerFileLanguage, filepath.FromSlash(name))
		if err := writeTranslationAsset(loader.AssetPath(path, loader.TranslationBundleVariant()), translations[st.PerFileLanguage][name]); err != nil {
			return err
		}
	}

	return pruneTranslations(versionDir, keptTranslationAssets(versionDir, keep))
}

// perFileTranslations are the names of en that go-pob reads one at a time: st.PerFileRoots and what they
// include.
func perFileTranslations(versionDir string, en map[string]*raw.TranslationFile) []string {
	return st.PerFileNames(st.PerFileRoots[string(extractor.GameOf(filepath.Base(versionDir)))], func(name string) ([]string, bool) {
		file, ok := en[name]
		if !ok {
			return nil, false
		}
		return file.Includes, true
	})
}

// keptTranslationAssets are the per-file assets of names versionDir keeps, relative to versionDir: the
// tuple asset, and the .min asset of a name the version's field list has.
func keptTranslationAssets(versionDir string, names []string) map[string]bool {
	version := filepath.Base(versionDir)
	keep := map[string]bool{}
	for _, name := range names {
		asset := "stat_translations/" + st.PerFileLanguage + "/" + name
		keep[loader.AssetPath(asset, loader.TranslationBundleVariant())] = true
		if variant, ok := fields.MinVariant(version, asset); ok {
			keep[loader.AssetPath(asset, variant)] = true
		}
	}
	return keep
}

func writeTranslationAsset(path string, file *raw.TranslationFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return errors.Wrap(err, "failed to make translation directories")
	}

	var buf bytes.Buffer
	bw := brotli.NewWriter(&buf)
	w := msgp.NewWriter(bw)
	if err := file.EncodeMsg(w); err != nil {
		return errors.Wrap(err, "failed to encode "+path)
	}
	if err := w.Flush(); err != nil {
		return errors.Wrap(err, "failed to encode "+path)
	}
	if err := bw.Close(); err != nil {
		return errors.Wrap(err, "failed to compress "+path)
	}

	return errors.Wrap(os.WriteFile(path, buf.Bytes(), 0o644), "failed to write "+path)
}

// translationBundle is lang's bundle of files, at its loader.TranslationBundlePath, compressed.
func translationBundle(lang string, files map[string]*raw.TranslationFile) (string, []byte, error) {
	encoded, err := loader.EncodeTranslationBundle(files)
	if err != nil {
		return "", nil, err
	}

	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	if _, err := w.Write(encoded); err != nil {
		return "", nil, errors.Wrap(err, "could not compress the "+lang+" translation bundle")
	}
	if err := w.Close(); err != nil {
		return "", nil, errors.Wrap(err, "could not compress the "+lang+" translation bundle")
	}
	return loader.TranslationBundlePath(lang, encoded), buf.Bytes(), nil
}

// translationBundles are versionDir's translation bundles of the current layout, by language, relative to
// versionDir.
func translationBundles(versionDir string) (map[string][]string, error) {
	paths, err := filepath.Glob(filepath.Join(versionDir, "stat_translations", "*."+loader.TranslationBundleVariant()+"-*.msgpack.br"))
	if err != nil {
		return nil, errors.Wrap(err, "failed listing translation bundles")
	}
	out := map[string][]string{}
	for _, path := range paths {
		lang, _, _ := strings.Cut(filepath.Base(path), ".")
		out[lang] = append(out[lang], "stat_translations/"+filepath.Base(path))
	}
	return out, nil
}

// writeTranslationBundle writes lang's bundle of files and removes the bundles of its layout it replaces.
func writeTranslationBundle(versionDir string, lang string, files map[string]*raw.TranslationFile) error {
	path, b, err := translationBundle(lang, files)
	if err != nil {
		return err
	}
	bundles, err := translationBundles(versionDir)
	if err != nil {
		return err
	}
	for _, old := range bundles[lang] {
		if old != path {
			if err := os.Remove(filepath.Join(versionDir, old)); err != nil {
				return errors.Wrap(err, "could not remove "+old)
			}
		}
	}
	return errors.Wrap(os.WriteFile(filepath.Join(versionDir, path), b, 0o644), "could not write "+path)
}

// readTranslationBundles reads versionDir's translation bundles of the current layout, by language then
// name. A language with more than one is an error.
func readTranslationBundles(versionDir string) (map[string]map[string]*raw.TranslationFile, error) {
	bundles, err := translationBundles(versionDir)
	if err != nil {
		return nil, err
	}

	out := map[string]map[string]*raw.TranslationFile{}
	for lang, paths := range bundles {
		if len(paths) != 1 {
			return nil, errors.Errorf("%s has %d %s translation bundles", versionDir, len(paths), lang)
		}
		b, err := readBrotliFile(filepath.Join(versionDir, paths[0]))
		if err != nil {
			return nil, err
		}
		n, b, err := msgp.ReadMapHeaderBytes(b)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read "+paths[0])
		}
		out[lang] = make(map[string]*raw.TranslationFile, n)
		for range n {
			var name string
			var entry []byte
			if name, b, err = msgp.ReadStringBytes(b); err != nil {
				return nil, errors.Wrap(err, "failed to read "+paths[0])
			}
			if entry, b, err = msgp.ReadBytesZC(b); err != nil {
				return nil, errors.Wrap(err, "failed to read "+name+" of "+paths[0])
			}
			var file raw.TranslationFile
			rest, err := file.UnmarshalMsg(entry)
			if err != nil || len(rest) > 0 {
				return nil, errors.Errorf("%s of %s does not decode: %v, %d trailing bytes", name, paths[0], err, len(rest))
			}
			out[lang][name] = &file
		}
		if len(b) > 0 {
			return nil, errors.Errorf("%s: %d trailing bytes", paths[0], len(b))
		}
	}
	return out, nil
}

// readPerFileTranslations reads every per-file translation versionDir has, by language then name: its tuple
// asset of the current layout, or else its json.gz. Older layouts and the frozen map assets are not read;
// every name of a version that has them also has one of the two.
func readPerFileTranslations(versionDir string) (map[string]map[string]*raw.TranslationFile, error) {
	root := filepath.Join(versionDir, "stat_translations")
	out := map[string]map[string]*raw.TranslationFile{}
	tupleSuffix := "." + loader.TranslationBundleVariant() + ".msgpack.br"
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return errors.Wrap(err, "failed resolving translation path")
		}
		lang, name, ok := strings.Cut(filepath.ToSlash(rel), "/")
		if !ok {
			return nil
		}

		var file raw.TranslationFile
		switch {
		case strings.HasSuffix(name, tupleSuffix):
			name = strings.TrimSuffix(name, tupleSuffix)
			b, err := readBrotliFile(path)
			if err != nil {
				return err
			}
			if rest, err := file.UnmarshalMsg(b); err != nil || len(rest) > 0 {
				return errors.Errorf("%s does not decode: %v, %d trailing bytes", path, err, len(rest))
			}
		case strings.HasSuffix(name, ".json.gz"):
			name = strings.TrimSuffix(name, ".json.gz")
			if _, err := os.Stat(filepath.Join(root, lang, name+tupleSuffix)); err == nil {
				return nil
			}
			b, err := readGzipFile(path)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(b, &file); err != nil {
				return errors.Wrap(err, "failed to decode "+path)
			}
		default:
			return nil
		}

		if out[lang] == nil {
			out[lang] = map[string]*raw.TranslationFile{}
		}
		out[lang][name] = &file
		return nil
	})
	if os.IsNotExist(err) {
		return out, nil
	}
	return out, errors.Wrap(err, "failed reading per-file translations")
}

// pruneTranslations removes every file under versionDir's stat_translations language directories that keep,
// relative to versionDir, does not name, and the directories that leaves empty.
func pruneTranslations(versionDir string, keep map[string]bool) error {
	root := filepath.Join(versionDir, "stat_translations")
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root {
				dirs = append(dirs, path)
			}
			return nil
		}
		rel, err := filepath.Rel(versionDir, path)
		if err != nil {
			return errors.Wrap(err, "failed resolving translation path")
		}
		if filepath.Dir(path) == root || keep[filepath.ToSlash(rel)] {
			return nil
		}
		return errors.Wrap(os.Remove(path), "could not remove "+path)
	})
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "failed pruning translations")
	}

	slices.Reverse(dirs)
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return errors.Wrap(err, "failed listing "+dir)
		}
		if len(entries) == 0 {
			if err := os.Remove(dir); err != nil {
				return errors.Wrap(err, "could not remove "+dir)
			}
		}
	}
	return nil
}

// bundleTranslations rebuilds versionDir's translation bundles from its per-file translations, or from the
// bundles themselves once those are pruned, and prunes the per-file assets go-pob does not read.
func bundleTranslations(versionDir string) error {
	translations, err := readPerFileTranslations(versionDir)
	if err != nil {
		return err
	}
	if len(translations) == 0 {
		if translations, err = readTranslationBundles(versionDir); err != nil {
			return err
		}
	}
	if len(translations) == 0 {
		// Nothing decodes as a TranslationFile (3.20's map assets predate it), so nothing is bundled.
		return pruneTranslations(versionDir, nil)
	}
	return writeTranslations(versionDir, translations)
}

func readBrotliFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "could not read "+path)
	}
	if bytes.HasPrefix(b, []byte("version https://git-lfs")) {
		return nil, errors.New(path + " is a git lfs pointer, run git lfs pull")
	}
	out, err := io.ReadAll(brotli.NewReader(bytes.NewReader(b)))
	return out, errors.Wrap(err, "could not decompress "+path)
}

func readGzipFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "could not read "+path)
	}
	defer f.Close()
	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, errors.Wrap(err, "could not decompress "+path)
	}
	out, err := io.ReadAll(r)
	return out, errors.Wrap(err, "could not decompress "+path)
}
