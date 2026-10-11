package raw

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type GrantedEffectsPerLevel struct {
	CostAmounts                []int `json:"CostAmounts"`
	CostTypes                  []int `json:"CostTypes"`
	LifeReservationFlat        int   `json:"LifeReservationFlat"`
	LifeReservationPercent     int   `json:"LifeReservationPercent"`
	CooldownGroup              int   `json:"CooldownGroup"`
	Cooldown                   int   `json:"Cooldown"`
	CostMultiplier             int   `json:"CostMultiplier"`
	AttackTime                 int   `json:"AttackTime"`
	GrantedEffect              int   `json:"GrantedEffect"`
	Level                      int   `json:"Level"`
	AttackSpeedMultiplier      int   `json:"AttackSpeedMultiplier"`
	CooldownBypassType         int   `json:"CooldownBypassType"`
	ManaReservationFlat        int   `json:"ManaReservationFlat"`
	ManaReservationPercent     int   `json:"ManaReservationPercent"`
	PlayerLevelReq             int   `json:"PlayerLevelReq"`
	SoulGainPreventionDuration int   `json:"SoulGainPreventionDuration"`
	StoredUses                 int   `json:"StoredUses"`
	VaalSouls                  int   `json:"VaalSouls"`
	VaalStoredUses             int   `json:"VaalStoredUses"`
	PvPDamageMultiplier        int   `json:"PvPDamageMultiplier"` // a MORE modifier, in percent
	Key                        int   `json:"_key"`
}

// UnmarshalJSON reads PvPDamageMultiplier as Var10 too, the name tables extracted before 3.29 give
// that column.
func (g *GrantedEffectsPerLevel) UnmarshalJSON(b []byte) error {
	type plain GrantedEffectsPerLevel
	var named struct {
		Var10 *int `json:"Var10"`
		plain
	}
	if err := json.Unmarshal(b, &named); err != nil {
		return errors.Wrap(err, "failed to decode granted effect level")
	}
	*g = GrantedEffectsPerLevel(named.plain)
	setFrom(&g.PvPDamageMultiplier, named.Var10)
	return nil
}
