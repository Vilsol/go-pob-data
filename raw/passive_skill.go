package raw

type PassiveSkill struct {
	MasteryGroup         *int64  `json:"MasteryGroup"`
	GrantedEffect        *int64  `json:"GrantedEffectsPerLevel"`
	Ascendancy           *int64  `json:"AscendancyKey"`
	Achievement          *int64  `json:"AchievementItem"`
	Flavour              string  `json:"FlavourText"`
	Icon                 string  `json:"Icon_DDSFile"`
	ID                   string  `json:"Id"`
	Name                 string  `json:"Name"`
	Stats                []int64 `json:"Stats"`
	Buffs                []int64 `json:"PassiveSkillBuffs"`
	ReminderTexts        []int64 `json:"ReminderStrings"`
	ClassStart           []int64 `json:"Characters"`
	Hash                 int64   `json:"PassiveSkillGraphId"`
	Stat5                int64   `json:"Stat5Value"`
	Key                  int64   `json:"_key"`
	Stat1                int64   `json:"Stat1Value"`
	Type                 int64   `json:"SkillType"`
	Stat2                int64   `json:"Stat2Value"`
	Stat4                int64   `json:"Stat4Value"`
	PassivePointsGranted int64   `json:"SkillPointsGranted"`
	Stat3                int64   `json:"Stat3Value"`
	MultipleChoiceOption bool    `json:"IsMultipleChoiceOption"`
	Mastery              bool    `json:"IsJustIcon"`
	MultipleChoice       bool    `json:"IsMultipleChoice"`
	AscendancyStart      bool    `json:"IsAscendancyStartingNode"`
	Blighted             bool    `json:"IsAnointmentOnly"`
	ClusterNode          bool    `json:"IsExpansion"`
	Proxy                bool    `json:"IsProxyPassive"`
	Keystone             bool    `json:"IsKeystone"`
	JewelSocket          bool    `json:"IsJewelSocket"`
	Notable              bool    `json:"IsNotable"`
}
