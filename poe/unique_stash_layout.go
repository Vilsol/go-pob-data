package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type UniqueStashLayout struct {
	raw.UniqueStashLayout
}

var UniqueStashLayouts []*UniqueStashLayout

func InitializeUniqueStashLayouts(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "UniqueStashLayout", &UniqueStashLayouts, nil, assetCache)
}

func (u *UniqueStashLayout) GetWord() *Word {
	return Words[u.WordsKey]
}

func (u *UniqueStashLayout) GetItemVisualIdentity() *ItemVisualIdentity {
	return ItemVisualIdentities[u.ItemVisualIdentityKey]
}

func (u *UniqueStashLayout) GetUniqueStashType() *UniqueStashType {
	return UniqueStashTypes[u.UniqueStashTypesKey]
}
