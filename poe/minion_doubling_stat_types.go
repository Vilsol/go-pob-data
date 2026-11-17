package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MinionDoublingStatType struct {
	raw.MinionDoublingStatType
}

var MinionDoublingStatTypes []*MinionDoublingStatType

func InitializeMinionDoublingStatTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MinionDoublingStatTypes", &MinionDoublingStatTypes, nil, assetCache)
}
