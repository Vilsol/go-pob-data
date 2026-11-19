package raw

type AdditionalLifeScaling struct {
	ID      string `json:"ID"`
	DatFile string `json:"DatFile"`
	IntID   int    `json:"IntId"`
	Key     int    `json:"_key"`
}

// LifeScalingPerLevel is a row of each per-level table AdditionalLifeScaling names whose columns are Level
// and MoreLife.
type LifeScalingPerLevel struct {
	Level    int `json:"Level"`
	MoreLife int `json:"MoreLife"`
	Key      int `json:"_key"`
}

// LifeScalingPerLevelTables are the tables of LifeScalingPerLevel rows.
var LifeScalingPerLevelTables = []string{
	"AncestralLifeScalingPerLevel",
	"ArchitectLifeScalingPerLevel",
	"AtlasUberMonsterLifeScalingPerLevel",
	"AzmeriLifeScalingPerLevel",
	"BetrayalTargetLifeScalingPerLevel",
	"BetrayalWallLifeScalingPerLevel",
	"BlightBossLifeScalingPerLevel",
	"ChieftainLifeScalingPerLevel",
	"CrucibleLifeScalingPerLevel",
	"FaridunLifeScalingPerLevel",
	"HarvestLifeScalingPerLevel",
	"HellscapeLifeScalingPerLevel",
	"LakeBossLifeScalingPerLevel",
	"PrimordialBossLifeScalingPerLevel",
	"RogueExileLifeScalingPerLevel",
	"SanctumFodderLifeScalingPerLevel",
	"SanctumLifeScalingPerLevel",
	"SettlersLifeScalingPerLevel",
}
