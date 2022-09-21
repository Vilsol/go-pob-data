package raw

type PantheonPanelLayout struct {
	GodName1         string `json:"GodName1"`
	CoverImage       string `json:"CoverImage"`
	SelectionImage   string `json:"SelectionImage"`
	ID               string `json:"Id"`
	GodName4         string `json:"GodName4"`
	GodName3         string `json:"GodName3"`
	GodName2         string `json:"GodName2"`
	Effect3StatsKeys []int  `json:"Effect3_StatsKeys"`
	Effect4StatsKeys []int  `json:"Effect4_StatsKeys"`
	Effect4Values    []int  `json:"Effect4_Values"`
	Effect3Values    []int  `json:"Effect3_Values"`
	AchievementItems []int  `json:"AchievementItems"`
	Effect2Values    []int  `json:"Effect2_Values"`
	Effect2StatsKeys []int  `json:"Effect2_StatsKeys"`
	Effect1Values    []int  `json:"Effect1_Values"`
	Effect1StatsKeys []int  `json:"Effect1_StatsKeys"`
	QuestState4      int    `json:"QuestState4"`
	QuestState1      int    `json:"QuestState1"`
	QuestState2      int    `json:"QuestState2"`
	QuestState3      int    `json:"QuestState3"`
	X                int    `json:"X"`
	Y                int    `json:"Y"`
	Key              int    `json:"_key"`
	IsMajorGod       bool   `json:"IsMajorGod"`
	IsDisabled       bool   `json:"IsDisabled"`
}
