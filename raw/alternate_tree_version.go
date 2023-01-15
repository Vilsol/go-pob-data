package raw

type AlternateTreeVersion struct {
	ConquerorType                          string `json:"Id"`
	Key                                    int    `json:"_key"`
	AreSmallAttributePassiveSkillsReplaced bool   `json:"Var1"`
	AreSmallNormalPassiveSkillsReplaced    bool   `json:"Var2"`
	MinimumAdditions                       int    `json:"Var5"`
	MaximumAdditions                       int    `json:"Var6"`
	NotableReplacementSpawnWeight          int    `json:"Var9"`
}
