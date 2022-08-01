package extractor

import (
	"fmt"
	"io"
	"io/fs"
)

const gameVersion = "3.18.1.4"

const cdnTemplate = "https://patchcdn.pathofexile.com/%s/"

// TODO Use a proxy when pulled from a browser
// const cdnTemplate = "https://poe-bundles.snos.workers.dev/%s/"

var _ fs.FS = (*WebFS)(nil)

type WebFS struct {
	BasePath    string
	CachedFiles map[string]*WebFile
}

func NewWebFS() *WebFS {
	return &WebFS{
		CachedFiles: make(map[string]*WebFile),
		BasePath:    fmt.Sprintf(cdnTemplate, gameVersion),
	}
}

func (w *WebFS) Open(name string) (fs.File, error) {
	if _, ok := w.CachedFiles[name]; !ok {
		w.CachedFiles[name] = &WebFile{
			BasePath: w.BasePath,
			Name:     name,
		}
	}

	return w.CachedFiles[name], nil
}

var _ io.ReaderAt = (*WebFile)(nil)

type WebFile struct {
	BasePath   string
	Name       string
	CachedFile []byte
}

func (w *WebFile) ReadAt(p []byte, off int64) (n int, err error) {
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
		w.CachedFile = Fetch(w.BasePath + w.Name)
	}

	return copy(bytes, w.CachedFile), nil
}

func (w *WebFile) Close() error {
	// Do nothing
	return nil
}
