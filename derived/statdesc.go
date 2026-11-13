package derived

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

// The stat description reader of PoB's src/Export/statdesc.lua, which every export script describes
// stats through.

type sdSpec struct {
	V any
	K string
}

type sdVariant struct {
	Text       string
	Limit      [][2]any
	Specs      []sdSpec
	GemQuality bool
}

var (
	gggMarkupRe = regexp.MustCompile(`<[^>]+>\{([^}]+)\}`)
	gggLinkRe   = regexp.MustCompile(`\[([^|\]]+)\]`)
	gggAliasRe  = regexp.MustCompile(`\[[^|]+\|([^|]+)\]`)
	includeRe   = regexp.MustCompile(`^Metadata/StatDescriptions/(.+)\.(?:txt|csd)$`)
)

//pob:src/Modules/Common.lua:1110-1116
func escapeGGGString(text string) string {
	text = gggMarkupRe.ReplaceAllString(text, "$1")
	text = gggLinkRe.ReplaceAllString(text, "$1")
	return gggAliasRe.ReplaceAllString(text, "$1")
}

//pob:block:src/Export/Scripts/statdesc.lua:48-69
func sdLimit(c raw.Condition) [2]any {
	end := func(p *int) any {
		if p == nil {
			return "#"
		}
		return *p
	}
	if c.Negated && c.Min != nil && c.Max != nil && *c.Min == *c.Max {
		return [2]any{"!", *c.Min}
	}
	return [2]any{end(c.Min), end(c.Max)}
}

//pob:block:src/Export/Scripts/statdesc.lua:70-97
func sdSpecs(handlers []raw.IndexHandler) []sdSpec {
	var tokens []string
	for _, h := range handlers {
		tokens = append(tokens, h.Name)
		if h.Arg != "" {
			tokens = append(tokens, h.Arg)
		}
	}
	specs := []sdSpec{}
	for i := 0; i < len(tokens); {
		if tokens[i] == "canonical_line" {
			specs = append(specs, sdSpec{K: "canonical_line", V: true})
			i++
			continue
		}
		if i+1 < len(tokens) {
			var v any = tokens[i+1]
			if n, err := strconv.ParseFloat(tokens[i+1], 64); err == nil {
				v = n
			}
			specs = append(specs, sdSpec{K: tokens[i], V: v})
			i += 2
		} else {
			i++
		}
	}
	return specs
}

type edDescriptor struct {
	stats    []string
	variants []sdVariant
	order    int
}

// source is where a describer reads its stat description files: version's English bundle.
type source struct {
	ctx        context.Context
	assetCache loader.AssetCache
	version    string
}

type exportDescriber struct {
	descriptors map[string]*edDescriptor
	src         source
}

type edRange struct {
	fmt        string
	min, max   float64
	minZ, maxZ bool
}

var (
	edIndexedRe = regexp.MustCompile(`\{(\d)\}`)
	edSignedRe  = regexp.MustCompile(`\{(\d?):(\+?)d?\}`)
)

//pob:src/Export/statdesc.lua:6-108
func (e *exportDescriber) parseStatFile(name string, order int) (int, error) {
	file, err := loader.LoadBundledTranslation(e.src.ctx, e.src.version, "en", name, e.src.assetCache)
	if err != nil {
		return order, fmt.Errorf("%s: %w", name, err)
	}
	for _, inc := range file.Includes {
		m := includeRe.FindStringSubmatch(inc)
		if m == nil {
			return order, fmt.Errorf("%s: include %q is not a stat description file", name, inc)
		}
		if order, err = e.parseStatFile(m[1], order); err != nil {
			return order, err
		}
	}
	for _, stat := range file.NoDescription {
		e.descriptors[stat] = &edDescriptor{}
	}
	for _, d := range file.Descriptors {
		desc := &edDescriptor{stats: d.IDs, order: order}
		order++
		for _, line := range d.List {
			if line.Qualifier == "table_only" {
				continue
			}
			v := sdVariant{Text: escapeGGGString(line.String), Specs: sdSpecs(line.Handlers), GemQuality: strings.Contains(line.Qualifier, "gem_quality")}
			for _, c := range line.Conditions {
				v.Limit = append(v.Limit, sdLimit(c))
			}
			desc.variants = append(desc.variants, v)
		}
		for _, stat := range d.IDs {
			e.descriptors[stat] = desc
		}
	}
	return order, nil
}

