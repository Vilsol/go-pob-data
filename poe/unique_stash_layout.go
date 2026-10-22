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

// uniqueStashLayoutByWord holds a unique's first layout row, preferring one that is not alternate art.
var uniqueStashLayoutByWord map[int]*UniqueStashLayout

func InitializeUniqueStashLayouts(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "UniqueStashLayout", &UniqueStashLayouts, func(count int64) {
		uniqueStashLayoutByWord = make(map[int]*UniqueStashLayout, count)
	}, assetCache, func(obj *UniqueStashLayout) {
		if prev, ok := uniqueStashLayoutByWord[obj.WordsKey]; !ok || (prev.IsAlternateArt && !obj.IsAlternateArt) {
			uniqueStashLayoutByWord[obj.WordsKey] = obj
		}
	})
}

// UniqueStashLayoutByName is the layout row of the unique named name, or nil. A unique with several
// variants (Impresence, Grand Spectrum) answers with its first. Needs Words and UniqueStashLayout.
func UniqueStashLayoutByName(name string) *UniqueStashLayout {
	for _, word := range wordsByText[name] {
		if layout, ok := uniqueStashLayoutByWord[word.Key]; ok {
			return layout
		}
	}
	return nil
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
