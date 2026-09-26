package treegen

import (
	"bytes"
	"cmp"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/pkg/errors"
	"gopkg.in/gographics/imagick.v3/imagick"
)

// Encoder settings are fixed so the same art always encodes to the same bytes.
const (
	jpegQuality = 86
	webpQuality = 90
)

type sheetEntry struct {
	art   *art
	group string
	name  string
	// sub is the part of the art the entry covers; empty means all of it.
	sub image.Rectangle
}

// sheet is one sprite sheet file; each art is packed once however many entries use it.
type sheet struct {
	image   *image.NRGBA
	rects   map[*art]image.Rectangle
	file    string
	entries []sheetEntry
}

func (s *sheet) add(group, name string, a *art) {
	s.entries = append(s.entries, sheetEntry{group: group, name: name, art: a})
}

// pack lays the arts out in shelves, tallest first, and draws the sheet.
func (s *sheet) pack() {
	arts := make([]*art, 0, len(s.entries))
	seen := make(map[*art]bool)
	area, maxW := 0, 0
	for _, e := range s.entries {
		if seen[e.art] {
			continue
		}
		seen[e.art] = true
		arts = append(arts, e.art)
		b := e.art.img.Bounds()
		area += b.Dx() * b.Dy()
		maxW = max(maxW, b.Dx())
	}
	slices.SortStableFunc(arts, func(a, b *art) int {
		ab, bb := a.img.Bounds(), b.img.Bounds()
		return cmp.Or(cmp.Compare(bb.Dy(), ab.Dy()), cmp.Compare(bb.Dx(), ab.Dx()), cmp.Compare(a.src.String(), b.src.String()))
	})

	width := max(maxW, int(math.Ceil(math.Sqrt(float64(area)))))
	s.rects = make(map[*art]image.Rectangle, len(arts))
	x, y, rowH, height := 0, 0, 0, 0
	for _, a := range arts {
		b := a.img.Bounds()
		if x+b.Dx() > width {
			x, y, rowH = 0, y+rowH, 0
		}
		s.rects[a] = image.Rect(x, y, x+b.Dx(), y+b.Dy())
		x += b.Dx()
		rowH = max(rowH, b.Dy())
		height = max(height, y+rowH)
	}

	s.image = image.NewNRGBA(image.Rect(0, 0, width, height))
	for _, a := range arts {
		draw.Draw(s.image, s.rects[a], a.img, image.Point{}, draw.Src)
	}
}

// derive is a sheet with the same layout whose pixels are transformed by fn.
func (s *sheet) derive(file string, fn func(r, g, b float64) (float64, float64, float64)) *sheet {
	d := &sheet{file: file, rects: s.rects, entries: s.entries, image: image.NewNRGBA(s.image.Rect)}
	for i := 0; i < len(s.image.Pix); i += 4 {
		p := s.image.Pix[i : i+4 : i+4]
		r, g, b := fn(float64(p[0])/255, float64(p[1])/255, float64(p[2])/255)
		d.image.Pix[i] = toByte(r)
		d.image.Pix[i+1] = toByte(g)
		d.image.Pix[i+2] = toByte(b)
		d.image.Pix[i+3] = p[3]
	}
	return d
}

func toByte(v float64) uint8 {
	return uint8(math.Round(255 * min(max(v, 0), 1)))
}

func (s *sheet) coord(e sheetEntry) Coord {
	r := s.rects[e.art]
	if !e.sub.Empty() {
		r = e.sub.Add(r.Min)
	}
	return Coord{X: r.Min.X, Y: r.Min.Y, W: r.Dx(), H: r.Dy()}
}

func (s *sheet) write(dir string) error {
	var buf bytes.Buffer
	switch path.Ext(s.file) {
	case ".png":
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		if err := enc.Encode(&buf, s.image); err != nil {
			return errors.Wrapf(err, "failed to encode %s", s.file)
		}
	case ".jpg":
		// JPEG has no alpha; the export's JPEG sheets are on black.
		opaque := image.NewRGBA(s.image.Rect)
		draw.Draw(opaque, opaque.Rect, image.NewUniform(color.Black), image.Point{}, draw.Src)
		draw.Draw(opaque, opaque.Rect, s.image, image.Point{}, draw.Over)
		if err := jpeg.Encode(&buf, opaque, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return errors.Wrapf(err, "failed to encode %s", s.file)
		}
	case ".webp":
		b, err := encodeWebP(s.image)
		if err != nil {
			return errors.Wrapf(err, "failed to encode %s", s.file)
		}
		buf.Write(b)
	default:
		return fmt.Errorf("unknown sheet format %s", s.file)
	}
	return errors.Wrapf(os.WriteFile(filepath.Join(dir, s.file), buf.Bytes(), 0o644), "failed to write %s", s.file)
}

func encodeWebP(img *image.NRGBA) ([]byte, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	b := img.Bounds()
	if err := mw.SetSize(uint(b.Dx()), uint(b.Dy())); err != nil {
		return nil, errors.Wrap(err, "failed to size image")
	}
	if err := mw.ReadImage("xc:none"); err != nil {
		return nil, errors.Wrap(err, "failed to create image")
	}
	if err := mw.ImportImagePixels(0, 0, uint(b.Dx()), uint(b.Dy()), "RGBA", imagick.PIXEL_CHAR, img.Pix); err != nil {
		return nil, errors.Wrap(err, "failed to set pixels")
	}
	if err := mw.SetImageFormat("webp"); err != nil {
		return nil, errors.Wrap(err, "failed to set format")
	}
	if err := mw.SetImageCompressionQuality(webpQuality); err != nil {
		return nil, errors.Wrap(err, "failed to set quality")
	}
	if err := mw.StripImage(); err != nil {
		return nil, errors.Wrap(err, "failed to strip metadata")
	}
	return mw.GetImageBlob(), nil
}