//pob:src/Export/statdesc.lua:120-169
func (src source) loadStatFile(name string) (*exportDescriber, error) {
	e := &exportDescriber{descriptors: map[string]*edDescriptor{}, src: src}
	if _, err := e.parseStatFile(name, 1); err != nil {
		return nil, err
	}
	return e, nil
}

//pob:src/Export/statdesc.lua:110-118
func (e *exportDescriber) nextOrder() int {
	nextOrder := 1
	for _, d := range e.descriptors {
		if d.order >= nextOrder {
			nextOrder = d.order + 1
		}
	}
	return nextOrder
}

//pob:src/Export/statdesc.lua:120-160
func (src source) loadStatFiles(name string, appends ...string) (*exportDescriber, error) {
	e, err := src.loadStatFile(name)
	if err != nil {
		return nil, err
	}
	for _, fileName := range appends {
		startOrder := e.nextOrder()
		added := &exportDescriber{descriptors: map[string]*edDescriptor{}, src: src}
		if _, err := added.parseStatFile(fileName, startOrder); err != nil {
			return nil, err
		}
		normalised := map[*edDescriptor]bool{}
		for stat, d := range added.descriptors {
			if d.order > 0 && !normalised[d] {
				d.order = max(d.order-startOrder+1, 1)
				normalised[d] = true
			}
			e.descriptors[stat] = d
		}
	}
	return e, nil
}

func edLimitNumber(v any) (float64, bool) {
	n, ok := v.(int)
	return float64(n), ok
}

//pob:src/Export/statdesc.lua:175-193
func edMatchLimit(variants []sdVariant, val []*edRange) *sdVariant {
	for i := range variants {
		desc := &variants[i]
		match := true
		for j, limit := range desc.Limit {
			if limit[0] == "!" {
				if n, _ := edLimitNumber(limit[1]); val[j].min == n {
					match = false
					break
				}
				continue
			}
			hi, hasHi := edLimitNumber(limit[1])
			lo, hasLo := edLimitNumber(limit[0])
			if (hasHi && val[j].min > hi) || (hasLo && val[j].min < lo) {
				match = false
				break
			}
		}
		if match {
			return desc
		}
	}
	return nil
}

//pob:src/Modules/Common.lua:709-715
func luaRound(val float64, dec int) float64 {
	p := math.Pow(10, float64(dec))
	return math.Floor(val*p+0.5) / p
}

func edFormat(v float64, f string, sign bool) string {
	var s string
	switch f {
	case "g":
		s = strconv.FormatFloat(v, 'g', 6, 64)
	case ".2f":
		s = strconv.FormatFloat(v, 'f', 2, 64)
	default:
		s = strconv.FormatFloat(math.Trunc(v), 'f', 0, 64)
		if s == "-0" {
			s = "0"
		}
	}
	if sign && !strings.HasPrefix(s, "-") {
		s = "+" + s
	}
	return s
}

func edValue(v *edRange) string {
	if v.min == v.max {
		return edFormat(v.min, v.fmt, false)
	}
	return "(" + edFormat(v.min, v.fmt, false) + "-" + edFormat(v.max, v.fmt, false) + ")"
}

