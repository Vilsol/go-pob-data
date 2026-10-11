package raw

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// MonsterMapBossDifficulty uses the 3.26+ schema names.
type MonsterMapBossDifficulty struct {
	Stat1                      *int `json:"Stat1"`
	Stat2                      *int `json:"Stat2"`
	Stat3                      *int `json:"Stat3"`
	Stat4                      *int `json:"Stat4"`
	Stat5                      *int `json:"Stat5"`
	MapLevel                   int  `json:"MapLevel"`
	BossLifePercentIncrease    int  `json:"BossLifePercentIncrease"`
	BossDamagePercentIncrease  int  `json:"BossDamagePercentIncrease"`
	BossIncItemQuantity        int  `json:"BossIncItemQuantity"`
	BossIncItemRarity          int  `json:"BossIncItemRarity"`
	BossAilmentPercentDecrease int  `json:"BossAilmentPercentDecrease"`
	Key                        int  `json:"_key"`
}

// UnmarshalJSON reads the tables extracted before 3.26, whose schema named StatN StatsKeyN and the
// values Stat1Value..Stat5Value, in the order of the fields above.
func (m *MonsterMapBossDifficulty) UnmarshalJSON(b []byte) error {
	type plain MonsterMapBossDifficulty
	var named struct {
		StatsKey1  *int `json:"StatsKey1"`
		StatsKey2  *int `json:"StatsKey2"`
		StatsKey3  *int `json:"StatsKey3"`
		StatsKey4  *int `json:"StatsKey4"`
		StatsKey5  *int `json:"StatsKey5"`
		Stat1Value *int `json:"Stat1Value"`
		Stat2Value *int `json:"Stat2Value"`
		Stat3Value *int `json:"Stat3Value"`
		Stat4Value *int `json:"Stat4Value"`
		Stat5Value *int `json:"Stat5Value"`
		plain
	}
	if err := json.Unmarshal(b, &named); err != nil {
		return errors.Wrap(err, "failed to decode monster map boss difficulty")
	}
	*m = MonsterMapBossDifficulty(named.plain)
	setPtrFrom(&m.Stat1, named.StatsKey1)
	setPtrFrom(&m.Stat2, named.StatsKey2)
	setPtrFrom(&m.Stat3, named.StatsKey3)
	setPtrFrom(&m.Stat4, named.StatsKey4)
	setPtrFrom(&m.Stat5, named.StatsKey5)
	setFrom(&m.BossLifePercentIncrease, named.Stat1Value)
	setFrom(&m.BossDamagePercentIncrease, named.Stat2Value)
	setFrom(&m.BossIncItemQuantity, named.Stat3Value)
	setFrom(&m.BossIncItemRarity, named.Stat4Value)
	setFrom(&m.BossAilmentPercentDecrease, named.Stat5Value)
	return nil
}
