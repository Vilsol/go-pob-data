package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ShaperApparitionAbility struct {
	raw.ShaperApparitionAbility
}

var ShaperApparitionAbilities []*ShaperApparitionAbility

func InitializeShaperApparitionAbilities(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ShaperApparitionAbility", &ShaperApparitionAbilities, nil, assetCache)
}
