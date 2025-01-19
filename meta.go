package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

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

	return errors.Wrap(os.WriteFile("data/meta.json", bs, 0o644), "failed writing metadata")
}

func IsNewHash(gameVersion string) bool {
	n, _ := strconv.Atoi(strings.Split(gameVersion, ".")[1])
	return n >= 22
}

func GetFileExtension(gameVersion string) string {
	n, _ := strconv.Atoi(strings.Split(gameVersion, ".")[1])
	if n >= 25 {
		return "datc64"
	}
	return "dat64"
}
