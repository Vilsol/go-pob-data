package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ComponentAttributeRequirement struct {
	raw.ComponentAttributeRequirement
}

var ComponentAttributeRequirements []*ComponentAttributeRequirement

func InitializeComponentAttributeRequirements(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ComponentAttributeRequirements", &ComponentAttributeRequirements, nil, assetCache)
}
