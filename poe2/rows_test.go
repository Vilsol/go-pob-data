package poe2

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/testdata"
)

const version = "0.5"

type nopCache struct{}

func (nopCache) Get(string) ([]byte, error) { return nil, nil }
func (nopCache) Set(string, []byte) error   { return nil }
func (nopCache) Exists(string) bool         { return false }

// checkRows decodes every row of table as T and as a map, and requires each column T declares to be
// present in the row with the same value, so a misspelt tag or a wrong type cannot decode as zero.
func checkRows[T any](t *testing.T, table string) {
	t.Helper()
	ctx := context.Background()

	rows, err := loader.LoadJSONTable[T](ctx, version, table, nopCache{})
	if err != nil {
		t.Fatal(err)
	}
	generic, err := loader.LoadJSONTable[map[string]any](ctx, version, table, nopCache{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 || len(rows) != len(generic) {
		t.Fatalf("%s: %d typed rows, %d rows", table, len(rows), len(generic))
	}

	for i, row := range rows {
		b, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		var declared map[string]any
		if err := json.Unmarshal(b, &declared); err != nil {
			t.Fatal(err)
		}
		if declared["_key"] != float64(i) {
			t.Fatalf("%s row %d has key %v", table, i, declared["_key"])
		}
		for column, value := range declared {
			got, ok := generic[i][column]
			if !ok {
				t.Fatalf("%s has no column %s", table, column)
			}
			if !reflect.DeepEqual(value, got) && !(isEmptyList(value) && isEmptyList(got)) {
				t.Fatalf("%s row %d column %s: decoded %v, published %v", table, i, column, value, got)
			}
		}
	}
	t.Logf("%s: %d rows", table, len(rows))
}

func isEmptyList(v any) bool {
	l, ok := v.([]any)
	return v == nil || (ok && len(l) == 0)
}

func checkAll(t *testing.T) {
	t.Helper()
	checkRows[GameConstant](t, GameConstantsTable)
	checkRows[PlayerMinionIntrinsicStat](t, PlayerMinionIntrinsicStatsTable)
	checkRows[SoulCore](t, SoulCoresTable)
	checkRows[SoulCoreStat](t, SoulCoreStatsTable)
	checkRows[SoulCoreType](t, SoulCoreTypesTable)
	checkRows[SoulCoreLimit](t, SoulCoreLimitsTable)
	checkRows[SoulCoreStatCategory](t, SoulCoreStatCategoriesTable)
	checkRows[QuestStaticReward](t, QuestStaticRewardsTable)
	checkRows[AlternatePassiveSkill](t, AlternatePassiveSkillsTable)
	checkRows[AlternatePassiveAddition](t, AlternatePassiveAdditionsTable)
	checkRows[Stat](t, StatsTable)
	checkRows[BaseItemType](t, BaseItemTypesTable)
	checkRows[ActiveSkillType](t, ActiveSkillTypeTable)
	checkRows[ActiveSkill](t, ActiveSkillsTable)
	checkRows[ActiveSkillWeaponRequirement](t, ActiveSkillWeaponRequirementTable)
	checkRows[WieldableClass](t, WieldableClassesTable)
	checkRows[ItemClass](t, ItemClassesTable)
	checkRows[ItemClassCategory](t, ItemClassCategoriesTable)
	checkRows[GrantedEffect](t, GrantedEffectsTable)
	checkRows[GrantedEffectPerLevel](t, GrantedEffectsPerLevelTable)
	checkRows[GrantedEffectStatSet](t, GrantedEffectStatSetsTable)
	checkRows[GrantedEffectStatSetPerLevel](t, GrantedEffectStatSetsPerLevelTable)
	checkRows[GrantedEffectQualityStat](t, GrantedEffectQualityStatsTable)
	checkRows[GrantedEffectLabel](t, GrantedEffectLabelsTable)
	checkRows[GemEffect](t, GemEffectsTable)
	checkRows[GemTag](t, GemTagsTable)
	checkRows[SkillGem](t, SkillGemsTable)
	checkRows[SupportGem](t, SupportGemsTable)
	checkRows[SupportGemFamily](t, SupportGemFamilyTable)
	checkRows[ItemExperiencePerLevel](t, ItemExperiencePerLevelTable)
	checkRows[CostType](t, CostTypesTable)
	checkRows[EffectivenessCostConstant](t, EffectivenessCostConstantsTable)
	checkRows[FlavourText](t, FlavourTextTable)
	checkRows[ClientString](t, ClientStringsTable)
	checkRows[ArmourType](t, ArmourTypesTable)
	checkRows[ShieldType](t, ShieldTypesTable)
	checkRows[WeaponType](t, WeaponTypesTable)
	checkRows[Flask](t, FlasksTable)
	checkRows[ComponentCharge](t, ComponentChargesTable)
	checkRows[UtilityFlaskBuff](t, UtilityFlaskBuffsTable)
	checkRows[BuffDefinition](t, BuffDefinitionsTable)
	checkRows[AttributeRequirement](t, AttributeRequirementsTable)
	checkRows[ItemSpirit](t, ItemSpiritTable)
	checkRows[ItemInherentSkill](t, ItemInherentSkillsTable)
	checkRows[ItemExperienceType](t, ItemExperienceTypesTable)
	checkRows[Tag](t, TagsTable)
	checkRows[SkillGemSupport](t, SkillGemSupportsTable)
}

func TestRows(t *testing.T) {
	testdata.ServeCommittedData(t)
	checkAll(t)
}

func TestRowsLive(t *testing.T) {
	if testing.Short() {
		t.Skip("reads data.pob.gg")
	}
	checkAll(t)
}
