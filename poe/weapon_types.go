package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type WeaponType struct {
	raw.WeaponType
}

var WeaponTypes []*WeaponType

func InitializeWeaponTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "WeaponTypes", &WeaponTypes, nil, assetCache)
}
