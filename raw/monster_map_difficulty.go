package raw

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// MonsterMapDifficulty uses the 3.26+ schema names.
type MonsterMapDifficulty struct {
	Stat1                 *int `json:"Stat1"`
	Stat2                 *int `json:"Stat2"`
	Stat3                 *int `json:"Stat3"`
	Stat4                 *int `json:"Stat4"`
	MapLevel              int  `json:"MapLevel"`
	LifePercentIncrease   int  `json:"LifePercentIncrease"`
	DamagePercentIncrease int  `json:"DamagePercentIncrease"`
	Stat3Value            int  `json:"Stat3Value"`
	Stat4Value            int  `json:"Stat4Value"`
	Key                   int  `json:"_key"`
}

// UnmarshalJSON reads the tables extracted before 3.26, whose schema named StatN StatsKeyN,
// LifePercentIncrease Stat1Value and DamagePercentIncrease Stat2Value.
func (m *MonsterMapDifficulty) UnmarshalJSON(b []byte) error {
	type plain MonsterMapDifficulty
	var named struct {
		StatsKey1  *int `json:"StatsKey1"`
		StatsKey2  *int `json:"StatsKey2"`
		StatsKey3  *int `json:"StatsKey3"`
		StatsKey4  *int `json:"StatsKey4"`
		Stat1Value *int `json:"Stat1Value"`
		Stat2Value *int `json:"Stat2Value"`
		plain
	}
	if err := json.Unmarshal(b, &named); err != nil {
		return errors.Wrap(err, "failed to decode monster map difficulty")
	}
	*m = MonsterMapDifficulty(named.plain)
	setPtrFrom(&m.Stat1, named.StatsKey1)
	setPtrFrom(&m.Stat2, named.StatsKey2)
	setPtrFrom(&m.Stat3, named.StatsKey3)
	setPtrFrom(&m.Stat4, named.StatsKey4)
	setFrom(&m.LifePercentIncrease, named.Stat1Value)
	setFrom(&m.DamagePercentIncrease, named.Stat2Value)
	return nil
}

func setPtrFrom(to **int, from *int) {
	if from != nil {
		*to = from
	}
}
