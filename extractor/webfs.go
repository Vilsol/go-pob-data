package extractor

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const cdnTemplate = "https://patch.poecdn.com/%s/"

var _ fs.FS = (*WebFS)(nil)

type WebFS struct {
	BasePath string
	// CacheDir keeps downloaded files on disk; empty disables caching.
	CacheDir string
}

func NewWebFS(gameVersion string) *WebFS {
	return &WebFS{
		BasePath: fmt.Sprintf(cdnTemplate, gameVersion),
		CacheDir: bundleCacheDir(gameVersion),
	}
}

// bundleCacheDir is $POB_BUNDLE_CACHE/<patch>, falling back to the user cache dir.
func bundleCacheDir(gameVersion string) string {
	root := os.Getenv("POB_BUNDLE_CACHE")
	if root == "" {
		userCache, err := os.UserCacheDir()
		if err != nil {
			return ""
		}
		root = filepath.Join(userCache, "go-pob-data", "bundles")
	}
	return filepath.Join(root, gameVersion)
}

func (w *WebFS) Open(name string) (fs.File, error) {
	return &WebFile{
		BasePath: w.BasePath,
		Name:     name,
		CacheDir: w.CacheDir,
	}, nil
}

var _ io.ReaderAt = (*WebFile)(nil)

type WebFile struct {
	BasePath   string
	Name       string
	CacheDir   string
	CachedFile []byte
}

func (w *WebFile) ReadAt(p []byte, off int64) (int, error) {
	if _, err := w.Read([]byte{}); err != nil {
		return 0, err
	}

	return copy(p, w.CachedFile[off:]), nil
}

func (w *WebFile) Stat() (fs.FileInfo, error) {
	// Do nothing
	return nil, nil
}

func (w *WebFile) Read(bytes []byte) (int, error) {
	if w.CachedFile == nil {
		w.CachedFile = w.fetch()
	}

	return copy(bytes, w.CachedFile), nil
}

func (w *WebFile) fetch() []byte {
	if w.CacheDir == "" {
		return Fetch(w.BasePath + w.Name)
	}

	cachePath := filepath.Join(w.CacheDir, filepath.FromSlash(w.Name))
	if data, err := os.ReadFile(cachePath); err == nil {
		return data
	}

	data := Fetch(w.BasePath + w.Name)

	// Write to a temp file and rename so an interrupted run never leaves a truncated bundle behind.
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return data
	}
	tmp, err := os.CreateTemp(filepath.Dir(cachePath), filepath.Base(cachePath)+".*.tmp")
	if err != nil {
		return data
	}
	_, err = tmp.Write(data)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), cachePath)
	}
	if err != nil {
		_ = os.Remove(tmp.Name())
	}

	return data
}

func (w *WebFile) Close() error {
	// Do nothing
	return nil
}
