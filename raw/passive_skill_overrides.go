package raw

type PassiveSkillOverride struct {
	Effect                *int   `json:"Effect"`
	Limit                 *int   `json:"Limit"`
	AllocatedPassiveSkill *int   `json:"AllocatedPassiveSkill"`
	ID                    string `json:"Id"`
	Name                  string `json:"Name"`
	NodeIcon              string `json:"NodeIcon"`
	PassiveBG             string `json:"PassiveBG"`
	Stats                 []int  `json:"Stats"`
	StatValues            []int  `json:"StatValues"`
	TattooBlockingPassive []int  `json:"TattooBlockingPassive"`
	Hash16                int    `json:"HASH16"`
	Type                  int    `json:"Type"`
	RequiresAdjacent      int    `json:"RequiresAdjacent"`
	MaxAdjacent           int    `json:"MaxAdjacent"`
	Key                   int    `json:"_key"`
}
