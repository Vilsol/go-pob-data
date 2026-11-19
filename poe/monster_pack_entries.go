package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MonsterPackEntry struct {
	raw.MonsterPackEntry
}

var MonsterPackEntries []*MonsterPackEntry

// MonsterPackEntriesByPack is keyed by MonsterPacksKey.
var MonsterPackEntriesByPack map[int][]*MonsterPackEntry

func InitializeMonsterPackEntries(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MonsterPackEntries", &MonsterPackEntries, func(int64) {
		MonsterPackEntriesByPack = make(map[int][]*MonsterPackEntry)
	}, assetCache, func(obj *MonsterPackEntry) {
		if obj.MonsterPacksKey != nil {
			MonsterPackEntriesByPack[*obj.MonsterPacksKey] = append(MonsterPackEntriesByPack[*obj.MonsterPacksKey], obj)
		}
	})
}

func (e *MonsterPackEntry) GetMonsterVariety() *MonsterVariety {
	if e.MonsterVarietiesKey == nil {
		return nil
	}

	return MonsterVarieties[*e.MonsterVarietiesKey]
}
