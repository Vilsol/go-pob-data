package raw

type MonsterVariety struct {
	ACTFiles                                      []string             `json:"ACTFiles"`
	AISFile                                       string               `json:"AISFile"`
	AOFiles                                       []string             `json:"AOFiles"`
	AttackSpeed                                   int                  `json:"AttackSpeed"`
	BackItemVisualIdentityKey                     *int                 `json:"Back_ItemVisualIdentityKey"`
	BaseMonsterTypeIndex                          string               `json:"BaseMonsterTypeIndex"`
	CriticalStrikeChance                          int                  `json:"CriticalStrikeChance"`
	DamageMultiplier                              int                  `json:"DamageMultiplier"`
	EPKFile                                       string               `json:"EPKFile"`
	EndgameModsKeys                               []int                `json:"Endgame_ModsKeys"`
	ExperienceMultiplier                          int                  `json:"ExperienceMultiplier"`
	GrantedEffectsKeys                            []int                `json:"GrantedEffectsKeys"`
	HelmetItemVisualIdentityKey                   *int                 `json:"Helmet_ItemVisualIdentityKey"`
	ID                                            string               `json:"Id"`
	KillRareAchievementItemsKeys                  []int                `json:"KillRare_AchievementItemsKeys"`
	KillSpecificMonsterCountAchievementItemsKeys  []int                `json:"KillSpecificMonsterCount_AchievementItemsKeys"`
	KillWhileOnslaughtIsActiveAchievementItemsKey *int                 `json:"KillWhileOnslaughtIsActive_AchievementItemsKey"`
	KillWhileTalismanIsActiveAchievementItemsKey  *int                 `json:"KillWhileTalismanIsActive_AchievementItemsKey"`
	LifeMultiplier                                int                  `json:"LifeMultiplier"`
	MainHandItemClassesKey                        *int                 `json:"MainHand_ItemClassesKey"`
	MaximumAttackDistance                         int                  `json:"MaximumAttackDistance"`
	MinimumAttackDistance                         int                  `json:"MinimumAttackDistance"`
	ModelSizeMultiplier                           int                  `json:"ModelSizeMultiplier"`
	ModsKeys                                      []int                `json:"ModsKeys"`
	ModsKeys2                                     []int                `json:"ModsKeys2"`
	MonsterArmoursKey                             *int                 `json:"MonsterArmoursKey"`
	MonsterConditionalEffectPacksKey              *int                 `json:"MonsterConditionalEffectPacksKey"`
	MonsterSegmentsKey                            *int                 `json:"MonsterSegmentsKey"`
	MonsterTypesKey                               int                  `json:"MonsterTypesKey"`
	Name                                          string               `json:"Name"`
	ObjectSize                                    int                  `json:"ObjectSize"`
	OffHandItemClassesKey                         *int                 `json:"OffHand_ItemClassesKey"`
	Part1ModsKeys                                 []int                `json:"Part1_ModsKeys"`
	Part2ModsKeys                                 []int                `json:"Part2_ModsKeys"`
	SinkAnimationAOFile                           string               `json:"SinkAnimation_AOFile"`
	SpecialModsKeys                               []int                `json:"Special_ModsKeys"`
	Stance                                        MonsterVarietyStance `json:"Stance"`
	TagsKeys                                      []int                `json:"TagsKeys"`
	Weapon1ItemVisualIdentityKeys                 []int                `json:"Weapon1_ItemVisualIdentityKeys"`
	Weapon2ItemVisualIdentityKeys                 []int                `json:"Weapon2_ItemVisualIdentityKeys"`
	Key                                           int                  `json:"_key"`
}

type MonsterVarietyStance string

const (
	Activated                   MonsterVarietyStance = "Activated"
	Bow                         MonsterVarietyStance = "Bow"
	Casterclone                 MonsterVarietyStance = "casterclone"
	Claw                        MonsterVarietyStance = "Claw"
	ClawClaw                    MonsterVarietyStance = "ClawClaw"
	ClawShield                  MonsterVarietyStance = "ClawShield"
	Dagger                      MonsterVarietyStance = "Dagger"
	DaggerDagger                MonsterVarietyStance = "DaggerDagger"
	DaggerShield                MonsterVarietyStance = "DaggerShield"
	Empty                       MonsterVarietyStance = ""
	Glaiveclone                 MonsterVarietyStance = "glaiveclone"
	NoHood                      MonsterVarietyStance = "NoHood"
	OneHand                     MonsterVarietyStance = "OneHand"
	OneHandAxe                  MonsterVarietyStance = "OneHandAxe"
	OneHandSword                MonsterVarietyStance = "OneHandSword"
	OneHandSwordDagger          MonsterVarietyStance = "OneHandSwordDagger"
	OneHandSwordOneHandSword    MonsterVarietyStance = "OneHandSwordOneHandSword"
	OneHandSwordShield          MonsterVarietyStance = "OneHandSwordShield"
	OneHandSwordThrusting       MonsterVarietyStance = "OneHandSwordThrusting"
	OneHandSwordThrustingShield MonsterVarietyStance = "OneHandSwordThrustingShield"
	PreFight                    MonsterVarietyStance = "PreFight"
	Relaxed                     MonsterVarietyStance = "Relaxed"
	Shield                      MonsterVarietyStance = "Shield"
	Staff                       MonsterVarietyStance = "Staff"
	Stance1                     MonsterVarietyStance = "stance1"
	Stance2                     MonsterVarietyStance = "stance2"
	Stance3                     MonsterVarietyStance = "stance3"
	Stance4                     MonsterVarietyStance = "stance4"
	TwoHandAxe                  MonsterVarietyStance = "TwoHandAxe"
	TwoHandMace                 MonsterVarietyStance = "TwoHandMace"
	TwoHandSword                MonsterVarietyStance = "TwoHandSword"
	Wand                        MonsterVarietyStance = "Wand"
	WandShield                  MonsterVarietyStance = "WandShield"
	WandWand                    MonsterVarietyStance = "WandWand"
)
