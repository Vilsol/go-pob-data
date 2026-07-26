package loader

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/raw"
)

// stubCache serves a fixed payload for every key it is told to know about,
// and records what LoadRaw writes back into it.
type stubCache struct {
	stored map[string][]byte
	setErr error
	get    []byte
	exists bool
}

func (s *stubCache) Get(_ string) ([]byte, error) {
	return s.get, nil
}

func (s *stubCache) Set(key string, value []byte) error {
	if s.stored == nil {
		s.stored = map[string][]byte{}
	}
	s.stored[key] = value
	return s.setErr
}

func (s *stubCache) Exists(_ string) bool {
	return s.exists
}

type stubTransport struct {
	body  []byte
	code  int
	calls int
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.calls++
	code := s.code
	if code == 0 {
		code = http.StatusOK
	}
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(bytes.NewReader(s.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

// serveCDN points http.DefaultClient at the given transport for the duration of the test.
func serveCDN(t *testing.T, transport *stubTransport) {
	t.Helper()
	previous := http.DefaultClient.Transport
	http.DefaultClient.Transport = transport
	t.Cleanup(func() {
		http.DefaultClient.Transport = previous
	})
}

// encodeRawPayload produces the exact wire format LoadRaw expects:
// brotli(msgpack(int64 count, elements...)).
func encodeRawPayload(t *testing.T, elements []*raw.CostType) []byte {
	t.Helper()

	var buf bytes.Buffer
	brotliWriter := brotli.NewWriter(&buf)
	writer := msgp.NewWriter(brotliWriter)

	testza.AssertNoError(t, writer.WriteInt64(int64(len(elements))))
	for _, element := range elements {
		testza.AssertNoError(t, element.EncodeMsg(writer))
	}

	testza.AssertNoError(t, writer.Flush())
	testza.AssertNoError(t, brotliWriter.Close())

	return buf.Bytes()
}

func ids(out []*raw.CostType) []string {
	result := make([]string, 0, len(out))
	for _, element := range out {
		result = append(result, element.ID)
	}
	return result
}

func healthyPayload(t *testing.T) []byte {
	t.Helper()
	return encodeRawPayload(t, []*raw.CostType{
		{ID: "Mana", Key: 1},
		{ID: "Life", Key: 2},
	})
}

// TestLoadRawEmptyCachedAssetIsNotSuccess pins the primary contract: a cached asset
// that cannot be read must never come back as a successful, empty table.
func TestLoadRawEmptyCachedAssetIsNotSuccess(t *testing.T) {
	serveCDN(t, &stubTransport{body: nil})

	cache := &stubCache{exists: true, get: []byte{}}

	out, err := LoadRaw[*raw.CostType](context.Background(), "3.18", "CostTypes", nil, cache)
	testza.AssertNotNil(t, err)
	testza.AssertEqual(t, 0, len(out))
}

// TestLoadRawEmptyCachedAssetRefetches proves the self-healing path: a 0-byte cache
// entry is treated as a miss, refetched, and the bad entry overwritten.
func TestLoadRawEmptyCachedAssetRefetches(t *testing.T) {
	payload := healthyPayload(t)
	transport := &stubTransport{body: payload}
	serveCDN(t, transport)

	cache := &stubCache{exists: true, get: []byte{}}

	out, err := LoadRaw[*raw.CostType](context.Background(), "3.18", "CostTypes", nil, cache)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, []string{"Mana", "Life"}, ids(out))
	testza.AssertEqual(t, 1, transport.calls)
	testza.AssertEqual(t, payload, cache.stored["https://data.pob.gg/data/3.18/raw/CostTypes.msgpack.br"])
}

// TestLoadRawCorruptCachedAsset pins that a non-empty but undecodable payload
// already errors, and still does.
func TestLoadRawCorruptCachedAsset(t *testing.T) {
	serveCDN(t, &stubTransport{body: nil})

	cache := &stubCache{exists: true, get: []byte("this is not brotli")}

	out, err := LoadRaw[*raw.CostType](context.Background(), "3.18", "CostTypes", nil, cache)
	testza.AssertNotNil(t, err)
	testza.AssertEqual(t, 0, len(out))
}

// TestLoadRawTruncatedCachedAsset pins the partial-truncation case: a valid prefix
// that runs out mid-element must error rather than return a short table.
func TestLoadRawTruncatedCachedAsset(t *testing.T) {
	serveCDN(t, &stubTransport{body: nil})

	payload := healthyPayload(t)
	cache := &stubCache{exists: true, get: payload[:len(payload)-4]}

	_, err := LoadRaw[*raw.CostType](context.Background(), "3.18", "CostTypes", nil, cache)
	testza.AssertNotNil(t, err)
}

// TestLoadRawHealthyCachedAsset is the regression guard on normal loading.
func TestLoadRawHealthyCachedAsset(t *testing.T) {
	transport := &stubTransport{body: nil}
	serveCDN(t, transport)

	cache := &stubCache{exists: true, get: healthyPayload(t)}

	var initCount int64
	var hooked []string

	out, err := LoadRaw[*raw.CostType](
		context.Background(),
		"3.18",
		"CostTypes",
		func(count int64) { initCount = count },
		cache,
		func(obj *raw.CostType) { hooked = append(hooked, obj.ID) },
	)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, []string{"Mana", "Life"}, ids(out))
	testza.AssertEqual(t, int64(2), initCount)
	testza.AssertEqual(t, []string{"Mana", "Life"}, hooked)
	testza.AssertEqual(t, 0, transport.calls)
}

// TestLoadRawLegitimatelyEmptyTable guards the fix against over-reach: a table that
// really does have zero rows decodes to an empty slice with no error.
func TestLoadRawLegitimatelyEmptyTable(t *testing.T) {
	serveCDN(t, &stubTransport{body: nil})

	cache := &stubCache{exists: true, get: encodeRawPayload(t, nil)}

	out, err := LoadRaw[*raw.CostType](context.Background(), "3.18", "CostTypes", nil, cache)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, 0, len(out))
}

