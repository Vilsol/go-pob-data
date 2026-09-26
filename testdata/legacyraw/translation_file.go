package legacyraw

type TranslationFile struct {
	Descriptors []*StatTranslation `json:"descriptors"`
	Includes    []string           `json:"includes"`
}
