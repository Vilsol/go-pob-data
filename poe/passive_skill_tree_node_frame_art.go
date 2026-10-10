package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveSkillTreeNodeFrameArt struct {
	raw.PassiveSkillTreeNodeFrameArt
}

var PassiveSkillTreeNodeFrameArts []*PassiveSkillTreeNodeFrameArt

var PassiveSkillTreeNodeFrameArtsByID map[string]*PassiveSkillTreeNodeFrameArt

func InitializePassiveSkillTreeNodeFrameArts(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveSkillTreeNodeFrameArt", &PassiveSkillTreeNodeFrameArts, func(count int64) {
		PassiveSkillTreeNodeFrameArtsByID = make(map[string]*PassiveSkillTreeNodeFrameArt, count)
	}, assetCache, func(obj *PassiveSkillTreeNodeFrameArt) {
		PassiveSkillTreeNodeFrameArtsByID[obj.ID] = obj
	})
}
