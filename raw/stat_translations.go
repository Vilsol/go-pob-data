package raw

// Condition is the limit of one stat, in the order of StatTranslation.IDs. A `#` limit has neither
// Min nor Max.
type Condition struct {
	Min     *int `json:"min,omitempty"`
	Max     *int `json:"max,omitempty"`
	Negated bool `json:"negated"`
}

// IndexHandler is one handler of a line, such as `negate 1` or `reminderstring ReminderTextBleeding`.
// Arg is a 1-based stat index, a ReminderText id, or empty for handlers like canonical_line.
type IndexHandler struct {
	Name string `json:"name"`
	Arg  string `json:"arg,omitempty"`
}

type LangTranslation struct {
	// IndexHandlers keeps one Arg per handler name, so it loses a handler applied to several stats;
	// prefer Handlers, which has every handler in order.
	IndexHandlers IndexHandlers `json:"index_handlers,omitempty"`
	String        string        `json:"string"`
	// Qualifier restricts the line to a context, such as gem_quality.
	Qualifier  string         `json:"qualifier,omitempty"`
	Conditions []Condition    `json:"conditions,omitempty"`
	Handlers   []IndexHandler `json:"handlers,omitempty"`
}

type StatTranslation struct {
	// Identifier names the description in files that declare has_identifiers.
	Identifier string   `json:"identifier,omitempty"`
	IDs        []string `json:"ids"`
	// OffHandIDs are the off hand stats of a handed_description; IDs are the main hand ones.
	OffHandIDs []string          `json:"off_hand_ids,omitempty"`
	List       []LangTranslation `json:"list"`
}
