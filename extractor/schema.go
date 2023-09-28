package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oriath-net/pogo/dat"
)

var (
	schemaFile *SchemaFile
	tableMap   map[string]Table
)

var dataPath = "data"

func LoadSchema(gameVersion string) {
	schemaRaw, err := os.ReadFile(filepath.Join(dataPath, gameVersion, "schema.min.json"))
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(schemaRaw, &schemaFile); err != nil {
		panic(err)
	}

	tableMap = make(map[string]Table, len(schemaFile.Tables))
	for _, table := range schemaFile.Tables {
		tableMap[table.Name] = table
	}
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
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

func (t Table) ToJSONFormat() dat.JsonFormat {
	out := dat.JsonFormat{
		File:   t.Name,
		Fields: make([]dat.JsonField, len(t.Columns)),
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
		if column.Array {
			tName += "[]"
		}

		out.Fields[i] = dat.JsonField{
			Name:        name,
			Type:        tName,
			Description: description,
			Unique:      column.Unique,
			Ref:         ref,
			RefField:    refField,
		}
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
)

func GetSchema(name string) Table {
	return tableMap[name]
}
