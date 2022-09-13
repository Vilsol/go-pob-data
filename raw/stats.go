package raw

type Stat struct {
	BelongsStatsKey       []string `json:"BelongsStatsKey"`
	Category              *int     `json:"Category"`
	ContextFlags          []int    `json:"ContextFlags"`
	Hash32                int      `json:"HASH32"`
	ID                    string   `json:"Id"`
	IsLocal               bool     `json:"IsLocal"`
	IsScalable            bool     `json:"IsScalable"`
	IsVirtual             bool     `json:"IsVirtual"`
	IsWeaponLocal         bool     `json:"IsWeaponLocal"`
	MainHandAliasStatsKey *int     `json:"MainHandAlias_StatsKey"`
	OffHandAliasStatsKey  *int     `json:"OffHandAlias_StatsKey"`
	Semantics             int      `json:"Semantics"`
	Text                  string   `json:"Text"`
	Key                   int      `json:"_key"`
}
