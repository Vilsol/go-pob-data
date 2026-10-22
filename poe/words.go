package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type Word struct {
	raw.Word
}

var Words []*Word

var wordsByText map[string][]*Word

func InitializeWords(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "Words", &Words, func(count int64) {
		wordsByText = make(map[string][]*Word, count)
	}, assetCache, func(obj *Word) {
		wordsByText[obj.Text] = append(wordsByText[obj.Text], obj)
	})
}
