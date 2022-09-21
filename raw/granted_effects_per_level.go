package raw

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
	Key                        int   `json:"_key"`
}
