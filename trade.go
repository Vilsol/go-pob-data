package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/loader"
)

// tradeSiteBase serves the trade site's API.
var tradeSiteBase = "https://www.pathofexile.com/"

// tradeStatsFile is the trade site's stats list endpoint for game, which is also the path its response is
// published at under files/ (with .json): the list is not in the game files, and the site serves only the
// live league's.
func tradeStatsFile(game extractor.Game) string {
	if game == extractor.PoE2 {
		return "api/trade2/data/stats"
	}
	return "api/trade/data/stats"
}

// tradeStats is the part of the response checked before it is published.
type tradeStats struct {
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
	Result []struct {
		ID      string            `json:"id"`
		Entries []json.RawMessage `json:"entries"`
	} `json:"result"`
}

// extractTradeStats fetches the trade site's stats list for gameVersion's game and publishes the response
// as it was served.
func extractTradeStats(game extractor.Game, gameVersion string) error {
	url := tradeSiteBase + tradeStatsFile(game)
	slog.Info("fetching", slog.String("url", url))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}
	// the trade site refuses requests without a user agent
	req.Header.Set("User-Agent", "go-pob-data (+https://github.com/Vilsol/go-pob-data)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not fetch "+url)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read "+url)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("fetching " + url + ": " + resp.Status)
	}

	var stats tradeStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return errors.Wrap(err, "could not decode "+url)
	}
	if stats.Error != nil {
		return errors.New("error received from " + url + ": " + stats.Error.Message)
	}
	entries := 0
	for _, category := range stats.Result {
		entries += len(category.Entries)
	}
	if entries == 0 {
		return errors.New(url + " served no stats")
	}

	var buf bytes.Buffer
	w := brotli.NewWriterLevel(&buf, brotli.BestCompression)
	if _, err := w.Write(body); err != nil {
		return errors.Wrap(err, "could not compress "+url)
	}
	if err := w.Close(); err != nil {
		return errors.Wrap(err, "could not compress "+url)
	}

	outPath := filepath.Join("data", gameVersion, loader.FilePath(tradeStatsFile(game)+".json"))
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return errors.Wrap(err, "could not create directory tree")
	}
	return errors.Wrap(os.WriteFile(outPath, buf.Bytes(), 0o644), "could not write "+outPath)
}
