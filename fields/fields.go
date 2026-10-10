// Package fields holds, per game version, the key paths a consumer reads from each asset (lists
// written by go-pob, format go-pob-data-fields/1), and names the .min assets that keep only those.
//
// A .min asset is its full tuple asset with every unlisted field set to its zero value, encoded by the
// same generated encoder, so it decodes with the same decoder as the full asset.
package fields

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
	st "github.com/Vilsol/go-pob-data/stat_translations"
)

// Format is the list format this package reads.
const Format = "go-pob-data-fields/1"

//go:embed *.json
var embedded embed.FS

// List is the list of one game version. Keys a projector ignores (wholeTypes, droppedKeys, notes, ...)
// are not decoded.
type List struct {
	Assets     map[string]Asset `json:"assets"`
	Format     string           `json:"format"`
	DataModule string           `json:"dataModule"`
	Version    string           `json:"version"`
}

// Asset lists the key paths to keep of one asset. A path is Go field names joined by dots: "A.B" keeps
// B inside the value under A, through slices, arrays and pointers. A path with no listed children keeps
// its whole value; one with children keeps only those. An empty Keep keeps only the row count.
type Asset struct {
	Type string   `json:"type"`
	Keep []string `json:"keep"`
}

var rawTables = map[string]reflect.Type{
	"ActiveSkillType":                   reflect.TypeFor[raw.ActiveSkillType](),
	"ActiveSkills":                      reflect.TypeFor[raw.ActiveSkill](),
	"AlternatePassiveAdditions":         reflect.TypeFor[raw.AlternatePassiveAddition](),
	"AlternatePassiveSkills":            reflect.TypeFor[raw.AlternatePassiveSkill](),
	"AlternateTreeVersions":             reflect.TypeFor[raw.AlternateTreeVersion](),
	"ArmourTypes":                       reflect.TypeFor[raw.ArmourType](),
	"AtlasNode":                         reflect.TypeFor[raw.AtlasNode](),
	"AtlasNodeDefinition":               reflect.TypeFor[raw.AtlasNodeDefinition](),
	"BaseItemTypes":                     reflect.TypeFor[raw.BaseItemType](),
	"ComponentAttributeRequirements":    reflect.TypeFor[raw.ComponentAttributeRequirement](),
	"ComponentCharges":                  reflect.TypeFor[raw.ComponentCharge](),
	"CostTypes":                         reflect.TypeFor[raw.CostType](),
	"CraftingBenchOptions":              reflect.TypeFor[raw.CraftingBenchOption](),
	"DefaultMonsterStats":               reflect.TypeFor[raw.DefaultMonsterStat](),
	"Essences":                          reflect.TypeFor[raw.Essence](),
	"Flasks":                            reflect.TypeFor[raw.Flask](),
	"GrantedEffectQualityStats":         reflect.TypeFor[raw.GrantedEffectQualityStat](),
	"GrantedEffectStatSets":             reflect.TypeFor[raw.GrantedEffectStatSet](),
	"GrantedEffectStatSetsPerLevel":     reflect.TypeFor[raw.GrantedEffectStatSetsPerLevel](),
	"GrantedEffects":                    reflect.TypeFor[raw.GrantedEffect](),
	"GrantedEffectsPerLevel":            reflect.TypeFor[raw.GrantedEffectsPerLevel](),
	"ItemClasses":                       reflect.TypeFor[raw.ItemClass](),
	"ItemExperiencePerLevel":            reflect.TypeFor[raw.ItemExperiencePerLevel](),
	"Mods":                              reflect.TypeFor[raw.Mod](),
	"MonsterMapBossDifficulty":          reflect.TypeFor[raw.MonsterMapBossDifficulty](),
	"MonsterMapDifficulty":              reflect.TypeFor[raw.MonsterMapDifficulty](),
	"MonsterVarieties":                  reflect.TypeFor[raw.MonsterVariety](),
	"PantheonPanelLayout":               reflect.TypeFor[raw.PantheonPanelLayout](),
	"PassiveSkills":                     reflect.TypeFor[raw.PassiveSkill](),
	"PassiveTreeExpansionJewels":        reflect.TypeFor[raw.PassiveTreeExpansionJewel](),
	"PassiveTreeExpansionSkills":        reflect.TypeFor[raw.PassiveTreeExpansionSkill](),
	"PassiveTreeExpansionSpecialSkills": reflect.TypeFor[raw.PassiveTreeExpansionSpecialSkill](),
	"ShieldTypes":                       reflect.TypeFor[raw.ShieldType](),
	"SkillGems":                         reflect.TypeFor[raw.SkillGem](),
	"SkillTotemVariations":              reflect.TypeFor[raw.SkillTotemVariation](),
	"SkillTotems":                       reflect.TypeFor[raw.SkillTotem](),
	"Stats":                             reflect.TypeFor[raw.Stat](),
	"Tags":                              reflect.TypeFor[raw.Tag](),
	"WeaponTypes":                       reflect.TypeFor[raw.WeaponType](),
}

