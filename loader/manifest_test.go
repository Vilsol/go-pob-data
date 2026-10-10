package loader

import (
	"context"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/testdata"
)

func TestHasTable(t *testing.T) {
	testdata.ServeCommittedData(t)
	ctx := context.Background()

	has, err := HasTable(ctx, testdata.LatestDataVersion(), "PassiveSkillTattoos")
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, has)

	has, err = HasTable(ctx, "3.28", "PassiveSkillTattoos")
	testza.AssertNoError(t, err)
	testza.AssertFalse(t, has, "3.28 was extracted before the table was published")

	has, err = HasTable(ctx, "3.28", "Mods")
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, has)

	_, err = HasTable(ctx, "0.0", "Mods")
	testza.AssertErrorIs(t, err, ErrAssetNotFound, "a version without a manifest cannot be told")

	m, err := LoadManifest(ctx, testdata.LatestDataVersion())
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, m.HasStatTranslation("skill_stat_descriptions"))
	testza.AssertTrue(t, m.HasFile("Metadata/Monsters/Monster.ot"))
}
