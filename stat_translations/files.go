package st

import (
	"path"
	"slices"
	"strings"
)

// Files are the stat description files under Metadata/StatDescriptions that are published, by name
// without extension: stat_descriptions.txt and every *_stat_descriptions.txt file of the game.
var Files = []string{
	"active_skill_gem_stat_descriptions",
	"advanced_mod_stat_descriptions",
	"atlas_relic_stat_descriptions",
	"atlas_stat_descriptions",
	"aura_skill_stat_descriptions",
	"banner_aura_skill_stat_descriptions",
	"beam_skill_stat_descriptions",
	"brand_skill_stat_descriptions",
	"buff_skill_stat_descriptions",
	"chest_stat_descriptions",
	"curse_skill_stat_descriptions",
	"debuff_skill_stat_descriptions",
	"deepwater_applied_adjacent_stat_descriptions",
	"deepwater_explicit_stat_descriptions",
	"deepwater_implicit_stat_descriptions",
	"deepwater_stat_descriptions",
	"expedition_relic_stat_descriptions",
	"gem_stat_descriptions",
	"graft_stat_descriptions",
	"heist_equipment_stat_descriptions",
	"leaguestone_stat_descriptions",
	"map_stat_descriptions",
	"mercenary_support_stat_descriptions",
	"minion_attack_skill_stat_descriptions",
	"minion_skill_stat_descriptions",
	"minion_spell_damage_skill_stat_descriptions",
	"minion_spell_skill_stat_descriptions",
	"mirage_stat_descriptions",
	"monster_stat_descriptions",
	"necropolis_stat_descriptions",
	"offering_skill_stat_descriptions",
	"passive_skill_aura_stat_descriptions",
	"passive_skill_stat_descriptions",
	"primordial_altar_stat_descriptions",
	"sanctum_relic_stat_descriptions",
	"secondary_debuff_skill_stat_descriptions",
	"sentinel_stat_descriptions",
	"single_minion_spell_skill_stat_descriptions",
	"skill_stat_descriptions",
	"stat_descriptions",
	"tincture_stat_descriptions",
	"vaal_side_area_stat_descriptions",
	"variable_duration_skill_stat_descriptions",
	"village_stat_descriptions",
}

// PoE2Files are PoE2's stat description files under Data/StatDescriptions (.csd), by name without
// extension; the files under SpecificSkillDir are found in the game rather than listed.
var PoE2Files = []string{
	"active_skill_gem_stat_descriptions",
	"advanced_mod_stat_descriptions",
	"atlas_stat_descriptions",
	"atlas_variant_stat_descriptions",
	"character_panel_gamepad_stat_descriptions",
	"character_panel_stat_descriptions",
	"chest_stat_descriptions",
	"endgame_map_stat_descriptions",
	"expedition_relic_special_stat_descriptions",
	"expedition_relic_stat_descriptions",
	"gem_stat_descriptions",
	"heist_equipment_stat_descriptions",
	"leaguestone_stat_descriptions",
	"map_stat_descriptions",
	"map_temple_room_stat_descriptions",
	"meta_gem_stat_descriptions",
	"monster_stat_descriptions",
	"passive_skill_aura_stat_descriptions",
	"passive_skill_stat_descriptions",
	"passive_skill_variant_stat_descriptions",
	"primordial_altar_stat_descriptions",
	"sanctum_relic_stat_descriptions",
	"sentinel_stat_descriptions",
	"skill_stat_descriptions",
	"stat_descriptions",
	"tablet_stat_descriptions",
	"utility_flask_buff_stat_descriptions",
}

// SpecificSkillDir holds PoE2's per-skill stat description files, some of them one directory deeper per
// stat set.
const SpecificSkillDir = "specific_skill_stat_descriptions"

// PerFileRoots are the English stat description files go-pob v0.1.0 to v0.7.0 load one at a time with
// loader.LoadTranslation, by game: data/raw and data/poe2 read stat_descriptions, and scripts/gamedatadump
// reads the rest and, through PerFileNames, what they include. Every other file is published only in the
// translation bundles.
var PerFileRoots = map[string][]string{
	"poe1": {
		"active_skill_gem_stat_descriptions",
		"aura_skill_stat_descriptions",
		"banner_aura_skill_stat_descriptions",
		"beam_skill_stat_descriptions",
		"brand_skill_stat_descriptions",
		"buff_skill_stat_descriptions",
		"curse_skill_stat_descriptions",
		"debuff_skill_stat_descriptions",
		"gem_stat_descriptions",
		"graft_stat_descriptions",
		"minion_attack_skill_stat_descriptions",
		"minion_skill_stat_descriptions",
		"minion_spell_damage_skill_stat_descriptions",
		"minion_spell_skill_stat_descriptions",
		"monster_stat_descriptions",
		"offering_skill_stat_descriptions",
		"passive_skill_stat_descriptions",
		"secondary_debuff_skill_stat_descriptions",
		"single_minion_spell_skill_stat_descriptions",
		"skill_stat_descriptions",
		"stat_descriptions",
		"tincture_stat_descriptions",
		"variable_duration_skill_stat_descriptions",
	},
	"poe2": {
		"passive_skill_stat_descriptions",
		"stat_descriptions",
	},
}

// PerFileLanguage is the only language published one file at a time.
const PerFileLanguage = "en"

// IncludeName is the name of the stat description file an include names by game path
// ("Metadata/StatDescriptions/stat_descriptions.txt" or "Data/StatDescriptions/stat_descriptions.csd").
func IncludeName(include string) string {
	name := include
	if i := strings.Index(strings.ToLower(name), "statdescriptions/"); i >= 0 {
		name = name[i+len("statdescriptions/"):]
	}
	return strings.TrimSuffix(name, path.Ext(name))
}

// PerFileNames are roots and every file they include, transitively, that includes knows: includes returns
// a file's includes and whether the version has it. Names the version does not have are left out.
func PerFileNames(roots []string, includes func(name string) ([]string, bool)) []string {
	seen := map[string]bool{}
	var out []string
	queue := slices.Clone(roots)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if seen[name] {
			continue
		}
		seen[name] = true
		incs, ok := includes(name)
		if !ok {
			continue
		}
		out = append(out, name)
		for _, include := range incs {
			queue = append(queue, IncludeName(include))
		}
	}
	slices.Sort(out)
	return out
}
