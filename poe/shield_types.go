package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ShieldType struct {
	raw.ShieldType
}

var ShieldTypes []*ShieldType

func InitializeShieldTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ShieldTypes", &ShieldTypes, nil, assetCache)
}
