package poe

import (
	"context"
	"path"
	"strings"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type AdditionalLifeScaling struct {
	// PerLevel is the rows of the table DatFile names, nil when the version does not publish it.
	PerLevel []*raw.LifeScalingPerLevel
	raw.AdditionalLifeScaling
}

var AdditionalLifeScalings []*AdditionalLifeScaling

// AdditionalLifeScalingsByIntID is keyed by IntID, the value of the stat additional_life_scaling_index.
var AdditionalLifeScalingsByIntID map[int]*AdditionalLifeScaling

func InitializeAdditionalLifeScalings(ctx context.Context, version string, assetCache loader.AssetCache) error {
	err := loader.InitHelper(ctx, version, "AdditionalLifeScaling", &AdditionalLifeScalings, func(count int64) {
		AdditionalLifeScalingsByIntID = make(map[int]*AdditionalLifeScaling, count)
	}, assetCache, func(obj *AdditionalLifeScaling) {
		AdditionalLifeScalingsByIntID[obj.IntID] = obj
	})
	if err != nil {
		return err
	}

	for _, scaling := range AdditionalLifeScalings {
		table := strings.TrimSuffix(path.Base(scaling.DatFile), path.Ext(scaling.DatFile))
		published, err := loader.HasTable(ctx, version, table)
		if err != nil {
			return err
		}
		if !published {
			continue
		}
		scaling.PerLevel, err = loader.LoadRaw[*raw.LifeScalingPerLevel](ctx, version, table, nil, assetCache)
		if err != nil {
			return err
		}
	}

	return nil
}
