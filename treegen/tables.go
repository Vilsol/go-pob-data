package treegen

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
)

// Row types list only the columns the generator reads; loadTable fails if any is missing.

//nolint:govet // fields follow the export key order
type passiveSkill struct {
	ID                       string `json:"Id"`
	Icon                     string `json:"Icon_DDSFile"`
	Stats                    []int  `json:"Stats"`
	Stat1Value               int    `json:"Stat1Value"`
	Stat2Value               int    `json:"Stat2Value"`
	Stat3Value               int    `json:"Stat3Value"`
	Stat4Value               int    `json:"Stat4Value"`
	Stat5Value               int    `json:"Stat5Value"`
	PassiveSkillGraphID      uint32 `json:"PassiveSkillGraphId"`
	Name                     string `json:"Name"`
	Characters               []int  `json:"Characters"`
	IsKeystone               bool   `json:"IsKeystone"`
	IsNotable                bool   `json:"IsNotable"`
	FlavourText              string `json:"FlavourText"`
	IsJewelSocket            bool   `json:"IsJewelSocket"`
	AscendancyKey            *int   `json:"AscendancyKey"`
	IsAscendancyStartingNode bool   `json:"IsAscendancyStartingNode"`
	ReminderStrings          []int  `json:"ReminderStrings"`
	SkillPointsGranted       int    `json:"SkillPointsGranted"`
	IsMultipleChoice         bool   `json:"IsMultipleChoice"`
	IsMultipleChoiceOption   bool   `json:"IsMultipleChoiceOption"`
	IsAnointmentOnly         bool   `json:"IsAnointmentOnly"`
	IsProxyPassive           bool   `json:"IsProxyPassive"`
	MasteryGroup             *int   `json:"MasteryGroup"`
	DescendancyKey           *int   `json:"DescendancyKey"`
	PassiveSkillBuffs        []int  `json:"PassiveSkillBuffs"`
}

func (p *passiveSkill) statValues() []int {
	return []int{p.Stat1Value, p.Stat2Value, p.Stat3Value, p.Stat4Value, p.Stat5Value}
}

type stat struct {
	ID string `json:"Id"`
}

//nolint:govet // fields follow the export key order
type character struct {
	Name             string `json:"Name"`
	IntegerID        int    `json:"IntegerId"`
	BaseStrength     int    `json:"BaseStrength"`
	BaseDexterity    int    `json:"BaseDexterity"`
	BaseIntelligence int    `json:"BaseIntelligence"`
	AttrsAsID        string `json:"AttrsAsId"`
	PassiveTreeImage string `json:"PassiveTreeImage"`
	// ArtOffsetX/Y place the PassiveTreeImage relative to the class start group (unnamed in the schema).
	ArtOffsetX          int    `json:"Var33"`
	ArtOffsetY          int    `json:"Var34"`
	SkillTreeBackground string `json:"SkillTreeBackground"`
}

//nolint:govet // fields follow the export key order
type ascendancy struct {
	ID                   string `json:"Id"`
	ClassNo              int    `json:"ClassNo"`
	Character            *int   `json:"Character"`
	CoordinateRect       string `json:"CoordinateRect"`
	RGBFlavourTextColour string `json:"RGBFlavourTextColour"`
	Name                 string `json:"Name"`
	FlavourText          string `json:"FlavourText"`
	BackgroundImage      string `json:"BackgroundImage"`
	LeagueBackground     string `json:"LeagueBackgroundImage"`
}

type descendancy struct {
	PassiveTreeUIArt     *int   `json:"PassiveTreeUIArt"`
	UIArt                *int   `json:"UIArt"`
	ID                   string `json:"Id"`
	Name                 string `json:"Name"`
	FlavourText          string `json:"FlavourText"`
	CoordinateRect       string `json:"CoordinateRect"`
	RGBFlavourTextColour string `json:"RGBFlavourTextColour"`
	BackgroundImage      string `json:"BackgroundImage"`
}

