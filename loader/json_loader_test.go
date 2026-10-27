package loader

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/testdata"
)

const jsonOnlyVersion = "0.5"

type gameConstant struct {
	ID      string `json:"Id"`
	Key     int    `json:"_key"`
	Value   int    `json:"Value"`
	Divisor int    `json:"Divisor"`
}

func TestLoadJSONTable(t *testing.T) {
	testdata.ServeCommittedData(t)
	ctx := context.Background()

	m, err := LoadManifest(ctx, jsonOnlyVersion)
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, m.JSONOnly)

	typed, err := LoadJSONTable[gameConstant](ctx, jsonOnlyVersion, "GameConstants", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertGreater(t, len(typed), 100)
	for i, row := range typed {
		testza.AssertEqual(t, i, row.Key)
		testza.AssertNotEqual(t, "", row.ID)
	}

	generic, err := LoadJSONTable[map[string]any](ctx, jsonOnlyVersion, "GameConstants", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, len(typed), len(generic))
	testza.AssertEqual(t, typed[0].ID, generic[0]["Id"])

	messages, err := LoadJSONTable[json.RawMessage](ctx, jsonOnlyVersion, "GameConstants", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, len(typed), len(messages))

	// PoE1 versions publish the json too.
	poe1, err := LoadJSONTable[gameConstant](ctx, testdata.LatestDataVersion(), "GameConstants", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertGreater(t, len(poe1), 0)
}

func TestLoadJSONTableNotInManifest(t *testing.T) {
	testdata.ServeCommittedData(t)

	rows, err := LoadJSONTable[gameConstant](context.Background(), jsonOnlyVersion, "NoSuchTable", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, rows)
	testza.AssertEqual(t, 0, len(rows))
}

// missingTable serves the committed data, less one url.
type missingTable struct {
	cdn  http.RoundTripper
	url  string
	hits int
}

func (m *missingTable) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.String() == m.url {
		m.hits++
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewReader(nil)), Header: make(http.Header), Request: req}, nil
	}
	return m.cdn.RoundTrip(req) //nolint:wrapcheck
}

func TestLoadJSONTableListedButMissing(t *testing.T) {
	cdn := testdata.ServeCommittedData(t)
	missing := &missingTable{cdn: cdn, url: JSONTableURL(jsonOnlyVersion, "GameConstants")}
	http.DefaultClient.Transport = missing

	_, err := LoadJSONTable[gameConstant](context.Background(), jsonOnlyVersion, "GameConstants", FakeCache{})
	testza.AssertErrorIs(t, err, ErrAssetNotFound)
	testza.AssertEqual(t, 1, missing.hits)
}

func TestLoadJSONTableNoManifest(t *testing.T) {
	testdata.ServeCommittedData(t)

	_, err := LoadJSONTable[gameConstant](context.Background(), "0.0", "GameConstants", FakeCache{})
	testza.AssertErrorIs(t, err, ErrAssetNotFound)
}

func gzipped(t *testing.T, s string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write([]byte(s))
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, w.Close())
	return buf.Bytes()
}

func TestLoadJSONTableCachedAsset(t *testing.T) {
	healthy := gzipped(t, `[{"_key":0,"Id":"A","Value":1,"Divisor":1},{"_key":1,"Id":"B","Value":2,"Divisor":10}]`)

	cases := []struct {
		name    string
		cached  []byte
		served  []byte
		fetches int
		wantErr bool
	}{
		{name: "healthy cache is used", cached: healthy, fetches: 0},
		{name: "empty cache refetches", cached: []byte{}, served: healthy, fetches: 1},
		{name: "corrupt cache refetches", cached: []byte("not gzip"), served: healthy, fetches: 1},
		{name: "truncated cache refetches", cached: healthy[:len(healthy)-4], served: healthy, fetches: 1},
		{name: "null is not a table", cached: gzipped(t, `null`), served: gzipped(t, `null`), fetches: 1, wantErr: true},
		{name: "unreadable everywhere", cached: []byte("x"), served: []byte("y"), fetches: 1, wantErr: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			transport := &stubTransport{body: c.served}
			serveCDN(t, transport)
			cache := &stubCache{exists: true, get: c.cached}

			rows, err := LoadJSONTable[gameConstant](context.Background(), jsonOnlyVersion, "GameConstants", cache)
			testza.AssertEqual(t, c.fetches, transport.calls)
			if c.wantErr {
				testza.AssertNotNil(t, err)
				return
			}
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, []gameConstant{{Key: 0, ID: "A", Value: 1, Divisor: 1}, {Key: 1, ID: "B", Value: 2, Divisor: 10}}, rows)
			if c.fetches > 0 {
				testza.AssertEqual(t, c.served, cache.stored[JSONTableURL(jsonOnlyVersion, "GameConstants")])
			}
		})
	}
}

// TestLoadJSONTableLive reads a table from data.pob.gg itself and compares it with the committed one.
func TestLoadJSONTableLive(t *testing.T) {
	if testing.Short() {
		t.Skip("reads data.pob.gg")
	}
	ctx := context.Background()

	live, err := LoadJSONTable[gameConstant](ctx, jsonOnlyVersion, "GameConstants", FakeCache{})
	testza.AssertNoError(t, err)

	testdata.ServeCommittedData(t)
	committed, err := LoadJSONTable[gameConstant](ctx, jsonOnlyVersion, "GameConstants", FakeCache{})
	testza.AssertNoError(t, err)

	testza.AssertEqual(t, committed, live)
}
