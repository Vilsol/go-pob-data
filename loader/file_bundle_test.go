package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/andybalholm/brotli"

	"github.com/Vilsol/go-pob-data/testdata"
)

// releasedFiles are the files released go-pob versions read one at a time with LoadFile: the trade site's
// stats lists. Every other file is read from its bundle.
var releasedFiles = map[string][]string{
	"0.5":  {"api/trade2/data/stats.json"},
	"3.29": {"api/trade/data/stats.json"},
}

// The game text files are published as one bundle per extension, read the way LoadFileBundle reads them
// since v0.5.17; LoadFile reads only releasedFiles, which the manifest lists as Files.
func TestFileBundlesAndFiles(t *testing.T) {
	cdn := testdata.ServeCommittedData(t)
	ctx := context.Background()

	for version, files := range releasedFiles {
		manifest, err := LoadManifest(ctx, version)
		testza.AssertNoError(t, err)
		testza.AssertEqual(t, files, manifest.Files, version)

		for _, path := range files {
			text, err := LoadFile(ctx, version, path, FakeCache{})
			testza.AssertNoError(t, err, path)
			testza.AssertTrue(t, json.Valid(text), path)
		}

		for _, ext := range []string{".ot", ".it", ".txt"} {
			bundle, err := LoadFileBundle(ctx, version, ext, FakeCache{})
			testza.AssertNoError(t, err, version+" "+ext)
			testza.AssertGreater(t, len(bundle), 0, version+" "+ext)
			for path, text := range bundle {
				testza.AssertTrue(t, strings.HasSuffix(path, ext), path)
				testza.AssertGreater(t, len(text), 0, path)
			}
		}
	}

	testza.AssertEqual(t, []string(nil), cdn.Misses())

	_, err := LoadFileBundle(ctx, "3.28", ".ot", FakeCache{})
	testza.AssertErrorIs(t, err, ErrAssetNotFound, "3.28 publishes no .ot files")
}

func TestFileBundleRoundTrip(t *testing.T) {
	files := map[string][]byte{"b.ot": []byte("two"), "a.ot": []byte("one\n"), "c.ot": {}}
	encoded := EncodeFileBundle(files)
	testza.AssertTrue(t, strings.HasPrefix(FileBundlePath(".ot", encoded), "files.ot-"))
	testza.AssertNotEqual(t, FileBundlePath(".ot", encoded), FileBundlePath(".ot", EncodeFileBundle(map[string][]byte{"a.ot": []byte("one")})))

	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	_, err := w.Write(encoded)
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, w.Close())

	got, err := decodeFileBundle(buf.Bytes())
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, len(files), len(got))
	for path, text := range files {
		testza.AssertTrue(t, bytes.Equal(text, got[path]), path)
	}

	_, err = decodeFileBundle(append(buf.Bytes()[:0:0], buf.Bytes()[:len(buf.Bytes())/2]...))
	testza.AssertNotNil(t, err, "a truncated bundle is an error")
}
