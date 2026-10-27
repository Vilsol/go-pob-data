package poe2

import (
	"context"
	"slices"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/loader"
)

// SkillTables holds the tables PoB-PoE2's skill exporter reads, with the lookups it makes
// (src/Export/Scripts/skills.lua). A foreign key is the index into the slice it names.
type SkillTables struct {
	grantedEffectByID   map[string]int
	statSetByID         map[string]int
	perLevel            map[int][]int
	statSetPerLevel     map[int][]int
	effectStatPerLevel  map[int][]int
	qualityStats        map[int]int
	gemEffectByEffect   map[int]int
	gemEffectByAddition map[int][]int
	skillGemByGemEffect map[int]int
	supportGemBySkill   map[int]int
	itemExperience      map[int][]int

	ActiveSkillTypes              []ActiveSkillType
	ActiveSkills                  []ActiveSkill
	ActiveSkillWeaponRequirements []ActiveSkillWeaponRequirement
	WieldableClasses              []WieldableClass
	ItemClasses                   []ItemClass
	ItemClassCategories           []ItemClassCategory
	GrantedEffects                []GrantedEffect
	GrantedEffectsPerLevel        []GrantedEffectPerLevel
	GrantedEffectStatSets         []GrantedEffectStatSet
	GrantedEffectStatSetsPerLevel []GrantedEffectStatSetPerLevel
	GrantedEffectQualityStats     []GrantedEffectQualityStat
	GrantedEffectLabels           []GrantedEffectLabel
	GemEffects                    []GemEffect
	GemTags                       []GemTag
	SkillGems                     []SkillGem
	SupportGems                   []SupportGem
	SupportGemFamilies            []SupportGemFamily
	ItemExperiencePerLevel        []ItemExperiencePerLevel
	CostTypes                     []CostType
	EffectivenessCostConstants    []EffectivenessCostConstant
	Stats                         []Stat
	BaseItemTypes                 []BaseItemType
}

func loadTable[T any](ctx context.Context, version string, name string, cache loader.AssetCache, out *[]T) error {
	rows, err := loader.LoadJSONTable[T](ctx, version, name, cache)
	if err != nil {
		return err
	}
	// Every skill table has rows; an empty one is a version that does not publish it.
	if len(rows) == 0 {
		return errors.New(version + " publishes no rows of " + name)
	}
	*out = rows
	return nil
}

// LoadSkillTables loads version's skill tables and indexes them.
func LoadSkillTables(ctx context.Context, version string, cache loader.AssetCache) (*SkillTables, error) {
	s := &SkillTables{}
	for _, load := range []func() error{
		func() error { return loadTable(ctx, version, ActiveSkillTypeTable, cache, &s.ActiveSkillTypes) },
		func() error { return loadTable(ctx, version, ActiveSkillsTable, cache, &s.ActiveSkills) },
		func() error {
			return loadTable(ctx, version, ActiveSkillWeaponRequirementTable, cache, &s.ActiveSkillWeaponRequirements)
		},
		func() error { return loadTable(ctx, version, WieldableClassesTable, cache, &s.WieldableClasses) },
		func() error { return loadTable(ctx, version, ItemClassesTable, cache, &s.ItemClasses) },
		func() error { return loadTable(ctx, version, ItemClassCategoriesTable, cache, &s.ItemClassCategories) },
		func() error { return loadTable(ctx, version, GrantedEffectsTable, cache, &s.GrantedEffects) },
		func() error {
			return loadTable(ctx, version, GrantedEffectsPerLevelTable, cache, &s.GrantedEffectsPerLevel)
		},
		func() error {
			return loadTable(ctx, version, GrantedEffectStatSetsTable, cache, &s.GrantedEffectStatSets)
		},
		func() error {
			return loadTable(ctx, version, GrantedEffectStatSetsPerLevelTable, cache, &s.GrantedEffectStatSetsPerLevel)
		},
		func() error {
			return loadTable(ctx, version, GrantedEffectQualityStatsTable, cache, &s.GrantedEffectQualityStats)
		},
		func() error { return loadTable(ctx, version, GrantedEffectLabelsTable, cache, &s.GrantedEffectLabels) },
		func() error { return loadTable(ctx, version, GemEffectsTable, cache, &s.GemEffects) },
		func() error { return loadTable(ctx, version, GemTagsTable, cache, &s.GemTags) },
		func() error { return loadTable(ctx, version, SkillGemsTable, cache, &s.SkillGems) },
		func() error { return loadTable(ctx, version, SupportGemsTable, cache, &s.SupportGems) },
		func() error { return loadTable(ctx, version, SupportGemFamilyTable, cache, &s.SupportGemFamilies) },
		func() error {
			return loadTable(ctx, version, ItemExperiencePerLevelTable, cache, &s.ItemExperiencePerLevel)
		},
		func() error { return loadTable(ctx, version, CostTypesTable, cache, &s.CostTypes) },
		func() error {
			return loadTable(ctx, version, EffectivenessCostConstantsTable, cache, &s.EffectivenessCostConstants)
		},
		func() error { return loadTable(ctx, version, StatsTable, cache, &s.Stats) },
		func() error { return loadTable(ctx, version, BaseItemTypesTable, cache, &s.BaseItemTypes) },
	} {
		if err := load(); err != nil {
			return nil, err
		}
	}
	s.index()
	return s, nil
}

