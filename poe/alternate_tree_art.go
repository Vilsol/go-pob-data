package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type AlternateTreeArt struct {
	raw.AlternateTreeArt
}

var AlternateTreeArts []*AlternateTreeArt

func InitializeAlternateTreeArts(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "AlternateTreeArt", &AlternateTreeArts, nil, assetCache)
}

func (a *AlternateTreeArt) GetConnectionArt() *PassiveSkillTreeConnectionArt {
	if a.ConnectionArt == nil {
		return nil
	}

	return PassiveSkillTreeConnectionArts[*a.ConnectionArt]
}

func (a *AlternateTreeArt) GetKeystoneFrameArt() *PassiveSkillTreeNodeFrameArt {
	if a.KeystoneFrameArt == nil {
		return nil
	}

	return PassiveSkillTreeNodeFrameArts[*a.KeystoneFrameArt]
}

func (a *AlternateTreeArt) GetNotableFrameArt() *PassiveSkillTreeNodeFrameArt {
	if a.NotableFrameArt == nil {
		return nil
	}

	return PassiveSkillTreeNodeFrameArts[*a.NotableFrameArt]
}

func (a *AlternateTreeArt) GetAscendancyFrameArt() *PassiveSkillTreeNodeFrameArt {
	if a.AscendancyFrameArt == nil {
		return nil
	}

	return PassiveSkillTreeNodeFrameArts[*a.AscendancyFrameArt]
}
