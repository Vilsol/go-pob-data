package main

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
)

type Metadata struct {
	Versions []string `json:"versions"`
}

func generateMeta() error {
	dir, err := os.ReadDir("data")
	if err != nil {
		return errors.Wrap(err, "failed reading data directory")
	}

	metadata := &Metadata{}
	for _, entry := range dir {
		if entry.IsDir() {
			metadata.Versions = append(metadata.Versions, entry.Name())
		}
	}

	bs, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(err, "failed marshaling metadata")
	}

	return errors.Wrap(os.WriteFile("data/meta.json", bs, 0o644), "failed writing metadata")
}

func IsNewHash(gameVersion string) bool {
	return extractor.IsNewHash(gameVersion)
}

func GetFileExtension(gameVersion string) string {
	return extractor.GetFileExtension(gameVersion)
}
