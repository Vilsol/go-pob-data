package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PantheonPanelLayout struct {
	raw.PantheonPanelLayout
}

var PantheonPanelLayouts []*PantheonPanelLayout

func InitializePantheonPanelLayouts(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PantheonPanelLayout", &PantheonPanelLayouts, nil, assetCache)
}
