package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
)

const GGGRepoBase = "https://raw.githubusercontent.com/grindinggear/skilltree-export/%s/"

var skillTreeSpriteGroups = []string{
	"background",
	"normalActive",
	"notableActive",
	"keystoneActive",
	"normalInactive",
	"notableInactive",
	"keystoneInactive",
	"mastery",
	"masteryConnected",
	"masteryActiveSelected",
	"masteryInactive",
	"masteryActiveEffect",
	"ascendancyBackground",
	"ascendancy",
	"startNode",
	"groupBackground",
	"frame",
	"jewel",
	"line",
	"jewelRadius",
}

func downloadTreeData(treeVersion string, gameVersion string) error {
	repoVersionBase := fmt.Sprintf(GGGRepoBase, treeVersion)
	response, err := http.DefaultClient.Get(repoVersionBase + "/data.json")
	if err != nil {
		return errors.Wrap(err, "failed to fetch skill tree data.json")
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("failed to fetch tree, status code: %d", response.StatusCode)
	}

	defer response.Body.Close()
	dataJSON, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read skill tree response body")
	}

	dataJSONOutPathGzip := filepath.Join("data", gameVersion, "tree", "data.json.gz")

	if err := os.MkdirAll(filepath.Dir(dataJSONOutPathGzip), 0o755); err != nil {
		if !os.IsExist(err) {
			return errors.Wrap(err, "failed to create directory tree")
		}
	}

	fGzip, err := os.OpenFile(dataJSONOutPathGzip, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return errors.Wrap(err, "could not open file for writing")
	}

	writerGzip := gzip.NewWriter(fGzip)
	if _, err := writerGzip.Write(dataJSON); err != nil {
		return errors.Wrap(err, "failed writing json to gzip")
	}

	_ = writerGzip.Close()
	_ = fGzip.Close()

	dataJSONOutPathBrotli := filepath.Join("data", gameVersion, "tree", "data.json.br")

	fBrotli, err := os.OpenFile(dataJSONOutPathBrotli, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return errors.Wrap(err, "could not open file for writing")
	}

	writerBrotli := brotli.NewWriter(fBrotli)
	if _, err := writerBrotli.Write(dataJSON); err != nil {
		return errors.Wrap(err, "failed writing json to brotli")
	}

	_ = writerBrotli.Close()
	_ = fBrotli.Close()

	var skillTreeData SkillTreeData
	if err := json.Unmarshal(dataJSON, &skillTreeData); err != nil {
		return errors.Wrap(err, "failed to unmarshal skill tree from json")
	}

	downloaded := make(map[string]bool)

	selectedResolution := skillTreeData.ImageZoomLevels[len(skillTreeData.ImageZoomLevels)-1]
	for _, group := range skillTreeSpriteGroups {
		spriteGroup := skillTreeData.Sprites[group]
		assetPath := spriteGroup[strconv.FormatFloat(selectedResolution, 'f', -1, 64)]
		if assetPath.Filename == "" {
			for _, p := range spriteGroup {
				assetPath = p
				break
			}
		}

		parsedURL, err := url.Parse(assetPath.Filename)
		if err != nil {
			return errors.Wrap(err, "failed to parse url")
		}

		fileName := path.Base(parsedURL.Path)
		if _, ok := downloaded[fileName]; ok {
			continue
		}
		downloaded[fileName] = true

		println("Downloading", fileName)

		response, err := http.DefaultClient.Get(repoVersionBase + "/assets/" + fileName)
		if err != nil {
			return errors.Wrap(err, "failed to download skill tree asset")
		}

		defer response.Body.Close()
		fileData, err := io.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "failed to read skill tree asset response body")
		}

		fileOutPath := filepath.Join("data", gameVersion, "tree", "assets", fileName)
		if err := os.MkdirAll(filepath.Dir(fileOutPath), 0o755); err != nil {
			if !os.IsExist(err) {
				return errors.Wrap(err, "failed to create directory ree")
			}
		}

		if err := os.WriteFile(fileOutPath, fileData, 0o755); err != nil {
			return errors.Wrap(err, "failed to write skill tree asset to file")
		}
	}

	return nil
}
