package loader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	"github.com/pkg/errors"
)

// flakyServer fails its first `fails` requests with fail, then serves "ok".
func flakyServer(t *testing.T, fails int32, fail func(w http.ResponseWriter)) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if n.Add(1) <= fails {
			fail(w)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)
	return srv, &n
}

// resetConn drops the connection without an answer, as a CDN edge reset does.
func resetConn(w http.ResponseWriter) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err == nil {
		_ = conn.Close()
	}
}

func fastBackoff(t *testing.T) {
	t.Helper()
	old := fetchBackoff
	fetchBackoff = time.Millisecond
	t.Cleanup(func() { fetchBackoff = old })
}

func TestFetchAssetRetriesAResetAndA5xx(t *testing.T) {
	fastBackoff(t)
	srv, n := flakyServer(t, 1, resetConn)
	b, err := fetchAsset(context.Background(), srv.URL, noCache{})
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, "ok", string(b))
	testza.AssertEqual(t, int32(2), n.Load())

	srv, n = flakyServer(t, 2, func(w http.ResponseWriter) { w.WriteHeader(http.StatusBadGateway) })
	b, err = fetchAsset(context.Background(), srv.URL, noCache{})
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, "ok", string(b))
	testza.AssertEqual(t, int32(3), n.Load())
}

func TestFetchAssetGivesUpAfterItsAttempts(t *testing.T) {
	fastBackoff(t)
	srv, n := flakyServer(t, 100, func(w http.ResponseWriter) { w.WriteHeader(http.StatusServiceUnavailable) })
	_, err := fetchAsset(context.Background(), srv.URL, noCache{})
	testza.AssertNotNil(t, err)
	testza.AssertEqual(t, int32(fetchAttempts), n.Load())
}

func TestFetchAssetNeverRetriesA404(t *testing.T) {
	fastBackoff(t)
	srv, n := flakyServer(t, 100, func(w http.ResponseWriter) { w.WriteHeader(http.StatusNotFound) })
	_, err := fetchAsset(context.Background(), srv.URL, noCache{})
	testza.AssertTrue(t, errors.Is(err, ErrAssetNotFound))
	testza.AssertEqual(t, int32(1), n.Load())

	exists, err := assetExists(context.Background(), srv.URL)
	testza.AssertNoError(t, err)
	testza.AssertFalse(t, exists)
	testza.AssertEqual(t, int32(2), n.Load())
}
