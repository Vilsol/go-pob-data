package derived

import (
	"context"
	"errors"
	"testing"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/testdata"
)

func TestTemplateLifeScaling(t *testing.T) {
	values, extends, err := templateLifeScaling("version 3\r\nextends \"Metadata/Monsters/A\"\r\nextends \"nothing\"\r\n\r\n" +
		"/* Stats\r\n{\r\n*/\r\nclient\r\n{\r\n\tStats\r\n\t{\r\n\t\ton_x = { f( \"}\" ); } // {\r\n\t}\r\n}\r\n\r\nStats\r\n{\r\n" +
		"\tset_additional_life_scaling_index = 8\r\n}\r\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 || values[0] != 8 || len(extends) != 2 || extends[0] != "Metadata/Monsters/A" || extends[1] != "nothing" {
		t.Errorf("values %v, extends %v", values, extends)
	}

	for _, text := range []string{
		"client\n{\n\tStats\n\t{\n\t\tset_additional_life_scaling_index = 8\n\t}\n}\n",
		"Stats\n{\n\tset_additional_life_scaling_index = 8 + 1\n}\n",
		"set_additional_life_scaling_index = 8\n",
		"Stats\n{\n\tx = \"{\n}\n",
	} {
		if values, _, err := templateLifeScaling(text); err == nil {
			t.Errorf("%q gives %v, want a refusal", text, values)
		}
	}
}

func TestMonsterLifeScaling(t *testing.T) {
	testdata.ServeCommittedData(t)
	rows, err := Load[MonsterLifeScaling](context.Background(), "3.30", noCache{})
	if err != nil {
		t.Fatal(err)
	}
	index := map[string]int{}
	perIndex := map[int]int{}
	for _, row := range rows {
		index[row.Variety] = row.Index
		perIndex[row.Index]++
	}

	for _, id := range []string{
		"Metadata/Monsters/AtlasInvaders/CleansingMonsters/CleansingBoss",
		"Metadata/Monsters/AtlasInvaders/ConsumeMonsters/ConsumeBoss",
		"Metadata/Monsters/AtlasMemory/AngerBossUBER",
		"Metadata/Monsters/AtlasMemory/BenevolenceBossUBER",
		"Metadata/Monsters/AtlasMemory/IgnoranceBoss/IgnoranceBossUBER",
	} {
		if got, ok := index[id]; !ok || got != 8 {
			t.Errorf("%s: %d, want the Primordial table (8)", id, got)
		}
	}
	if got := index["Metadata/Monsters/Cannibal/CannibalMaleChampionAtlasUber"]; got != 21 {
		t.Errorf("an AtlasUber addon variety: %d, want 21", got)
	}
	for _, id := range []string{
		"Metadata/Monsters/AtlasBosses/TheShaperBoss",
		// Its base sets 25 and its addon 23.
		"Metadata/Monsters/FaridunLeague/FaridunHumanShieldSettlers",
	} {
		if got, ok := index[id]; ok {
			t.Errorf("%s: %d, want none", id, got)
		}
	}
	if len(rows) != 1237 || perIndex[8] != 30 || perIndex[21] != 89 {
		t.Errorf("%d rows, %d of index 8, %d of index 21", len(rows), perIndex[8], perIndex[21])
	}

	if _, err := Load[MonsterLifeScaling](context.Background(), "3.29", noCache{}); !errors.Is(err, loader.ErrAssetNotFound) {
		t.Error("3.29 publishes no AdditionalLifeScaling, so no MonsterLifeScaling")
	}
}
