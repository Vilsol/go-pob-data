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
	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/sidecar"
)

func main() {
	ggpkPath := flag.String("ggpk", "", "the Content.ggpk without a bundle index")
	refPath := flag.String("ref", "", "the reference patch: a directory holding Bundles2/, or cdn:<patch>")
	schema := flag.String("schema", "", "data/<version>/schema.min.json to read the reference tables' columns from")
	out := flag.String("out", "", "directory to write sidecar.tsv and unnamed.tsv into")
	inputs := flag.String("inputs", "", "paths the extractor opens, one per line (a GOPOB_OPENLOG file works): their bundles are indexed too")
	dump := flag.String("dump", "", "write this target bundle, decompressed, to -out and exit")
	flag.Parse()
	if *dump != "" && *ggpkPath != "" && *out != "" {
		if err := dumpBundle(*ggpkPath, *dump, *out); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		return
	}
	if *ggpkPath == "" || *refPath == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	if *inputs != "" {
		if err := readInputs(*inputs); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}
	if err := run(*ggpkPath, *refPath, *schema, *out); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(ggpkPath, refPath, schema, out string) error {
	f, err := os.Open(ggpkPath)
	if err != nil {
		return errors.Wrap(err, "could not open the ggpk")
	}
	defer f.Close()
	target, err := ggpk.NewLoader(f)
	if err != nil {
		return errors.Wrap(err, "could not read the ggpk")
	}

	readRef := func(name string) ([]byte, error) { return readAll(os.DirFS(refPath), name) }
	if patch, ok := strings.CutPrefix(refPath, "cdn:"); ok {
		web := extractor.NewWebFS(extractor.PoE1, patch)
		readRef = func(name string) ([]byte, error) {
			// WebFile.Read never reports EOF: take the whole file it fetched
			f, _ := web.Open(name)
			wf := f.(*extractor.WebFile)
			if _, err := wf.Read(nil); err != nil {
				return nil, err
			}
			return wf.CachedFile, nil
		}
	}
	raw, err := readRef("Bundles2/_.index.bin")
	if err != nil {
		return fmt.Errorf("reference index: %w", err)
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
				// the parser reads either graph layout from the file (3.30's PoE1 graphs use PoE2's)
				_, err := extractor.ParsePassiveSkillGraph(bytes.NewReader(b))
				return err
			}
			return nil
		},
	}
	if schema != "" {
		extractor.LoadSchema(extractor.PoE1, schema)
		names := extractor.SchemaTables()
		opt.Columns = func(table string) []sidecar.Column {
			columns := extractor.SchemaColumns(names[table])
			cols := make([]sidecar.Column, 0, len(columns))
			for _, c := range columns {
				cols = append(cols, sidecar.Column{Type: c.Type, Ref: c.Ref, Offset: c.Offset, Size: c.Size})
			}
			return cols
		}
	}

	refBundles := map[string]bool{}
	for _, f := range refFiles {
		refBundles[f.Bundle] = true
	}
	bundles := targetBundles(refFiles, target)
	slog.Info("target bundles", "count", len(bundles))
	entries, spans, err := sidecar.Rebuild(
		func(name string) ([]byte, error) { return sidecar.ReadBundle(target, name) },
		func(name string) ([]byte, error) {
			if !refBundles[name] {
				return nil, fs.ErrNotExist
			}
			raw, err := readRef("Bundles2/" + name + ".bundle.bin")
			if err != nil {
				return nil, err
			}
			return sidecar.DecodeBundle(raw)
		},
		refFiles, bundles, opt)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(out, 0o755); err != nil {
		return errors.Wrap(err, "could not create "+out)
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

var inputSet = map[string]bool{}

// readInputs loads -inputs: lines of a path, or of a status and a path separated by a tab.
func readInputs(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "could not read "+path)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if i := strings.LastIndexByte(line, '\t'); i >= 0 {
			line = line[i+1:]
		}
		if line = strings.TrimSpace(line); line != "" {
			inputSet[strings.ToLower(line)] = true
		}
	}
	return nil
}

func readAll(fsys fs.FS, name string) ([]byte, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, errors.Wrap(err, "could not open "+name)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	return b, errors.Wrap(err, "could not read "+name)
}

func dumpBundle(ggpkPath, name, out string) error {
	f, err := os.Open(ggpkPath)
	if err != nil {
		return errors.Wrap(err, "could not open the ggpk")
	}
	defer f.Close()
	target, err := ggpk.NewLoader(f)
	if err != nil {
		return errors.Wrap(err, "could not read the ggpk")
	}
	b, err := sidecar.ReadBundle(target, name)
	if err != nil {
		return err
	}
	return errors.Wrap(os.WriteFile(out, b, 0o644), "could not write "+out)
}

func wanted(p string) bool {
	if inputSet[p] {
		return true
	}
	switch p {
	case "metadata/passiveskillgraph.psg", "metadata/characters/character.ot", "metadata/monsters/monster.ot", "art/uiimages1.txt":
		return true
	}
	return englishTable.MatchString(p) || strings.HasPrefix(p, "metadata/statdescriptions/")
}

func writeFile(p string, fn func(io.Writer) error) error {
	f, err := os.Create(p)
	if err != nil {
		return errors.Wrap(err, "could not create "+p)
	}
	if err := fn(f); err != nil {
		_ = f.Close()
		return err
	}
	return errors.Wrap(f.Close(), "could not close "+p)
}