//pob:block:src/Export/statdesc.lua:252-388
//nolint:gocyclo // Mirrors the Lua elseif chain.
func edApplySpec(val []*edRange, spec sdSpec) error {
	idx, ok := spec.V.(float64)
	if !ok {
		switch spec.K {
		case "reminderstring", "canonical_line", "canonical_stat", "_stat":
			return nil
		}
		return fmt.Errorf("handler %s has no stat index", spec.K)
	}
	if int(idx) < 1 || int(idx) > len(val) {
		return fmt.Errorf("handler %s indexes stat %v of %d", spec.K, idx, len(val))
	}
	v := val[int(idx)-1]
	scale := func(f func(float64) float64, format string) {
		v.min, v.max = f(v.min), f(v.max)
		if format != "" {
			v.fmt = format
		}
	}
	switch spec.K {
	case "negate":
		v.max, v.min = -v.min, -v.max
	case "invert_chance":
		v.max, v.min = 100-v.min, 100-v.max
	case "negate_and_double":
		v.max, v.min = -2*v.min, -2*v.max
	case "passive_hash":
		if v.min < 0 {
			v.min += 65536
			v.max += 65536
		}
	case "divide_by_two_0dp":
		scale(func(x float64) float64 { return x / 2 }, "")
	case "divide_by_ten_0dp":
		scale(func(x float64) float64 { return x / 10 }, "")
	case "divide_by_fifteen_0dp":
		scale(func(x float64) float64 { return x / 15 }, "")
	case "divide_by_four":
		scale(func(x float64) float64 { return x / 4 }, "g")
	case "divide_by_five":
		scale(func(x float64) float64 { return x / 5 }, "g")
	case "divide_by_six":
		scale(func(x float64) float64 { return x / 6 }, "g")
	case "divide_by_ten_1dp_if_required", "divide_by_ten_1dp":
		scale(func(x float64) float64 { return luaRound(x/10, 1) }, "g")
	case "divide_by_twelve":
		scale(func(x float64) float64 { return x / 12 }, "g")
	case "divide_by_twenty":
		scale(func(x float64) float64 { return x / 20 }, "g")
	case "divide_by_one_hundred":
		scale(func(x float64) float64 { return x / 100 }, "g")
	case "divide_by_one_hundred_1dp":
		scale(func(x float64) float64 { return luaRound(x/100, 1) }, "g")
	case "divide_by_one_hundred_2dp", "divide_by_one_hundred_2dp_if_required":
		scale(func(x float64) float64 { return luaRound(x/100, 2) }, "g")
	case "divide_by_one_thousand":
		scale(func(x float64) float64 { return luaRound(x/1000, 1) }, "g")
	case "per_minute_to_per_second":
		scale(func(x float64) float64 { return luaRound(x/60, 1) }, "g")
	case "permyriad_per_minute_to_%_per_second":
		scale(func(x float64) float64 { return luaRound(x/60/100, 1) }, "g")
	case "per_minute_to_per_second_2dp_if_required", "per_minute_to_per_second_2dp":
		scale(func(x float64) float64 { return luaRound(x/60, 2) }, "g")
	case "per_minute_to_per_second_0dp":
		scale(func(x float64) float64 { return x / 60 }, "")
	case "per_minute_to_per_second_1dp":
		scale(func(x float64) float64 { return luaRound(x/60, 1) }, "g")
	case "milliseconds_to_seconds":
		scale(func(x float64) float64 { return x / 1000 }, "g")
	case "milliseconds_to_seconds_0dp":
		scale(func(x float64) float64 { return x / 1000 }, "")
	case "milliseconds_to_seconds_2dp_if_required", "milliseconds_to_seconds_2dp":
		scale(func(x float64) float64 { return luaRound(x/1000, 2) }, "g")
	case "deciseconds_to_seconds":
		scale(func(x float64) float64 { return x / 10 }, ".2f")
	case "locations_to_metres":
		scale(func(x float64) float64 { return x / 10 }, "g")
	case "60%_of_value":
		scale(func(x float64) float64 { return x * 0.6 }, "")
	case "30%_of_value":
		scale(func(x float64) float64 { return x * 0.3 }, "")
	case "mod_value_to_item_class":
		return fmt.Errorf("handler %s needs ItemClasses, which no ported export reads", spec.K)
	case "multiplicative_damage_modifier":
		scale(func(x float64) float64 { return 100 + x }, "")
	case "multiply_by_four":
		scale(func(x float64) float64 { return x * 4 }, "")
	case "times_one_point_five":
		scale(func(x float64) float64 { return x * 1.5 }, "")
	case "times_twenty":
		scale(func(x float64) float64 { return x * 20 }, "")
	case "double":
		scale(func(x float64) float64 { return x * 2 }, "")
	case "plus_two_hundred":
		scale(func(x float64) float64 { return x + 200 }, "")
	}
	return nil
}

