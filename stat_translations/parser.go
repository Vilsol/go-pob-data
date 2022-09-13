package stat_translations

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/unicode"
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
	order       int
	descriptors map[string]*Description
	loader      fs.FS
	loaded      map[string]bool
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
	Text    string            `json:"text"`
	Limit   [][2]string       `json:"limit,omitempty"`
	Special map[string]string `json:"special,omitempty"`
}

type Description struct {
	Lang  map[string][]Translation
	Order int
	Stats []string
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

	os.WriteFile(path.Base(name), all, 0755)

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
						desc.Special[string(match[0][1])] = string(match[0][2])
					}

					curDescriptor.Lang[curLang] = append(curDescriptor.Lang[curLang], desc)
				}
			}
		}
	}

	return nil
}
