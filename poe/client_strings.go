package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ClientString struct {
	raw.ClientString
}

var ClientStrings []*ClientString

var ClientStringsByID map[string]*ClientString

func InitializeClientStrings(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ClientStrings", &ClientStrings, func(count int64) {
		ClientStringsByID = make(map[string]*ClientString, count)
	}, assetCache, func(obj *ClientString) {
		ClientStringsByID[obj.ID] = obj
	})
}
