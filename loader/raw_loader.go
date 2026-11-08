package loader

import (
	"bytes"
	"context"
	"log/slog"
	"reflect"

	"github.com/Vilsol/slox"
	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// LoadRaw loads a raw brotli-compressed json dump from remote source
//
// Returns data from cache if found
func LoadRaw[T msgp.Decodable](ctx context.Context, version string, name string, onInit func(count int64), assetCache AssetCache, hooks ...func(obj T)) ([]T, error) {
	var row T
	url := VariantURL(ctx, version, "raw/"+name, row)

	slox.From(ctx).Debug(
		"loading raw asset",
		slog.String("version", version),
		slog.String("name", name),
	)

	b, fromCache, err := loadAsset(ctx, url, assetCache)
	if errors.Is(err, ErrAssetNotFound) {
		// A version that has the table always has its json. Without it, the table is absent from
		// this version; with it, T's layout was never published for it, and an empty table would hide that.
		hasTable, probeErr := assetExists(ctx, cdnBase+version+"/raw/"+name+".json.gz")
		if probeErr != nil {
			return nil, probeErr
		}
		if hasTable {
			return nil, errors.Wrap(err, "table exists but has no asset for this layout")
		}

		slox.From(ctx).Warn(
			"no such asset for this version",
			slog.String("version", version),
			slog.String("name", name),
		)

		if onInit != nil {
			onInit(0)
		}

		return []T{}, nil
	}

	if err != nil {
		return nil, err
	}

	return decodeRows[T](ctx, url, b, fromCache, onInit, assetCache, hooks...)
}

// LoadAssetRows loads the rows of the tuple asset at path under version's directory, such as
// "pob/ItemMods.1a2b3c4d.msgpack.br": an int64 row count, then the rows. A missing asset is
// ErrAssetNotFound.
func LoadAssetRows[T msgp.Decodable](ctx context.Context, version string, path string, assetCache AssetCache) ([]T, error) {
	url := cdnBase + version + "/" + path

	slox.From(ctx).Debug(
		"loading asset rows",
		slog.String("version", version),
		slog.String("path", path),
	)

	b, fromCache, err := loadAsset(ctx, url, assetCache)
	if err != nil {
		return nil, err
	}

	return decodeRows[T](ctx, url, b, fromCache, nil, assetCache)
}

// decodeRows decodes b, the brotli tuple asset at url: an int64 row count, then the rows.
func decodeRows[T msgp.Decodable](ctx context.Context, url string, b []byte, fromCache bool, onInit func(count int64), assetCache AssetCache, hooks ...func(obj T)) ([]T, error) {
	r := msgp.NewReader(brotli.NewReader(bytes.NewReader(b)))

	elementCount, err := r.ReadInt64()
	if err != nil && fromCache {
		// Nothing has been emitted yet, so a cached asset we cannot even read a
		// header out of is safe to throw away and fetch again.
		slox.From(ctx).Warn(
			"discarding unreadable cached asset",
			slog.String("url", url),
			slog.Any("error", err),
		)

		b, err = fetchAsset(ctx, url, assetCache)
		if err != nil {
			return nil, err
		}

		r = msgp.NewReader(brotli.NewReader(bytes.NewReader(b)))
		elementCount, err = r.ReadInt64()
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to read element count: "+url)
	}

	if elementCount < 0 {
		return nil, errors.Errorf("negative element count %d: %s", elementCount, url)
	}

	if onInit != nil {
		onInit(elementCount)
	}

	out := make([]T, elementCount)
	if elementCount > 0 {
		elemType := reflect.ValueOf(out[0]).Type().Elem()
		for i := int64(0); i < elementCount; i++ {
			out[i] = reflect.New(elemType).Interface().(T)
			if err := out[i].DecodeMsg(r); err != nil {
				return nil, errors.Wrap(err, "failed to read message")
			}
			for _, hook := range hooks {
				hook(out[i])
			}
		}
	}

	return out, nil
}

func InitHelper[T msgp.Decodable](ctx context.Context, version string, name string, target *[]T, onInit func(count int64), assetCache AssetCache, hooks ...func(obj T)) error {
	loadedRaw, err := LoadRaw[T](ctx, version, name, onInit, assetCache, hooks...)
	if err != nil {
		return err
	}

	*target = loadedRaw

	return nil
}
