package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/loader"
)

func readBrotli(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.HasPrefix(b, []byte("version https://git-lfs")) {
		t.Skip("git lfs pointer, run git lfs pull")
	}
	out, err := io.ReadAll(brotli.NewReader(bytes.NewReader(b)))
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return out
}

func serveTradeStats(t *testing.T, status int, body string) *string {
	t.Helper()
	var agent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/trade2/data/stats" {
			http.NotFound(w, r)
			return
		}
		agent = r.UserAgent()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	prev := tradeSiteBase
	tradeSiteBase = srv.URL + "/"
	t.Cleanup(func() { tradeSiteBase = prev })
	return &agent
}

// The response is published byte for byte where LoadFile reads it, and fetched with a user agent.
func TestExtractTradeStats(t *testing.T) {
	const body = `{"result":[{"id":"explicit","label":"Explicit","entries":[{"id":"explicit.stat_1","text":"# to\nLife","type":"explicit"}]}]}`
	agent := serveTradeStats(t, http.StatusOK, body)
	inTempDir(t)

	if err := extractTradeStats(extractor.PoE2, "0.5"); err != nil {
		t.Fatal(err)
	}
	got := readBrotli(t, filepath.Join("data", "0.5", loader.FilePath("api/trade2/data/stats.json")))
	if string(got) != body {
		t.Errorf("published %q, want the response %q", got, body)
	}
	if *agent == "" || strings.HasPrefix(*agent, "Go-http-client") {
		t.Errorf("fetched with user agent %q", *agent)
	}
}

// A response that is not a stats list publishes nothing.
func TestExtractTradeStatsRefuses(t *testing.T) {
	for name, c := range map[string]struct {
		body   string
		status int
	}{
		"status":      {status: http.StatusServiceUnavailable, body: `{"result":[{"id":"explicit","entries":[{}]}]}`},
		"api error":   {status: http.StatusOK, body: `{"error":{"code":1,"message":"Resource not found"}}`},
		"no stats":    {status: http.StatusOK, body: `{"result":[{"id":"explicit","entries":[]}]}`},
		"not json":    {status: http.StatusOK, body: `<html>`},
		"empty array": {status: http.StatusOK, body: `{"result":[]}`},
	} {
		t.Run(name, func(t *testing.T) {
			serveTradeStats(t, c.status, c.body)
			inTempDir(t)
			if err := extractTradeStats(extractor.PoE2, "0.5"); err == nil {
				t.Fatal("published a response that is not a stats list")
			}
			if _, err := os.Stat("data"); !os.IsNotExist(err) {
				t.Error("wrote data for a refused response")
			}
		})
	}
}

type publishedTradeStats struct {
	Result []struct {
		ID      string `json:"id"`
		Label   string `json:"label"`
		Entries []struct {
			Option *struct {
				Options []struct {
					ID   json.Number `json:"id"`
					Text string      `json:"text"`
				} `json:"options"`
			} `json:"option"`
			ID   string `json:"id"`
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"entries"`
	} `json:"result"`
}

// Every published stats list is its own game's, and every entry has an id under its category, a text and
// a type.
func TestTradeStatsPublished(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("data", "*", "files", "api", "*", "data", "stats.json.br"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no trade stats published")
	}
	games := map[extractor.Game]bool{}
	for _, path := range paths {
		version := strings.Split(filepath.ToSlash(path), "/")[1]
		t.Run(version, func(t *testing.T) {
			game := extractor.GameOf(version)
			games[game] = true
			if want := filepath.Join("data", version, loader.FilePath(tradeStatsFile(game)+".json")); path != want {
				t.Fatalf("%s is not where %s's list is published (%s)", path, game, want)
			}
			dec := json.NewDecoder(bytes.NewReader(readBrotli(t, path)))
			dec.UseNumber()
			var stats publishedTradeStats
			if err := dec.Decode(&stats); err != nil {
				t.Fatal(err)
			}
			entries := 0
			for _, category := range stats.Result {
				if category.ID == "" || category.Label == "" {
					t.Errorf("category %q has no id or label", category.ID)
				}
				for _, entry := range category.Entries {
					entries++
					if !strings.HasPrefix(entry.ID, category.ID+".") || entry.Text == "" || entry.Type == "" {
						t.Errorf("%s: entry %+v", category.ID, entry)
					}
				}
			}
			if entries == 0 {
				t.Error("no entries")
			}
		})
	}
	if !games[extractor.PoE1] || !games[extractor.PoE2] {
		t.Errorf("trade stats published for %v, want both games", games)
	}
}
