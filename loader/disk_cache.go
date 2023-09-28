//go:build !js

package loader

import (
	"io"
	"log/slog"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/djherbis/fscache.v0"
)

type diskCache struct {
	cache    *fscache.FSCache
	basePath string
}

func DiskCache(basePath string, cacheTime time.Duration) (AssetCache, error) {
	cache, err := fscache.New(basePath, 0o755, cacheTime)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cache")
	}

	return &diskCache{
		basePath: basePath,
		cache:    cache,
	}, nil
}

func (d *diskCache) Get(key string) ([]byte, error) {
	slog.Debug("loading from cache", slog.String("key", key))

	r, _, err := d.cache.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get key from cache: "+key)
	}

	if r == nil {
		return nil, nil
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from cache")
	}

	return b, nil
}

func (d *diskCache) Set(key string, value []byte) error {
	slog.Debug("storing in cache", slog.String("key", key), slog.Int("len", len(value)))

	_ = d.cache.Remove(key)

	_, w, err := d.cache.Get(key)
	if err != nil {
		return errors.Wrap(err, "failed to set key on cache: "+key)
	}

	if w == nil {
		return errors.New("could not write to cache")
	}

	defer w.Close()

	if _, err := w.Write(value); err != nil {
		return errors.Wrap(err, "failed to write to cache")
	}

	return nil
}

func (d *diskCache) Exists(key string) bool {
	return d.cache.Exists(key)
}
