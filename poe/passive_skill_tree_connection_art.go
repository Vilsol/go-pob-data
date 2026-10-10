package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveSkillTreeConnectionArt struct {
	raw.PassiveSkillTreeConnectionArt
}

var PassiveSkillTreeConnectionArts []*PassiveSkillTreeConnectionArt

var PassiveSkillTreeConnectionArtsByID map[string]*PassiveSkillTreeConnectionArt

func InitializePassiveSkillTreeConnectionArts(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveSkillTreeConnectionArt", &PassiveSkillTreeConnectionArts, func(count int64) {
		PassiveSkillTreeConnectionArtsByID = make(map[string]*PassiveSkillTreeConnectionArt, count)
	}, assetCache, func(obj *PassiveSkillTreeConnectionArt) {
		PassiveSkillTreeConnectionArtsByID[obj.ID] = obj
	})
}
