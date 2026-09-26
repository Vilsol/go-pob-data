package loader

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strconv"
	"strings"

	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/raw"
)

const cdnBase = "https://data.pob.gg/data/"

var decodable = reflect.TypeOf((*msgp.Decodable)(nil)).Elem()

// LayoutHash identifies the tuple wire layout of v's row type: every field name and type, in order,
// through nested types. Top-level pointers are ignored, so a row type and its pointer agree.
func LayoutHash(v any) string {
	var b strings.Builder
	writeLayout(&b, rowType(reflect.TypeOf(v)), map[reflect.Type]bool{})

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:4])
}

// rowType is the type whose DecodeMsg runs. A wrapper that embeds a row (as poe's types embed raw's)
// decodes through the row's promoted DecodeMsg, so it has the row's layout.
func rowType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct {
		for i := range t.NumField() {
			if f := t.Field(i); f.Anonymous && reflect.PointerTo(f.Type).Implements(decodable) {
				return rowType(f.Type)
			}
		}
	}

	return t
}

func writeLayout(b *strings.Builder, t reflect.Type, open map[reflect.Type]bool) {
	switch t.Kind() {
	case reflect.Pointer:
		b.WriteString("*")
		writeLayout(b, t.Elem(), open)
	case reflect.Slice:
		b.WriteString("[]")
		writeLayout(b, t.Elem(), open)
	case reflect.Array:
		b.WriteString("[" + strconv.Itoa(t.Len()) + "]")
		writeLayout(b, t.Elem(), open)
	case reflect.Map:
		b.WriteString("map[")
		writeLayout(b, t.Key(), open)
		b.WriteString("]")
		writeLayout(b, t.Elem(), open)
	case reflect.Struct:
		if open[t] {
			b.WriteString(t.String())
			return
		}
		open[t] = true
		defer delete(open, t)

		b.WriteString("struct{")
		for i := range t.NumField() {
			f := t.Field(i)
			b.WriteString(f.Name + " ")
			writeLayout(b, f.Type, open)
			b.WriteString(";")
		}
		b.WriteString("}")
	default:
		b.WriteString(t.Kind().String())
	}
}

// TupleVariant is the asset variant of tuple-encoded rows of v's type.
func TupleVariant(v any) string {
	return "t-" + LayoutHash(v)
}

// AssetPath is where an asset lives under its version directory: "raw/Mods" in variant
// "t-1a2b3c4d" is "raw/Mods.t-1a2b3c4d.msgpack.br". The empty variant is the frozen map-encoded
// asset that older library versions read.
func AssetPath(path string, variant string) string {
	if variant == "" {
		return path + ".msgpack.br"
	}
	return path + "." + variant + ".msgpack.br"
}

// AssetURL is the CDN URL of AssetPath(path, variant) for version.
func AssetURL(version string, path string, variant string) string {
	return cdnBase + version + "/" + AssetPath(path, variant)
}

// RawURL is the URL LoadRaw reads table name of version from, for rows of type T.
func RawURL[T any](version string, name string) string {
	var row T
	return AssetURL(version, "raw/"+name, TupleVariant(row))
}

// TranslationURL is the URL LoadTranslation reads from.
func TranslationURL(version string, language string, name string) string {
	return AssetURL(version, "stat_translations/"+language+"/"+name, TupleVariant(raw.TranslationFile{}))
}
