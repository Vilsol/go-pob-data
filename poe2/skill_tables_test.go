package poe2

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"testing"

	"github.com/Vilsol/go-pob-data/testdata"
)

// pobSkill is a skill as PoB-PoE2's src/Data/Skills/*.lua holds it. testdata/pob_skill_levels.json is
// three skills of each act_*/other/sup_* file of the pin, dumped with luajit.
type pobSkill struct {
	Set *struct {
		ID                             string     `json:"id"`
		Stats                          []string   `json:"stats"`
		ConstantStats                  [][]any    `json:"constantStats"`
		BaseEffectiveness              *float64   `json:"baseEffectiveness"`
		IncrementalEffectiveness       *float64   `json:"incrementalEffectiveness"`
		DamageIncrementalEffectiveness *float64   `json:"damageIncrementalEffectiveness"`
		Levels                         []pobLevel `json:"levels"`
	} `json:"set"`
	ID           string           `json:"id"`
	File         string           `json:"file"`
	Levels       []map[string]any `json:"levels"`
	QualityStats [][]any          `json:"qualityStats"`
	CastTime     float64          `json:"castTime"`
	Support      bool             `json:"support"`
}

type pobLevel struct {
	Values     []float64 `json:"values"`
	ActorLevel float64   `json:"actorLevel"`
	Level      int       `json:"level"`
}

func near(a, b float64) bool {
	return float32(a) == float32(b) || math.Abs(a-b) <= 1e-9*math.Max(1, math.Abs(b))
}

// levelExtras is skills.lua:238-300 for one level: the numbers PoB writes beside each level.
func levelExtras(s *SkillTables, granted *GrantedEffect, levelRow *GrantedEffectPerLevel, statRow *GrantedEffectStatSetPerLevel, levelRequirement int) map[string]any {
	extra := map[string]any{"level": float64(levelRow.Level), "levelRequirement": float64(levelRequirement)}
	cost := map[string]any{}
	for i, costType := range granted.CostTypes {
		cost[s.CostTypes[costType].ID] = float64(levelRow.CostAmounts[i])
	}
	extra["cost"] = cost
	set := func(name string, ok bool, v float64) {
		if ok {
			extra[name] = v
		}
	}
	set("spiritReservationFlat", levelRow.Reservation != 0, float64(levelRow.Reservation))
	set("reservationMultiplier", levelRow.EffectOnPlayer != 100, float64(levelRow.EffectOnPlayer-100))
	set("manaMultiplier", levelRow.CostMultiplier != 100, float64(levelRow.CostMultiplier-100))
	set("attackSpeedMultiplier", levelRow.AttackSpeedMultiplier != 0, float64(levelRow.AttackSpeedMultiplier))
	set("attackTime", levelRow.AttackTime != 0, float64(levelRow.AttackTime))
	set("cooldown", levelRow.Cooldown != 0, float64(levelRow.Cooldown)/1000)
	set("PvPDamageMultiplier", levelRow.PvPDamageMultiplier != 0, float64(levelRow.PvPDamageMultiplier))
	set("storedUses", levelRow.StoredUses != 0, float64(levelRow.StoredUses))
	if statRow != nil {
		set("critChance", statRow.SpellCritChance != 0, float64(statRow.SpellCritChance)/100)
		set("critChance", statRow.AttackCritChance != 0, float64(statRow.AttackCritChance)/100)
		set("baseMultiplier", statRow.BaseMultiplier != 0, float64(statRow.BaseMultiplier)/10000+1)
	}
	return extra
}

func sameExtras(got map[string]any, want map[string]any) error {
	keys := map[string]bool{}
	for k := range got {
		keys[k] = true
	}
	for k := range want {
		keys[k] = true
	}
	for k := range keys {
		g, w := got[k], want[k]
		if k == "cost" {
			gc, _ := g.(map[string]any)
			wc, _ := w.(map[string]any) // an empty Lua table dumps as []
			if len(gc) != len(wc) {
				return fmt.Errorf("cost %v, PoB %v", gc, w)
			}
			for r, v := range gc {
				if wv, ok := wc[r].(float64); !ok || !near(v.(float64), wv) {
					return fmt.Errorf("cost %v, PoB %v", gc, wc)
				}
			}
			continue
		}
		gf, gok := g.(float64)
		wf, wok := w.(float64)
		if gok != wok || !near(gf, wf) {
			return fmt.Errorf("%s %v, PoB %v", k, g, w)
		}
	}
	return nil
}

