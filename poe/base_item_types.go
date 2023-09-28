package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type BaseItemType struct {
	raw.BaseItemType
}

var BaseItemTypes []*BaseItemType

var (
	BaseItemTypeByIDMap   map[string]*BaseItemType
	BaseItemTypeByNameMap map[string]*BaseItemType
)

func InitializeBaseItemTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "BaseItemTypes", &BaseItemTypes, func(count int64) {
		BaseItemTypeByIDMap = make(map[string]*BaseItemType, count)
		BaseItemTypeByNameMap = make(map[string]*BaseItemType, count)
	}, assetCache, func(obj *BaseItemType) {
		BaseItemTypeByIDMap[obj.ID] = obj
		BaseItemTypeByNameMap[obj.Name] = obj
	})
}

func (b *BaseItemType) SkillGem() *SkillGem {
	return skillGemsByBaseItemTypeMap[b.Key]
}
