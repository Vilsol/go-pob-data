package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveJewelSlot struct {
	raw.PassiveJewelSlot
}

var PassiveJewelSlots []*PassiveJewelSlot

func InitializePassiveJewelSlots(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveJewelSlots", &PassiveJewelSlots, nil, assetCache)
}

func (p *PassiveJewelSlot) GetClusterJewelSize() *PassiveTreeExpansionJewelSize {
	if p.ClusterJewelSize == nil {
		return nil
	}

	return PassiveTreeExpansionJewelSizes[*p.ClusterJewelSize]
}
