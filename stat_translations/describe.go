package st

import (
	"fmt"
	"io/fs"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	placeholderRegex = regexp.MustCompile(`\{(\d*)(?::([^}]*))?\}`)
	decimalsRegex    = regexp.MustCompile(`_(\d)dp(_if_required)?$`)
)

// valueDivisors are the index handlers that only rescale a value.
var valueDivisors = map[string]float64{
	"per_minute_to_per_second": 60,
	"milliseconds_to_seconds":  1000,
	"deciseconds_to_seconds":   10,
	"divide_by_one_hundred":    100,
	"locations_to_metres":      10,
	"divide_by_ten":            10,
}

// defaultDecimals is how many decimals a rescaled value keeps when its handler doesn't say.
const defaultDecimals = 2

type limit struct {
	min, max int
	not      bool
}

func (l limit) matches(v int) bool {
	return (v >= l.min && v <= l.max) != l.not
}

type handler struct {
	name, arg string
}

type variant struct {
	text string
	// qualifier limits the line to a context such as gem_quality, which Describe never renders in.
	qualifier string
	limits    []limit
	handlers  []handler
}

func isInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// parseLimit parses `#`, `N`, `!N` and `min|max` where either side may be `#`.
func parseLimit(s string) limit {
	l := limit{min: math.MinInt, max: math.MaxInt}
	if rest, ok := strings.CutPrefix(s, "!"); ok {
		l.not = true
		s = rest
	}
	lo, hi, ranged := strings.Cut(s, "|")
	if !ranged {
		hi = lo
	}
	if n, err := strconv.Atoi(lo); err == nil {
		l.min = n
	}
	if n, err := strconv.Atoi(hi); err == nil {
		l.max = n
	}
	return l
}

// NewDescriber parses name and, first, every file it includes, for Describe.
func NewDescriber(loader fs.FS, name string) (*TranslationParser, error) {
	t := NewTranslationParser(loader)
	t.followIncludes = true
	if err := t.ParseFile(name); err != nil {
		return nil, err
	}
	return t, nil
}

// Described is what the game shows for a set of stats.
type Described struct {
	// Lines are in description file order, which is the order the game lists them in.
	Lines []string
	// Reminders are the ReminderText ids the rendered lines reference, in order.
	Reminders []string
	// Undescribed are the non-zero stats no description mentions; the game shows nothing for them.
	Undescribed []string
}

// Describe renders stats the way the game does: every description matching a non-zero stat is
// rendered once with the values of all its stats, using the first variant whose limits match.
func (t *TranslationParser) Describe(ids []string, values []int) (*Described, error) {
	if len(ids) != len(values) {
		return nil, fmt.Errorf("%d stats but %d values", len(ids), len(values))
	}

	out := &Described{}
	byID := make(map[string]int, len(ids))
	var descriptions []*Description
	for i, id := range ids {
		if values[i] == 0 {
			continue
		}
		byID[id] += values[i]

		d, ok := t.descriptors[id]
		switch {
		case !ok:
			out.Undescribed = append(out.Undescribed, id)
		case d.Stats == nil:
			// no_description
		case !slices.Contains(descriptions, d):
			descriptions = append(descriptions, d)
		}
	}
	slices.SortFunc(descriptions, func(a, b *Description) int { return a.Order - b.Order })

	for _, d := range descriptions {
		values := make([]int, len(d.Stats))
		for i, id := range d.Stats {
			values[i] = byID[id]
		}
		line, reminders, err := d.render(values)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", strings.Join(d.Stats, " "), err)
		}
		if line != "" {
			out.Lines = append(out.Lines, line)
			out.Reminders = append(out.Reminders, reminders...)
		}
	}
	return out, nil
}

type number struct {
	value    float64
	decimals int
	fixed    bool
}

func (n number) format(sign bool) string {
	s := strconv.FormatFloat(n.value, 'f', -1, 64)
	if n.fixed {
		s = strconv.FormatFloat(n.value, 'f', n.decimals, 64)
	} else if n.value != math.Trunc(n.value) {
		p := math.Pow10(n.decimals)
		s = strconv.FormatFloat(math.Round(n.value*p)/p, 'f', -1, 64)
	}
	if sign && n.value >= 0 {
		return "+" + s
	}
	return s
}

// render returns the empty string when no variant matches: the game hides the stats then.
func (d *Description) render(values []int) (string, []string, error) {
	idx := slices.IndexFunc(d.variants, func(v variant) bool {
		if v.qualifier != "" {
			return false
		}
		for i, l := range v.limits {
			if i >= len(values) || !l.matches(values[i]) {
				return false
			}
		}
		return true
	})
	if idx < 0 {
		return "", nil, nil
	}
	v := d.variants[idx]

	numbers := make([]number, len(values))
	for i, value := range values {
		numbers[i] = number{value: float64(value), decimals: defaultDecimals}
	}

	var reminders []string
	for _, h := range v.handlers {
		switch h.name {
		case "reminderstring":
			reminders = append(reminders, h.arg)
			continue
		case "canonical_line":
			continue
		}

		i, err := strconv.Atoi(h.arg)
		if err != nil || i < 1 || i > len(numbers) {
			return "", nil, fmt.Errorf("handler %s has bad stat index %q", h.name, h.arg)
		}
		n := &numbers[i-1]

		name := h.name
		if m := decimalsRegex.FindStringSubmatch(name); m != nil {
			n.decimals, _ = strconv.Atoi(m[1])
			n.fixed = m[2] == ""
			name = strings.TrimSuffix(name, m[0])
		}
		switch divisor, ok := valueDivisors[name]; {
		case ok:
			n.value /= divisor
		case name == "negate":
			n.value = -n.value
		case name == "add_one":
			n.value++
		case name == "canonical_stat":
		default:
			return "", nil, fmt.Errorf("unsupported index handler %s", h.name)
		}
	}

	next := 0
	var bad error
	text := placeholderRegex.ReplaceAllStringFunc(v.text, func(p string) string {
		m := placeholderRegex.FindStringSubmatch(p)
		i := next
		if m[1] != "" {
			i, _ = strconv.Atoi(m[1])
		} else {
			next++
		}
		if i >= len(numbers) {
			bad = fmt.Errorf("placeholder %s but %d stats", p, len(numbers))
			return p
		}
		return numbers[i].format(strings.Contains(m[2], "+"))
	})
	return text, reminders, bad
}
