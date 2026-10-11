package raw

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// DefaultMonsterStat uses the 3.28+ schema names.
type DefaultMonsterStat struct {
	DisplayLevel               string  `json:"DisplayLevel"`
	Damage                     float64 `json:"Damage"`
	MinionDamage               float64 `json:"MinionDamage"`
	AltDamage1                 float64 `json:"AltDamage1"`
	AltDamage2                 float64 `json:"AltDamage2"`
	Evasion                    int     `json:"Evasion"`
	Accuracy                   int     `json:"Accuracy"`
	MonsterLife                int     `json:"MonsterLife"`
	Experience                 int     `json:"Experience"`
	MinionLife                 int     `json:"MinionLife"`
	Armour                     int     `json:"Armour"`
	Difficulty                 int     `json:"Difficulty"`
	AltLife1                   int     `json:"AltLife1"`
	AltLife2                   int     `json:"AltLife2"`
	EvasiveEvasion             int     `json:"EvasiveEvasion"`
	AilmentThreshold           int     `json:"AilmentThreshold"`
	MonsterPhysConversionMulti int     `json:"MonsterPhysConversionMulti"` // 3.28+
	Key                        int     `json:"_key"`
}

// UnmarshalJSON reads the tables extracted before 3.28, whose schema named MonsterLife Life, MinionLife
// AllyLife, MinionDamage Damage2, AilmentThreshold Armour, and left Armour and the Alt columns unnamed.
func (d *DefaultMonsterStat) UnmarshalJSON(b []byte) error {
	type plain DefaultMonsterStat
	var named struct {
		Life     *int     `json:"Life"`
		AllyLife *int     `json:"AllyLife"`
		Damage2  *float64 `json:"Damage2"`
		Var7     *int     `json:"Var7"`
		Var10    *int     `json:"Var10"`
		Var11    *float64 `json:"Var11"`
		Var12    *float64 `json:"Var12"`
		Var13    *int     `json:"Var13"`
		Var14    *int     `json:"Var14"`
		plain
	}
	if err := json.Unmarshal(b, &named); err != nil {
		return errors.Wrap(err, "failed to decode default monster stat")
	}
	*d = DefaultMonsterStat(named.plain)
	if named.Life == nil {
		return nil
	}

	d.AilmentThreshold, d.Armour = d.Armour, 0
	setFrom(&d.MonsterLife, named.Life)
	setFrom(&d.MinionLife, named.AllyLife)
	setFrom(&d.MinionDamage, named.Damage2)
	setFrom(&d.Armour, named.Var7)
	setFrom(&d.AltLife1, named.Var10)
	setFrom(&d.AltDamage1, named.Var11)
	setFrom(&d.AltDamage2, named.Var12)
	setFrom(&d.AltLife2, named.Var13)
	setFrom(&d.EvasiveEvasion, named.Var14)
	return nil
}

func setFrom[T any](to *T, from *T) {
	if from != nil {
		*to = *from
	}
}
