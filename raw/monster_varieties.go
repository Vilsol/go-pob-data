package raw

type MonsterVariety struct {
	HelmetItemVisualIdentityKey                   *int                 `json:"Helmet_ItemVisualIdentityKey"`
	MonsterArmoursKey                             *int                 `json:"MonsterArmoursKey"`
	KillWhileTalismanIsActiveAchievementItemsKey  *int                 `json:"KillWhileTalismanIsActive_AchievementItemsKey"`
	MonsterSegmentsKey                            *int                 `json:"MonsterSegmentsKey"`
	OffHandItemClassesKey                         *int                 `json:"OffHand_ItemClassesKey"`
	MainHandItemClassesKey                        *int                 `json:"MainHand_ItemClassesKey"`
	KillWhileOnslaughtIsActiveAchievementItemsKey *int                 `json:"KillWhileOnslaughtIsActive_AchievementItemsKey"`
	MonsterConditionalEffectPacksKey              *int                 `json:"MonsterConditionalEffectPacksKey"`
	BackItemVisualIdentityKey                     *int                 `json:"Back_ItemVisualIdentityKey"`
	EPKFile                                       string               `json:"EPKFile"`
	Stance                                        MonsterVarietyStance `json:"Stance"`
	SinkAnimationAOFile                           string               `json:"SinkAnimation_AOFile"`
	AISFile                                       string               `json:"AISFile"`
	ID                                            string               `json:"Id"`
	BaseMonsterTypeIndex                          string               `json:"BaseMonsterTypeIndex"`
	Name                                          string               `json:"Name"`
	TagsKeys                                      []int                `json:"TagsKeys"`
	KillSpecificMonsterCountAchievementItemsKeys  []int                `json:"KillSpecificMonsterCount_AchievementItemsKeys"`
	SpecialModsKeys                               []int                `json:"Special_ModsKeys"`
	KillRareAchievementItemsKeys                  []int                `json:"KillRare_AchievementItemsKeys"`
	ACTFiles                                      []string             `json:"ACTFiles"`
	GrantedEffectsKeys                            []int                `json:"GrantedEffectsKeys"`
	Part2ModsKeys                                 []int                `json:"Part2_ModsKeys"`
	ModsKeys                                      []int                `json:"ModsKeys"`
	ModsKeys2                                     []int                `json:"ModsKeys2"`
	EndgameModsKeys                               []int                `json:"Endgame_ModsKeys"`
	Weapon1ItemVisualIdentityKeys                 []int                `json:"Weapon1_ItemVisualIdentityKeys"`
	Weapon2ItemVisualIdentityKeys                 []int                `json:"Weapon2_ItemVisualIdentityKeys"`
	Part1ModsKeys                                 []int                `json:"Part1_ModsKeys"`
	AOFiles                                       []string             `json:"AOFiles"`
	MaximumAttackDistance                         int                  `json:"MaximumAttackDistance"`
	ObjectSize                                    int                  `json:"ObjectSize"`
	MonsterTypesKey                               int                  `json:"MonsterTypesKey"`
	ModelSizeMultiplier                           int                  `json:"ModelSizeMultiplier"`
	MinimumAttackDistance                         int                  `json:"MinimumAttackDistance"`
	LifeMultiplier                                int                  `json:"LifeMultiplier"`
	ExperienceMultiplier                          int                  `json:"ExperienceMultiplier"`
	DamageMultiplier                              int                  `json:"DamageMultiplier"`
	CriticalStrikeChance                          int                  `json:"CriticalStrikeChance"`
	AttackSpeed                                   int                  `json:"AttackSpeed"`
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
