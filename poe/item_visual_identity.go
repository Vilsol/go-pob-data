package poe

import (
	"context"
	"strings"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ItemVisualIdentity struct {
	raw.ItemVisualIdentity
}

var ItemVisualIdentities []*ItemVisualIdentity

var ItemVisualIdentitiesByID map[string]*ItemVisualIdentity

func InitializeItemVisualIdentities(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ItemVisualIdentity", &ItemVisualIdentities, func(count int64) {
		ItemVisualIdentitiesByID = make(map[string]*ItemVisualIdentity, count)
	}, assetCache, func(obj *ItemVisualIdentity) {
		ItemVisualIdentitiesByID[obj.ID] = obj
	})
}

// ArtPath is the item's 2D art without the .dds extension, e.g. "Art/2DItems/Belts/Headhunter"; image
// hosts that mirror the game's art serve it under that path.
func (v *ItemVisualIdentity) ArtPath() string {
	return strings.TrimSuffix(v.DDSFile, ".dds")
}
