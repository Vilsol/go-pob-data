package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type Tag struct {
	raw.Tag
}

var (
	Tags        []*Tag
	TagIDsToTag map[string]*Tag
)

func InitializeTags(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "Tags", &Tags, func(count int64) {
		TagIDsToTag = make(map[string]*Tag, count)
	}, assetCache, func(tag *Tag) {
		TagIDsToTag[tag.ID] = tag
	})
}