// checkSkill compares what the rows give for one skill with PoB's numbers, and returns the first
// disagreement.
func checkSkill(s *SkillTables, want pobSkill) error {
	granted := s.GrantedEffectByID(want.ID)
	if granted == nil {
		return fmt.Errorf("no granted effect")
	}
	if !near(float64(granted.CastTime)/1000, want.CastTime) && !granted.IsSupport {
		return fmt.Errorf("castTime %v, PoB %v", float64(granted.CastTime)/1000, want.CastTime)
	}

	perLevel := s.PerLevel(granted.Key)
	gemLevels := len(perLevel)
	var skillGem *SkillGem
	if gemEffect := s.GemEffect(granted.Key); gemEffect != nil {
		skillGem = s.SkillGem(gemEffect.Key)
	}
	if skillGem != nil && granted.IsSupport {
		gemLevels = 1
	}
	var progression []*ItemExperiencePerLevel
	if skillGem != nil {
		progression = s.GemLevelProgression(skillGem.Key)
	}
	statRows := s.LevelStatRows(granted.Key, gemLevels)

	if gemLevels != len(want.Levels) {
		return fmt.Errorf("%d levels, PoB %d", gemLevels, len(want.Levels))
	}
	requirement := 0
	for i := 0; i < gemLevels; i++ {
		if i < len(progression) {
			requirement = max(progression[i].Level, requirement)
		}
		var statRow *GrantedEffectStatSetPerLevel
		if i < len(statRows) {
			statRow = statRows[i]
		}
		if err := sameExtras(levelExtras(s, granted, perLevel[i], statRow, requirement), want.Levels[i]); err != nil {
			return fmt.Errorf("level %d: %w", perLevel[i].Level, err)
		}
	}

	if !(skillGem != nil && granted.IsSupport) {
		var quality [][]any
		if q := s.QualityStats(granted.Key); q != nil {
			for i, stat := range q.Stats {
				quality = append(quality, []any{s.Stats[stat].ID, float64(q.StatsValuesPermille[i]) / 1000, q.ApplyToStatSets})
			}
		}
		if len(quality) != len(want.QualityStats) {
			return fmt.Errorf("quality stats %v, PoB %v", quality, want.QualityStats)
		}
		for i, q := range quality {
			w := want.QualityStats[i]
			if q[0] != w[0] || !near(q[1].(float64), w[1].(float64)) || fmt.Sprint(q[2]) != fmt.Sprint(w[2]) {
				return fmt.Errorf("quality stat %d %v, PoB %v", i, q, w)
			}
		}
	}

	if want.Set == nil {
		return nil
	}
	return checkFirstSet(s, want)
}

