package treegen

import (
	"fmt"
	"image"
	"io/fs"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/Vilsol/go-pob-data/extractor"
)

// zoomLevel2 is the scale of the PoE2 sprites: PoB-PoE2 draws the game textures at their size
// (PassiveTree.lua's scaleImage = 1).
const zoomLevel2 = 1

// groupZoom2 scales down the groups of large textures, which at their size would make sheets
// past what browsers load as one texture.
var groupZoom2 = map[string]float64{
	"ascendancyBackground": zoomLevel,
	"masteryActiveEffect":  0.5,
	"line":                 0.5,
}

func zoomOf2(group string) float64 {
	if z, ok := groupZoom2[group]; ok {
		return z
	}
	return zoomLevel2
}

// render2 renders src at its size in the game files scaled by zoom, rounding down like scaled.
func (b *artBuilder2) render2(src source, zoom float64) (*art, error) {
	w, h, err := b.r.nativeSize(src)
	if err != nil {
		return nil, err
	}
	return b.r.render(src, int(math.Floor(float64(w)*zoom+1e-9)), int(math.Floor(float64(h)*zoom+1e-9)))
}

// mainCircleArt2 are the tree's centre circles, which passivetree.lua:382-383 names by path.
var mainCircleArt2 = [][2]string{
	{"BGTree", "Art/2DArt/UIImages/InGame/PassiveTree/PassiveTreeMainCircle"},
	{"BGTreeActive", "Art/2DArt/UIImages/InGame/PassiveTree/PassiveTreeMainCircleActive2"},
}

// legionArt2 is passivetree.lua:419-433.
var legionArt2 = []string{
	"Art/2DArt/UIImages/InGame/Abyss/AbyssPassiveSkillScreenJewelCircle1",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenEternalEmpireJewelCircle1",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenEternalEmpireJewelCircle2",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenKalguuranJewelCircle1",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenKalguuranJewelCircle2",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenKaruiJewelCircle1",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenKaruiJewelCircle2",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenMarakethJewelCircle1",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenMarakethJewelCircle2",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenTemplarJewelCircle1",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenTemplarJewelCircle2",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenVaalJewelCircle1",
	"Art/2DArt/UIImages/InGame/PassiveSkillScreenVaalJewelCircle2",
}

const (
	ascendancyMiddleArt2 = "Art/2DArt/UIImages/InGame/PassiveSkillScreenAscendancyMiddle"
	startNodeArt2        = "Art/2DArt/UIImages/InGame/PassiveSkillScreenStartNodeBackgroundInactive"
)

// connectionArt2 are the connection art rows passivetree.lua:263-268 decomposes, plus the main tree's.
var connectionArt2 = []string{"Character", "CharacterAscendancy", "CharacterPlanned"}

type passiveSkillTreeArt2 struct {
	UIArt *int   `json:"UIArt"`
	ID    string `json:"Id"`
}

type connectionArt2Row struct {
	ID           string `json:"Id"`
	Normal       string `json:"Normal"`
	Intermediate string `json:"Intermediate"`
	Active       string `json:"Active"`
	Mask         string `json:"Mask"`
}

type jewelArt2 struct {
	JewelArt     string `json:"SocketedImage"`
	BaseItemType int    `json:"BaseItemTypesKey"`
}

type uniqueJewelArt2 struct {
	Words    *int   `json:"Name"`
	JewelArt string `json:"JewelArt"`
}

type word2 struct {
	Text string `json:"Text"`
}

type alternatePassive2 struct {
	DDSIcon string `json:"DDSIcon"`
}

type baseItemName2 struct {
	Name string `json:"Name"`
}

type artData2 struct {
	trees          []passiveSkillTreeArt2
	connections    []connectionArt2Row
	jewelArt       []jewelArt2
	uniqueJewelArt []uniqueJewelArt2
	words          []word2
	alternates     []alternatePassive2
	baseItems      []baseItemName2
}

func loadArtData2(loader fs.FS, gameVersion string) (*artData2, error) {
	a := &artData2{}
	l := &tableLoader{loader: loader, game: extractor.PoE2, gameVersion: gameVersion}
	loadTable(l, "PassiveSkillTrees", &a.trees)
	loadTable(l, "PassiveSkillTreeConnectionArt", &a.connections)
	loadTable(l, "PassiveJewelArt", &a.jewelArt)
	loadTable(l, "PassiveJewelUniqueArt", &a.uniqueJewelArt)
	loadTable(l, "Words", &a.words)
	loadTable(l, "AlternatePassiveSkills", &a.alternates)
	loadTable(l, "BaseItemTypes", &a.baseItems)
	return a, l.err
}

