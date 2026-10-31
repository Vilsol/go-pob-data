package poe2

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/testdata"
)

// pobMinion is one entry of the pin's src/Data/Minions.lua or Spectres.lua, dumped with luajit, with the
// variety its template names and the stats its modList comments carry.
type pobMinion struct {
	Name                         string   `json:"name"`
	Variety                      string   `json:"variety"`
	WeaponType1                  string   `json:"weaponType1"`
	WeaponType2                  string   `json:"weaponType2"`
	MonsterCategory              string   `json:"monsterCategory"`
	MonsterTags                  []string `json:"monsterTags"`
	SpawnLocation                []string `json:"spawnLocation"`
	SkillList                    []string `json:"skillList"`
	Mods                         [][3]any `json:"mods"`
	DamageSpread                 float64  `json:"damageSpread"`
	Armour                       float64  `json:"armour"`
	Evasion                      float64  `json:"evasion"`
	Damage                       float64  `json:"damage"`
	Life                         float64  `json:"life"`
	AttackTime                   float64  `json:"attackTime"`
	CritChance                   float64  `json:"critChance"`
	DamageFixup                  float64  `json:"damageFixup"`
	CompanionReservation         float64  `json:"companionReservation"`
	EnergyShield                 float64  `json:"energyShield"`
	ColdResist                   int      `json:"coldResist"`
	FireResist                   int      `json:"fireResist"`
	LightningResist              int      `json:"lightningResist"`
	ChaosResist                  int      `json:"chaosResist"`
	CompanionFireResist          int      `json:"companionFireResist"`
	CompanionColdResist          int      `json:"companionColdResist"`
	CompanionLightningResist     int      `json:"companionLightningResist"`
	CompanionChaosResist         int      `json:"companionChaosResist"`
	AttackRange                  int      `json:"attackRange"`
	BaseMovementSpeed            int      `json:"baseMovementSpeed"`
	SpectreReservation           int      `json:"spectreReservation"`
	BaseDamageIgnoresAttackSpeed bool     `json:"baseDamageIgnoresAttackSpeed"`
}

func loadPoBMinions(t *testing.T) map[string]pobMinion {
	t.Helper()
	b, err := os.ReadFile("testdata/pob_minions.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]pobMinion
	if err := json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}
	return want
}

func readOT(ctx context.Context, t *testing.T) func(string) (string, bool) {
	t.Helper()
	manifest, err := loader.LoadManifest(ctx, version)
	if err != nil {
		t.Fatal(err)
	}
	return func(path string) (string, bool) {
		if !manifest.HasFile(path) {
			return "", false
		}
		b, err := loader.LoadFile(ctx, version, path, nopCache{})
		if err != nil {
			t.Fatal(err)
		}
		return string(b), true
	}
}

func sameFloat(a, b float64) bool {
	return a == b || math.Abs(a-b) <= 1e-9*math.Max(math.Abs(a), math.Abs(b))
}

