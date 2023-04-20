package main

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
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

	return errors.Wrap(os.WriteFile("data/meta.json", bs, 0644), "failed writing metadata")
}
