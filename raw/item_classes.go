package raw

type ItemClass struct {
	ItemStance           *int   `json:"ItemStance"`
	ItemClassCategory    *int   `json:"ItemClassCategory"`
	Name                 string `json:"Name"`
	ID                   string `json:"Id"`
	Flags                []int  `json:"Flags"`
	Key                  int    `json:"_key"`
	CanBeDoubleCorrupted bool   `json:"CanBeDoubleCorrupted"`
	CanHaveInfluence     bool   `json:"CanHaveInfluence"`
	CanHaveVeiledMods    bool   `json:"CanHaveVeiledMods"`
	CanScourge           bool   `json:"CanScourge"`
	CanTransferSkin      bool   `json:"CanTransferSkin"`
	CanHaveIncubators    bool   `json:"CanHaveIncubators"`
	CanHaveAspects       bool   `json:"CanHaveAspects"`
	AllocateToMapOwner   bool   `json:"AllocateToMapOwner"`
	CanBeCorrupted       bool   `json:"CanBeCorrupted"`
	AlwaysShow           bool   `json:"AlwaysShow"`
	RemovedIfLeavesArea  bool   `json:"RemovedIfLeavesArea"`
	AlwaysAllocate       bool   `json:"AlwaysAllocate"`
}