// index builds the lookups in row order, so a list is what DatFile:GetRowList returns and a single
// row is the first match, as DatFile:GetRow returns.
func (s *SkillTables) index() {
	s.grantedEffectByID = make(map[string]int, len(s.GrantedEffects))
	for i, row := range s.GrantedEffects {
		if _, ok := s.grantedEffectByID[row.ID]; !ok {
			s.grantedEffectByID[row.ID] = i
		}
	}
	s.statSetByID = make(map[string]int, len(s.GrantedEffectStatSets))
	for i, row := range s.GrantedEffectStatSets {
		if _, ok := s.statSetByID[row.ID]; !ok {
			s.statSetByID[row.ID] = i
		}
	}
	s.perLevel = make(map[int][]int)
	for i, row := range s.GrantedEffectsPerLevel {
		s.perLevel[row.GrantedEffect] = append(s.perLevel[row.GrantedEffect], i)
	}
	s.statSetPerLevel = make(map[int][]int)
	s.effectStatPerLevel = make(map[int][]int)
	for i, row := range s.GrantedEffectStatSetsPerLevel {
		s.statSetPerLevel[row.StatSet] = append(s.statSetPerLevel[row.StatSet], i)
		for _, effect := range uniq(row.GrantedEffects) {
			s.effectStatPerLevel[effect] = append(s.effectStatPerLevel[effect], i)
		}
	}
	s.qualityStats = make(map[int]int)
	for i, row := range s.GrantedEffectQualityStats {
		if _, ok := s.qualityStats[row.GrantedEffect]; !ok {
			s.qualityStats[row.GrantedEffect] = i
		}
	}
	s.gemEffectByEffect = make(map[int]int)
	s.gemEffectByAddition = make(map[int][]int)
	for i, row := range s.GemEffects {
		if _, ok := s.gemEffectByEffect[row.GrantedEffect]; !ok {
			s.gemEffectByEffect[row.GrantedEffect] = i
		}
		for _, effect := range uniq(row.AdditionalGrantedEffects) {
			s.gemEffectByAddition[effect] = append(s.gemEffectByAddition[effect], i)
		}
	}
	s.skillGemByGemEffect = make(map[int]int)
	for i, row := range s.SkillGems {
		for _, effect := range row.GemEffects {
			if _, ok := s.skillGemByGemEffect[effect]; !ok {
				s.skillGemByGemEffect[effect] = i
			}
		}
	}
	s.supportGemBySkill = make(map[int]int)
	for i, row := range s.SupportGems {
		if _, ok := s.supportGemBySkill[row.SkillGem]; !ok {
			s.supportGemBySkill[row.SkillGem] = i
		}
	}
	s.itemExperience = make(map[int][]int)
	for i, row := range s.ItemExperiencePerLevel {
		s.itemExperience[row.ItemExperienceType] = append(s.itemExperience[row.ItemExperienceType], i)
	}
}

// uniq keeps a row listed once when its list column names a key twice; GetRowList breaks on the first match.
func uniq(keys []int) []int {
	out := make([]int, 0, len(keys))
	for _, k := range keys {
		if !slices.Contains(out, k) {
			out = append(out, k)
		}
	}
	return out
}

func rows[T any](table []T, indices []int) []*T {
	out := make([]*T, len(indices))
	for i, index := range indices {
		out[i] = &table[index]
	}
	return out
}

func row[T any](table []T, lookup map[int]int, key int) *T {
	index, ok := lookup[key]
	if !ok {
		return nil
	}
	return &table[index]
}

// GrantedEffectByID is dat("GrantedEffects"):GetRow("Id", id) (skills.lua:134).
func (s *SkillTables) GrantedEffectByID(id string) *GrantedEffect {
	index, ok := s.grantedEffectByID[id]
	if !ok {
		return nil
	}
	return &s.GrantedEffects[index]
}