// testdata/pob_minions.json is the pin's src/Data/Minions.lua and Spectres.lua, dumped with luajit; the pin
// exported them from 0.5.5, so a field the live patch changed disagrees.
func TestMinionsMatchPoB(t *testing.T) {
	testdata.ServeCommittedData(t)
	ctx := context.Background()
	m, err := LoadMinionTables(ctx, version, nopCache{})
	if err != nil {
		t.Fatal(err)
	}
	read := readOT(ctx, t)
	want := loadPoBMinions(t)

	names := make([]string, 0, len(want))
	for name := range want {
		names = append(names, name)
	}
	slices.Sort(names)

	agree := map[string]int{}
	disagree := map[string][]string{}
	for _, name := range names {
		pob := want[name]
		got, err := m.Minion(pob.Variety, read)
		if err != nil {
			disagree["variety"] = append(disagree["variety"], name+": "+err.Error())
			continue
		}
		mods := make([][3]any, len(got.ModList))
		for i, stat := range got.ModList {
			mods[i] = [3]any{stat.ModID, stat.StatID, stat.Value}
		}
		if pob.Mods == nil {
			pob.Mods = [][3]any{}
		}
		for _, field := range []struct {
			got, want any
			name      string
		}{
			{got.Name, pob.Name, "name"},
			{got.MonsterTags, pob.MonsterTags, "monsterTags"},
			{got.BaseDamageIgnoresAttackSpeed, pob.BaseDamageIgnoresAttackSpeed, "baseDamageIgnoresAttackSpeed"},
			{got.Life, pob.Life, "life"},
			{got.EnergyShield, pob.EnergyShield, "energyShield"},
			{got.Armour, pob.Armour, "armour"},
			{got.Evasion, pob.Evasion, "evasion"},
			{got.FireResist, pob.FireResist, "fireResist"},
			{got.ColdResist, pob.ColdResist, "coldResist"},
			{got.LightningResist, pob.LightningResist, "lightningResist"},
			{got.ChaosResist, pob.ChaosResist, "chaosResist"},
			{got.CompanionFireResist, pob.CompanionFireResist, "companionFireResist"},
			{got.CompanionColdResist, pob.CompanionColdResist, "companionColdResist"},
			{got.CompanionLightningResist, pob.CompanionLightningResist, "companionLightningResist"},
			{got.CompanionChaosResist, pob.CompanionChaosResist, "companionChaosResist"},
			{got.Damage, pob.Damage, "damage"},
			{got.DamageSpread, pob.DamageSpread, "damageSpread"},
			{got.AttackTime, pob.AttackTime, "attackTime"},
			{got.AttackRange, pob.AttackRange, "attackRange"},
			{got.CritChance, pob.CritChance, "critChance"},
			{got.DamageFixup, pob.DamageFixup, "damageFixup"},
			{got.WeaponType1, pob.WeaponType1, "weaponType1"},
			{got.WeaponType2, pob.WeaponType2, "weaponType2"},
			{got.BaseMovementSpeed, pob.BaseMovementSpeed, "baseMovementSpeed"},
			{got.SpectreReservation, pob.SpectreReservation, "spectreReservation"},
			{got.CompanionReservation, pob.CompanionReservation, "companionReservation"},
			{got.MonsterCategory, pob.MonsterCategory, "monsterCategory"},
			{got.SpawnLocation, pob.SpawnLocation, "spawnLocation"},
			{got.SkillList, pob.SkillList, "skillList"},
			{fmt.Sprint(mods), fmt.Sprint(pob.Mods), "modList"},
		} {
			same := false
			switch g := field.got.(type) {
			case float64:
				same = sameFloat(g, field.want.(float64))
			case []string:
				w := field.want.([]string)
				same = slices.Equal(g, w)
			default:
				same = reflect.DeepEqual(field.got, field.want)
			}
			if same {
				agree[field.name]++
			} else {
				disagree[field.name] = append(disagree[field.name], fmt.Sprintf("%s (%s): %v, PoB %v", name, pob.Variety, field.got, field.want))
			}
		}
	}
	fields := make([]string, 0, len(agree)+len(disagree))
	for field := range agree {
		fields = append(fields, field)
	}
	for field := range disagree {
		if _, ok := agree[field]; !ok {
			fields = append(fields, field)
		}
	}
	slices.Sort(fields)
	for _, field := range fields {
		t.Logf("%-28s agree %4d  disagree %4d", field, agree[field], len(disagree[field]))
		for _, d := range disagree[field] {
			name := strings.SplitN(d, " ", 2)[0]
			if _, ok := patchDrift[name+"/"+field]; !ok {
				t.Errorf("%s: %s", field, d)
			}
		}
	}
	for key := range patchDrift {
		name, field, _ := strings.Cut(key, "/")
		if !slices.ContainsFunc(disagree[field], func(d string) bool { return strings.HasPrefix(d, name+" ") }) {
			t.Errorf("%s now agrees with PoB; drop it from patchDrift", key)
		}
	}
}

// patchDrift is what the live patch changed since the pin exported its minions, by entry/field.
var patchDrift = map[string]string{
	"ManifestWeapon/modList": "AnimatedWeaponPlayer.ot gained set_usable_while_shapeshifted",
}
