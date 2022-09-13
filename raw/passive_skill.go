package raw

type PassiveSkill struct {
	ID                   string  `json:"Id"`
	Icon                 string  `json:"Icon_DDSFile"`
	Stats                []int64 `json:"Stats"`
	Stat1                int64   `json:"Stat1Value"`
	Stat2                int64   `json:"Stat2Value"`
	Stat3                int64   `json:"Stat3Value"`
	Stat4                int64   `json:"Stat4Value"`
	Hash                 int64   `json:"PassiveSkillGraphId"`
	Name                 string  `json:"Name"`
	ClassStart           []int64 `json:"Characters"`
	Keystone             bool    `json:"IsKeystone"`
	Notable              bool    `json:"IsNotable"`
	Flavour              string  `json:"FlavourText"`
	Mastery              bool    `json:"IsJustIcon"`
	Achievement          *int64  `json:"AchievementItem"`
	JewelSocket          bool    `json:"IsJewelSocket"`
	Ascendancy           *int64  `json:"AscendancyKey"`
	AscendancyStart      bool    `json:"IsAscendancyStartingNode"`
	ReminderTexts        []int64 `json:"ReminderStrings"`
	PassivePointsGranted int64   `json:"SkillPointsGranted"`
	MultipleChoice       bool    `json:"IsMultipleChoice"`
	MultipleChoiceOption bool    `json:"IsMultipleChoiceOption"`
	Stat5                int64   `json:"Stat5Value"`
	Buffs                []int64 `json:"PassiveSkillBuffs"`
	GrantedEffect        *int64  `json:"GrantedEffectsPerLevel"`
	Blighted             bool    `json:"IsAnointmentOnly"`
	ClusterNode          bool    `json:"IsExpansion"`
	Proxy                bool    `json:"IsProxyPassive"`
	Type                 int64   `json:"SkillType"`
	MasteryGroup         *int64  `json:"MasteryGroup"`
	Key                  int64   `json:"_key"`
}
