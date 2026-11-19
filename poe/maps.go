package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type Map struct {
	raw.Map
}

var Maps []*Map

func InitializeMaps(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "Maps", &Maps, nil, assetCache)
}

func (m *Map) GetRegularWorldArea() *WorldArea {
	if m.RegularWorldAreasKey == nil {
		return nil
	}

	return WorldAreas[*m.RegularWorldAreasKey]
}
