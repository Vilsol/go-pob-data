package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type AlternatePassiveSkill struct {
	raw.AlternatePassiveSkill
}

var AlternatePassiveSkills []*AlternatePassiveSkill

func InitializeAlternatePassiveSkills(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "AlternatePassiveSkills", &AlternatePassiveSkills, nil, assetCache)
}
