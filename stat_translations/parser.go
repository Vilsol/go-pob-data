package st

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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
	statLimitPipeRegex   = regexp.MustCompile(`^([\d\-#]+?)\|([\d\-#]+?)(?:\|.+?)?$`)
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

type Description struct {
	Lang  map[string][]raw.LangTranslation
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
				Lang: map[string][]raw.LangTranslation{
					curLang: make([]raw.LangTranslation, 0),
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
				curDescriptor.Lang[curLang] = make([]raw.LangTranslation, 0)
			} else {
				statMatches := statLineRegex.FindAllSubmatch(line, -1)
				if len(statMatches) > 0 {
					statLimits := statMatches[0][1]
					special := statMatches[0][3]

					desc := raw.LangTranslation{
						String:        string(statMatches[0][2]),
						Conditions:    make([]raw.Condition, 0),
						IndexHandlers: make(map[string]string),
					}

					statLimitMatch := statLimitRegex.FindAllSubmatch(statLimits, -1)
					for _, match := range statLimitMatch {
						statLimit := string(match[0])

						if statLimit == "#" {
							// Do nothing
							continue
						}

						if statLimitNumberRegex.MatchString(statLimit) {
							n, _ := strconv.Atoi(statLimit)
							desc.Conditions = append(desc.Conditions, raw.Condition{
								Min: &n,
								Max: &n,
							})
						} else {
							negate := statLimitNegateRegex.FindAllStringSubmatch(statLimit, -1)
							if len(negate) > 0 {
								n, _ := strconv.Atoi(negate[0][1])
								desc.Conditions = append(desc.Conditions, raw.Condition{
									Negated: true,
									Min:     &n,
									Max:     &n,
								})
							} else {
								pipeMatch := statLimitPipeRegex.FindAllStringSubmatch(statLimit, -1)
								n1, _ := strconv.Atoi(pipeMatch[0][1])
								n2, _ := strconv.Atoi(pipeMatch[0][2])

								condition := raw.Condition{}

								if pipeMatch[0][1] != "#" {
									condition.Min = &n1
								}

								if pipeMatch[0][2] != "#" {
									condition.Min = &n2
								}

								desc.Conditions = append(desc.Conditions, condition)
							}
						}
					}

					specialMatch := specialRegex.FindAllSubmatch(special, -1)
					for _, match := range specialMatch {
						desc.IndexHandlers[string(match[1])] = string(match[2])
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

			fullOut[lang] = append(fullOut[lang], &raw.StatTranslation{
				IDs:  description.Stats,
				List: translations,
			})
		}
	}

	for lang, translations := range fullOut {
		b, err := json.Marshal(translations)
		if err != nil {
			return errors.Wrap(err, "failed to marshal translations file")
		}

		fullPath := filepath.Join(outDir, languageMap[lang], translationName+".msgpack.br")
		println("writing to", fullPath)

		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return errors.Wrap(err, "failed to make translation directories")
		}

		f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
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

		outPathGzip := filepath.Join(outDir, languageMap[lang], translationName+".json.gz")
		println("writing to", outPathGzip)

		fGzip, err := os.OpenFile(outPathGzip, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
		if err != nil {
			return errors.Wrap(err, "could not open file for writing")
		}

		writerGzip := gzip.NewWriter(fGzip)
		if _, err := writerGzip.Write(b); err != nil {
			return errors.Wrap(err, "could not compress to gzip")
		}

		_ = writerGzip.Close()
		_ = fGzip.Close()
	}

	return nil
}
