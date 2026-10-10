package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type GemTag struct {
	raw.GemTag
}

var GemTags []*GemTag

var GemTagsByID map[string]*GemTag

func InitializeGemTags(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "GemTags", &GemTags, func(count int64) {
		GemTagsByID = make(map[string]*GemTag, count)
	}, assetCache, func(obj *GemTag) {
		GemTagsByID[obj.ID] = obj
	})
}
