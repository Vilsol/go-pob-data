package raw

type Stat struct {
	MainHandAliasStatsKey *int     `json:"MainHandAlias_StatsKey"`
	Category              *int     `json:"Category"`
	OffHandAliasStatsKey  *int     `json:"OffHandAlias_StatsKey"`
	ID                    string   `json:"Id"`
	Text                  string   `json:"Text"`
	ContextFlags          []int    `json:"ContextFlags"`
	BelongsStatsKey       []string `json:"BelongsStatsKey"`
	Hash32                int      `json:"HASH32"`
	Semantics             int      `json:"Semantics"`
	Key                   int      `json:"_key"`
	IsWeaponLocal         bool     `json:"IsWeaponLocal"`
	IsVirtual             bool     `json:"IsVirtual"`
	IsScalable            bool     `json:"IsScalable"`
	IsLocal               bool     `json:"IsLocal"`
}
