package derived

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/poe2/pobdata"
)

// Row is a row type of this package.
type Row interface {
	Table() string
}

// Variant is the variant of T's assets: its tuple layout.
func Variant[T Row]() string {
	var row T
	return loader.TupleVariant(row)
}

// AssetPath is where T's asset of the pobdata.Encode bytes encoded lives under its version directory,
// named by T's layout and the hash of encoded: "derived/ModScalability.t-1a2b3c4d-5e6f7a8b.msgpack.br".
func AssetPath[T Row](encoded []byte) string {
	var row T
	sum := sha256.Sum256(encoded)
	return loader.AssetPath("derived/"+row.Table(), Variant[T]()+"-"+hex.EncodeToString(sum[:4]))
}

// Load is T's rows of version, from the asset its manifest names for T's layout. A version that has no
// such asset is loader.ErrAssetNotFound.
func Load[T Row, PT interface {
	*T
	msgp.Decodable
}](ctx context.Context, version string, assetCache loader.AssetCache) ([]T, error) {
	manifest, err := loader.LoadManifest(ctx, version)
	if err != nil {
		return nil, err
	}

	var row T
	path, ok := manifest.Derived[row.Table()][Variant[T]()]
	if !ok {
		return nil, errors.Wrap(loader.ErrAssetNotFound, "no "+row.Table()+" "+Variant[T]()+" in "+version)
	}

	return pobdata.Load[T, PT](ctx, version, path, assetCache)
}
