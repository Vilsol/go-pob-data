package extractor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/oriath-net/pogo/dat"
	"github.com/pkg/errors"
)

var (
	schemaFile  *SchemaFile
	tableMap    map[string]Table
	tableSource map[string]string
)

var dataPath = "data"

func LoadSchema(game Game, gameVersion string) {
	schemaPath := filepath.Join(dataPath, gameVersion, "schema.min.json")
	var err error
	schemaFile, err = readSchema(schemaPath)
	if err != nil {
		panic(err)
	}

	tableMap = schemaFile.gameTables(game)
	tableSource = make(map[string]string, len(tableMap))
	for name := range tableMap {
		tableSource[name] = schemaPath
	}
}

func readSchema(schemaPath string) (*SchemaFile, error) {
	schemaRaw, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read schema")
	}

	var s *SchemaFile
	if err := json.Unmarshal(schemaRaw, &s); err != nil {
		return nil, errors.Wrap(err, "failed to decode schema")
	}
	return s, nil
}

// gameTables is every table the schema describes for game.
func (s *SchemaFile) gameTables(game Game) map[string]Table {
	out := make(map[string]Table, len(s.Tables))
	for _, table := range s.Tables {
		if table.ValidFor&game.validFor() == 0 {
			continue
		}
		if t, ok := out[table.Name]; ok {
			if t.ValidFor < table.ValidFor {
				// Ignore newer versions of the table
				continue
			}
		}

		out[table.Name] = table
	}
	return out
}

type SchemaFile struct {
	Tables       []Table       `json:"tables"`
	Enumerations []Enumeration `json:"enumerations"`
	Version      int64         `json:"version"`
	CreatedAt    int64         `json:"createdAt"`
}

type Enumeration struct {
	Name        string    `json:"name"`
	Enumerators []*string `json:"enumerators"`
	Indexing    int64     `json:"indexing"`
}

type Table struct {
	Name     string   `json:"name"`
	Columns  []Column `json:"columns"`
	ValidFor int      `json:"validFor"`
}

func (t Table) ToJSONFormat() dat.JsonFormat {
	out := dat.JsonFormat{
		File:   t.Name,
		Fields: make([]dat.JsonField, 0, len(t.Columns)),
	}

	for i, column := range t.Columns {
		name := ""
		if column.Name != nil {
			name = *column.Name
		} else {
			name = fmt.Sprintf("Var%d", i)
		}

		description := ""
		if column.Description != nil {
			description = *column.Description
		}

		ref := ""
		refField := ""
		if column.References != nil {
			ref = column.References.Table
			if column.References.Column != nil {
				refField = *column.References.Column
			}
		}

		tName := typeToType[column.Type]
		if tName == "" {
			slog.Error("unknown type", slog.String("type", string(column.Type)), slog.String("file", t.Name))
		}

		if column.Array {
			tName += "[]"
		}

		field := dat.JsonField{
			Name:        name,
			Type:        tName,
			Description: description,
			Unique:      column.Unique,
			Ref:         ref,
			RefField:    refField,
		}

		// An interval is a min and a max of its type; pogo has no interval type.
		if column.Interval {
			field.Name = name + "Min"
			out.Fields = append(out.Fields, field)
			field.Name = name + "Max"
		}

		out.Fields = append(out.Fields, field)
	}

	return out
}

var typeToType = map[Type]string{
	TypeArray:      "void",
	TypeBool:       "bool",
	TypeEnumRow:    "i32",
	TypeF32:        "f32",
	TypeForeignRow: "longid",
	TypeI32:        "i32",
	TypeRow:        "shortid",
	TypeString:     "string",
	TypeU16:        "u16",
	TypeI16:        "i16",
	TypeU32:        "u32",
}

type Column struct {
	Until       interface{} `json:"until"`
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	References  *References `json:"references"`
	File        *string     `json:"file"`
	Type        Type        `json:"type"`
	Files       []File      `json:"files"`
	Array       bool        `json:"array"`
	Unique      bool        `json:"unique"`
	Localized   bool        `json:"localized"`
	Interval    bool        `json:"interval"`
}

type References struct {
	Column *string `json:"column,omitempty"`
	Table  string  `json:"table"`
}

type File string

type Type string

const (
	TypeArray      Type = "array"
	TypeBool       Type = "bool"
	TypeEnumRow    Type = "enumrow"
	TypeF32        Type = "f32"
	TypeForeignRow Type = "foreignrow"
	TypeI32        Type = "i32"
	TypeRow        Type = "row"
	TypeString     Type = "string"
	TypeU16        Type = "u16"
	TypeI16        Type = "i16"
	TypeU32        Type = "u32"
)

func GetSchema(name string) Table {
	return tableMap[name]
}

// SchemaColumn is one decoded column of a table's datc64 row: its pogo type, byte offset and size.
type SchemaColumn struct {
	Name   string
	Type   string
	Ref    string
	Offset int
	Size   int
}

// SchemaColumns lays out table name's datc64 row as pogo would decode it; nil when the schema has no such table.
func SchemaColumns(name string) []SchemaColumn {
	t, ok := tableMap[name]
	if !ok {
		return nil
	}
	fields := t.ToJSONFormat().Fields
	out := make([]SchemaColumn, 0, len(fields))
	off := 0
	for _, f := range fields {
		ft := dat.FieldType(f.Type)
		if !ft.Valid() {
			continue
		}
		size := ft.Size(dat.Datc64)
		out = append(out, SchemaColumn{Name: f.Name, Type: f.Type, Ref: f.Ref, Offset: off, Size: size})
		off += size
	}
	return out
}

// SchemaTables maps every loaded table's lower-case name to its schema name.
func SchemaTables() map[string]string {
	out := make(map[string]string, len(tableMap))
	for n := range tableMap {
		out[strings.ToLower(n)] = n
	}
	return out
}
