package raw

type PantheonPanelLayout struct {
	AchievementItems []int  `json:"AchievementItems"`
	CoverImage       string `json:"CoverImage"`
	Effect1StatsKeys []int  `json:"Effect1_StatsKeys"`
	Effect1Values    []int  `json:"Effect1_Values"`
	Effect2StatsKeys []int  `json:"Effect2_StatsKeys"`
	Effect2Values    []int  `json:"Effect2_Values"`
	Effect3StatsKeys []int  `json:"Effect3_StatsKeys"`
	Effect3Values    []int  `json:"Effect3_Values"`
	Effect4StatsKeys []int  `json:"Effect4_StatsKeys"`
	Effect4Values    []int  `json:"Effect4_Values"`
	GodName1         string `json:"GodName1"`
	GodName2         string `json:"GodName2"`
	GodName3         string `json:"GodName3"`
	GodName4         string `json:"GodName4"`
	ID               string `json:"Id"`
	IsDisabled       bool   `json:"IsDisabled"`
	IsMajorGod       bool   `json:"IsMajorGod"`
	QuestState1      int    `json:"QuestState1"`
	QuestState2      int    `json:"QuestState2"`
	QuestState3      int    `json:"QuestState3"`
	QuestState4      int    `json:"QuestState4"`
	SelectionImage   string `json:"SelectionImage"`
	X                int    `json:"X"`
	Y                int    `json:"Y"`
	Key              int    `json:"_key"`
}