func (e *exportDescriber) describeStats(stats map[string][2]float64) ([]string, []float64, error) {
	ranges := make(map[string]*edRange, len(stats))
	for s, v := range stats {
		ranges[s] = &edRange{min: v[0], max: v[1]}
	}
	return e.describeRanges(ranges)
}

//pob:src/Export/statdesc.lua:212-429
func (e *exportDescriber) describeRanges(stats map[string]*edRange) ([]string, []float64, error) {
	var descOrdered []*edDescriptor
	seen := map[*edDescriptor]bool{}
	for s, v := range stats {
		d := e.descriptors[s]
		if (v.min != 0 || v.max != 0) && d != nil && d.stats != nil && !seen[d] {
			seen[d] = true
			descOrdered = append(descOrdered, d)
		}
	}
	sort.Slice(descOrdered, func(i, j int) bool { return descOrdered[i].order < descOrdered[j].order })

	var out []string
	var orders []float64
	for _, descriptor := range descOrdered {
		val := make([]*edRange, len(descriptor.stats))
		for i, s := range descriptor.stats {
			val[i] = stats[s]
			if val[i] == nil {
				val[i] = &edRange{}
			}
			val[i].fmt = "d"
		}
		desc := edMatchLimit(descriptor.variants, val)
		// Hack to handle ranges starting or ending at 0 where no descriptor is defined for 0
		if desc == nil {
			for _, s := range val {
				if s.min == 0 && s.max > 0 {
					s.min, s.minZ = 1, true
				} else if s.min < 0 && s.max == 0 {
					s.max, s.maxZ = -1, true
				}
			}
			desc = edMatchLimit(descriptor.variants, val)
			for _, s := range val {
				if s.minZ {
					s.min = 0
				}
				if s.maxZ {
					s.max = 0
				}
			}
		}
		if desc == nil {
			continue
		}
		for _, spec := range desc.Specs {
			if err := edApplySpec(val, spec); err != nil {
				return nil, nil, fmt.Errorf("%s: %w", strings.Join(descriptor.stats, " "), err)
			}
		}
		at := func(n string) (*edRange, error) {
			i := 0
			if n != "" {
				i, _ = strconv.Atoi(n)
			}
			if i >= len(val) {
				return nil, fmt.Errorf("%s: %q reads value %d of %d", strings.Join(descriptor.stats, " "), desc.Text, i+1, len(val))
			}
			return val[i], nil
		}
		var fail error
		statDesc := edIndexedRe.ReplaceAllStringFunc(desc.Text, func(m string) string {
			v, err := at(edIndexedRe.FindStringSubmatch(m)[1])
			if err != nil {
				fail = err
				return m
			}
			return edValue(v)
		})
		statDesc = strings.ReplaceAll(statDesc, "{}", edValue(val[0]))
		statDesc = edSignedRe.ReplaceAllStringFunc(statDesc, func(m string) string {
			sub := edSignedRe.FindStringSubmatch(m)
			v, err := at(sub[1])
			if err != nil {
				fail = err
				return m
			}
			sign := sub[2] == "+"
			switch {
			case v.min == v.max:
				return edFormat(v.min, v.fmt, sign)
			case sign && v.max < 0:
				return "-(" + edFormat(-v.min, v.fmt, false) + "-" + edFormat(-v.max, v.fmt, false) + ")"
			case sign:
				return "+(" + edFormat(v.min, v.fmt, false) + "-" + edFormat(v.max, v.fmt, false) + ")"
			default:
				return "(" + edFormat(v.min, v.fmt, false) + "-" + edFormat(v.max, v.fmt, false) + ")"
			}
		})
		if fail != nil {
			return nil, nil, fail
		}
		statDesc = strings.ReplaceAll(statDesc, "%%", "%")
		order := float64(descriptor.order)
		for _, line := range strings.Split(statDesc, `\n`) {
			if line == "" || strings.Contains(line, `\`) {
				continue
			}
			out = append(out, line)
			orders = append(orders, order)
			order += 0.1
		}
	}
	return out, orders, nil
}
