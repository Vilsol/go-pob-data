package pobdata

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/loader"
)

// Encode is rows as a tuple asset holds them before compression: an int64 row count, then each row.
func Encode[T any, PT interface {
	*T
	msgp.Encodable
}](rows []T) ([]byte, error) {
	var buf bytes.Buffer
	w := msgp.NewWriter(&buf)
	if err := w.WriteInt64(int64(len(rows))); err != nil {
		return nil, errors.Wrap(err, "failed to encode msgpack")
	}
	for i := range rows {
		if err := PT(&rows[i]).EncodeMsg(w); err != nil {
			return nil, errors.Wrap(err, "failed to encode msgpack")
		}
	}
	if err := w.Flush(); err != nil {
		return nil, errors.Wrap(err, "failed to flush msgpack")
	}
	return buf.Bytes(), nil
}

// Compress is encoded as the published asset's bytes.
func Compress(encoded []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	if _, err := w.Write(encoded); err != nil {
		return nil, errors.Wrap(err, "failed to compress")
	}
	if err := w.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to compress")
	}
	return buf.Bytes(), nil
}

// AssetPath is where table's asset of Encode's bytes encoded lives under its version directory, named
// by their hash: "pob/ItemMods.1a2b3c4d.msgpack.br".
func AssetPath(table string, encoded []byte) string {
	sum := sha256.Sum256(encoded)
	return loader.AssetPath("pob/"+table, hex.EncodeToString(sum[:4]))
}

// Load is the rows of the asset at path, an AssetPath, under version's directory.
func Load[T any, PT interface {
	*T
	msgp.Decodable
}](ctx context.Context, version string, path string, assetCache loader.AssetCache) ([]T, error) {
	rows, err := loader.LoadAssetRows[PT](ctx, version, path, assetCache)
	if err != nil {
		return nil, errors.Wrap(err, path)
	}
	out := make([]T, len(rows))
	for i, row := range rows {
		out[i] = *row
	}
	return out, nil
}
