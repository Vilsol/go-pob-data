package testdata

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var GameVersion = sync.OnceValue(func() string {
	resp, err := http.DefaultClient.Get("https://raw.githubusercontent.com/poe-tool-dev/latest-patch-version/main/latest.txt")
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body)
})

func ShortGameVersion() string {
	return strings.Join(strings.Split(GameVersion(), ".")[:2], ".")
}

// LatestDataVersion is the newest version committed in data/meta.json, so tests against data.pob.gg
// only move when this repo adds data, not whenever a new patch goes live.
var LatestDataVersion = sync.OnceValue(func() string {
	_, file, _, _ := runtime.Caller(0)
	b, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "data", "meta.json"))
	if err != nil {
		panic(err)
	}

	var meta struct {
		Versions []string `json:"versions"`
	}
	if err := json.Unmarshal(b, &meta); err != nil {
		panic(err)
	}

	return meta.Versions[len(meta.Versions)-1]
})
