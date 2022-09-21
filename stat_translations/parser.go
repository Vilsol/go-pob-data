package st

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"golang.org/x/text/encoding/unicode"

	"github.com/Vilsol/go-pob-data/raw"
)

var (
	lineRegex            = regexp.MustCompile(`[^\r\n]+`)
	includeRegex         = regexp.MustCompile(`include "(Metadata/StatDescriptions/.+\.txt)"$`)
	noDescRegex          = regexp.MustCompile(`no_description ([\w_%+\-]+)`)
	langRegex            = regexp.MustCompile(`lang "(.+)"`)
	statLineRegex        = regexp.MustCompile(`([\d%\-#| ]+) "(.*)" *(.*)`)
	statLimitRegex       = regexp.MustCompile(`[!\d\-#|]+`)
	statLimitNumberRegex = regexp.MustCompile(`^-?\d+$`)
	statLimitNegateRegex = regexp.MustCompile(`^!(-?\d+)$`)
	statLimitPipeRegex   = regexp.MustCompile(`^([\d\-#]+)|([\d\-#]+)$`)
	specialRegex         = regexp.MustCompile(`([\w%_]+) (\w+)`)
	statsRegex           = regexp.MustCompile(`\d+\s+([\w_+\-% ]+)`)
	statRegex            = regexp.MustCompile(`[\w_+\-%]+`)
)

type TranslationParser struct {
	loader      fs.FS
	descriptors map[string]*Description
	loaded      map[string]bool
	order       int
}

func NewTranslationParser(loader fs.FS) *TranslationParser {
	return &TranslationParser{
		order:       1,
		descriptors: make(map[string]*Description),
		loader:      loader,
		loaded:      make(map[string]bool),
	}
}

type Translation struct {
	Special map[string]string `json:"special,omitempty"`
	Text    string            `json:"text"`
	Limit   [][2]string       `json:"limit,omitempty"`
}

type Description struct {
	Lang  map[string][]Translation
	Stats []string
	Order int
}

func (t *TranslationParser) ParseFile(name string) error {
	if _, ok := t.loaded[name]; ok {
		return nil
	}
	t.loaded[name] = true

	println("parsing", name)

	data, err := t.loader.Open(name)
	if err != nil {
		return errors.Wrap(err, "failed to open input file")
	}

	parsed := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Reader(data)
	all, err := io.ReadAll(parsed)
	if err != nil {
		return errors.Wrap(err, "failed to parse input file")
	}

	var curDescriptor *Description
	curLang := "English"

	for _, line := range lineRegex.FindAll(all, -1) {
		include := includeRegex.FindAllSubmatch(line, -1)
		if len(include) > 0 {
			if err := t.ParseFile(string(include[0][1])); err != nil {
				return err
			}
			continue
		}

		noDesc := noDescRegex.FindAllSubmatch(line, -1)
		if len(noDesc) > 0 {
			t.descriptors[string(noDesc[0][1])] = &Description{
				Order: 0,
			}
		} else if bytes.Contains(line, []byte("handed_description")) || (bytes.Contains(line, []byte("description")) && !bytes.Contains(line, []byte("_description"))) {
			curLang = "English"
			curDescriptor = &Description{
				Lang: map[string][]Translation{
					curLang: make([]Translation, 0),
				},
				Order: t.order,
			}
			t.order++
		} else if curDescriptor.Stats == nil {
			stats := statsRegex.FindAllSubmatch(line, -1)
			if len(stats) > 0 {
				statNames := statRegex.FindAll(stats[0][1], -1)
				curDescriptor.Stats = make([]string, len(statNames))
				for i, statName := range statNames {
					statNameStr := string(statName)
					curDescriptor.Stats[i] = statNameStr
					t.descriptors[statNameStr] = curDescriptor
				}
			}
		} else {
			langName := langRegex.FindAllSubmatch(line, -1)
			if len(langName) > 0 {
				curLang = string(langName[0][1])
				curDescriptor.Lang[curLang] = make([]Translation, 0)
			} else {
				statMatches := statLineRegex.FindAllSubmatch(line, -1)
				if len(statMatches) > 0 {
					statLimits := statMatches[0][1]
					special := statMatches[0][3]

					desc := Translation{
						Text:    string(statMatches[0][2]),
						Limit:   make([][2]string, 0),
						Special: make(map[string]string),
					}

					statLimitMatch := statLimitRegex.FindAllSubmatch(statLimits, -1)
					for _, match := range statLimitMatch {
						statLimit := string(match[0])

						if statLimit == "#" {
							// Do nothing
						} else if statLimitNumberRegex.MatchString(statLimit) {
							desc.Limit = append(desc.Limit, [2]string{
								statLimit,
								statLimit,
							})
						} else {
							negate := statLimitNegateRegex.FindAllStringSubmatch(statLimit, -1)
							if len(negate) > 0 {
								desc.Limit = append(desc.Limit, [2]string{
									"!",
									negate[0][1],
								})
							} else {
								pipeMatch := statLimitPipeRegex.FindAllStringSubmatch(statLimit, -1)
								desc.Limit = append(desc.Limit, [2]string{
									pipeMatch[0][1],
									pipeMatch[0][2],
								})
							}
						}
					}

					specialMatch := specialRegex.FindAllSubmatch(special, -1)
					for _, match := range specialMatch {
						desc.Special[string(match[1])] = string(match[2])
					}

					curDescriptor.Lang[curLang] = append(curDescriptor.Lang[curLang], desc)
				}
			}
		}
	}

	return nil
}

