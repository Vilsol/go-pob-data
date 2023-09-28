package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type Essence struct {
	raw.Essence
}

var Essences []*Essence

func InitializeEssences(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "Essences", &Essences, nil, assetCache)
}
