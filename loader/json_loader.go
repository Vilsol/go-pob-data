package loader

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/Vilsol/slox"
	"github.com/pkg/errors"
)

// JSONTableURL is the URL LoadJSONTable reads the json form of version's table name from.
func JSONTableURL(version string, name string) string {
	return cdnBase + version + "/raw/" + name + ".json.gz"
}

// LoadJSONTable loads the json form of a raw table, which every version publishes, including the
// json-only ones (Manifest.JSONOnly) that LoadRaw cannot read. T is a row: a struct with json tags
// naming the columns it needs, map[string]any, or json.RawMessage. Foreign keys are row indices.
//
// A table the version's manifest does not list loads as empty; a listed table that cannot be fetched
// is an error.
//
// Returns data from cache if found
func LoadJSONTable[T any](ctx context.Context, version string, name string, assetCache AssetCache) ([]T, error) {
	url := JSONTableURL(version, name)

	slox.From(ctx).Debug(
		"loading json table",
		slog.String("version", version),
		slog.String("name", name),
	)

	b, fromCache, err := loadAsset(ctx, url, assetCache)
	if errors.Is(err, ErrAssetNotFound) {
		hasTable, manifestErr := HasTable(ctx, version, name)
		if manifestErr != nil {
			return nil, errors.Wrap(manifestErr, "could not tell whether "+version+" has "+name)
		}
		if hasTable {
			return nil, errors.Wrap(err, "the manifest lists the table but its json is missing")
		}

		slox.From(ctx).Warn(
			"no such table for this version",
			slog.String("version", version),
			slog.String("name", name),
		)

		return []T{}, nil
	}

	if err != nil {
		return nil, err
	}

	rows, err := decodeJSONTable[T](b)
	if err != nil && fromCache {
		// Nothing is returned to the caller until the whole table decodes, so a
		// cached asset that fails to decode is safe to throw away and fetch again.
		slox.From(ctx).Warn(
			"discarding unreadable cached asset",
			slog.String("url", url),
			slog.Any("error", err),
		)

		b, err = fetchAsset(ctx, url, assetCache)
		if err != nil {
			return nil, err
		}

		rows, err = decodeJSONTable[T](b)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json table: "+url)
	}

	return rows, nil
}

func decodeJSONTable[T any](b []byte) ([]T, error) {
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decompress")
	}

	var rows []T
	if err := json.NewDecoder(r).Decode(&rows); err != nil {
		return nil, errors.Wrap(err, "failed to decode")
	}

	// A truncated gzip stream can still hold a whole json array; the checksum is only read at EOF.
	if _, err := io.Copy(io.Discard, r); err != nil {
		return nil, errors.Wrap(err, "failed to decompress")
	}

	if rows == nil {
		return nil, errors.New("not a json array")
	}

	return rows, nil
}
