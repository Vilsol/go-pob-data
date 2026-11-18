package loader

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Vilsol/slox"
	"github.com/pkg/errors"
)

// ErrAssetNotFound is returned when the CDN has no such asset. The set of tables the
// game ships changes between versions, so this is not fatal for every caller.
var ErrAssetNotFound = errors.New("asset not found")

// loadAsset returns the bytes of url, preferring the cache over the network.
//
// A cache entry that exists but holds no bytes is treated as a miss: it can only
// have come from a writer that died mid-write, and decoding it would otherwise
// look like a legitimately empty asset.
func loadAsset(ctx context.Context, url string, assetCache AssetCache) ([]byte, bool, error) {
	if assetCache.Exists(url) {
		b, err := assetCache.Get(url)
		if err != nil {
			return nil, false, errors.Wrap(err, "failed to retrieve url from cache: "+url)
		}

		if len(b) > 0 {
			return b, true, nil
		}

		slox.From(ctx).Warn(
			"discarding empty cached asset",
			slog.String("url", url),
		)
	}

	b, err := fetchAsset(ctx, url, assetCache)
	if err != nil {
		return nil, false, err
	}

	return b, false, nil
}

// fetchAttempts is how many times a download is tried before giving up. A CDN
// edge can reset a connection or answer 5xx for a moment; one such blip must not
// fail the whole data load.
const fetchAttempts = 3

// fetchBackoff is the wait before the second attempt; it doubles after that.
var fetchBackoff = 250 * time.Millisecond

// errRetryable marks a failure another attempt may cure: a transport error, a
// 5xx answer or a body cut short. A 404 or any other status is final.
type errRetryable struct{ error }

func (e errRetryable) Unwrap() error { return e.error }

// withRetry runs attempt up to fetchAttempts times while it fails retryably,
// waiting between attempts unless ctx ends first.
func withRetry[T any](ctx context.Context, attempt func() (T, error)) (T, error) {
	wait := fetchBackoff
	for i := 1; ; i++ {
		v, err := attempt()
		var r errRetryable
		if err == nil || !errors.As(err, &r) || i == fetchAttempts {
			if errors.As(err, &r) {
				err = r.error
			}
			return v, err
		}
		select {
		case <-ctx.Done():
			var zero T
			return zero, errors.Wrap(ctx.Err(), "waiting to retry")
		case <-time.After(wait):
		}
		wait *= 2
	}
}

// fetchAsset downloads url and stores it in the cache, replacing any entry
// already held for that key.
func fetchAsset(ctx context.Context, url string, assetCache AssetCache) ([]byte, error) {
	slox.From(ctx).Debug(
		"fetching",
		slog.String("url", url),
	)

	b, err := withRetry(ctx, func() ([]byte, error) {
		request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create request")
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, errRetryable{errors.Wrap(err, "failed to fetch url: "+url)}
		}
		defer response.Body.Close()

		// An error page is not an asset. Returning before the read keeps it out of the
		// cache, where it would decode as an empty table on every subsequent run.
		if response.StatusCode == http.StatusNotFound {
			return nil, errors.Wrap(ErrAssetNotFound, url)
		}

		if response.StatusCode >= http.StatusInternalServerError {
			return nil, errRetryable{errors.Errorf("unexpected status %d fetching url: %s", response.StatusCode, url)}
		}

		if response.StatusCode != http.StatusOK {
			return nil, errors.Errorf("unexpected status %d fetching url: %s", response.StatusCode, url)
		}

		b, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errRetryable{errors.Wrap(err, "failed to read response body")}
		}
		return b, nil
	})
	if err != nil {
		return nil, err
	}

	// Caching is best-effort: the caller already has valid bytes either way.
	if err := assetCache.Set(url, b); err != nil {
		slox.From(ctx).Warn(
			"failed to store asset in cache",
			slog.String("url", url),
			slog.Any("error", err),
		)
	}

	return b, nil
}

// assetExists asks the CDN whether url exists, without downloading it.
func assetExists(ctx context.Context, url string) (bool, error) {
	return withRetry(ctx, func() (bool, error) {
		return probeAsset(ctx, url)
	})
}

// probeAsset is one HEAD request for assetExists.
func probeAsset(ctx context.Context, url string) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false, errors.Wrap(err, "failed to create request")
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, errRetryable{errors.Wrap(err, "failed to fetch url: "+url)}
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusInternalServerError {
		return false, errRetryable{errors.Errorf("unexpected status %d probing url: %s", response.StatusCode, url)}
	}

	switch response.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}

	return false, errors.Errorf("unexpected status %d probing url: %s", response.StatusCode, url)
}
