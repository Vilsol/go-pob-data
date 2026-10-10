package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveTreeExpansionJewelSize struct {
	raw.PassiveTreeExpansionJewelSize
}

var PassiveTreeExpansionJewelSizes []*PassiveTreeExpansionJewelSize

func InitializePassiveTreeExpansionJewelSizes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveTreeExpansionJewelSizes", &PassiveTreeExpansionJewelSizes, nil, assetCache)
}
