package raw

type ArmourType struct {
	ArmourMax              int `json:"ArmourMax"`
	ArmourMin              int `json:"ArmourMin"`
	BaseItemTypesKey       int `json:"BaseItemTypesKey"`
	EnergyShieldMax        int `json:"EnergyShieldMax"`
	EnergyShieldMin        int `json:"EnergyShieldMin"`
	EvasionMax             int `json:"EvasionMax"`
	EvasionMin             int `json:"EvasionMin"`
	IncreasedMovementSpeed int `json:"IncreasedMovementSpeed"`
	WardMax                int `json:"WardMax"`
	WardMin                int `json:"WardMin"`
	Key                    int `json:"_key"`
}
