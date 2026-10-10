package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every image a published tree references is published itself: the class illustrations under raw/ and
// the sprite sheets under tree/assets/.
func TestTreeImagesPublished(t *testing.T) {
	trees, err := filepath.Glob("data/*/tree/data.json.gz")
	if err != nil {
		t.Fatal(err)
	}

	for _, tree := range trees {
		version := filepath.Base(filepath.Dir(filepath.Dir(tree)))
		t.Run(version, func(t *testing.T) {
			var data struct {
				ExtraImages map[string]struct {
					Image string `json:"image"`
				} `json:"extraImages"`
				Sprites any `json:"sprites"`
			}
			if err := json.Unmarshal(readGzip(t, tree), &data); err != nil {
				t.Fatal(err)
			}

			// Node icons are sprite sheet coordinates, not files, so only the sheets' URLs are files.
			images := sheetURLs(data.Sprites)
			for _, extra := range data.ExtraImages {
				images = append(images, extra.Image)
			}

			for _, image := range images {
				var path string
				switch {
				case strings.HasPrefix(image, "https://data.pob.gg/data/"):
					path = filepath.Join("data", strings.TrimPrefix(image, "https://data.pob.gg/data/"))
				case strings.Contains(image, "://"):
					continue
				default:
					path = filepath.Join("data", version, "raw", image)
				}
				if _, err := os.Stat(path); err != nil {
					t.Errorf("%s references %s, which is not published", tree, image)
				}
			}
		})
	}
}

func sheetURLs(v any) []string {
	var out []string
	switch v := v.(type) {
	case string:
		if strings.Contains(v, "://") {
			out = append(out, v)
		}
	case map[string]any:
		for k, e := range v {
			if k != "coords" {
				out = append(out, sheetURLs(e)...)
			}
		}
	case []any:
		for _, e := range v {
			out = append(out, sheetURLs(e)...)
		}
	}
	return out
}
