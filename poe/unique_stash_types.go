package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type UniqueStashType struct {
	raw.UniqueStashType
}

var UniqueStashTypes []*UniqueStashType

var UniqueStashTypesByID map[string]*UniqueStashType

func InitializeUniqueStashTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "UniqueStashTypes", &UniqueStashTypes, func(count int64) {
		UniqueStashTypesByID = make(map[string]*UniqueStashType, count)
	}, assetCache, func(obj *UniqueStashType) {
		UniqueStashTypesByID[obj.ID] = obj
	})
}
