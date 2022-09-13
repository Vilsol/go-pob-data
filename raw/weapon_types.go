package raw

type WeaponType struct {
	BaseItemTypesKey int `json:"BaseItemTypesKey"`
	Critical         int `json:"Critical"`
	DamageMax        int `json:"DamageMax"`
	DamageMin        int `json:"DamageMin"`
	Null6            int `json:"Null6"`
	RangeMax         int `json:"RangeMax"`
	Speed            int `json:"Speed"`
	Key              int `json:"_key"`
}
