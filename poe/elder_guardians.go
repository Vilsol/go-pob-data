package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ElderGuardian struct {
	raw.ElderGuardian
}

var ElderGuardians []*ElderGuardian

func InitializeElderGuardians(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ElderGuardians", &ElderGuardians, nil, assetCache)
}
