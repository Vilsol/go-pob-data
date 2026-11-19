package derived

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/poe"
)

// MonsterLifeScaling is a monster variety's additional life scaling index, sorted by variety: the
// AdditionalLifeScaling IntID its templates' Stats set set_additional_life_scaling_index to, through their
// extends lines, from its BaseMonsterTypeIndex and every AddonMonsterTypeIndex. A variety none of whose
// templates sets it has no row, and so does one whose templates set it to different values. PoB has no
// counterpart.
type MonsterLifeScaling struct {
	// Variety is the MonsterVarieties Id.
	Variety string
	Index   int
}

func (MonsterLifeScaling) Table() string { return MonsterLifeScalingTable }

var (
	templateExtendsRe     = regexp.MustCompile(`^extends "([^"]+)"$`)
	templateLifeScalingRe = regexp.MustCompile(`^set_additional_life_scaling_index\s*=\s*(-?\d+)$`)
)

// templateLifeScaling is the values the top-level Stats section of an .ot template sets
// set_additional_life_scaling_index to, and the templates it extends. Braces inside strings and comments do
// not open or close sections; a mention of the stat anywhere else refuses the template.
func templateLifeScaling(text string) ([]int, []string, error) {
	var values []int
	var extends []string
	depth := 0
	section := ""
	previous := ""
	inString, inBlockComment := false, false
	for _, line := range strings.Split(text, "\n") {
		startDepth := depth
		var code strings.Builder
		for i := 0; i < len(line); i++ {
			c := line[i]
			switch {
			case inBlockComment:
				if c == '*' && i+1 < len(line) && line[i+1] == '/' {
					inBlockComment = false
					i++
				}
				continue
			case inString:
				if c == '"' {
					inString = false
				}
			case c == '"':
				inString = true
			case c == '/' && i+1 < len(line) && line[i+1] == '/':
				i = len(line)
				continue
			case c == '/' && i+1 < len(line) && line[i+1] == '*':
				inBlockComment = true
				i++
				continue
			case c == '{':
				if depth == 0 {
					section = strings.TrimSpace(code.String())
					if section == "" {
						section = previous
					}
				}
				depth++
			case c == '}':
				depth--
				if depth < 0 {
					return nil, nil, errors.New("unbalanced }")
				}
			}
			code.WriteByte(c)
		}
		trimmed := strings.TrimSpace(code.String())
		m := templateExtendsRe.FindStringSubmatch(trimmed)
		switch {
		case m != nil && startDepth == 0 && depth == 0:
			extends = append(extends, m[1])
		case strings.Contains(trimmed, "additional_life_scaling_index"):
			m = templateLifeScalingRe.FindStringSubmatch(trimmed)
			if m == nil || startDepth != 1 || depth != 1 || section != "Stats" {
				return nil, nil, fmt.Errorf("unrecognised line %q in section %s", trimmed, section)
			}
			value, err := strconv.Atoi(m[1])
			if err != nil {
				return nil, nil, err //nolint:wrapcheck
			}
			values = append(values, value)
		}
		if trimmed != "" && depth == 0 {
			previous = trimmed
		}
	}
	if depth != 0 || inString || inBlockComment {
		return nil, nil, errors.New("unclosed section, string or comment")
	}
	return values, extends, nil
}

// buildMonsterLifeScaling is every monster variety's additional life scaling index of src.version.
func buildMonsterLifeScaling(src source) ([]MonsterLifeScaling, error) {
	if err := poe.InitializeMonsterVarieties(src.ctx, src.version, src.assetCache); err != nil {
		return nil, err
	}
	templates, err := loader.LoadFileBundle(src.ctx, src.version, ".ot", src.assetCache)
	if err != nil {
		return nil, err
	}

	type parsed struct {
		values  []int
		extends []string
	}
	cache := map[string]parsed{}
	var walk func(name string, seen map[string]bool, values map[int]bool) error
	walk = func(name string, seen map[string]bool, values map[int]bool) error {
		if name == "" || name == "nothing" || seen[name] {
			return nil
		}
		seen[name] = true
		p, ok := cache[name]
		if !ok {
			text, ok := templates[name+".ot"]
			if !ok {
				return fmt.Errorf("%s.ot is not in the .ot bundle", name)
			}
			values, extends, err := templateLifeScaling(string(text))
			if err != nil {
				return fmt.Errorf("%s.ot: %w", name, err)
			}
			p = parsed{values: values, extends: extends}
			cache[name] = p
		}
		for _, v := range p.values {
			values[v] = true
		}
		for _, parent := range p.extends {
			if err := walk(parent, seen, values); err != nil {
				return err
			}
		}
		return nil
	}

	var rows []MonsterLifeScaling
	for _, variety := range poe.MonsterVarieties {
		values := map[int]bool{}
		for _, root := range append([]string{variety.BaseMonsterTypeIndex}, variety.AddonMonsterTypeIndex...) {
			if err := walk(root, map[string]bool{}, values); err != nil {
				return nil, fmt.Errorf("%s: %w", variety.ID, err)
			}
		}
		if len(values) != 1 {
			continue
		}
		for index := range values {
			rows = append(rows, MonsterLifeScaling{Variety: variety.ID, Index: index})
		}
	}
	slices.SortFunc(rows, func(a, b MonsterLifeScaling) int { return strings.Compare(a.Variety, b.Variety) })
	return rows, nil
}
