package raw

type PassiveJewelRadius struct {
	ClientString    *int   `json:"ClientString"`
	ID              string `json:"ID"`
	RingOuterRadius int    `json:"RingOuterRadius"`
	RingInnerRadius int    `json:"RingInnerRadius"`
	Radius          int    `json:"Radius"`
	Key             int    `json:"_key"`
}
