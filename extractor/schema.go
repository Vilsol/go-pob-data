package extractor

import (
	"encoding/json"
	"fmt"

	"github.com/oriath-net/pogo/dat"
)

const schemaURL = "https://github.com/poe-tool-dev/dat-schema/releases/download/latest/schema.min.json"

var schemaFile *SchemaFile
var tableMap map[string]Table

func LoadSchema() {
	schemaRaw := Fetch(schemaURL)

	if err := json.Unmarshal(schemaRaw, &schemaFile); err != nil {
		panic(err)
	}

	tableMap = make(map[string]Table, len(schemaFile.Tables))
	for _, table := range schemaFile.Tables {
		tableMap[table.Name] = table
	}
}

type SchemaFile struct {
	Version      int64         `json:"version"`
	CreatedAt    int64         `json:"createdAt"`
	Tables       []Table       `json:"tables"`
	Enumerations []Enumeration `json:"enumerations"`
}

type Enumeration struct {
	Name        string    `json:"name"`
	Indexing    int64     `json:"indexing"`
	Enumerators []*string `json:"enumerators"`
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
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Array       bool        `json:"array"`
	Type        Type        `json:"type"`
	Unique      bool        `json:"unique"`
	Localized   bool        `json:"localized"`
	References  *References `json:"references"`
	Until       interface{} `json:"until"`
	File        *string     `json:"file"`
	Files       []File      `json:"files"`
}

type References struct {
	Table  string  `json:"table"`
	Column *string `json:"column,omitempty"`
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
