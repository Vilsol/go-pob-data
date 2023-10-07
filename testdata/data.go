package testdata

import (
	"io"
	"net/http"
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
