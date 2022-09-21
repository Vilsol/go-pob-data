package raw

type Condition struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

type LangTranslation struct {
	IndexHandlers map[string]string `json:"index_handlers,omitempty"`
	String        string            `json:"string"`
	Conditions    []Condition       `json:"conditions,omitempty"`
}

type StatTranslation struct {
	IDs  []string          `json:"ids"`
	List []LangTranslation `json:"list"`
}
