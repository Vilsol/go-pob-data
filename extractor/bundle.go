package extractor

import (
	"io/fs"

	"github.com/oriath-net/pogo/poefs/bundle"
	"github.com/pkg/errors"
)

func GetBundleLoader(source fs.FS) (fs.FS, error) {
	loader, err := bundle.NewLoader(source)
	return loader, errors.Wrap(err, "failed to make a new bundle loader")
}
