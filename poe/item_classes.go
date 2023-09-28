package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ItemClass struct {
	raw.ItemClass
}

var ItemClasses []*ItemClass

func InitializeItemClasses(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ItemClasses", &ItemClasses, nil, assetCache)
}