// StatSetByID is dat("GrantedEffectStatSets"):GetRow("Id", id) (skills.lua:230, 489).
func (s *SkillTables) StatSetByID(id string) *GrantedEffectStatSet {
	index, ok := s.statSetByID[id]
	if !ok {
		return nil
	}
	return &s.GrantedEffectStatSets[index]
}

// PerLevel is dat("GrantedEffectsPerLevel"):GetRowList("GrantedEffect", granted) (skills.lua:174, 227).
func (s *SkillTables) PerLevel(grantedEffect int) []*GrantedEffectPerLevel {
	return rows(s.GrantedEffectsPerLevel, s.perLevel[grantedEffect])
}

// StatSetPerLevel is dat("GrantedEffectStatSetsPerLevel"):GetRowList("GrantedEffectStatSets", set)
// (skills.lua:230, 494).
func (s *SkillTables) StatSetPerLevel(statSet int) []*GrantedEffectStatSetPerLevel {
	return rows(s.GrantedEffectStatSetsPerLevel, s.statSetPerLevel[statSet])
}

// EffectStatSetPerLevel is dat("GrantedEffectStatSetsPerLevel"):GetRowList("GrantedEffect", granted), a
// match on the GrantedEffects list column (skills.lua:228).
func (s *SkillTables) EffectStatSetPerLevel(grantedEffect int) []*GrantedEffectStatSetPerLevel {
	return rows(s.GrantedEffectStatSetsPerLevel, s.effectStatPerLevel[grantedEffect])
}

// LevelStatRows are the stat rows of a granted effect's levels: those of the stat set sharing its Id when
// that set has at least gemLevels rows, else EffectStatSetPerLevel (skills.lua:228-233).
func (s *SkillTables) LevelStatRows(grantedEffect int, gemLevels int) []*GrantedEffectStatSetPerLevel {
	statsPerLevel := s.EffectStatSetPerLevel(grantedEffect)
	if set := s.StatSetByID(s.GrantedEffects[grantedEffect].ID); set != nil {
		if primary := s.StatSetPerLevel(set.Key); len(primary) >= gemLevels {
			statsPerLevel = primary
		}
	}
	return statsPerLevel
}

// QualityStats is dat("GrantedEffectQualityStats"):GetRow("GrantedEffect", granted) (skills.lua:312).
func (s *SkillTables) QualityStats(grantedEffect int) *GrantedEffectQualityStat {
	return row(s.GrantedEffectQualityStats, s.qualityStats, grantedEffect)
}

// GemEffect is dat("GemEffects"):GetRow("GrantedEffect", granted) (skills.lua:139). skills.lua:142 falls
// back to GetRow("AdditionalGrantedEffects", granted), which never matches: GetRow compares a list cell
// to a row with ==.
func (s *SkillTables) GemEffect(grantedEffect int) *GemEffect {
	return row(s.GemEffects, s.gemEffectByEffect, grantedEffect)
}

// GemEffectsAdding is dat("GemEffects"):GetRowList("AdditionalGrantedEffects", granted) (skills.lua:760).
func (s *SkillTables) GemEffectsAdding(grantedEffect int) []*GemEffect {
	return rows(s.GemEffects, s.gemEffectByAddition[grantedEffect])
}

// SkillGem is the first SkillGems row listing gemEffect (skills.lua:150-170).
func (s *SkillTables) SkillGem(gemEffect int) *SkillGem {
	return row(s.SkillGems, s.skillGemByGemEffect, gemEffect)
}

// SupportGem is dat("SupportGems"):GetRow("SkillGem", skillGem) (skills.lua:349, 985).
func (s *SkillTables) SupportGem(skillGem int) *SupportGem {
	return row(s.SupportGems, s.supportGemBySkill, skillGem)
}

// GemLevelProgression is dat("ItemExperiencePerLevel"):GetRowList("ItemExperienceType",
// skillGem.GemLevelProgression) (skills.lua:236, 1002).
func (s *SkillTables) GemLevelProgression(skillGem int) []*ItemExperiencePerLevel {
	return rows(s.ItemExperiencePerLevel, s.itemExperience[s.SkillGems[skillGem].ItemExperienceType])
}

// WeaponClasses are the item classes of an active skill's WeaponRestrictions.WeaponClass (skills.lua:415-419, 975).
func (s *SkillTables) WeaponClasses(activeSkill int) []*ItemClass {
	requirement := s.ActiveSkills[activeSkill].WeaponRequirements
	if requirement == nil {
		return nil
	}
	wieldable := s.ActiveSkillWeaponRequirements[*requirement].WieldableClasses
	out := make([]*ItemClass, len(wieldable))
	for i, class := range wieldable {
		out[i] = &s.ItemClasses[s.WieldableClasses[class].ItemClass]
	}
	return out
}
