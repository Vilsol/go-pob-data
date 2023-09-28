package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveTreeExpansionJewel struct {
	raw.PassiveTreeExpansionJewel
}

var PassiveTreeExpansionJewels []*PassiveTreeExpansionJewel

func InitializePassiveTreeExpansionJewels(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveTreeExpansionJewels", &PassiveTreeExpansionJewels, nil, assetCache)
}
