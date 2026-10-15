package st

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"golang.org/x/text/encoding/unicode"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

var (
	includeRegex     = regexp.MustCompile(`^include "(Metadata/StatDescriptions/.+\.txt|Data/StatDescriptions/.+\.csd)"$`)
	noDescRegex      = regexp.MustCompile(`^no_description (\S+)$`)
	descriptionRegex = regexp.MustCompile(`^(handed_)?description(?: (\S+))?$`)
	langRegex        = regexp.MustCompile(`^lang "(.+)"$`)
	countRegex       = regexp.MustCompile(`^\d+$`)
	statsRegex       = regexp.MustCompile(`^(\d+)\s+(.+)$`)
	variantRegex     = regexp.MustCompile(`^([-\d#|!]+(?:\s+[-\d#|!]+)*)\s+(?:([a-z_]+)\s+)?"(.*)"\s*(.*)$`)
	// A third part, as in a `1|1|#` typo of the game, is ignored like the game and PoB do.
	limitRegex = regexp.MustCompile(`^(!)?(-?\d+|#)(?:\|(-?\d+|#))?(?:\|(?:-?\d+|#))*$`)
)

// defaultLang is the language of a description's lines before any lang block.
const defaultLang = "English"

type TranslationParser struct {
	loader      fs.FS
	descriptors map[string]*Description
	loaded      map[string]bool
	// ordered are this file's descriptions in file order, without those of included files.
	ordered       []*Description
	includes      []string
	noDescription []string
	order         int
	// followIncludes parses included files in place, so a file's own descriptions override theirs.
	followIncludes bool
}

func NewTranslationParser(loader fs.FS) *TranslationParser {
	return &TranslationParser{
		order:       1,
		descriptors: make(map[string]*Description),
		includes:    make([]string, 0),
		loader:      loader,
		loaded:      make(map[string]bool),
	}
}

type Description struct {
	Lang       map[string][]raw.LangTranslation
	Identifier string
	Stats      []string
	OffHand    []string
	// variants are the English lines as Describe renders them.
	variants []variant
	Order    int
}

// descriptionState is what the next line of a description must be.
type descriptionState int

const (
	wantStats descriptionState = iota
	wantMainHandStats
	wantOffHandStats
	wantCount
	wantVariant
)

// ParseFile parses a stat description file. Every line must be one the format defines: anything else
// fails the parse rather than being dropped.
func (t *TranslationParser) ParseFile(name string) error {
	if _, ok := t.loaded[name]; ok {
		return nil
	}
	t.loaded[name] = true

	slog.Info("parsing", slog.String("name", name))

	data, err := t.loader.Open(name)
	if err != nil {
		return errors.Wrap(err, "failed to open input file")
	}

	parsed := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Reader(data)
	all, err := io.ReadAll(parsed)
	if err != nil {
		return errors.Wrap(err, "failed to parse input file")
	}

	var cur *Description
	var handed bool
	state := wantStats
	curLang := defaultLang
	remaining := 0
	statCount := 0
	// repeated is the first block of a language the current block repeats, which PoE2 files do.
	var repeated []raw.LangTranslation

	for i, rawLine := range bytes.Split(all, []byte("\n")) {
		line := strings.TrimSpace(strings.TrimPrefix(string(rawLine), "\ufeff"))
		if line == "" {
			continue
		}

		fail := func(reason string) error {
			return errors.Errorf("%s:%d: %s: %q", name, i+1, reason, line)
		}
		endRepeat := func() error {
			if remaining > 0 || repeated == nil {
				return nil
			}
			if !reflect.DeepEqual(repeated, cur.Lang[curLang]) {
				return fail("language repeated with different lines")
			}
			repeated = nil
			return nil
		}

		// A description ends after its last variant, so these only start a new one there.
		topLevel := cur == nil || (state == wantVariant && remaining == 0)

		switch {
		case topLevel && includeRegex.MatchString(line):
			include := includeRegex.FindStringSubmatch(line)[1]
			t.includes = append(t.includes, include)
			if t.followIncludes {
				if err := t.ParseFile(strings.ToLower(include)); err != nil {
					return err
				}
			}
			cur = nil
		case topLevel && noDescRegex.MatchString(line):
			stat := noDescRegex.FindStringSubmatch(line)[1]
			t.descriptors[stat] = &Description{}
			t.noDescription = append(t.noDescription, stat)
			cur = nil
		case topLevel && (line == "has_identifiers" || line == "no_identifiers"):
			cur = nil
		case topLevel && descriptionRegex.MatchString(line):
			m := descriptionRegex.FindStringSubmatch(line)
			handed = m[1] != ""
			cur = &Description{
				Lang:       map[string][]raw.LangTranslation{},
				Identifier: m[2],
				Order:      t.order,
			}
			t.order++
			t.ordered = append(t.ordered, cur)
			state, curLang = wantStats, defaultLang
		case cur != nil && state == wantVariant && remaining == 0 && langRegex.MatchString(line):
			curLang = langRegex.FindStringSubmatch(line)[1]
			if first, ok := cur.Lang[curLang]; ok {
				if curLang == defaultLang {
					return fail("language repeated")
				}
				repeated = first
			}
			state = wantCount
		case cur != nil && state == wantStats && handed && countRegex.MatchString(line):
			// A handed description lists its count, then the main hand stats, then the off hand ones.
			statCount, _ = strconv.Atoi(line)
			state = wantMainHandStats
		case cur != nil && state == wantMainHandStats:
			cur.Stats = strings.Fields(line)
			if len(cur.Stats) != statCount {
				return fail("stat count does not match the stats")
			}
			for _, stat := range cur.Stats {
				t.descriptors[stat] = cur
			}
			state = wantOffHandStats
		case cur != nil && state == wantStats:
			m := statsRegex.FindStringSubmatch(line)
			if m == nil {
				return fail("expected the stat list")
			}
			cur.Stats = strings.Fields(m[2])
			if n, _ := strconv.Atoi(m[1]); n != len(cur.Stats) {
				return fail("stat count does not match the stats")
			}
			for _, stat := range cur.Stats {
				t.descriptors[stat] = cur
			}
			state = wantCount
			if handed {
				state = wantOffHandStats
			}
		case cur != nil && state == wantOffHandStats:
			cur.OffHand = strings.Fields(line)
			if len(cur.OffHand) != len(cur.Stats) {
				return fail("off hand stats do not match the main hand stats")
			}
			state = wantCount
		case cur != nil && state == wantCount && countRegex.MatchString(line):
			remaining, _ = strconv.Atoi(line)
			cur.Lang[curLang] = make([]raw.LangTranslation, 0, remaining)
			state = wantVariant
			if err := endRepeat(); err != nil {
				return err
			}
		case cur != nil && state == wantVariant && remaining > 0:
			translation, v, err := parseLine(line, len(cur.Stats))
			if err != nil {
				return fail(err.Error())
			}
			cur.Lang[curLang] = append(cur.Lang[curLang], translation)
			if curLang == defaultLang {
				cur.variants = append(cur.variants, v)
			}
			remaining--
			if err := endRepeat(); err != nil {
				return err
			}
		default:
			return fail("unexpected line")
		}
	}

	if cur != nil && (state != wantVariant || remaining != 0) {
		return errors.Errorf("%s: last description is incomplete", name)
	}

	return nil
}

