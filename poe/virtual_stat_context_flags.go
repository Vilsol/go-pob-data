package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type VirtualStatContextFlag struct {
	raw.VirtualStatContextFlag
}

var VirtualStatContextFlags []*VirtualStatContextFlag

var VirtualStatContextFlagsByID map[string]*VirtualStatContextFlag

func InitializeVirtualStatContextFlags(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "VirtualStatContextFlags", &VirtualStatContextFlags, func(count int64) {
		VirtualStatContextFlagsByID = make(map[string]*VirtualStatContextFlag, count)
	}, assetCache, func(obj *VirtualStatContextFlag) {
		VirtualStatContextFlagsByID[obj.ID] = obj
	})
}