// RowType is the type the loader decodes asset into: a raw table's row ("raw/Mods") or a
// TranslationFile ("stat_translations/en/stat_descriptions").
func RowType(asset string) (reflect.Type, bool) {
	if name, ok := strings.CutPrefix(asset, "raw/"); ok {
		t, ok := rawTables[name]
		return t, ok
	}

	parts := strings.Split(asset, "/")
	if len(parts) == 3 && parts[0] == "stat_translations" && parts[1] != "" && slices.Contains(st.Files, parts[2]) {
		return reflect.TypeFor[raw.TranslationFile](), true
	}

	return nil, false
}

// Parse reads and validates a list: its format, that every asset is one the loader reads with the
// listed type, and that every key path names fields of that type.
func Parse(b []byte) (*List, error) {
	var l List
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, errors.Wrap(err, "failed to decode list")
	}

	if l.Format != Format {
		return nil, errors.Errorf("list format %q, this package reads %q", l.Format, Format)
	}
	if l.Version == "" {
		return nil, errors.New("list has no version")
	}

	for _, asset := range slices.Sorted(maps.Keys(l.Assets)) {
		a := l.Assets[asset]
		t, ok := RowType(asset)
		if !ok {
			return nil, errors.Errorf("%s %s: unknown asset", l.Version, asset)
		}
		if a.Type != t.Name() {
			return nil, errors.Errorf("%s %s: type %q, the loader decodes it as %s", l.Version, asset, a.Type, t.Name())
		}
		if _, err := keepTree(t, a.Keep); err != nil {
			return nil, errors.Wrapf(err, "%s %s", l.Version, asset)
		}
	}

	return &l, nil
}

var lists = func() map[string]*List {
	files, err := embedded.ReadDir(".")
	if err != nil {
		panic(err)
	}

	out := map[string]*List{}
	for _, file := range files {
		b, err := embedded.ReadFile(file.Name())
		if err != nil {
			panic(err)
		}

		// TestEmbeddedLists keeps these valid.
		l, err := Parse(b)
		if err != nil {
			panic(file.Name() + ": " + err.Error())
		}
		if l.Version+".json" != file.Name() {
			panic(file.Name() + ": list of version " + l.Version)
		}
		out[l.Version] = l
	}
	return out
}()

// Versions are the game versions that have a list, sorted.
func Versions() []string {
	versions := make([]string, 0, len(lists))
	for version := range lists {
		versions = append(versions, version)
	}
	slices.Sort(versions)
	return versions
}

// For returns the list of version.
func For(version string) (*List, bool) {
	l, ok := lists[version]
	return l, ok
}

// MinVariant is the variant of asset's .min in version ("min-" and 8 hex digits), if version's list
// has asset. It fits loader.VariantFunc.
func MinVariant(version string, asset string) (string, bool) {
	l, ok := lists[version]
	if !ok {
		return "", false
	}
	return l.MinVariant(asset)
}

// MinVariant is the variant of asset's .min under l, if l has asset. It hashes the row type's layout
// and the sorted keep list, so a list revision or a struct change moves the .min to a new name.
func (l *List) MinVariant(asset string) (string, bool) {
	a, ok := l.Assets[asset]
	if !ok {
		return "", false
	}
	t, ok := RowType(asset)
	if !ok {
		return "", false
	}

	keep := slices.Compact(slices.Sorted(slices.Values(a.Keep)))
	sum := sha256.Sum256([]byte(loader.LayoutHash(reflect.New(t).Interface()) + "\n" + strings.Join(keep, "\n")))
	return "min-" + hex.EncodeToString(sum[:4]), true
}

type node struct {
	children map[string]*node
}

// keepTree rejects a path that names no field, or descends into a value that has no fields.
func keepTree(t reflect.Type, keep []string) (*node, error) {
	root := &node{}
	for _, p := range keep {
		n, ft := root, t
		for _, seg := range strings.Split(p, ".") {
			st := ft
			for st.Kind() == reflect.Pointer || st.Kind() == reflect.Slice || st.Kind() == reflect.Array {
				st = st.Elem()
			}
			if st.Kind() != reflect.Struct {
				return nil, errors.Errorf("key path %q: %s has no fields, cannot select %s", p, ft, seg)
			}
			f, ok := st.FieldByName(seg)
			if !ok || !f.IsExported() {
				return nil, errors.Errorf("key path %q: %s has no field %s", p, st.Name(), seg)
			}

			if n.children == nil {
				n.children = map[string]*node{}
			}
			if n.children[seg] == nil {
				n.children[seg] = &node{}
			}
			n, ft = n.children[seg], f.Type
		}
	}
	return root, nil
}