// TestLoadRawFailingCacheWrite pins that caching stays best-effort: a cache that
// cannot store the fetched asset must not fail the load.
func TestLoadRawFailingCacheWrite(t *testing.T) {
	serveCDN(t, &stubTransport{body: healthyPayload(t)})

	cache := &stubCache{exists: false, setErr: errors.New("disk full")}

	out, err := LoadRaw[*raw.CostType](context.Background(), "3.18", "CostTypes", nil, cache)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, []string{"Mana", "Life"}, ids(out))
}

// TestLoadRawNegativeElementCount pins that a decodable but nonsensical header is an
// error rather than a panic in make().
func TestLoadRawNegativeElementCount(t *testing.T) {
	serveCDN(t, &stubTransport{body: nil})

	var buf bytes.Buffer
	brotliWriter := brotli.NewWriter(&buf)
	writer := msgp.NewWriter(brotliWriter)
	testza.AssertNoError(t, writer.WriteInt64(-1))
	testza.AssertNoError(t, writer.Flush())
	testza.AssertNoError(t, brotliWriter.Close())

	cache := &stubCache{exists: true, get: buf.Bytes()}

	_, err := LoadRaw[*raw.CostType](context.Background(), "3.18", "CostTypes", nil, cache)
	testza.AssertNotNil(t, err)
}

// TestLoadRawMissingAsset pins the absent-table policy: a 404 is an empty table, and
// the error body never reaches the cache.
func TestLoadRawMissingAsset(t *testing.T) {
	serveCDN(t, &stubTransport{code: http.StatusNotFound, body: []byte("<html>404</html>")})

	cache := &stubCache{exists: false}

	var initCount int64 = -1

	out, err := LoadRaw[*raw.CostType](context.Background(), "3.29", "SkillTotems", func(count int64) { initCount = count }, cache)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, 0, len(out))
	testza.AssertEqual(t, int64(0), initCount)
	testza.AssertEqual(t, 0, len(cache.stored))
}

// TestLoadRawServerError pins that a 5xx body is an error and is never cached.
func TestLoadRawServerError(t *testing.T) {
	serveCDN(t, &stubTransport{code: http.StatusInternalServerError, body: []byte("<html>500</html>")})

	cache := &stubCache{exists: false}

	_, err := LoadRaw[*raw.CostType](context.Background(), "3.18", "CostTypes", nil, cache)
	testza.AssertNotNil(t, err)
	testza.AssertEqual(t, 0, len(cache.stored))
}

func encodeTranslationPayload(t *testing.T, file *raw.TranslationFile) []byte {
	t.Helper()

	var buf bytes.Buffer
	brotliWriter := brotli.NewWriter(&buf)
	writer := msgp.NewWriter(brotliWriter)

	testza.AssertNoError(t, file.EncodeMsg(writer))
	testza.AssertNoError(t, writer.Flush())
	testza.AssertNoError(t, brotliWriter.Close())

	return buf.Bytes()
}

func healthyTranslation(t *testing.T) []byte {
	t.Helper()
	return encodeTranslationPayload(t, &raw.TranslationFile{Includes: []string{"other.txt"}})
}

// TestLoadTranslationEmptyCachedAssetRefetches mirrors the raw loader: an unreadable
// cache entry heals rather than bricking the install.
func TestLoadTranslationEmptyCachedAssetRefetches(t *testing.T) {
	payload := healthyTranslation(t)
	transport := &stubTransport{body: payload}
	serveCDN(t, transport)

	cache := &stubCache{exists: true, get: []byte{}}

	out, err := LoadTranslation(context.Background(), "3.18", "en", "stat_descriptions", cache)
	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, out)
	testza.AssertEqual(t, []string{"other.txt"}, out.Includes)
	testza.AssertEqual(t, 1, transport.calls)
}

// TestLoadTranslationUnreadableEverywhere pins that an unreadable asset is an error,
// not an empty translation file.
func TestLoadTranslationUnreadableEverywhere(t *testing.T) {
	serveCDN(t, &stubTransport{body: nil})

	cache := &stubCache{exists: true, get: []byte("this is not brotli")}

	_, err := LoadTranslation(context.Background(), "3.18", "en", "stat_descriptions", cache)
	testza.AssertNotNil(t, err)
}

// TestLoadTranslationHealthyCachedAsset is the regression guard on normal loading.
func TestLoadTranslationHealthyCachedAsset(t *testing.T) {
	transport := &stubTransport{body: nil}
	serveCDN(t, transport)

	cache := &stubCache{exists: true, get: healthyTranslation(t)}

	out, err := LoadTranslation(context.Background(), "3.18", "en", "stat_descriptions", cache)
	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, out)
	testza.AssertEqual(t, []string{"other.txt"}, out.Includes)
	testza.AssertEqual(t, 0, transport.calls)
}
