package fields

import (
	"bytes"
	"io"
	"reflect"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// Project turns the full tuple asset of asset into its .min under keep: every row is decoded, every
// unlisted field set to its zero value, and the rows re-encoded and compressed the way the extractor
// writes the full asset.
func Project(asset string, keep []string, full []byte) ([]byte, error) {
	t, ok := RowType(asset)
	if !ok {
		return nil, errors.Errorf("%s: unknown asset", asset)
	}
	tree, err := keepTree(t, keep)
	if err != nil {
		return nil, errors.Wrap(err, asset)
	}

	r := msgp.NewReader(brotli.NewReader(bytes.NewReader(full)))
	var buf bytes.Buffer
	bw := brotli.NewWriter(&buf)
	w := msgp.NewWriter(bw)

	// A raw table is a row count and that many rows, a translation asset is one TranslationFile.
	count := int64(1)
	if strings.HasPrefix(asset, "raw/") {
		if count, err = r.ReadInt64(); err != nil {
			return nil, errors.Wrap(err, asset+": failed to read row count")
		}
		if err := w.WriteInt64(count); err != nil {
			return nil, errors.Wrap(err, "failed to write row count")
		}
	}

	for i := range count {
		row := reflect.New(t)
		if err := row.Interface().(msgp.Decodable).DecodeMsg(r); err != nil {
			return nil, errors.Wrapf(err, "%s: failed to decode row %d", asset, i)
		}
		zeroUnkept(row.Elem(), tree)
		if err := row.Interface().(msgp.Encodable).EncodeMsg(w); err != nil {
			return nil, errors.Wrapf(err, "%s: failed to encode row %d", asset, i)
		}
	}

	if _, err := r.R.ReadByte(); !errors.Is(err, io.EOF) {
		return nil, errors.Errorf("%s: trailing bytes after the last row", asset)
	}
	if err := w.Flush(); err != nil {
		return nil, errors.Wrap(err, "failed to flush msgpack")
	}
	if err := bw.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to compress msgpack")
	}

	return buf.Bytes(), nil
}

func zeroUnkept(v reflect.Value, n *node) {
	if n.all {
		return
	}

	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			zeroUnkept(v.Elem(), n)
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			zeroUnkept(v.Index(i), n)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			c := n.children[v.Type().Field(i).Name]
			switch {
			case c == nil:
				v.Field(i).SetZero()
			case len(c.children) > 0:
				zeroUnkept(v.Field(i), c)
			}
		}
	default:
	}
}
