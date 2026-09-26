package treegen

import (
	"bufio"
	"fmt"
	"image"
	"io"
	"io/fs"
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/unicode"
	"gopkg.in/gographics/imagick.v3/imagick"
)

// uiAtlasFile lists the named UI images and where each sits on its atlas page.
const uiAtlasFile = "Art/UIImages1.txt"

type atlasSprite struct {
	page       string
	x, y, w, h int
}

// loadUIAtlas parses lines of the form `"name" "page.dds" x1 y1 x2 y2`; the corners are inclusive.
func loadUIAtlas(loader fs.FS, gameVersion string) (map[string]atlasSprite, error) {
	f, err := loader.Open(gamePath(gameVersion, uiAtlasFile))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", uiAtlasFile)
	}
	defer f.Close()

	atlas := make(map[string]atlasSprite)
	scanner := bufio.NewScanner(unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Reader(f))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, `"`)
		if len(parts) != 5 {
			return nil, fmt.Errorf("bad %s line %q", uiAtlasFile, line)
		}
		corners := strings.Fields(parts[4])
		if len(corners) != 4 {
			return nil, fmt.Errorf("bad %s line %q", uiAtlasFile, line)
		}
		var c [4]int
		for i, s := range corners {
			if c[i], err = strconv.Atoi(s); err != nil {
				return nil, errors.Wrapf(err, "bad %s line %q", uiAtlasFile, line)
			}
		}
		atlas[parts[1]] = atlasSprite{page: parts[3], x: c[0], y: c[1], w: c[2] - c[0] + 1, h: c[3] - c[1] + 1}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", uiAtlasFile)
	}
	return atlas, nil
}

// source is a game texture: a whole .dds file, or a named sprite of the UI atlas.
type source struct {
	file  string
	atlas string
}

func (s source) String() string {
	return s.file + s.atlas
}

// art is a source rendered at a fixed size, straight (non-premultiplied) RGBA.
type art struct {
	img *image.NRGBA
	src source
}

type artKey struct {
	src  source
	w, h int
}

// renderer decodes game textures with ImageMagick and resizes them to sprite size.
type renderer struct {
	loader      fs.FS
	atlas       map[string]atlasSprite
	pages       map[string]*imagick.MagickWand
	arts        map[artKey]*art
	gameVersion string
}

func newRenderer(loader fs.FS, gameVersion string) (*renderer, error) {
	// ImageMagick can't be initialized again after imagick.Terminate, so terminating is left to the caller.
	imagick.Initialize()

	atlas, err := loadUIAtlas(loader, gameVersion)
	if err != nil {
		return nil, err
	}
	return &renderer{
		loader:      loader,
		gameVersion: gameVersion,
		atlas:       atlas,
		pages:       make(map[string]*imagick.MagickWand),
		arts:        make(map[artKey]*art),
	}, nil
}

func (r *renderer) close() {
	for _, mw := range r.pages {
		mw.Destroy()
	}
}

func (r *renderer) decode(path string) (*imagick.MagickWand, error) {
	f, err := r.loader.Open(gamePath(r.gameVersion, path))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", path)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", path)
	}
	mw := imagick.NewMagickWand()
	if err := mw.ReadImageBlob(b); err != nil {
		mw.Destroy()
		return nil, errors.Wrapf(err, "failed to decode %s", path)
	}
	return mw, nil
}

// load returns a new wand holding the source at its native size.
func (r *renderer) load(src source) (*imagick.MagickWand, error) {
	if src.file != "" {
		return r.decode(src.file)
	}

	s, ok := r.atlas[src.atlas]
	if !ok {
		return nil, fmt.Errorf("no UI atlas image %s", src.atlas)
	}
	page, ok := r.pages[s.page]
	if !ok {
		var err error
		if page, err = r.decode(s.page); err != nil {
			return nil, err
		}
		r.pages[s.page] = page
	}
	mw := page.Clone()
	if err := mw.CropImage(uint(s.w), uint(s.h), s.x, s.y); err != nil {
		mw.Destroy()
		return nil, errors.Wrapf(err, "failed to crop %s", src.atlas)
	}
	if err := mw.ResetImagePage(""); err != nil {
		mw.Destroy()
		return nil, errors.Wrapf(err, "failed to crop %s", src.atlas)
	}
	return mw, nil
}

// nativeSize is the source's size in the game files.
func (r *renderer) nativeSize(src source) (int, int, error) {
	if src.atlas != "" {
		s, ok := r.atlas[src.atlas]
		if !ok {
			return 0, 0, fmt.Errorf("no UI atlas image %s", src.atlas)
		}
		return s.w, s.h, nil
	}
	mw, err := r.load(src)
	if err != nil {
		return 0, 0, err
	}
	defer mw.Destroy()
	return int(mw.GetImageWidth()), int(mw.GetImageHeight()), nil
}

// render returns the source resized to w x h; the same source and size always yields the same *art.
func (r *renderer) render(src source, w, h int) (*art, error) {
	key := artKey{src: src, w: w, h: h}
	if a, ok := r.arts[key]; ok {
		return a, nil
	}

	mw, err := r.load(src)
	if err != nil {
		return nil, err
	}
	defer mw.Destroy()

	if int(mw.GetImageWidth()) != w || int(mw.GetImageHeight()) != h {
		if err := mw.ResizeImage(uint(w), uint(h), imagick.FILTER_CATROM); err != nil {
			return nil, errors.Wrapf(err, "failed to resize %s", src)
		}
	}
	pix, err := mw.ExportImagePixels(0, 0, uint(w), uint(h), "RGBA", imagick.PIXEL_CHAR)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read pixels of %s", src)
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pix.([]byte))

	a := &art{src: src, img: img}
	r.arts[key] = a
	return a, nil
}

// scaled converts a size at zoom 1 to the sprite zoom level, rounding down like the export.
func scaled(n float64) int {
	return int(math.Floor(n*zoomLevel + 1e-9))
}

// renderScaled renders the source at its native size scaled to the sprite zoom level.
func (r *renderer) renderScaled(src source) (*art, error) {
	w, h, err := r.nativeSize(src)
	if err != nil {
		return nil, err
	}
	return r.render(src, scaled(float64(w)), scaled(float64(h)))
}
