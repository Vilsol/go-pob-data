package testdata

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

const cdnBase = "https://data.pob.gg/data/"

// CDN serves data.pob.gg from the committed data/ tree, so tests read the artifacts of this
// checkout rather than whatever was last published.
type CDN struct {
	dir    string
	misses []string
	mu     sync.Mutex
}

// ServeCommittedData points http.DefaultClient at the committed data/ tree for the duration of t.
func ServeCommittedData(t testing.TB) *CDN {
	t.Helper()

	_, file, _, _ := runtime.Caller(0)
	cdn := &CDN{dir: filepath.Join(filepath.Dir(file), "..", "data")}

	previous := http.DefaultClient.Transport
	http.DefaultClient.Transport = cdn
	t.Cleanup(func() {
		http.DefaultClient.Transport = previous
	})

	return cdn
}

// Misses lists every URL that was answered with a 404.
func (c *CDN) Misses() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.misses...)
}

func (c *CDN) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	code := http.StatusOK

	var b []byte
	if !strings.HasPrefix(url, cdnBase) {
		code = http.StatusNotFound
	} else {
		var err error
		b, err = os.ReadFile(filepath.Join(c.dir, filepath.FromSlash(strings.TrimPrefix(url, cdnBase))))
		if os.IsNotExist(err) {
			code = http.StatusNotFound
		} else if err != nil {
			return nil, err
		}
	}

	if bytes.HasPrefix(b, []byte("version https://git-lfs")) {
		panic("git lfs pointer, run git lfs pull: " + url)
	}

	if code == http.StatusNotFound {
		c.mu.Lock()
		c.misses = append(c.misses, url)
		c.mu.Unlock()
		b = nil
	}

	if req.Method == http.MethodHead {
		b = nil
	}

	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}