// parseLine parses a line like `1|# # gem_quality "{0}% increased {1}" negate 1 reminderstring ReminderTextX`
// of a description of statCount stats.
func parseLine(line string, statCount int) (raw.LangTranslation, variant, error) {
	m := variantRegex.FindStringSubmatch(line)
	if m == nil {
		return raw.LangTranslation{}, variant{}, errors.New("expected a variant")
	}

	limits := strings.Fields(m[1])
	if len(limits) != statCount {
		return raw.LangTranslation{}, variant{}, errors.Errorf("%d limits for %d stats", len(limits), statCount)
	}

	out := raw.LangTranslation{
		String:        m[3],
		Qualifier:     m[2],
		Conditions:    make([]raw.Condition, 0, len(limits)),
		IndexHandlers: make(raw.IndexHandlers),
	}
	v := variant{text: strings.ReplaceAll(m[3], `\n`, "\n"), qualifier: m[2]}

	for _, l := range limits {
		lm := limitRegex.FindStringSubmatch(l)
		if lm == nil {
			return raw.LangTranslation{}, variant{}, errors.Errorf("bad limit %q", l)
		}
		lo, hi := lm[2], lm[3]
		if lm[3] == "" {
			hi = lo
		}
		c := raw.Condition{Negated: lm[1] != ""}
		if n, err := strconv.Atoi(lo); err == nil {
			c.Min = &n
		}
		if n, err := strconv.Atoi(hi); err == nil {
			c.Max = &n
		}
		out.Conditions = append(out.Conditions, c)
		v.limits = append(v.limits, parseLimit(l))
	}

	tokens := strings.Fields(m[4])
	for i := 0; i < len(tokens); i++ {
		h := raw.IndexHandler{Name: tokens[i]}
		// Handlers take a 1-based stat index, reminderstring a ReminderText id, canonical_line nothing.
		if i+1 < len(tokens) && (h.Name == "reminderstring" || isInt(tokens[i+1])) {
			i++
			h.Arg = tokens[i]
		}
		out.Handlers = append(out.Handlers, h)
		out.IndexHandlers[h.Name] = h.Arg
		v.handlers = append(v.handlers, handler{name: h.Name, arg: h.Arg})
	}

	return out, v, nil
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
	// Every language gets the file, so a file without descriptions of its own still carries its
	// includes and hidden stats.
	fullOut := make(map[string]*raw.TranslationFile, len(languageMap))
	for lang := range languageMap {
		fullOut[lang] = &raw.TranslationFile{
			Descriptors:   make([]*raw.StatTranslation, 0),
			Includes:      t.includes,
			NoDescription: t.noDescription,
		}
	}

	for _, description := range t.ordered {
		for lang, translations := range description.Lang {
			if _, ok := fullOut[lang]; !ok {
				return errors.Errorf("unknown language %q", lang)
			}

			fullOut[lang].Descriptors = append(fullOut[lang].Descriptors, &raw.StatTranslation{
				Identifier: description.Identifier,
				IDs:        description.Stats,
				OffHandIDs: description.OffHand,
				List:       translations,
			})
		}
	}

	for lang, translations := range fullOut {
		b, err := json.Marshal(translations)
		if err != nil {
			return errors.Wrap(err, "failed to marshal translations file")
		}

		fullPath := loader.AssetPath(filepath.Join(outDir, languageMap[lang], translationName), loader.TupleVariant(translations))
		slog.Info("writing to", slog.String("path", fullPath))

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
		slog.Info("writing to", slog.String("path", outPathGzip))

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
