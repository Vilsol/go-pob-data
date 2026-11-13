package extractor

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

const (
	// refitAgree is the share of matched rows a column must agree on to be placed.
	refitAgree = 0.7
	// refitMargin is how far the best placement must beat the runner-up.
	refitMargin = 0.1
	// refitMinRows is the fewest rows matched by Id that say anything about a table.
	refitMinRows = 10
)

// RefitTable is one published table: its rows' cells by column, as comparable strings, and the row IDs.
type RefitTable struct {
	Cells   map[string][]string
	Kinds   map[string]string
	IDs     []string
	Columns []string
}

// RefitFinding is a column whose values did not stay under its name from one version to the next.
type RefitFinding struct {
	Column string
	// Kind is moved (its values now sit under To), retyped (its JSON type changed), changed (no column agrees
	// better than its own, but that one disagrees on most rows) or lost (placed nowhere).
	Kind      string
	To        string
	Agreement float64
}

func (f RefitFinding) String() string {
	switch f.Kind {
	case "moved":
		return fmt.Sprintf("%s moved to %s (%.0f%% of rows agree)", f.Column, f.To, f.Agreement*100)
	case "retyped":
		return fmt.Sprintf("%s retyped %s", f.Column, f.To)
	case "changed":
		return fmt.Sprintf("%s changed in place (%.0f%% of rows agree)", f.Column, f.Agreement*100)
	}
	return fmt.Sprintf("%s lost (best placement %s agrees on %.0f%% of rows)", f.Column, f.To, f.Agreement*100)
}

// LoadRefitTables reads the json.gz tables of a published version under dataDir, every one or only those
// named. Foreign keys are resolved to the Id of the row they point at, so a key keeps its value when the
// referenced table gains rows.
func LoadRefitTables(dataDir string, gameVersion string, only ...string) (map[string]*RefitTable, error) {
	schema, err := readSchema(filepath.Join(dataDir, gameVersion, "schema.min.json"))
	if err != nil {
		return nil, err
	}
	tables := schema.gameTables(GameOf(gameVersion))

	paths, err := filepath.Glob(filepath.Join(dataDir, gameVersion, "raw", "*.json.gz"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to list tables")
	}

	wanted := map[string]bool{}
	for _, name := range only {
		wanted[name] = true
		for _, f := range tables[name].ToJSONFormat().Fields {
			wanted[f.Ref] = true
		}
	}

	rows := make(map[string][]map[string]json.RawMessage, len(paths))
	ids := map[string]map[int64]string{}
	for _, p := range paths {
		name := strings.TrimSuffix(filepath.Base(p), ".json.gz")
		if len(only) > 0 && !wanted[name] {
			continue
		}
		r, err := readRows(p)
		if err != nil {
			return nil, err
		}
		rows[name] = r
		ids[name] = rowIDs(r)
	}

	out := make(map[string]*RefitTable, len(rows))
	for name, r := range rows {
		if len(only) > 0 && !contains(only, name) {
			continue
		}
		out[name] = refitTable(tables[name], r, ids)
	}
	return out, nil
}

func readRows(p string) ([]map[string]json.RawMessage, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open table")
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open table "+p)
	}

	var r []map[string]json.RawMessage
	if err := json.NewDecoder(gz).Decode(&r); err != nil {
		return nil, errors.Wrap(err, "failed to decode table "+p)
	}
	return r, nil
}

// rowIDs maps each row's key to its Id; nil when the table has no string Id.
func rowIDs(rows []map[string]json.RawMessage) map[int64]string {
	out := map[int64]string{}
	for _, row := range rows {
		var key int64
		var id string
		if json.Unmarshal(row["_key"], &key) != nil || json.Unmarshal(row["Id"], &id) != nil {
			return nil
		}
		out[key] = id
	}
	return out
}

