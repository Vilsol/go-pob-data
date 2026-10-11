package raw

import (
	"encoding/json"
	"testing"
)

// The same level 70 rows as 3.27 (before the renames) and 3.29 publish them.
func TestMonsterTablesReadOldNames(t *testing.T) {
	var old, current DefaultMonsterStat
	if err := json.Unmarshal([]byte(`{"_key":69,"DisplayLevel":"70","Damage":413.01,"Evasion":5081,"Accuracy":314,"Life":6937,"Experience":10720,"AllyLife":1478,"Var7":33714,"Difficulty":3,"Damage2":332.05,"Var10":6453,"Var11":413.01,"Var12":413.01,"Var13":6820,"Var14":7513,"Armour":6937}`), &old); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"_key":69,"DisplayLevel":"70","Damage":413.01,"Evasion":5081,"Accuracy":314,"MonsterLife":6937,"Experience":10720,"MinionLife":1478,"Armour":33714,"Difficulty":3,"MinionDamage":332.05,"AltLife1":6453,"AltDamage1":413.01,"AltDamage2":413.01,"AltLife2":6820,"EvasiveEvasion":7513,"AilmentThreshold":6937,"MonsterPhysConversionMulti":250}`), &current); err != nil {
		t.Fatal(err)
	}
	current.MonsterPhysConversionMulti = 0
	if old != current {
		t.Errorf("old names decode to %+v, current to %+v", old, current)
	}

	var oldMap, currentMap MonsterMapBossDifficulty
	if err := json.Unmarshal([]byte(`{"_key":69,"MapLevel":70,"Stat1Value":216,"Stat2Value":0,"StatsKey1":2448,"StatsKey2":2449,"StatsKey3":11,"Stat3Value":0,"StatsKey4":10,"Stat4Value":0,"StatsKey5":10707,"Stat5Value":-42}`), &oldMap); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"_key":69,"MapLevel":70,"BossLifePercentIncrease":216,"BossDamagePercentIncrease":0,"Stat1":2448,"Stat2":2449,"Stat3":11,"BossIncItemQuantity":0,"Stat4":10,"BossIncItemRarity":0,"Stat5":10707,"BossAilmentPercentDecrease":-42}`), &currentMap); err != nil {
		t.Fatal(err)
	}
	if b1, _ := json.Marshal(oldMap); string(b1) != mustJSON(t, currentMap) {
		t.Errorf("old names decode to %s, current to %s", b1, mustJSON(t, currentMap))
	}

	var oldArea, currentArea MonsterMapDifficulty
	if err := json.Unmarshal([]byte(`{"_key":4,"MapLevel":70,"Stat1Value":13,"Stat2Value":2,"StatsKey1":2448,"StatsKey2":2449,"StatsKey3":null,"Stat3Value":0,"StatsKey4":null,"Stat4Value":0}`), &oldArea); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"_key":4,"MapLevel":70,"LifePercentIncrease":13,"DamagePercentIncrease":2,"Stat1":2448,"Stat2":2449,"Stat3":null,"Stat3Value":0,"Stat4":null,"Stat4Value":0}`), &currentArea); err != nil {
		t.Fatal(err)
	}
	if b1, _ := json.Marshal(oldArea); string(b1) != mustJSON(t, currentArea) {
		t.Errorf("old names decode to %s, current to %s", b1, mustJSON(t, currentArea))
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
