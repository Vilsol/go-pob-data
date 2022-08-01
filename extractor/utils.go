package extractor

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/djherbis/fscache.v0"
)

var downloadClient *http.Client
var cache *fscache.FSCache

func init() {
	dir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	baseCacheDir := filepath.Join(dir, "go-pob", "bundle-cache")
	if err := os.MkdirAll(baseCacheDir, 0777); err != nil {
		if !os.IsExist(err) {
			panic(err)
		}
	}

	cache, err = fscache.New(baseCacheDir, 0755, time.Hour*24*30) // 30 day cache
	if err != nil {
		panic(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	downloadClient = &http.Client{Transport: tr}
}

func Fetch(url string) []byte {
	r, w, err := cache.Get(url)
	if err != nil {
		panic(err)
	}

	if w == nil {
		all, err := io.ReadAll(r)
		if err != nil {
			panic(err)
		}
		return all
	}

	resp, err := downloadClient.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	defer w.Close()
	if _, err := w.Write(all); err != nil {
		panic(err)
	}

	return all
}
