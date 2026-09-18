package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/tinylib/msgp/msgp"
)

const staleSchema = "older dat schema than the struct (3.28+), re-encoding would drop its columns"

var update = flag.Bool("update", false, "rewrite msgpack.br files that do not match their json.gz")

// Known mismatches, predating this test. Drop an entry and run with -update to rebuild it.
var knownStale = map[string]string{
	"3.18/AtlasNode":           staleSchema,
	"3.18/AtlasNodeDefinition": staleSchema,
	"3.24/AtlasNode":           staleSchema,
	"3.24/AtlasNodeDefinition": staleSchema,
	"3.25/AtlasNode":           staleSchema,
	"3.25/AtlasNodeDefinition": staleSchema,
	"3.26/AtlasNode":           staleSchema,
	"3.26/AtlasNodeDefinition": staleSchema,
	"3.27/AtlasNode":           staleSchema,
	"3.27/AtlasNodeDefinition": staleSchema,
}

// A struct that gains a field without its artifacts being rebuilt silently ships msgpack without that column.
func TestMsgpackMatchesJSON(t *testing.T) {
	dirs, err := filepath.Glob("data/*/raw")
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range dirs {
		for _, file := range filesToExport {
			name := filepath.Base(file.path)
			jsonPath := filepath.Join(dir, name+".json.gz")
			msgpPath := filepath.Join(dir, name+".msgpack.br")

			key := filepath.Base(filepath.Dir(dir)) + "/" + name
			t.Run(key, func(t *testing.T) {
				committed, err := os.ReadFile(msgpPath)
				if os.IsNotExist(err) {
					t.Skip("not extracted for this version")
				}
				if err != nil {
					t.Fatal(err)
				}
				if bytes.HasPrefix(committed, []byte("version https://git-lfs")) {
					t.Skip("git lfs pointer, run git lfs pull")
				}

				f, err := os.Open(jsonPath)
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				gz, err := gzip.NewReader(f)
				if err != nil {
					t.Fatal(err)
				}
				jsonData, err := io.ReadAll(gz)
				if err != nil {
					t.Fatal(err)
				}

				var rows []json.RawMessage
				if err := json.Unmarshal(jsonData, &rows); err != nil {
					t.Fatal(err)
				}

				reason, stale := knownStale[key]

				var buf bytes.Buffer
				bw := brotli.NewWriter(&buf)
				w := msgp.NewWriter(bw)
				if err := w.WriteInt64(int64(len(rows))); err != nil {
					t.Fatal(err)
				}
				if err := file.Fn(jsonData, w); err != nil {
					if stale {
						t.Skip("known stale: " + reason)
					}
					t.Fatal(err)
				}
				if err := w.Flush(); err != nil {
					t.Fatal(err)
				}
				if err := bw.Close(); err != nil {
					t.Fatal(err)
				}

				switch equal := bytes.Equal(buf.Bytes(), committed); {
				case !equal && !stale && *update:
					if err := os.WriteFile(msgpPath, buf.Bytes(), 0o755); err != nil {
						t.Fatal(err)
					}
				case !equal && !stale:
					t.Errorf("%s is stale: re-encode it from %s", msgpPath, filepath.Base(jsonPath))
				case equal && stale:
					t.Errorf("%s is no longer stale, remove it from knownStale", key)
				case stale:
					t.Skip("known stale: " + reason)
				}
			})
		}
	}
}
