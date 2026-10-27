package main

import (
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/treegen"
)

func generateTreeData(loader fs.FS, game extractor.Game, gameVersion string) error {
	dir := filepath.Join("data", gameVersion, "tree")

	// Start clean so sheets that are no longer produced don't linger.
	if err := os.RemoveAll(filepath.Join(dir, "assets")); err != nil {
		return errors.Wrap(err, "failed to remove old tree assets")
	}

	var err error
	if game == extractor.PoE2 {
		// The PoE2 tree has no sprite sheets yet.
		_, err = treegen.WritePoE2(loader, gameVersion, dir)
	} else {
		_, err = treegen.Write(loader, gameVersion, dir)
	}
	if err != nil {
		return errors.Wrap(err, "failed to generate tree")
	}

	dataJSONPath := filepath.Join(dir, "data.json")
	dataJSON, err := os.ReadFile(dataJSONPath)
	if err != nil {
		return errors.Wrap(err, "failed to read tree data.json")
	}

	if err := writeCompressed(filepath.Join(dir, "data.json.gz"), dataJSON, func(w io.Writer) io.WriteCloser {
		return gzip.NewWriter(w)
	}); err != nil {
		return err
	}

	if err := writeCompressed(filepath.Join(dir, "data.json.br"), dataJSON, func(w io.Writer) io.WriteCloser {
		return brotli.NewWriter(w)
	}); err != nil {
		return err
	}

	return errors.Wrap(os.Remove(dataJSONPath), "failed to remove uncompressed tree data.json")
}

func writeCompressed(outPath string, data []byte, newWriter func(io.Writer) io.WriteCloser) error {
	f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return errors.Wrap(err, "could not open file for writing")
	}
	defer f.Close()

	w := newWriter(f)
	if _, err := w.Write(data); err != nil {
		return errors.Wrap(err, "failed writing compressed json")
	}

	if err := w.Close(); err != nil {
		return errors.Wrap(err, "failed closing compressed writer")
	}

	return errors.Wrap(f.Close(), "failed closing file")
}
