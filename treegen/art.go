package treegen

import (
	"fmt"
	"image"
	"io/fs"
	"math"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// zoomLevel is the scale the sprites are rendered at: the largest of the export's imageZoomLevels.
const zoomLevel = 0.3835

// jewelRadiusZoom: the export keeps the jewel radius circles at their native size.
const jewelRadiusZoom = 1

// assetsURL is where data.pob.gg serves the sheets; consumers only use the file's name.
const assetsURL = "https://data.pob.gg/data/%s/tree/assets/"

// altAscendancyTreeID is the PassiveSkillTrees row of the custom ascendancy passives the export has icons for.
const altAscendancyTreeID = "DefaultAltAscendancies"

// alternateMasteryType is the PassiveSkillOverrideTypes row whose overrides replace masteries (rune grafts).
const alternateMasteryType = "AlternateMastery"

const inGameArt = "Art/2DArt/UIImages/InGame/"

// Node icon sizes on the tree are game-client values absent from the game files; these give the export's sprite sizes.
var (
	normalIconSize   = [2]float64{69, 69}
	notableIconSize  = [2]float64{99, 99}
	keystoneIconSize = [2]float64{137, 139}
	masteryIconSize  = [2]float64{226, 226}
)

// Inactive icons aren't in the game files; the export darkens and desaturates the active ones, and
// these rules reproduce its sheets (measured against skills-disabled and mastery-disabled of 3.29).
// Both first pull each channel toward the HSL lightness (max+min)/2.
//
// inactiveSkill keeps 48% of the saturation, then compresses each channel with a*v/(1+k*v).
func inactiveSkill(r, g, b float64) (float64, float64, float64) {
	const s, a, k = 0.48, 0.928, 0.5
	curve := func(v float64) float64 { return a * v / (1 + k*v) }
	r, g, b = desaturate(r, g, b, s)
	return curve(r), curve(g), curve(b)
}

// inactiveMastery keeps 40% of the saturation and scales by 1/sqrt(3).
func inactiveMastery(r, g, b float64) (float64, float64, float64) {
	r, g, b = desaturate(r, g, b, 0.4)
	return r / math.Sqrt(3), g / math.Sqrt(3), b / math.Sqrt(3)
}

func desaturate(r, g, b, s float64) (float64, float64, float64) {
	l := (max(r, g, b) + min(r, g, b)) / 2
	return l + s*(r-l), l + s*(g-l), l + s*(b-l)
}

// Named UI images the export uses, keyed by the export's name, that no table references.
var (
	groupBackgrounds = [][2]string{
		{"PSGroupBackground1", "PassiveSkillScreenGroupBackgroundSmall"},
		{"PSGroupBackground2", "PassiveSkillScreenGroupBackgroundMedium"},
		{"PSGroupBackground3", "PassiveSkillScreenGroupBackgroundLargeHalf"},
		{"GroupBackgroundSmallAlt", "PassiveSkillScreenGroupBackgroundSmallAlt"},
		{"GroupBackgroundMediumAlt", "PassiveSkillScreenGroupBackgroundMediumAlt"},
		{"GroupBackgroundLargeHalfAlt", "PassiveSkillScreenGroupBackgroundLargeHalfAlt"},
	}
	extraFrames = [][2]string{
		{"JewelFrameUnallocated", "PassiveSkillScreenJewelSocketNormal"},
		{"JewelFrameCanAllocate", "PassiveSkillScreenJewelSocketCanAllocate"},
		{"JewelFrameAllocated", "PassiveSkillScreenJewelSocketActive"},
		{"BlightedNotableFrameUnallocated", "PassiveSkillScreenBlightedNotableFrameNormal"},
		{"BlightedNotableFrameCanAllocate", "PassiveSkillScreenBlightedNotableFrameCanAllocate"},
		{"BlightedNotableFrameAllocated", "PassiveSkillScreenBlightedNotableFrameActive"},
		{"JewelSocketAltNormal", "PassiveSkillScreenJewelSocketAltNormal"},
		{"JewelSocketAltCanAllocate", "PassiveSkillScreenJewelSocketAltCanAllocate"},
		{"JewelSocketAltActive", "PassiveSkillScreenJewelSocketAltActive"},
		{"JewelSocketClusterAltNormal1Small", "JewelSocketClusterAltNormal1Small"},
		{"JewelSocketClusterAltNormal1Medium", "JewelSocketClusterAltNormal1Medium"},
		{"JewelSocketClusterAltNormal1Large", "JewelSocketClusterAltNormal1Large"},
		{"JewelSocketClusterAltCanAllocate1Small", "JewelSocketClusterAltCanAllocate1Small"},
		{"JewelSocketClusterAltCanAllocate1Medium", "JewelSocketClusterAltCanAllocate1Medium"},
		{"JewelSocketClusterAltCanAllocate1Large", "JewelSocketClusterAltCanAllocate1Large"},
	}
	jewelRadii = [][2]string{
		{"JewelCircle1", "PassiveSkillScreenJewelCircle1"},
		{"JewelCircle1Inverse", "PassiveSkillScreenJewelCircle1inverse"},
		{"VaalJewelCircle1", "PassiveSkillScreenVaalJewelCircle1"},
		{"VaalJewelCircle2", "PassiveSkillScreenVaalJewelCircle2"},
		{"KaruiJewelCircle1", "PassiveSkillScreenKaruiJewelCircle1"},
		{"KaruiJewelCircle2", "PassiveSkillScreenKaruiJewelCircle2"},
		{"MarakethJewelCircle1", "PassiveSkillScreenMarakethJewelCircle1"},
		{"MarakethJewelCircle2", "PassiveSkillScreenMarakethJewelCircle2"},
		{"TemplarJewelCircle1", "PassiveSkillScreenTemplarJewelCircle1"},
		{"TemplarJewelCircle2", "PassiveSkillScreenTemplarJewelCircle2"},
		{"EternalEmpireJewelCircle1", "PassiveSkillScreenEternalEmpireJewelCircle1"},
		{"EternalEmpireJewelCircle2", "PassiveSkillScreenEternalEmpireJewelCircle2"},
		{"KalguurJewelCircle1", "PassiveSkillScreenKalguuranJewelCircle1"},
		{"KalguurJewelCircle2", "PassiveSkillScreenKalguuranJewelCircle2"},
	}
	lineDecorations = [][2]string{
		{"PSLineDeco", "PassiveSkillScreenOrnament1"},
		{"PSLineDecoHighlighted", "PassiveSkillScreenOrnament1"},
	}
)

const (
	backgroundArt          = "Art/2DArt/UIImages/Common/Background2"
	startNodeInactiveArt   = inGameArt + "PassiveSkillScreenStartNodeBackgroundInactive"
	masteryButtonArt       = inGameArt + "PassiveMastery/PassiveMasteryButtonInactiveDefault"
	connectionCurvesFormat = "Art/2DArt/PassiveTree/PassiveSkillScreenCurves%sTogether.dds"
)

// nodeFrames maps the export's frame names to PassiveSkillTreeNodeFrameArt rows.
var nodeFrames = []struct {
	row                         string
	normal, canAllocate, active string
}{
	{"CharacterSmall", "PSSkillFrame", "PSSkillFrameHighlighted", "PSSkillFrameActive"},
	{"CharacterNotable", "NotableFrameUnallocated", "NotableFrameCanAllocate", "NotableFrameAllocated"},
	{"CharacterKeystone", "KeystoneFrameUnallocated", "KeystoneFrameCanAllocate", "KeystoneFrameAllocated"},
}

type artData struct {
	altAscendancies map[uint32]bool
	nodeFrames      []nodeFrameArt
	jewelArt        []jewelArt
	treeArt         []ascendancyTreeArt
	uiArt           []ascendancyUIArt
	overrideTypes   []passiveSkillOverrideType
}

// spriteRef is one entry of the sprites object: a group of names on a sheet at a zoom level.
type spriteRef struct {
	sheet *sheet
	key   string
	group string
	zoom  float64
}

type artBuilder struct {
	r       *renderer
	d       *gameData
	a       *artData
	tree    *Tree
	names   map[[2]string]bool
	sheets  []*sheet
	sprites []spriteRef
}

func (b *artBuilder) newSheet(file string) *sheet {
	s := &sheet{file: file}
	b.sheets = append(b.sheets, s)
	return s
}

// sprite registers group on s as the sprites entry key.
func (b *artBuilder) sprite(key string, s *sheet, group string) {
	b.sprites = append(b.sprites, spriteRef{key: key, sheet: s, group: group, zoom: zoomLevel})
}

// add puts a name on a sheet once per group.
func (b *artBuilder) add(s *sheet, group, name string, a *art) {
	k := [2]string{group, name}
	if b.names[k] {
		return
	}
	b.names[k] = true
	s.add(group, name, a)
}

func (b *artBuilder) addScaled(s *sheet, group, name string, src source) error {
	a, err := b.r.renderScaled(src)
	if err != nil {
		return err
	}
	b.add(s, group, name, a)
	return nil
}

func (b *artBuilder) addSized(s *sheet, group, name string, src source, size [2]float64) error {
	a, err := b.r.render(src, scaled(size[0]), scaled(size[1]))
	if err != nil {
		return err
	}
	b.add(s, group, name, a)
	return nil
}

func loadArtData(loader fs.FS, gameVersion string, d *gameData) (*artData, error) {
	a := &artData{}
	l := &tableLoader{loader: loader, gameVersion: gameVersion}
	loadTable(l, "PassiveSkillTreeNodeFrameArt", &a.nodeFrames)
	loadTable(l, "PassiveJewelArt", &a.jewelArt)
	loadTable(l, "PassiveSkillTreeUIArtAscendancy", &a.treeArt)
	loadTable(l, "UIArtAscendancy", &a.uiArt)
	loadTable(l, "PassiveSkillOverrideTypes", &a.overrideTypes)
	if l.err != nil {
		return nil, l.err
	}

	graph, err := loadGraph(loader, gameVersion, d, altAscendancyTreeID)
	if err != nil {
		return nil, err
	}
	a.altAscendancies = make(map[uint32]bool)
	for _, g := range graph.Groups {
		for _, n := range g.Nodes {
			a.altAscendancies[n.ID] = true
		}
	}
	return a, nil
}

// buildArt renders the sprite sheets into dir and sets the tree's sprites, zoom levels and class images.
func buildArt(tree *Tree, d *gameData, loader fs.FS, gameVersion string, dir string) error {
	a, err := loadArtData(loader, gameVersion, d)
	if err != nil {
		return err
	}
	r, err := newRenderer(loader, gameVersion)
	if err != nil {
		return err
	}
	defer r.close()

	b := &artBuilder{r: r, d: d, a: a, tree: tree, names: make(map[[2]string]bool)}
	for _, step := range []func() error{
		b.background, b.nodeIcons, b.masteries, b.tattoos, b.ascendancies, b.bloodlines,
		b.groupBackgrounds, b.frames, b.jewels, b.lines, b.jewelRadii, b.extraImages,
	} {
		if err := step(); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrap(err, "failed to create assets directory")
	}
	for _, s := range b.sheets {
		if s.image == nil {
			s.pack()
		}
		if err := s.write(dir); err != nil {
			return err
		}
	}

	base := fmt.Sprintf(assetsURL, gameVersion)
	for _, ref := range b.sprites {
		sprite := &Sprite{Filename: base + ref.sheet.file, W: ref.sheet.image.Rect.Dx(), H: ref.sheet.image.Rect.Dy()}
		for _, e := range ref.sheet.entries {
			if e.group == ref.group {
				sprite.Coords = append(sprite.Coords, Entry[Coord]{Key: e.name, Value: ref.sheet.coord(e)})
			}
		}
		if len(sprite.Coords) == 0 {
			return fmt.Errorf("sprite group %s is empty", ref.key)
		}
		zoom := strconv.FormatFloat(ref.zoom, 'f', -1, 64)
		tree.Sprites = append(tree.Sprites, Entry[Ordered[*Sprite]]{Key: ref.key, Value: Ordered[*Sprite]{{Key: zoom, Value: sprite}}})
	}
	tree.ImageZoomLevels = []float64{zoomLevel}
	return nil
}

func (b *artBuilder) background() error {
	s := b.newSheet("background-3.png")
	b.sprite("background", s, "background")
	return b.addScaled(s, "background", "Background2", source{atlas: backgroundArt})
}

func (b *artBuilder) isAlternateMastery(o passiveSkillOverride) bool {
	return o.Type != nil && b.a.overrideTypes[*o.Type].ID == alternateMasteryType
}

// nodeIcons adds the icons of the tree's nodes, of the custom ascendancy passives and of the
// passive overrides (tattoos): the node icons the export's sheets have.
func (b *artBuilder) nodeIcons() error {
	skills := b.newSheet("skills-3.jpg")
	mastery := b.newSheet("mastery-3.png")

	// either is for icons that can be drawn at normal or notable size: the export has notable
	// sprites of some multiple choice options and tattoos, so they get both.
	add := func(name string, keystone, notable, isMastery, either bool) error {
		src := source{file: strings.TrimSuffix(name, ".png") + ".dds"}
		switch {
		case keystone:
			return b.addSized(skills, "keystoneActive", name, src, keystoneIconSize)
		case notable:
			return b.addSized(skills, "notableActive", name, src, notableIconSize)
		case isMastery:
			return b.addSized(mastery, "mastery", name, src, masteryIconSize)
		case either:
			if err := b.addSized(skills, "notableActive", name, src, notableIconSize); err != nil {
				return err
			}
		}
		return b.addSized(skills, "normalActive", name, src, normalIconSize)
	}

	for _, n := range b.tree.Nodes.List {
		if n.Icon == "" {
			continue
		}
		if err := add(n.Icon, n.IsKeystone, n.IsNotable, n.IsMastery, n.IsMultipleChoiceOption); err != nil {
			return errors.Wrapf(err, "icon of node %d", n.Skill)
		}
	}
	for _, p := range b.d.passiveSkills {
		if !b.a.altAscendancies[p.PassiveSkillGraphID] || p.Icon == "" {
			continue
		}
		if err := add(iconPath(p.Icon), p.IsKeystone, p.IsNotable, p.MasteryGroup != nil, p.IsMultipleChoiceOption); err != nil {
			return errors.Wrapf(err, "icon of passive %s", p.ID)
		}
	}
	for _, o := range b.d.overrides {
		if o.NodeIcon == "" {
			continue
		}
		if err := add(iconPath(o.NodeIcon), false, false, b.isAlternateMastery(o), !b.isAlternateMastery(o)); err != nil {
			return errors.Wrap(err, "icon of passive override")
		}
	}

	skills.pack()
	disabled := skills.derive("skills-disabled-3.jpg", inactiveSkill)
	b.sheets = append(b.sheets, disabled)
	b.sprite("normalActive", skills, "normalActive")
	b.sprite("notableActive", skills, "notableActive")
	b.sprite("keystoneActive", skills, "keystoneActive")
	b.sprite("normalInactive", disabled, "normalActive")
	b.sprite("notableInactive", disabled, "notableActive")
	b.sprite("keystoneInactive", disabled, "keystoneActive")
	b.sprite("mastery", mastery, "mastery")
	return nil
}

// masteries adds every mastery group's inactive, active and effect art, and the rune graft
// overrides that replace masteries.
func (b *artBuilder) masteries() error {
	connected := b.newSheet("mastery-connected-3.png")
	selected := b.newSheet("mastery-active-selected-3.png")
	effect := b.newSheet("mastery-active-effect-3.png")

	if err := b.addSized(selected, "masteryActiveSelected", "PassiveMasteryConnectedButton", source{atlas: masteryButtonArt}, masteryIconSize); err != nil {
		return err
	}
	for _, mg := range b.d.masteryGroups {
		if err := b.addSized(connected, "masteryConnected", iconPath(mg.InactiveIcon), source{file: mg.InactiveIcon}, masteryIconSize); err != nil {
			return err
		}
		if err := b.addSized(selected, "masteryActiveSelected", iconPath(mg.ActiveIcon), source{file: mg.ActiveIcon}, masteryIconSize); err != nil {
			return err
		}
		if err := b.addScaled(effect, "masteryActiveEffect", iconPath(mg.ActiveEffectImage), source{atlas: mg.ActiveEffectImage}); err != nil {
			return err
		}
	}
	for _, o := range b.d.overrides {
		if !b.isAlternateMastery(o) {
			continue
		}
		if o.NodeIcon != "" {
			src := source{file: o.NodeIcon}
			if err := b.addSized(connected, "masteryConnected", iconPath(o.NodeIcon), src, masteryIconSize); err != nil {
				return err
			}
			if err := b.addSized(selected, "masteryActiveSelected", iconPath(o.NodeIcon), src, masteryIconSize); err != nil {
				return err
			}
		}
		if o.PassiveBG != "" {
			if err := b.addScaled(effect, "masteryActiveEffect", iconPath(o.PassiveBG), source{atlas: o.PassiveBG}); err != nil {
				return err
			}
		}
	}

	connected.pack()
	disabled := connected.derive("mastery-disabled-3.png", inactiveMastery)
	b.sheets = append(b.sheets, disabled)
	b.sprite("masteryConnected", connected, "masteryConnected")
	b.sprite("masteryActiveSelected", selected, "masteryActiveSelected")
	b.sprite("masteryInactive", disabled, "masteryConnected")
	b.sprite("masteryActiveEffect", effect, "masteryActiveEffect")
	return nil
}

// tattoos adds the backgrounds the passive overrides draw behind their node.
func (b *artBuilder) tattoos() error {
	s := b.newSheet("tattoo-active-effect-3.png")
	b.sprite("tattooActiveEffect", s, "tattooActiveEffect")
	for _, o := range b.d.overrides {
		if o.PassiveBG == "" {
			continue
		}
		if err := b.addScaled(s, "tattooActiveEffect", iconPath(o.PassiveBG), source{atlas: o.PassiveBG}); err != nil {
			return err
		}
	}
	return nil
}

func (b *artBuilder) treeArt(id string) (*ascendancyTreeArt, *ascendancyUIArt, error) {
	i := slices.IndexFunc(b.a.treeArt, func(a ascendancyTreeArt) bool { return a.ID == id })
	j := slices.IndexFunc(b.a.uiArt, func(a ascendancyUIArt) bool { return a.ID == id })
	if i < 0 || j < 0 {
		return nil, nil, fmt.Errorf("no ascendancy art %q", id)
	}
	return &b.a.treeArt[i], &b.a.uiArt[j], nil
}

// ascendancyFrames adds an ascendancy's buttons, node frames and charm sockets.
func (b *artBuilder) ascendancyFrames(s *sheet, group string, t *ascendancyTreeArt, u *ascendancyUIArt) error {
	art := [][2]string{
		{"AscendancyButton", u.ButtonNormal},
		{"AscendancyButtonHighlight", u.ButtonHighlight},
		{"AscendancyButtonPressed", u.ButtonPressed},
		{"AscendancyFrameLargeNormal", t.NotableFrameNormal},
		{"AscendancyFrameLargeCanAllocate", t.NotableFrameCanAllocate},
		{"AscendancyFrameLargeAllocated", t.NotableFrameActive},
		{"AscendancyFrameSmallNormal", t.PassiveFrameNormal},
		{"AscendancyFrameSmallCanAllocate", t.PassiveFrameCanAllocate},
		{"AscendancyFrameSmallAllocated", t.PassiveFrameActive},
		{"AscendancyMiddle", t.StartNode},
	}
	if t.SocketFrameNormal != "" {
		art = append(art,
			[2]string{"CharmFrameNormal", t.SocketFrameNormal},
			[2]string{"CharmFrameCanAllocate", t.SocketFrameCanAllocate},
			[2]string{"CharmFrameAllocated", t.SocketFrameActive},
		)
		// Charms are the jewel art whose base item is an <Attribute>AnimalCharm.
		for _, j := range b.a.jewelArt {
			if attr, ok := strings.CutSuffix(path.Base(b.d.baseItemTypes[j.BaseItemType].ID), "AnimalCharm"); ok {
				art = append(art, [2]string{"CharmSocketActive" + attr, j.SocketedImage})
			}
		}
	}
	for _, a := range art {
		if a[1] == "" {
			return fmt.Errorf("ascendancy art %s has no %s", t.ID, a[0])
		}
		if err := b.addScaled(s, group, a[0], source{atlas: a[1]}); err != nil {
			return err
		}
	}
	return nil
}

// ascendancies adds the ascendancy backgrounds, and the default ascendancy buttons and frames.
func (b *artBuilder) ascendancies() error {
	s := b.newSheet("ascendancy-3.webp")
	b.sprite("ascendancy", s, "ascendancy")

	own := make(map[string]bool)
	for _, a := range b.d.ascendancies {
		if err := b.addScaled(s, "ascendancy", "Classes"+a.ID, source{atlas: a.BackgroundImage}); err != nil {
			return err
		}
		own[a.BackgroundImage] = true
	}
	// The export names the league's custom ascendancy background, the one no ascendancy has as its own, ClassesAlternate.
	var league []string
	for _, a := range b.d.ascendancies {
		if !own[a.LeagueBackground] && !slices.Contains(league, a.LeagueBackground) {
			league = append(league, a.LeagueBackground)
		}
	}
	if len(league) != 1 {
		return fmt.Errorf("want one league ascendancy background, got %q", league)
	}
	if err := b.addScaled(s, "ascendancy", "ClassesAlternate", source{atlas: league[0]}); err != nil {
		return err
	}

	t, u, err := b.treeArt("Default")
	if err != nil {
		return err
	}
	return b.ascendancyFrames(s, "ascendancy", t, u)
}

// bloodlines adds one group per non-default ascendancy art, named after it, e.g. DescendancyAzmeri
// is azmeriBloodline, with the backgrounds of the bloodlines that use it.
func (b *artBuilder) bloodlines() error {
	s := b.newSheet("bloodline-3.webp")
	for i, t := range b.a.treeArt {
		if t.ID == "Default" {
			continue
		}
		name, ok := strings.CutPrefix(t.ID, "Descendancy")
		if !ok || name == "" {
			return fmt.Errorf("unexpected ascendancy art %s", t.ID)
		}
		group := strings.ToLower(name[:1]) + name[1:] + "Bloodline"
		b.sprite(group, s, group)

		var u *int
		for _, desc := range b.d.descendancies {
			if desc.PassiveTreeUIArt == nil || *desc.PassiveTreeUIArt != i {
				continue
			}
			if u != nil && desc.UIArt != nil && *desc.UIArt != *u {
				return fmt.Errorf("bloodlines of %s use different UI art", t.ID)
			}
			u = desc.UIArt
			if err := b.addScaled(s, group, "Classes"+desc.ID, source{atlas: desc.BackgroundImage}); err != nil {
				return err
			}
		}
		if u == nil {
			return fmt.Errorf("no bloodline uses ascendancy art %s", t.ID)
		}
		if err := b.ascendancyFrames(s, group, &b.a.treeArt[i], &b.a.uiArt[*u]); err != nil {
			return err
		}
	}
	return nil
}

func (b *artBuilder) groupBackgrounds() error {
	s := b.newSheet("group-background-3.png")
	b.sprite("startNode", s, "startNode")
	b.sprite("groupBackground", s, "groupBackground")
	for _, c := range b.d.characters {
		if err := b.addScaled(s, "startNode", "center"+strings.ToLower(c.Name), source{atlas: c.SkillTreeBackground}); err != nil {
			return err
		}
	}
	if err := b.addScaled(s, "startNode", "PSStartNodeBackgroundInactive", source{atlas: startNodeInactiveArt}); err != nil {
		return err
	}
	for _, g := range groupBackgrounds {
		if err := b.addScaled(s, "groupBackground", g[0], source{atlas: inGameArt + g[1]}); err != nil {
			return err
		}
	}
	return nil
}

func (b *artBuilder) frames() error {
	s := b.newSheet("frame-3.png")
	b.sprite("frame", s, "frame")
	for _, f := range nodeFrames {
		i := slices.IndexFunc(b.a.nodeFrames, func(a nodeFrameArt) bool { return a.ID == f.row })
		if i < 0 {
			return fmt.Errorf("no PassiveSkillTreeNodeFrameArt row %s", f.row)
		}
		row := b.a.nodeFrames[i]
		for _, a := range [][2]string{{f.normal, row.Normal}, {f.canAllocate, row.CanAllocate}, {f.active, row.Active}} {
			if err := b.addScaled(s, "frame", a[0], source{atlas: a[1]}); err != nil {
				return err
			}
		}
	}
	for _, f := range extraFrames {
		if err := b.addScaled(s, "frame", f[0], source{atlas: inGameArt + f[1]}); err != nil {
			return err
		}
	}
	return nil
}

// jewels adds the socketed jewel art, named like PassiveSkillScreenJewelSocketActive_Blue without
// the prefix and underscore; charms are in their bloodline's group.
func (b *artBuilder) jewels() error {
	s := b.newSheet("jewel-3.png")
	b.sprite("jewel", s, "jewel")
	const prefix = inGameArt + "PassiveSkillScreen"
	for _, j := range b.a.jewelArt {
		for _, img := range []string{j.SocketedImage, j.SocketedImageLarge} {
			name, ok := strings.CutPrefix(img, prefix)
			if !ok {
				continue
			}
			if err := b.addScaled(s, "jewel", strings.ReplaceAll(name, "_", ""), source{atlas: img}); err != nil {
				return err
			}
		}
	}
	return nil
}

// lines stacks the active, normal and intermediate connection art. Each texture is a quarter of
// every orbit's circle, centred on its bottom right corner, with the straight connector along its top.
func (b *artBuilder) lines() error {
	s := b.newSheet("line-3.png")
	b.sprite("line", s, "line")

	for _, state := range []string{"Active", "Normal", "Intermediate"} {
		a, err := b.r.renderScaled(source{file: fmt.Sprintf(connectionCurvesFormat, state)})
		if err != nil {
			return err
		}
		size := a.img.Rect.Dx()

		// Each orbit's arc crosses the bottom row; the outermost crossing is the largest orbit.
		arcs := opaqueRuns(a.img, image.Pt(1, 0), image.Pt(0, size-1), size)
		if len(arcs) != len(orbitRadii)-1 {
			return fmt.Errorf("found %d orbit arcs in the %s connection art, want %d", len(arcs), state, len(orbitRadii)-1)
		}
		for i, arc := range arcs {
			orbit := len(arcs) - i
			s.entries = append(s.entries, sheetEntry{group: "line", name: "Orbit" + strconv.Itoa(orbit) + state, art: a, sub: image.Rect(arc[0], arc[0], size, size)})
		}

		connector := opaqueRuns(a.img, image.Pt(0, 1), image.Pt(0, 0), size)
		if len(connector) != 1 {
			return fmt.Errorf("found %d connectors in the %s connection art, want 1", len(connector), state)
		}
		s.entries = append(s.entries, sheetEntry{group: "line", name: "LineConnector" + state, art: a, sub: image.Rect(0, connector[0][0], size, connector[0][1])})
	}
	for _, d := range lineDecorations {
		if err := b.addScaled(s, "line", d[0], source{atlas: inGameArt + d[1]}); err != nil {
			return err
		}
	}
	return nil
}

// opaqueRuns walks n pixels from start in step and returns the [first, last+1) offsets of each run of visible pixels.
func opaqueRuns(img *image.NRGBA, step, start image.Point, n int) [][2]int {
	const visible = 16
	var runs [][2]int
	in := false
	for i := range n {
		p := start.Add(step.Mul(i))
		a := img.NRGBAAt(p.X, p.Y).A
		switch {
		case a >= visible && !in:
			runs = append(runs, [2]int{i, n})
			in = true
		case a < visible && in:
			runs[len(runs)-1][1] = i
			in = false
		}
	}
	return runs
}

func (b *artBuilder) jewelRadii() error {
	s := b.newSheet("jewel-radius.png")
	b.sprites = append(b.sprites, spriteRef{key: "jewelRadius", sheet: s, group: "jewelRadius", zoom: jewelRadiusZoom})
	for _, j := range jewelRadii {
		src := source{atlas: inGameArt + j[1]}
		w, h, err := b.r.nativeSize(src)
		if err != nil {
			return err
		}
		a, err := b.r.render(src, w, h)
		if err != nil {
			return err
		}
		b.add(s, "jewelRadius", j[0], a)
	}
	return nil
}

// extraImages places each class's illustration at its start group plus the class's art offset.
// The images themselves are extracted with the raw data.
func (b *artBuilder) extraImages() error {
	starts := make(map[int]*Group)
	for _, n := range b.tree.Nodes.List {
		if n.ClassStartIndex != nil && n.Position != nil {
			starts[*n.ClassStartIndex] = b.tree.Groups[n.Group-1]
		}
	}
	chars := slices.Clone(b.d.characters)
	slices.SortFunc(chars, func(a, b character) int { return a.IntegerID - b.IntegerID })
	for _, c := range chars {
		if c.PassiveTreeImage == "" {
			continue
		}
		g, ok := starts[c.IntegerID]
		if !ok {
			return fmt.Errorf("class %s has no start node", c.Name)
		}
		b.tree.ExtraImages = append(b.tree.ExtraImages, Entry[*ExtraImage]{Key: strconv.Itoa(c.IntegerID), Value: &ExtraImage{
			X:     Float(float64(g.X) + float64(c.ArtOffsetX)),
			Y:     Float(float64(g.Y) + float64(c.ArtOffsetY)),
			Image: iconPath(c.PassiveTreeImage),
		}})
	}
	if len(b.tree.ExtraImages) == 0 {
		return errors.New("no class illustrations")
	}
	return nil
}

// Write generates the tree with its sprite sheets and writes dir/data.json and the sheets into dir/assets.
func Write(loader fs.FS, gameVersion string, dir string) (*Tree, error) {
	tree, d, err := generate(loader, gameVersion)
	if err != nil {
		return nil, err
	}
	if err := buildArt(tree, d, loader, gameVersion, filepath.Join(dir, "assets")); err != nil {
		return nil, err
	}
	b, err := tree.Marshal()
	if err != nil {
		return nil, err
	}
	return tree, errors.Wrap(os.WriteFile(filepath.Join(dir, "data.json"), b, 0o644), "failed to write data.json")
}
