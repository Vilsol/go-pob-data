package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/exelayout"
	"github.com/Vilsol/go-pob-data/extractor"
)

// checkExe compares the layouts the client executable loads against data/<gameVersion>/schema.min.json.
// A table the extractor exports that disagrees fails the check; any other only warns.
func checkExe(game extractor.Game, source string, gameVersion string) error {
	patch, ok := strings.CutPrefix(source, cdnPrefix)
	if !ok {
		return fmt.Errorf("-check-exe reads the client from the patch CDN: pass %s<patch>, not %q", cdnPrefix, source)
	}
	layout, err := exelayout.Load(game, patch)
	if err != nil {
		return err
	}

	raw, err := os.ReadFile(filepath.Join("data", gameVersion, "schema.min.json"))
	if err != nil {
		return errors.Wrap(err, "reading the schema")
	}
	var schema extractor.SchemaFile
	if err := json.Unmarshal(raw, &schema); err != nil {
		return errors.Wrap(err, "decoding the schema")
	}
	tables := exelayout.SchemaTables(&schema, game)

	exported := map[string]string{}
	for _, f := range tablesToExport(game) {
		exported[strings.ToLower(path.Base(f.path))] = path.Base(f.path)
	}

	failed, warned, compared := 0, 0, 0
	for _, t := range layout.Tables {
		key := strings.ToLower(t.Name)
		_, isExported := exported[key]
		delete(exported, key)
		st, ok := tables[key]
		if !ok {
			continue
		}
		compared++
		for _, d := range exelayout.Compare(t, st) {
			if isExported {
				slog.Error(d.String())
				failed++
			} else {
				slog.Warn(d.String())
				warned++
			}
		}
	}
	for _, name := range slices.Sorted(maps.Values(exported)) {
		slog.Warn("the client loads no exported table of this name", slog.String("table", name))
	}

	slog.Info("checked the client's table layouts",
		slog.String("patch", patch),
		slog.Int("client_tables", len(layout.Tables)),
		slog.Int("compared", compared),
		slog.Int("exported_disagreements", failed),
		slog.Int("other_disagreements", warned))
	if failed > 0 {
		return fmt.Errorf("%d disagreements in exported tables", failed)
	}
	return nil
}
