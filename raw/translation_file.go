package raw

type TranslationFile struct {
	Descriptors []*StatTranslation `json:"descriptors"`
	Includes    []string           `json:"includes"`
	// NoDescription are the stats the file hides, including ones an included file describes.
	NoDescription []string `json:"no_description,omitempty"`
}
