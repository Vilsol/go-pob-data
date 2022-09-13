package raw

type GrantedEffectsPerLevel struct {
	AttackSpeedMultiplier      int   `json:"AttackSpeedMultiplier"`
	AttackTime                 int   `json:"AttackTime"`
	Cooldown                   int   `json:"Cooldown"`
	CooldownBypassType         int   `json:"CooldownBypassType"`
	CooldownGroup              int   `json:"CooldownGroup"`
	CostAmounts                []int `json:"CostAmounts"`
	CostMultiplier             int   `json:"CostMultiplier"`
	CostTypes                  []int `json:"CostTypes"`
	GrantedEffect              int   `json:"GrantedEffect"`
	Level                      int   `json:"Level"`
	LifeReservationFlat        int   `json:"LifeReservationFlat"`
	LifeReservationPercent     int   `json:"LifeReservationPercent"`
	ManaReservationFlat        int   `json:"ManaReservationFlat"`
	ManaReservationPercent     int   `json:"ManaReservationPercent"`
	PlayerLevelReq             int   `json:"PlayerLevelReq"`
	SoulGainPreventionDuration int   `json:"SoulGainPreventionDuration"`
	StoredUses                 int   `json:"StoredUses"`
	VaalSouls                  int   `json:"VaalSouls"`
	VaalStoredUses             int   `json:"VaalStoredUses"`
	Key                        int   `json:"_key"`
}