type artBuilder2 struct {
	r      *renderer
	d      *gameData2
	a      *artData2
	tree   *Tree2
	ui     *treeUIArt2
	names  map[[2]string]bool
	byLow  map[string]string
	byPage map[string]string
	sheets []*sheet
	refs   []spriteRef
}

func (b *artBuilder2) newSheet(file string) *sheet {
	s := &sheet{file: file}
	b.sheets = append(b.sheets, s)
	return s
}

func (b *artBuilder2) sprite(key string, s *sheet, group string) {
	b.refs = append(b.refs, spriteRef{key: key, sheet: s, group: group, zoom: zoomOf2(group)})
}

// addNative puts src on s at its group's zoom, once per group and name.
func (b *artBuilder2) addNative(s *sheet, group, name string, src source) error {
	k := [2]string{group, name}
	if b.names[k] {
		return nil
	}
	a, err := b.render2(src, zoomOf2(group))
	if err != nil {
		return err
	}
	b.names[k] = true
	s.add(group, name, a)
	return nil
}

// uiImage is the UI image named name, found ignoring case like AssetSheets.lua's parseUIImages.
func (b *artBuilder2) uiImage(name string) (source, error) {
	exact, ok := b.byLow[strings.ToLower(name)]
	if !ok {
		return source{}, fmt.Errorf("no ui image %q", name)
	}
	return source{atlas: exact}, nil
}

// uiPath is the UI image whose texture is path: AssetSheets.lua:154-164 crops a texture to the
// rect every UI image on it shares.
func (b *artBuilder2) uiPath(path string) (source, error) {
	name, ok := b.byPage[strings.ToLower(path)]
	if !ok {
		return source{}, fmt.Errorf("no ui image on %q", path)
	}
	if name == "" {
		return source{file: path}, nil
	}
	return source{atlas: name}, nil
}

func (b *artBuilder2) frameArt(row *int) (*nodeFrameArt2, error) {
	if row == nil {
		return nil, fmt.Errorf("missing node frame art")
	}
	return &b.d.frameArt[*row], nil
}

// buildArt2 renders the PoE2 sprite sheets into dir and sets the tree's sprites and zoom levels,
// from the textures passivetree.lua puts on its sheets.
func buildArt2(tree *Tree2, d *gameData2, loader fs.FS, gameVersion string, dir string) error {
	a, err := loadArtData2(loader, gameVersion)
	if err != nil {
		return err
	}
	r, err := newRenderer(loader, extractor.PoE2, gameVersion)
	if err != nil {
		return err
	}
	defer r.close()

	b := &artBuilder2{r: r, d: d, a: a, tree: tree, names: make(map[[2]string]bool), byLow: make(map[string]string), byPage: make(map[string]string)}
	names := make([]string, 0, len(r.atlas))
	for name := range r.atlas {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		b.byLow[strings.ToLower(name)] = name
		s := r.atlas[name]
		page := strings.ToLower(s.page)
		if other, ok := b.byPage[page]; !ok {
			b.byPage[page] = name
		} else if other != "" && r.atlas[other] != s {
			b.byPage[page] = ""
		}
	}

	ti := slices.IndexFunc(a.trees, func(t passiveSkillTreeArt2) bool { return t.ID == mainTreeID })
	if ti < 0 || a.trees[ti].UIArt == nil {
		return fmt.Errorf("no ui art for passive tree %q", mainTreeID)
	}
	b.ui = &d.uiArt[*a.trees[ti].UIArt]

	for _, step := range []func() error{
		b.nodeIcons, b.masteries, b.ascendancyBackgrounds, b.groupBackgrounds, b.frames,
		b.jewels, b.legion, b.lines,
	} {
		if err := step(); err != nil {
			return err
		}
	}

	tree.Sprites, err = writeSheets(dir, gameVersion, b.sheets, b.refs)
	if err != nil {
		return err
	}
	tree.ImageZoomLevels = []float64{zoomLevel2}
	return nil
}

