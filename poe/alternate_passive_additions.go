package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type AlternatePassiveAddition struct {
	raw.AlternatePassiveAddition
}

var AlternatePassiveAdditions []*AlternatePassiveAddition

func InitializeAlternatePassiveAdditions(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "AlternatePassiveAdditions", &AlternatePassiveAdditions, nil, assetCache)
}
