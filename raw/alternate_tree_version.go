package raw

type AlternateTreeVersion struct {
	ConquerorType                          string `json:"Id"`
	Key                                    int    `json:"_key"`
	AreSmallAttributePassiveSkillsReplaced bool   `json:"Var1"`
	AreSmallNormalPassiveSkillsReplaced    bool   `json:"Var2"`
	MinimumAdditionsSmallAttribute         int    `json:"Var3"`
	MaximumAdditionsSmallAttribute         int    `json:"Var4"`
	MinimumAdditions                       int    `json:"Var5"`
	MaximumAdditions                       int    `json:"Var6"`
	MinimumAdditionsNotable                int    `json:"Var7"`
	MaximumAdditionsNotable                int    `json:"Var8"`
	NotableReplacementSpawnWeight          int    `json:"Var9"`
}
