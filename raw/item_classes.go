package raw

type ItemClass struct {
	AllocateToMapOwner   bool   `json:"AllocateToMapOwner"`
	AlwaysAllocate       bool   `json:"AlwaysAllocate"`
	AlwaysShow           bool   `json:"AlwaysShow"`
	CanBeCorrupted       bool   `json:"CanBeCorrupted"`
	CanBeDoubleCorrupted bool   `json:"CanBeDoubleCorrupted"`
	CanHaveAspects       bool   `json:"CanHaveAspects"`
	CanHaveIncubators    bool   `json:"CanHaveIncubators"`
	CanHaveInfluence     bool   `json:"CanHaveInfluence"`
	CanHaveVeiledMods    bool   `json:"CanHaveVeiledMods"`
	CanScourge           bool   `json:"CanScourge"`
	CanTransferSkin      bool   `json:"CanTransferSkin"`
	Flags                []int  `json:"Flags"`
	ID                   string `json:"Id"`
	ItemClassCategory    *int   `json:"ItemClassCategory"`
	ItemStance           *int   `json:"ItemStance"`
	Name                 string `json:"Name"`
	RemovedIfLeavesArea  bool   `json:"RemovedIfLeavesArea"`
	Key                  int    `json:"_key"`
}
