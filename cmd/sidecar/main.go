// Command sidecar writes a sidecar index for a Content.ggpk that ships its bundles without Bundles2/_.index.bin,
// using a previous patch (its index and bundles, from a directory or the patch CDN) as the map.
//
//	sidecar -ggpk Content.ggpk -ref cdn:3.29.3.6 -schema 3.29 -out <dir>
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/oriath-net/pogo/poefs/ggpk"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/sidecar"
)

func main() {
	ggpkPath := flag.String("ggpk", "", "the Content.ggpk without a bundle index")
	refPath := flag.String("ref", "", "the reference patch: a directory holding Bundles2/, or cdn:<patch>")
	schema := flag.String("schema", "", "data/<version>/schema.min.json to read the reference tables' columns from")
	out := flag.String("out", "", "directory to write sidecar.tsv and unnamed.tsv into")
	flag.Parse()
	if *ggpkPath == "" || *refPath == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*ggpkPath, *refPath, *schema, *out); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(ggpkPath, refPath, schema, out string) error {
	f, err := os.Open(ggpkPath)
	if err != nil {
		return err
	}
	defer f.Close()
	target, err := ggpk.NewLoader(f)
	if err != nil {
		return err
	}

	var ref fs.FS = os.DirFS(refPath)
	if patch, ok := strings.CutPrefix(refPath, "cdn:"); ok {
		ref = extractor.NewWebFS(extractor.PoE1, patch)
	}
	ih, err := ref.Open("Bundles2/_.index.bin")
	if err != nil {
		return fmt.Errorf("reference index: %w", err)
	}
	raw, err := io.ReadAll(ih)
	_ = ih.Close()
	if err != nil {
		return err
	}
	refFiles, err := sidecar.ParseIndex(raw)
	if err != nil {
		return err
	}

	opt := sidecar.Options{
		MovedText: func(p string) bool { return strings.HasPrefix(p, "metadata/statdescriptions/") },
		Logf:      func(format string, args ...any) { slog.Info(fmt.Sprintf(format, args...)) },
		Check: func(p string, b []byte) error {
			if strings.HasSuffix(p, ".psg") {
				// 3.30's PoE1 graphs use the PoE2 layout (u64 roots, edges with an orbit); accept either
				if _, err := extractor.ParsePassiveSkillGraph(bytes.NewReader(b), extractor.PoE1); err == nil {
					return nil
				}
				_, err := extractor.ParsePassiveSkillGraph(bytes.NewReader(b), extractor.PoE2)
				return err
			}
			return nil
		},
	}
	if schema != "" {
		extractor.LoadSchema(extractor.PoE1, schema)
		names := extractor.SchemaTables()
		opt.Columns = func(table string) []sidecar.Column {
			var cols []sidecar.Column
			for _, c := range extractor.SchemaColumns(names[table]) {
				cols = append(cols, sidecar.Column{Type: c.Type, Ref: c.Ref, Offset: c.Offset, Size: c.Size})
			}
			return cols
		}
	}

	bundles := targetBundles(refFiles, target)
	slog.Info("target bundles", "count", len(bundles))
	entries, spans, err := sidecar.Rebuild(
		func(name string) ([]byte, error) { return sidecar.ReadBundle(target, name) },
		func(name string) ([]byte, error) { return sidecar.ReadBundle(ref, name) },
		refFiles, bundles, opt)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(out, "sidecar.tsv"), func(w io.Writer) error { return sidecar.WriteEntries(w, entries) }); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(out, "unnamed.tsv"), func(w io.Writer) error { return sidecar.WriteSpans(w, spans) }); err != nil {
		return err
	}
	counts := map[sidecar.Confidence]int{}
	for _, e := range entries {
		counts[e.Confidence]++
	}
	var unnamed int64
	for _, s := range spans {
		unnamed += s.Size
	}
	fmt.Printf("entries: high %d, medium %d, low %d; unnamed spans %d (%d bytes)\n", counts[sidecar.High], counts[sidecar.Medium], counts[sidecar.Low], len(spans), unnamed)
	return nil
}

// targetBundles are the target's bundles that continue a reference bundle holding a file the extractor reads:
// the same name, or the same name with a _N split suffix.
func targetBundles(refFiles []sidecar.RefFile, target fs.FS) []string {
	bases := map[string]bool{}
	for _, f := range refFiles {
		if wanted(f.Path) && strings.HasPrefix(f.Bundle, "Folders/") {
			b := f.Bundle
			if i := strings.LastIndexByte(b, '_'); i > strings.LastIndexByte(b, '/') {
				b = b[:i]
			}
			bases[b] = true
		}
	}
	var out []string
	for b := range bases {
		for i := range 5 {
			name := b
			if i > 0 {
				name = fmt.Sprintf("%s_%d", b, i)
			}
			if h, err := target.Open("Bundles2/" + name + ".bundle.bin"); err == nil {
				_ = h.Close()
				out = append(out, name)
			}
		}
	}
	slices.Sort(out)
	return out
}

var englishTable = regexp.MustCompile(`^data/[^/]+\.datc(l)?64$`)

func wanted(p string) bool {
	switch p {
	case "metadata/passiveskillgraph.psg", "metadata/characters/character.ot", "metadata/monsters/monster.ot", "art/uiimages1.txt":
		return true
	}
	return englishTable.MatchString(p) || strings.HasPrefix(p, "metadata/statdescriptions/")
}

func writeFile(p string, fn func(io.Writer) error) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	if err := fn(f); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