type masteryGroup struct {
	InactiveIcon      string `json:"InactiveIcon"`
	ActiveIcon        string `json:"ActiveIcon"`
	ActiveEffectImage string `json:"ActiveEffectImage"`
	MasteryEffects    []int  `json:"MasteryEffects"`
}

// masteryGroupWithArt is PassiveSkillMasteryGroups from 3.30, whose art is a PassiveSkillTreeMasteryArt row,
// as in PoE2.
type masteryGroupWithArt struct {
	Art            *int  `json:"Art"`
	MasteryEffects []int `json:"MasteryEffects"`
}

type masteryArt struct {
	InactiveIcon      string `json:"InactiveIcon"`
	ActiveIcon        string `json:"ActiveIcon"`
	ActiveEffectImage string `json:"ActiveEffectImage"`
}

type masteryEffect struct {
	Stats      []int  `json:"Stats"`
	Hash16     uint16 `json:"HASH16"`
	Stat1Value int    `json:"Stat1Value"`
	Stat2Value int    `json:"Stat2Value"`
	Stat3Value int    `json:"Stat3Value"`
}

func (m *masteryEffect) statValues() []int {
	return []int{m.Stat1Value, m.Stat2Value, m.Stat3Value}
}

type jewelSlot struct {
	Slot             *int `json:"Slot"`
	ClusterJewelSize *int `json:"ClusterJewelSize"`
	ReplacesSlot     *int `json:"ReplacesSlot"`
	ProxySlot        *int `json:"ProxySlot"`
	ClusterIndex     int  `json:"ClusterIndex"`
}

type reminderText struct {
	ID   string `json:"Id"`
	Text string `json:"Text"`
}

type buffTemplate struct {
	BuffDefinition *int  `json:"BuffDefinitionsKey"`
	StatValues     []int `json:"Buff_StatValues"`
}

type buffDefinition struct {
	Stats        []int `json:"StatsKeys"`
	GrantedFlags []int `json:"GrantedFlags"`
}

type clientString struct {
	ID   string `json:"Id"`
	Text string `json:"Text"`
}

type blightRecipe struct {
	Result *int  `json:"BlightCraftingResultsKey"`
	Items  []int `json:"BlightCraftingItemsKeys"`
}

type blightResult struct {
	PassiveSkill *int `json:"PassiveSkill"`
}

type blightItem struct {
	Oil *int `json:"Oil"`
}

type baseItemType struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
}

type expansionSkill struct {
	Mastery      *int `json:"Mastery_PassiveSkillsKey"`
	PassiveSkill int  `json:"PassiveSkillsKey"`
}

type expansionSpecialSkill struct {
	PassiveSkill int `json:"PassiveSkillsKey"`
}

type passiveSkillOverride struct {
	AllocatedPassiveSkill *int   `json:"AllocatedPassiveSkill"`
	Type                  *int   `json:"Type"`
	NodeIcon              string `json:"NodeIcon"`
	PassiveBG             string `json:"PassiveBG"`
}

type passiveSkillOverrideType struct {
	ID string `json:"Id"`
}

type nodeFrameArt struct {
	ID          string `json:"Id"`
	Normal      string `json:"Normal"`
	CanAllocate string `json:"CanAllocate"`
	Active      string `json:"Active"`
}

type jewelArt struct {
	SocketedImage      string `json:"SocketedImage"`
	SocketedImageLarge string `json:"SocketedImageLarge"`
	BaseItemType       int    `json:"BaseItemTypesKey"`
}

// ascendancyTreeArt is a PassiveSkillTreeUIArtAscendancy row: the frames of an ascendancy's nodes.
type ascendancyTreeArt struct {
	ID                      string `json:"Id"`
	StartNode               string `json:"StartNode"`
	PassiveFrameNormal      string `json:"PassiveFrameNormal"`
	PassiveFrameCanAllocate string `json:"PassiveFrameCanAllocate"`
	PassiveFrameActive      string `json:"PassiveFrameActive"`
	NotableFrameNormal      string `json:"NotableFrameNormal"`
	NotableFrameCanAllocate string `json:"NotableFrameCanAllocate"`
	NotableFrameActive      string `json:"NotableFrameActive"`
	SocketFrameNormal       string `json:"SocketFrameNormal"`
	SocketFrameCanAllocate  string `json:"SocketFrameCanAllocate"`
	SocketFrameActive       string `json:"SocketFrameActive"`
}