// checkFirstSet is skills.lua's #set for the first set of a skill (skills.lua:487-) on sets whose stats
// do not change between levels.
func checkFirstSet(s *SkillTables, want pobSkill) error {
	set := s.StatSetByID(want.Set.ID)
	if set == nil {
		return fmt.Errorf("no stat set %s", want.Set.ID)
	}
	for _, e := range []struct {
		pobSide *float64
		name    string
		got     float64
		skip    float64
	}{
		{want.Set.BaseEffectiveness, "baseEffectiveness", set.BaseEffectiveness, 1},
		{want.Set.IncrementalEffectiveness, "incrementalEffectiveness", set.IncrementalEffectiveness, 0},
		{want.Set.DamageIncrementalEffectiveness, "damageIncrementalEffectiveness", set.DamageIncrementalEffectiveness, 0},
	} {
		if (e.got != e.skip) != (e.pobSide != nil) || (e.pobSide != nil && !near(e.got, *e.pobSide)) {
			return fmt.Errorf("%s %v, PoB %v", e.name, e.got, e.pobSide)
		}
	}

	if len(set.ConstantStats) != len(want.Set.ConstantStats) {
		return fmt.Errorf("%d constant stats, PoB %d", len(set.ConstantStats), len(want.Set.ConstantStats))
	}
	for i, stat := range set.ConstantStats {
		w := want.Set.ConstantStats[i]
		if s.Stats[stat].ID != w[0] || float64(set.ConstantStatsValues[i]) != w[1] {
			return fmt.Errorf("constant stat %d %s %d, PoB %v", i, s.Stats[stat].ID, set.ConstantStatsValues[i], w)
		}
	}

	levels := s.StatSetPerLevel(set.Key)
	if len(levels) != len(want.Set.Levels) {
		return fmt.Errorf("%d set levels, PoB %d", len(levels), len(want.Set.Levels))
	}
	var stats []string
	for i, row := range levels {
		var values []float64
		var rowStats []string
		for j, stat := range row.FloatStats {
			rowStats = append(rowStats, s.Stats[stat].ID)
			if len(levels) > 5 {
				values = append(values, float64(row.BaseResolvedValues[j]))
			} else {
				values = append(values, row.FloatStatsValues[j]/math.Max(s.EffectivenessCostConstants[row.InterpolationBases[j]].Multiplier, 0.00001))
			}
		}
		for j, stat := range row.AdditionalStats {
			rowStats = append(rowStats, s.Stats[stat].ID)
			values = append(values, float64(row.AdditionalStatsValues[j]))
		}
		for _, stat := range row.AdditionalFlags {
			if !slices.Contains(rowStats, s.Stats[stat].ID) {
				rowStats = append(rowStats, s.Stats[stat].ID)
			}
		}
		if i == 0 {
			stats = rowStats
		}
		w := want.Set.Levels[i]
		if row.GemLevel != w.Level || !near(row.ActorLevel, w.ActorLevel) {
			return fmt.Errorf("set level %d actor level %v, PoB level %d actor level %v", row.GemLevel, row.ActorLevel, w.Level, w.ActorLevel)
		}
		if len(values) != len(w.Values) {
			return fmt.Errorf("set level %d values %v, PoB %v", row.GemLevel, values, w.Values)
		}
		for j := range values {
			if !near(values[j], w.Values[j]) {
				return fmt.Errorf("set level %d values %v, PoB %v", row.GemLevel, values, w.Values)
			}
		}
	}
	for _, stat := range set.ImplicitStats {
		if !slices.Contains(stats, s.Stats[stat].ID) {
			stats = append(stats, s.Stats[stat].ID)
		}
	}
	if !slices.Equal(stats, want.Set.Stats) {
		return fmt.Errorf("stats %v, PoB %v", stats, want.Set.Stats)
	}
	return nil
}

