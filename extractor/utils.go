package extractor

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

var downloadClient = &http.Client{Transport: &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}}

func Fetch(url string) []byte {
	slog.Debug("fetching", slog.String("url", url))
	resp, err := downloadClient.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("fetching %s: %s", url, resp.Status))
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return all
}