// ascendancyUIArt is a UIArtAscendancy row: the ascendancy's buttons.
type ascendancyUIArt struct {
	ID              string `json:"Id"`
	ButtonNormal    string `json:"AscendancyButtonNormal"`
	ButtonHighlight string `json:"AscendancyButtonHighlight"`
	ButtonPressed   string `json:"AscendancyButtonPressed"`
}

type passiveSkillTree struct {
	ID                string `json:"Id"`
	PassiveSkillGraph string `json:"PassiveSkillGraph"`
}

func gamePath(gameVersion string, path string) string {
	return gameFilePath(extractor.PoE1, gameVersion, path)
}

func gameFilePath(game extractor.Game, gameVersion string, path string) string {
	if extractor.IsNewHash(game, gameVersion) {
		return strings.ToLower(path)
	}
	return path
}

type tableLoader struct {
	loader      fs.FS
	err         error
	game        extractor.Game
	gameVersion string
}

var errNoColumn = errors.New("missing column")

// loadMasteryGroups reads each mastery group's art from its own columns, or from 3.30 on from the row its Art
// column names, in the table the schema says it references.
func loadMasteryGroups(l *tableLoader, dst *[]masteryGroup) {
	if l.err != nil {
		return
	}
	groups, err := decodeTable[masteryGroupWithArt](l.loader, l.game, l.gameVersion, "PassiveSkillMasteryGroups")
	if errors.Is(err, errNoColumn) {
		loadTable(l, "PassiveSkillMasteryGroups", dst)
		return
	}
	if err != nil {
		l.err = err
		return
	}
	artTable := "PassiveSkillTreeMasteryArt"
	for _, c := range extractor.SchemaColumns("PassiveSkillMasteryGroups") {
		if c.Name == "Art" && c.Ref != "" {
			artTable = c.Ref
		}
	}
	var art []masteryArt
	if loadTable(l, artTable, &art); l.err != nil {
		return
	}
	*dst = make([]masteryGroup, len(groups))
	for i, g := range groups {
		if g.Art == nil || *g.Art >= len(art) {
			l.err = fmt.Errorf("mastery group %d has no art row", i)
			return
		}
		a := art[*g.Art]
		(*dst)[i] = masteryGroup{
			InactiveIcon:      a.InactiveIcon,
			ActiveIcon:        a.ActiveIcon,
			ActiveEffectImage: a.ActiveEffectImage,
			MasteryEffects:    g.MasteryEffects,
		}
	}
}

// loadTable decodes table name into dst; after the first failure it does nothing and l.err holds the error.
func loadTable[T any](l *tableLoader, name string, dst *[]T) {
	if l.err == nil {
		*dst, l.err = decodeTable[T](l.loader, l.game, l.gameVersion, name)
	}
}

func decodeTable[T any](loader fs.FS, game extractor.Game, gameVersion string, name string) ([]T, error) {
	f, err := loader.Open(gameFilePath(game, gameVersion, game.TableDir()+name+"."+extractor.GetFileExtension(game, gameVersion)))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", name)
	}

	rows, err := extractor.ParseDat(f, name+"."+extractor.GetFileExtension(game, gameVersion))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", name)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("table %s is empty", name)
	}

	b, err := json.Marshal(rows)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal %s", name)
	}

	var columns []map[string]json.RawMessage
	if err := json.Unmarshal(b, &columns); err != nil {
		return nil, errors.Wrapf(err, "failed to decode %s columns", name)
	}
	rt := reflect.TypeFor[T]()
	for i := range rt.NumField() {
		column := strings.Split(rt.Field(i).Tag.Get("json"), ",")[0]
		if _, ok := columns[0][column]; !ok {
			return nil, fmt.Errorf("table %s has no column %s: %w", name, column, errNoColumn)
		}
	}

	var out []T
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, errors.Wrapf(err, "failed to decode %s", name)
	}
	return out, nil
}
