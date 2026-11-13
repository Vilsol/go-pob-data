package sidecar

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"
	"time"
)

// FS serves the files a sidecar index names out of the bundles under lower's Bundles2/, decompressing each bundle
// once. Names are matched lower-case, as the bundle loader does for paths since 3.22.
type FS struct {
	lower   fs.FS
	files   map[string]Entry
	bundles map[string][]byte
	mu      sync.Mutex
}

// NewFS keeps only entries at or above minConf.
func NewFS(lower fs.FS, entries []Entry, minConf Confidence) *FS {
	files := make(map[string]Entry, len(entries))
	for _, e := range entries {
		if e.Confidence >= minConf {
			files[strings.ToLower(e.Path)] = e
		}
	}
	return &FS{lower: lower, files: files, bundles: map[string][]byte{}}
}

func (f *FS) Open(name string) (fs.File, error) {
	e, ok := f.files[strings.ToLower(name)]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	blob, err := f.bundle(e.Bundle)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	if e.Offset+e.Size > int64(len(blob)) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fmt.Errorf("entry runs past bundle %s", e.Bundle)}
	}
	return &file{Reader: bytes.NewReader(blob[e.Offset : e.Offset+e.Size]), name: path.Base(name), size: e.Size}, nil
}

func (f *FS) bundle(name string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if b, ok := f.bundles[name]; ok {
		return b, nil
	}
	b, err := ReadBundle(f.lower, name)
	if err != nil {
		return nil, err
	}
	f.bundles[name] = b
	return b, nil
}

// ReadBundle reads and decompresses Bundles2/<name>.bundle.bin from lower.
func ReadBundle(lower fs.FS, name string) ([]byte, error) {
	h, err := lower.Open("Bundles2/" + name + ".bundle.bin")
	if err != nil {
		return nil, fmt.Errorf("bundle %s: %w", name, err)
	}
	defer h.Close()
	raw, err := io.ReadAll(h)
	if err != nil {
		return nil, fmt.Errorf("bundle %s: %w", name, err)
	}
	b, err := DecodeBundle(raw)
	if err != nil {
		return nil, fmt.Errorf("bundle %s: %w", name, err)
	}
	return b, nil
}

type file struct {
	*bytes.Reader
	name string
	size int64
}

func (f *file) Stat() (fs.FileInfo, error) { return info{f.name, f.size}, nil }
func (f *file) Close() error               { return nil }

type info struct {
	name string
	size int64
}

func (i info) Name() string       { return i.name }
func (i info) Size() int64        { return i.size }
func (i info) Mode() fs.FileMode  { return 0o444 }
func (i info) ModTime() time.Time { return time.Time{} }
func (i info) IsDir() bool        { return false }
func (i info) Sys() any           { return nil }