var languageMap = map[string]string{
	"English":             "en",
	"German":              "de",
	"Korean":              "kr",
	"Russian":             "ru",
	"Portuguese":          "po",
	"Spanish":             "es",
	"French":              "fr",
	"Traditional Chinese": "tw",
	"Simplified Chinese":  "cn",
	"Thai":                "th",
	"Japanese":            "jp",
}

func (t *TranslationParser) SaveTo(outDir string, translationName string) error {
	fullOut := make(map[string][]*raw.StatTranslation)
	for _, description := range t.descriptors {
		for lang, translations := range description.Lang {
			if _, ok := fullOut[lang]; !ok {
				fullOut[lang] = make([]*raw.StatTranslation, 0)
			}

			outTranslations := make([]raw.LangTranslation, len(translations))
			for i, trans := range translations {
				conditions := make([]raw.Condition, 0)
				for _, limits := range trans.Limit {
					if limits[0] != "" || limits[1] != "" {
						cond := raw.Condition{}
						if limits[0] != "" {
							temp, _ := strconv.Atoi(limits[0])
							cond.Min = &temp
						}
						if limits[1] != "" {
							temp, _ := strconv.Atoi(limits[1])
							cond.Max = &temp
						}
						conditions = append(conditions, cond)
					}
				}

				outTranslations[i] = raw.LangTranslation{
					Conditions:    conditions,
					IndexHandlers: trans.Special,
					String:        trans.Text,
				}
			}

			fullOut[lang] = append(fullOut[lang], &raw.StatTranslation{
				IDs:  description.Stats,
				List: outTranslations,
			})
		}
	}

	for lang, translations := range fullOut {
		fullPath := filepath.Join(outDir, languageMap[lang], translationName+".msgpack.br")
		println("writing to", fullPath)

		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return errors.Wrap(err, "failed to make translation directories")
		}

		f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0o755)
		if err != nil {
			return errors.Wrap(err, "failed to open translation file")
		}

		writerMsgpBrotli := brotli.NewWriter(f)

		msg := msgp.NewWriter(writerMsgpBrotli)
		if err := msg.WriteIntf(translations); err != nil {
			return errors.Wrap(err, "failed to encode msgpack")
		}

		if err := msg.Flush(); err != nil {
			return errors.Wrap(err, "failed to flush msgpack")
		}

		_ = writerMsgpBrotli.Close()
		_ = f.Close()
	}

	return nil
}
