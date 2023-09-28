package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ArmourType struct {
	raw.ArmourType
}

var ArmourTypes []*ArmourType

func InitializeArmourTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ArmourTypes", &ArmourTypes, nil, assetCache)
}
