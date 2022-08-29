package main

type SkillTreeData struct {
	Sprites         map[string]map[string]AssetPath `json:"sprites"`
	ImageZoomLevels []float64                       `json:"imageZoomLevels"`
}

type AssetPath struct {
	Filename string `json:"filename"`
}
