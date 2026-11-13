package exelayout

import (
	"fmt"
	"strings"

	"github.com/Vilsol/go-pob-data/extractor"
)

// Disagreement is one way a schema table differs from what the client loads.
type Disagreement struct {
	Table   string
	Message string
}

func (d Disagreement) String() string { return d.Table + ": " + d.Message }

// SchemaTables picks game's tables from a dat-schema file by lower-case name, preferring a
// game-specific table over a shared one as extractor.LoadSchema does.
func SchemaTables(schema *extractor.SchemaFile, game extractor.Game) map[string]extractor.Table {
	bit := 1
	if game == extractor.PoE2 {
		bit = 2
	}
	out := map[string]extractor.Table{}
	for _, t := range schema.Tables {
		if t.ValidFor&bit == 0 {
			continue
		}
		key := strings.ToLower(t.Name)
		if prev, ok := out[key]; ok && prev.ValidFor < t.ValidFor {
			continue
		}
		out[key] = t
	}
	return out
}

func columnSize(c extractor.Column) int {
	if c.Array {
		return 16
	}
	var n int
	switch c.Type {
	case extractor.TypeBool:
		n = 1
	case extractor.TypeI16, extractor.TypeU16:
		n = 2
	case extractor.TypeI32, extractor.TypeU32, extractor.TypeF32, extractor.TypeEnumRow:
		n = 4
	case extractor.TypeString, extractor.TypeRow, "i64", "u64":
		n = 8
	case extractor.TypeForeignRow, extractor.TypeArray:
		n = 16
	}
	if c.Interval {
		n *= 2
	}
	return n
}

// RowSize is the width of a schema table's row.
func RowSize(t extractor.Table) int {
	w := 0
	for _, c := range t.Columns {
		w += columnSize(c)
	}
	return w
}

func schemaKind(c extractor.Column) Kind {
	if c.Array {
		return KindList
	}
	switch c.Type { //nolint:exhaustive // only pointer columns have a kind
	case extractor.TypeString:
		return KindString
	case extractor.TypeForeignRow:
		return KindKey
	case extractor.TypeRow:
		return KindRow
	}
	return ""
}

func columnName(c extractor.Column) string {
	n := "unnamed"
	if c.Name != nil && *c.Name != "" {
		n = *c.Name
	}
	if c.Array {
		return n + " ([" + string(c.Type) + "])"
	}
	return n + " (" + string(c.Type) + ")"
}

// Compare lists how schema table st differs from client table t: its row size, and when the fixup was
// walked, every pointer column whose offset, kind or proven key target the schema has otherwise.
func Compare(t Table, st extractor.Table) []Disagreement {
	var out []Disagreement
	add := func(format string, args ...any) {
		out = append(out, Disagreement{Table: t.Name, Message: fmt.Sprintf(format, args...)})
	}
	if w := RowSize(st); w != t.RowSize {
		add("row is %d bytes, schema has %d", t.RowSize, w)
		return out
	}
	if !t.Walked {
		return out
	}

	type span struct {
		c        extractor.Column
		start, n int
	}
	spans := make([]span, 0, len(st.Columns))
	at := 0
	for _, c := range st.Columns {
		n := columnSize(c)
		spans = append(spans, span{c, at, n})
		at += n
	}
	for _, p := range t.Pointers {
		for _, s := range spans {
			if p.Offset < s.start || p.Offset >= s.start+s.n {
				continue
			}
			isKey := s.c.Type == extractor.TypeForeignRow
			switch {
			case s.start != p.Offset:
				add("byte %d: client reads %s, schema has %s from byte %d", p.Offset, p.Kind, columnName(s.c), s.start)
			case schemaKind(s.c) != p.Kind || (p.Element == 16 && !isKey && s.c.Type != extractor.TypeArray):
				add("byte %d: client reads %s (element %d), schema has %s", p.Offset, p.Kind, p.Element, columnName(s.c))
			case p.Proven && isKey && s.c.References != nil && s.c.References.Table != "" &&
				!strings.EqualFold(s.c.References.Table, p.Target):
				add("byte %d: client resolves %s against %s, schema references %s", p.Offset, columnName(s.c), p.Target, s.c.References.Table)
			}
			break
		}
	}
	for _, s := range spans {
		k := schemaKind(s.c)
		if k != KindString && k != KindList {
			continue
		}
		touched := false
		for _, p := range t.Pointers {
			if p.Offset < s.start+s.n && p.Offset+p.Kind.width() > s.start {
				touched = true
				break
			}
		}
		if !touched {
			add("byte %d: schema has %s, client fixes up nothing there", s.start, columnName(s.c))
		}
	}
	return out
}
