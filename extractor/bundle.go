package extractor

import (
	"io/fs"

	"github.com/oriath-net/pogo/poefs/bundle"
	"github.com/pkg/errors"
)

func GetBundleLoader(source fs.FS, newHashFunc bool) (fs.FS, error) {
	loader, err := bundle.NewLoader(source, newHashFunc)
	return loader, errors.Wrap(err, "failed to make a new bundle loader")
}
