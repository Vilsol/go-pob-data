package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type FlavourText struct {
	raw.FlavourText
}

var FlavourTexts []*FlavourText

var FlavourTextsByID map[string]*FlavourText

func InitializeFlavourTexts(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "FlavourText", &FlavourTexts, func(count int64) {
		FlavourTextsByID = make(map[string]*FlavourText, count)
	}, assetCache, func(obj *FlavourText) {
		FlavourTextsByID[obj.ID] = obj
	})
}
