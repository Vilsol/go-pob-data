package loader

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/andybalholm/brotli"

	"github.com/Vilsol/go-pob-data/testdata"
)

// The .ot bundle holds exactly the .ot files the manifest lists, each as LoadFile reads it.
func TestFileBundleMatchesFiles(t *testing.T) {
	cdn := testdata.ServeCommittedData(t)
	ctx := context.Background()

	for _, version := range []string{"0.5", testdata.LatestDataVersion()} {
		manifest, err := LoadManifest(ctx, version)
		testza.AssertNoError(t, err)

		bundle, err := LoadFileBundle(ctx, version, ".ot", FakeCache{})
		testza.AssertNoError(t, err)

		var want []string
		for _, path := range manifest.Files {
			if strings.HasSuffix(path, ".ot") {
				want = append(want, path)
			}
		}
		got := make([]string, 0, len(bundle))
		for path := range bundle {
			got = append(got, path)
		}
		slices.Sort(got)
		testza.AssertGreater(t, len(want), 1)
		testza.AssertEqual(t, want, got, version)

		for _, path := range want {
			text, err := LoadFile(ctx, version, path, FakeCache{})
			testza.AssertNoError(t, err)
			testza.AssertTrue(t, bytes.Equal(text, bundle[path]), path)
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
