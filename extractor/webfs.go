package extractor

import (
	"fmt"
	"io"
	"io/fs"
)

const gameVersion = "3.18.1.5"

const cdnTemplate = "https://patchcdn.pathofexile.com/%s/"

var _ fs.FS = (*WebFS)(nil)

type WebFS struct {
	BasePath string
}

func NewWebFS() *WebFS {
	return &WebFS{
		BasePath: fmt.Sprintf(cdnTemplate, gameVersion),
	}
}

func (w *WebFS) Open(name string) (fs.File, error) {
	return &WebFile{
		BasePath: w.BasePath,
		Name:     name,
	}, nil
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