// nodeIcons adds each node's icon by its type (passivetree.lua:699-716) and the icons of the
// options it can switch to (:649, :937, :973). PoB-PoE2 draws unallocated icons as they are, so the
// inactive groups share the active sheet.
func (b *artBuilder2) nodeIcons() error {
	s := b.newSheet("skills-3.webp")
	add := func(group, icon string) error {
		if icon == "" {
			return nil
		}
		return b.addNative(s, group, icon, source{file: icon})
	}
	options := func(n *Node2) []NodeOption2 {
		switch o := n.Options.(type) {
		case []NodeOption2:
			return o
		case map[string]NodeOption2:
			keys := make([]string, 0, len(o))
			for k := range o {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			out := make([]NodeOption2, 0, len(o))
			for _, k := range keys {
				out = append(out, o[k])
			}
			return out
		}
		return nil
	}

	for _, n := range b.sortedNodes() {
		if n.Icon != nil && !n.IsOnlyImage {
			group := "normalActive"
			switch {
			case n.IsKeystone:
				group = "keystoneActive"
			case n.IsNotable:
				group = "notableActive"
			case n.IsJewelSocket:
				group = "socketActive"
			}
			if err := add(group, *n.Icon); err != nil {
				return fmt.Errorf("icon of node %d: %w", n.Skill, err)
			}
		}
		for _, o := range options(n) {
			if err := add("normalActive", o.Icon); err != nil {
				return fmt.Errorf("option icon of node %d: %w", n.Skill, err)
			}
		}
	}

	for _, t := range []string{"normal", "notable", "keystone", "socket"} {
		b.sprite(t+"Active", s, t+"Active")
	}
	for _, t := range []string{"normal", "notable", "keystone", "socket"} {
		b.sprite(t+"Inactive", s, t+"Active")
	}
	return nil
}

func (b *artBuilder2) sortedNodes() []*Node2 {
	ids := make([]uint32, 0, len(b.tree.Nodes))
	for _, n := range b.tree.Nodes {
		ids = append(ids, n.Skill)
	}
	slices.Sort(ids)
	out := make([]*Node2, len(ids))
	for i, id := range ids {
		out[i] = b.tree.Nodes[strconv.FormatUint(uint64(id), 10)]
	}
	return out
}

// masteries adds the mastery effect art (passivetree.lua:851-854).
func (b *artBuilder2) masteries() error {
	s := b.newSheet("mastery-active-effect-3.webp")
	b.sprite("masteryActiveEffect", s, "masteryActiveEffect")
	for _, n := range b.sortedNodes() {
		if n.ActiveEffectImage == "" {
			continue
		}
		src, err := b.uiImage(n.ActiveEffectImage)
		if err != nil {
			return err
		}
		if err := b.addNative(s, "masteryActiveEffect", n.ActiveEffectImage, src); err != nil {
			return err
		}
	}
	return nil
}

// ascendancyBackgrounds adds the tree's centre circles (passivetree.lua:382-383) and each class's
// and ascendancy's background (:548, :578).
func (b *artBuilder2) ascendancyBackgrounds() error {
	s := b.newSheet("ascendancy-background-3.webp")
	b.sprite("ascendancyBackground", s, "ascendancyBackground")
	for _, c := range mainCircleArt2 {
		src, err := b.uiImage(c[1])
		if err != nil {
			return err
		}
		if err := b.addNative(s, "ascendancyBackground", c[0], src); err != nil {
			return err
		}
	}
	for _, class := range b.tree.Classes {
		ci := slices.IndexFunc(b.d.characters, func(c character2) bool { return c.Name == class.Name })
		if err := b.addBackground(s, class.Background.Image, b.d.characters[ci].PassiveTreeImage); err != nil {
			return err
		}
		for _, asc := range class.Ascendancies {
			ai := slices.IndexFunc(b.d.ascendancies, func(a ascendancy2) bool { return a.ID == asc.InternalID })
			if err := b.addBackground(s, asc.Background.Image, b.d.ascendancies[ai].PassiveTreeImage); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *artBuilder2) addBackground(s *sheet, name, path string) error {
	if path == "" {
		return nil
	}
	src, err := b.uiPath(path)
	if err != nil {
		src = source{file: path}
	}
	return b.addNative(s, "ascendancyBackground", name, src)
}

// groupBackgrounds is passivetree.lua:312-319, :352-359 and :378-379; the Blank names reuse the
// group backgrounds as PoB-PoE2 does.
func (b *artBuilder2) groupBackgrounds() error {
	s := b.newSheet("group-background-3.png")
	b.sprite("groupBackground", s, "groupBackground")
	b.sprite("startNode", s, "startNode")
	for _, g := range [][2]string{
		{"PSGroupBackground1", b.ui.GroupBackgroundSmall},
		{"PSGroupBackground2", b.ui.GroupBackgroundMedium},
		{"PSGroupBackground3", b.ui.GroupBackgroundLarge},
		{"PSGroupBackgroundSmallBlank", b.ui.GroupBackgroundSmall},
		{"PSGroupBackgroundMediumBlank", b.ui.GroupBackgroundMedium},
		{"PSGroupBackgroundLargeBlank", b.ui.GroupBackgroundLarge},
	} {
		src, err := b.uiImage(g[1])
		if err != nil {
			return err
		}
		if err := b.addNative(s, "groupBackground", g[0], src); err != nil {
			return err
		}
	}
	src, err := b.uiImage(startNodeArt2)
	if err != nil {
		return err
	}
	return b.addNative(s, "startNode", "PSStartNodeBackgroundInactive", src)
}

// frames adds the named frames of passivetree.lua:322-376 and :580-597, and every frame a node
// overlay names by its texture path (:743-752, :780-789, :991-1000).
func (b *artBuilder2) frames() error {
	s := b.newSheet("frame-3.png")
	b.sprite("frame", s, "frame")

	named := func(row *int, normal, canAllocate, active string) error {
		f, err := b.frameArt(row)
		if err != nil {
			return err
		}
		for _, a := range [][2]string{{normal, f.Normal}, {active, f.Active}, {canAllocate, f.CanAllocate}} {
			src, err := b.uiImage(a[1])
			if err != nil {
				return err
			}
			if err := b.addNative(s, "frame", a[0], src); err != nil {
				return err
			}
		}
		return nil
	}
	if err := named(b.ui.PassiveFrame, "PSSkillFrame", "PSSkillFrameHighlighted", "PSSkillFrameActive"); err != nil {
		return err
	}
	if err := named(b.ui.KeystoneFrame, "KeystoneFrameUnallocated", "KeystoneFrameCanAllocate", "KeystoneFrameAllocated"); err != nil {
		return err
	}
	if err := named(b.ui.NotableFrame, "NotableFrameUnallocated", "NotableFrameCanAllocate", "NotableFrameAllocated"); err != nil {
		return err
	}
	if err := named(b.ui.JewelFrame, "JewelFrameUnallocated", "JewelFrameCanAllocate", "JewelFrameAllocated"); err != nil {
		return err
	}
	src, err := b.uiImage(ascendancyMiddleArt2)
	if err != nil {
		return err
	}
	if err := b.addNative(s, "frame", "AscendancyMiddle", src); err != nil {
		return err
	}

	for _, class := range b.tree.Classes {
		for _, asc := range class.Ascendancies {
			a := b.d.ascendancies[slices.IndexFunc(b.d.ascendancies, func(a ascendancy2) bool { return a.ID == asc.InternalID })]
			if a.UIArt == nil {
				return fmt.Errorf("ascendancy %s has no ui art", a.ID)
			}
			ui := b.d.uiArt[*a.UIArt]
			if err := named(ui.PassiveFrame, asc.Name+"FrameSmallNormal", asc.Name+"FrameSmallCanAllocate", asc.Name+"FrameSmallAllocated"); err != nil {
				return err
			}
			if err := named(ui.NotableFrame, asc.Name+"FrameLargeNormal", asc.Name+"FrameLargeCanAllocate", asc.Name+"FrameLargeAllocated"); err != nil {
				return err
			}
		}
	}

	overlay := func(o *Overlay) error {
		if o == nil {
			return nil
		}
		for _, name := range []string{o.Unalloc, o.Alloc, o.Path} {
			if b.names[[2]string{"frame", name}] {
				continue
			}
			src, err := b.uiPath(name)
			if err != nil {
				return err
			}
			if err := b.addNative(s, "frame", name, src); err != nil {
				return err
			}
		}
		return nil
	}
	for _, n := range b.sortedNodes() {
		if err := overlay(n.NodeOverlay); err != nil {
			return fmt.Errorf("overlay of node %d: %w", n.Skill, err)
		}
		if o, ok := n.Options.(map[string]NodeOption2); ok {
			keys := make([]string, 0, len(o))
			for k := range o {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			for _, k := range keys {
				if err := overlay(o[k].NodeOverlay); err != nil {
					return fmt.Errorf("option overlay of node %d: %w", n.Skill, err)
				}
			}
		}
	}
	return nil
}

// jewels adds the socketed jewel art by base item name and the unique jewel art by unique name
// (passivetree.lua:385-411).
func (b *artBuilder2) jewels() error {
	s := b.newSheet("jewel-3.png")
	b.sprite("jewel", s, "jewel")
	add := func(name, art string) error {
		if ignoreFilter2.MatchString(name) || art == "" {
			return nil
		}
		src, err := b.uiImage(art)
		if err != nil {
			return err
		}
		return b.addNative(s, "jewel", name, src)
	}
	for _, j := range b.a.jewelArt {
		if err := add(b.a.baseItems[j.BaseItemType].Name, j.JewelArt); err != nil {
			return err
		}
	}
	for _, j := range b.a.uniqueJewelArt {
		if j.Words == nil {
			continue
		}
		if err := add(b.a.words[*j.Words].Text, j.JewelArt); err != nil {
			return err
		}
	}
	return nil
}

// legion is passivetree.lua:413-438: the alternate passives' icons by path and the jewel radius
// circles by lowercased texture path.
func (b *artBuilder2) legion() error {
	s := b.newSheet("legion-3.png")
	b.sprite("legion", s, "legion")
	for _, p := range b.a.alternates {
		if p.DDSIcon == "" {
			continue
		}
		if err := b.addNative(s, "legion", p.DDSIcon, source{file: p.DDSIcon}); err != nil {
			return err
		}
	}
	for _, art := range legionArt2 {
		src, err := b.uiImage(art)
		if err != nil {
			return err
		}
		if err := b.addNative(s, "legion", strings.ToLower(b.r.atlas[src.atlas].page), src); err != nil {
			return err
		}
	}
	return nil
}

// lines cuts each connection art's curves texture like PoE1's lines: a quarter of every orbit's
// circle centred on its bottom right corner, with the straight connector along its top. The
// textures nest the arcs, so each sprite keeps only the pixels of its arc's region of the
// connection art's mask, as passivetree.lua:1252 (extract_lines.scm) does. The names are
// passivetree.lua:1256-1283's, prefixed by the connection art.
func (b *artBuilder2) lines() error {
	s := b.newSheet("line-3.png")
	b.sprite("line", s, "line")

	// Each arc is the orbit of that radius rank; orbit 0 has no arc.
	byRadius := make([]int, 0, len(orbitRadii2)-1)
	for i := 1; i < len(orbitRadii2); i++ {
		byRadius = append(byRadius, i)
	}
	slices.SortFunc(byRadius, func(x, y int) int { return int(orbitRadii2[y] - orbitRadii2[x]) })

	ids := slices.Clone(connectionArt2)
	if b.ui.Connection != nil && !slices.Contains(ids, b.a.connections[*b.ui.Connection].ID) {
		ids = append(ids, b.a.connections[*b.ui.Connection].ID)
	}
	for _, id := range ids {
		ci := slices.IndexFunc(b.a.connections, func(c connectionArt2Row) bool { return c.ID == id })
		if ci < 0 {
			return fmt.Errorf("no connection art %s", id)
		}
		c := b.a.connections[ci]
		mask, err := b.render2(source{file: c.Mask}, zoomOf2("line"))
		if err != nil {
			return err
		}
		owners := maskOwners(mask.img)
		for _, state := range [][2]string{{"Normal", c.Normal}, {"Intermediate", c.Intermediate}, {"Active", c.Active}} {
			a, err := b.render2(source{file: state[1]}, zoomOf2("line"))
			if err != nil {
				return err
			}
			if a.img.Rect != mask.img.Rect {
				return fmt.Errorf("the %s %s connection art is %v, its mask %v", id, state[0], a.img.Rect, mask.img.Rect)
			}
			// The textures carry a padding row and column past the corner the arcs are centred on, and
			// compression speckles their faint shadows, so arcs are found by their solid cores.
			size := a.img.Rect.Dx()
			for size > 1 && len(solidRuns(a.img, image.Pt(1, 0), image.Pt(0, size-1), size)) == 0 {
				size--
			}

			arcs := solidRuns(a.img, image.Pt(1, 0), image.Pt(0, size-1), size)
			if len(arcs) != len(byRadius) {
				return fmt.Errorf("found %d orbit arcs in the %s %s connection art, want %d", len(arcs), id, state[0], len(byRadius))
			}
			cut := lineCutter{art: a, owners: owners, used: make(map[int]string)}
			for i, arc := range arcs {
				name := id + "Orbit" + strconv.Itoa(byRadius[i]) + state[0]
				if err := cut.add(s, name, image.Rect(arc[0], arc[0], size, size), image.Pt(1, 0), image.Pt(arc[0], size-1), arc[1]-arc[0]); err != nil {
					return err
				}
			}

			// Left of the largest arc only the connector crosses a column.
			x := arcs[0][0] / 2
			connector := solidRuns(a.img, image.Pt(0, 1), image.Pt(x, 0), size)
			if len(connector) != 1 {
				return fmt.Errorf("found %d connectors in the %s %s connection art, want 1", len(connector), id, state[0])
			}
			name := id + "LineConnector" + state[0]
			if err := cut.add(s, name, image.Rect(0, connector[0][0], size, connector[0][1]), image.Pt(0, 1), image.Pt(x, connector[0][0]), connector[0][1]-connector[0][0]); err != nil {
				return err
			}
		}
	}
	return nil
}

// lineCutter cuts one connection art texture into a sprite per line, each the pixels of its own
// region of the mask.
type lineCutter struct {
	art    *art
	used   map[int]string
	owners []int
}

// add puts the pixels of sub that share the mask region of the line's most opaque pixel among the
// n from start along step on s as its own art.
func (c *lineCutter) add(s *sheet, name string, sub image.Rectangle, step, start image.Point, n int) error {
	img := c.art.img
	w := img.Rect.Dx()
	at, best := start, -1
	for i := 0; i < n; i++ {
		p := start.Add(step.Mul(i))
		if a := int(img.NRGBAAt(p.X, p.Y).A); a > best {
			at, best = p, a
		}
	}
	region := c.owners[at.Y*w+at.X]
	if region == 0 {
		return fmt.Errorf("%s is in no region of its mask", name)
	}
	if other, ok := c.used[region]; ok {
		return fmt.Errorf("%s and %s share a region of their mask", name, other)
	}
	c.used[region] = name

	out := image.NewNRGBA(image.Rect(0, 0, sub.Dx(), sub.Dy()))
	for y := sub.Min.Y; y < sub.Max.Y; y++ {
		for x := sub.Min.X; x < sub.Max.X; x++ {
			if c.owners[y*w+x] == region {
				out.SetNRGBA(x-sub.Min.X, y-sub.Min.Y, img.NRGBAAt(x, y))
			}
		}
	}
	s.add("line", name, &art{img: out, src: c.art.src})
	return nil
}

// maskOwners labels each separate lit area of a connection art mask. extract_lines.scm selects the
// black between the areas with GIMP's default threshold of 15.
func maskOwners(mask *image.NRGBA) []int {
	const lit = 16
	w, h := mask.Rect.Dx(), mask.Rect.Dy()
	isLit := func(i int) bool {
		p := mask.Pix[i*4 : i*4+4 : i*4+4]
		return p[3] > 0 && max(p[0], p[1], p[2]) >= lit
	}

	owners := make([]int, w*h)
	label := 0
	for i := range owners {
		if owners[i] != 0 || !isLit(i) {
			continue
		}
		label++
		owners[i] = label
		stack := []int{i}
		for len(stack) > 0 {
			j := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			x, y := j%w, j/w
			for _, k := range [4]int{j - 1, j + 1, j - w, j + w} {
				if (k == j-1 && x == 0) || (k == j+1 && x == w-1) || (k == j-w && y == 0) || (k == j+w && y == h-1) {
					continue
				}
				if owners[k] == 0 && isLit(k) {
					owners[k] = label
					stack = append(stack, k)
				}
			}
		}
	}
	return owners
}

// solidRuns are the runs of opaqueRuns that hold a solid pixel, each widened over the faint
// pixels around it up to its neighbours.
func solidRuns(img *image.NRGBA, step, start image.Point, n int) [][2]int {
	const solid, faint = 128, 16
	alpha := func(i int) uint8 {
		p := start.Add(step.Mul(i))
		return img.NRGBAAt(p.X, p.Y).A
	}
	var runs [][2]int
	for i := 0; i < n; i++ {
		if alpha(i) < solid {
			continue
		}
		lo := i
		floor := 0
		if len(runs) > 0 {
			floor = runs[len(runs)-1][1]
		}
		for lo > floor && alpha(lo-1) >= faint {
			lo--
		}
		hi := i
		for hi < n && alpha(hi) >= solid {
			hi++
		}
		for hi < n && alpha(hi) >= faint && alpha(hi) < solid {
			hi++
		}
		runs = append(runs, [2]int{lo, hi})
		i = hi - 1
	}
	return runs
}
