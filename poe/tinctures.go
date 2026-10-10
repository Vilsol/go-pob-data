package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type Tincture struct {
	raw.Tincture
}

var Tinctures []*Tincture

func InitializeTinctures(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "Tinctures", &Tinctures, nil, assetCache)
}

func (t *Tincture) GetBaseItem() *BaseItemType {
	return BaseItemTypes[t.BaseItem]
}