func TestSkillLevelsMatchPoB(t *testing.T) {
	testdata.ServeCommittedData(t)

	b, err := os.ReadFile("testdata/pob_skill_levels.json")
	if err != nil {
		t.Fatal(err)
	}
	var want []pobSkill
	if err := json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}

	s, err := LoadSkillTables(context.Background(), version, nopCache{})
	if err != nil {
		t.Fatal(err)
	}

	for _, skill := range want {
		t.Run(skill.ID, func(t *testing.T) {
			if err := checkSkill(s, skill); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSkillForeignKeys(t *testing.T) {
	testdata.ServeCommittedData(t)

	s, err := LoadSkillTables(context.Background(), version, nopCache{})
	if err != nil {
		t.Fatal(err)
	}

	checked := 0
	in := func(table string, n int, keys ...int) {
		t.Helper()
		for _, k := range keys {
			checked++
			if k < 0 || k >= n {
				t.Errorf("%s key %d out of %d rows", table, k, n)
			}
		}
	}
	opt := func(table string, n int, key *int) {
		t.Helper()
		if key != nil {
			in(table, n, *key)
		}
	}
	stats, types := len(s.Stats), len(s.ActiveSkillTypes)
	for _, r := range s.ActiveSkillTypes {
		opt(StatsTable, stats, r.FlagStat)
	}
	for _, r := range s.ActiveSkills {
		in(ActiveSkillTypeTable, types, r.ActiveSkillTypes...)
		in(ActiveSkillTypeTable, types, r.MinionActiveSkillTypes...)
		opt(ActiveSkillWeaponRequirementTable, len(s.ActiveSkillWeaponRequirements), r.WeaponRequirements)
	}
	for _, r := range s.ActiveSkillWeaponRequirements {
		in(WieldableClassesTable, len(s.WieldableClasses), r.WieldableClasses...)
	}
	for _, r := range s.WieldableClasses {
		in(ItemClassesTable, len(s.ItemClasses), r.ItemClass)
	}
	for _, r := range s.ItemClasses {
		opt(ItemClassCategoriesTable, len(s.ItemClassCategories), r.ItemClassCategory)
	}
	for _, r := range s.GrantedEffects {
		opt(ActiveSkillsTable, len(s.ActiveSkills), r.ActiveSkill)
		in(ActiveSkillTypeTable, types, r.AllowedActiveSkillTypes...)
		in(ActiveSkillTypeTable, types, r.AddedActiveSkillTypes...)
		in(ActiveSkillTypeTable, types, r.ExcludedActiveSkillTypes...)
		in(GrantedEffectStatSetsTable, len(s.GrantedEffectStatSets), r.StatSet)
		in(GrantedEffectStatSetsTable, len(s.GrantedEffectStatSets), r.AdditionalStatSets...)
		in(CostTypesTable, len(s.CostTypes), r.CostTypes...)
		in(ItemClassesTable, len(s.ItemClasses), r.SupportWeaponRestrictions...)
	}
	for _, r := range s.GrantedEffectsPerLevel {
		in(GrantedEffectsTable, len(s.GrantedEffects), r.GrantedEffect)
	}
	for _, r := range s.GrantedEffectStatSets {
		opt(GrantedEffectLabelsTable, len(s.GrantedEffectLabels), r.Label)
		in(StatsTable, stats, r.ImplicitStats...)
		in(StatsTable, stats, r.ConstantStats...)
		in(StatsTable, stats, r.IgnoredStats...)
	}
	for _, r := range s.GrantedEffectStatSetsPerLevel {
		in(GrantedEffectStatSetsTable, len(s.GrantedEffectStatSets), r.StatSet)
		in(GrantedEffectsTable, len(s.GrantedEffects), r.GrantedEffects...)
		in(StatsTable, stats, r.AdditionalFlags...)
		in(StatsTable, stats, r.FloatStats...)
		in(StatsTable, stats, r.AdditionalStats...)
		in(EffectivenessCostConstantsTable, len(s.EffectivenessCostConstants), r.InterpolationBases...)
	}
	for _, r := range s.GrantedEffectQualityStats {
		in(GrantedEffectsTable, len(s.GrantedEffects), r.GrantedEffect)
		in(StatsTable, stats, r.Stats...)
		in(StatsTable, stats, r.AltStats...)
	}
	for _, r := range s.GemEffects {
		in(GrantedEffectsTable, len(s.GrantedEffects), r.GrantedEffect)
		in(GrantedEffectsTable, len(s.GrantedEffects), r.AdditionalGrantedEffects...)
		in(GemTagsTable, len(s.GemTags), r.GemTags...)
	}
	for _, r := range s.SkillGems {
		in(BaseItemTypesTable, len(s.BaseItemTypes), r.BaseItemType)
		in(GemEffectsTable, len(s.GemEffects), r.GemEffects...)
	}
	for _, r := range s.SupportGems {
		in(SkillGemsTable, len(s.SkillGems), r.SkillGem)
		in(SupportGemFamilyTable, len(s.SupportGemFamilies), r.Family...)
	}
	for _, r := range s.CostTypes {
		in(StatsTable, stats, r.Stat)
	}
	t.Logf("%d foreign keys resolve", checked)
}
