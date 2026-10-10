package raw

type CurrencyExchange struct {
	Item                     int  `json:"Item"`
	Category                 int  `json:"Category"`
	SubCategory              int  `json:"SubCategory"`
	GoldPurchaseFee          int  `json:"GoldPurchaseFee"`
	Key                      int  `json:"_key"`
	EnabledInStandardLeague  bool `json:"EnabledInStandardLeague"`
	EnabledInChallengeLeague bool `json:"EnabledInChallengeLeague"`
}