func refitTable(schema Table, rows []map[string]json.RawMessage, ids map[string]map[int64]string) *RefitTable {
	refs := map[string]string{}
	var columns []string
	for _, f := range schema.ToJSONFormat().Fields {
		columns = append(columns, f.Name)
		switch {
		case f.Ref != "":
			refs[f.Name] = f.Ref
		case strings.HasPrefix(f.Type, "shortid"):
			refs[f.Name] = schema.Name
		}
	}
	if len(rows) > 0 {
		// Columns the json has but this schema does not name, so a schema mismatch still compares every value.
		for c := range rows[0] {
			if c != "_key" && !contains(columns, c) {
				columns = append(columns, c)
			}
		}
	}

	t := &RefitTable{Cells: map[string][]string{}, Kinds: map[string]string{}, Columns: columns}
	for _, row := range rows {
		var id string
		if json.Unmarshal(row["Id"], &id) != nil {
			id = ""
		}
		t.IDs = append(t.IDs, id)
		for _, c := range columns {
			raw := row[c]
			if k := jsonKind(raw); k != "null" && k != "" {
				t.Kinds[c] = k
			}
			t.Cells[c] = append(t.Cells[c], resolveKeys(raw, ids[refs[c]]))
		}
	}
	return t
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func jsonKind(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	switch raw[0] {
	case 'n':
		return "null"
	case 't', 'f':
		return "bool"
	case '"':
		return "string"
	case '[':
		return "list"
	case '{':
		return "object"
	}
	return "number"
}

// resolveKeys turns a key or list of keys into the IDs of the rows they point at, leaving anything else as is.
func resolveKeys(raw json.RawMessage, ids map[int64]string) string {
	if ids == nil {
		return string(raw)
	}
	var key int64
	if json.Unmarshal(raw, &key) == nil {
		return "#" + ids[key]
	}
	var keys []int64
	if json.Unmarshal(raw, &keys) == nil && len(keys) > 0 {
		out := make([]string, len(keys))
		for i, k := range keys {
			out[i] = ids[k]
		}
		return "#[" + strings.Join(out, ",") + "]"
	}
	return string(raw)
}

// Refit matches rows by Id between two versions of a table and reports, per column of the older one, where
// its values went: a column whose values now sit under another name (at least refitAgree of the rows, beating
// every other placement by refitMargin) moved, one placed nowhere is lost, and one that agrees with nothing
// better than its own name changed in place, which a key into an unpublished table does whenever that table
// gains rows. A column holding a single value says nothing about where it went
// and is skipped. Nil when too few rows share an Id.
func Refit(older, newer *RefitTable) []RefitFinding {
	newRow := map[string]int{}
	for i, id := range newer.IDs {
		if id == "" {
			continue
		}
		if _, dup := newRow[id]; dup {
			newRow[id] = -1
		} else {
			newRow[id] = i
		}
	}
	type pair struct{ old, new int }
	var pairs []pair
	seenOld := map[string]int{}
	for _, id := range older.IDs {
		seenOld[id]++
	}
	for i, id := range older.IDs {
		if j, ok := newRow[id]; ok && j >= 0 && seenOld[id] == 1 {
			pairs = append(pairs, pair{i, j})
		}
	}
	if len(pairs) < refitMinRows {
		return nil
	}

	var findings []RefitFinding
	for _, c := range older.Columns {
		oldCells := older.Cells[c]
		if oldCells == nil {
			continue
		}
		counts := map[string]int{}
		mode := ""
		for _, p := range pairs {
			v := oldCells[p.old]
			counts[v]++
			if counts[v] > counts[mode] {
				mode = v
			}
		}

		if newKind, oldKind := newer.Kinds[c], older.Kinds[c]; newKind != "" && oldKind != "" && newKind != oldKind {
			findings = append(findings, RefitFinding{Column: c, Kind: "retyped", To: oldKind + " -> " + newKind})
		}
		if len(counts) < 2 {
			continue
		}

		// A column that mostly holds one value agrees with every other column that does; only its other values
		// tell where it went, so a placement scores the lower of its agreement over all rows and over those.
		agree := make(map[string]float64, len(newer.Columns))
		for _, d := range newer.Columns {
			newCells := newer.Cells[d]
			all, rare := 0, 0
			for _, p := range pairs {
				if oldCells[p.old] == newCells[p.new] {
					all++
					if oldCells[p.old] != mode {
						rare++
					}
				}
			}
			agree[d] = min(float64(all)/float64(len(pairs)), float64(rare)/float64(len(pairs)-counts[mode]))
		}
		best, second := "", ""
		for _, d := range newer.Columns {
			switch {
			case best == "" || agree[d] > agree[best]:
				best, second = d, best
			case second == "" || agree[d] > agree[second]:
				second = d
			}
		}

		self, hasSelf := agree[c]
		switch {
		case hasSelf && self >= refitAgree && self >= agree[best]-refitMargin:
		case best != c && agree[best] >= refitAgree && agree[best]-agree[second] >= refitMargin:
			findings = append(findings, RefitFinding{Column: c, Kind: "moved", To: best, Agreement: agree[best]})
		case hasSelf && self >= agree[best]:
			// Nothing agrees better than the column itself: its values changed in place, or it is lost under its name.
			findings = append(findings, RefitFinding{Column: c, Kind: "changed", To: c, Agreement: self})
		default:
			findings = append(findings, RefitFinding{Column: c, Kind: "lost", To: best, Agreement: agree[best]})
		}
	}
	sort.SliceStable(findings, func(i, j int) bool { return findings[i].Kind < findings[j].Kind })
	return findings
}
