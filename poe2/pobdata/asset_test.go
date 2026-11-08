package pobdata

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/testdata"
)

func decodeAll[T any, PT interface {
	*T
	msgp.Decodable
}](t *testing.T, encoded []byte) []T {
	t.Helper()
	r := msgp.NewReader(bytes.NewReader(encoded))
	n, err := r.ReadInt64()
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]T, n)
	for i := range rows {
		if err := PT(&rows[i]).DecodeMsg(r); err != nil {
			t.Fatalf("row %d: %v", i, err)
		}
	}
	if _, err := r.R.Peek(1); !errors.Is(err, io.EOF) {
		t.Fatalf("trailing bytes after %d rows", n)
	}
	return rows
}

func reencode[T any, PT interface {
	*T
	msgp.Decodable
	msgp.Encodable
}](t *testing.T, encoded []byte) ([]byte, int) {
	t.Helper()
	rows := decodeAll[T, PT](t, encoded)
	out, err := Encode[T, PT](rows)
	if err != nil {
		t.Fatal(err)
	}
	return out, len(rows)
}

// Every published asset decodes as its table's rows, re-encodes to the same bytes, and is named by
// their hash.
func TestAssetsDecodeAndAreNamedByContent(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "data", "*", "pob", "*.msgpack.br"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no published assets")
	}

	seen := map[string]bool{}
	for _, path := range paths {
		compressed, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.HasPrefix(compressed, []byte("version https://git-lfs")) {
			t.Fatalf("%s is a git-lfs pointer; run git lfs pull", path)
		}
		encoded, err := io.ReadAll(brotli.NewReader(bytes.NewReader(compressed)))
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}

		table, _, _ := strings.Cut(filepath.Base(path), ".")
		var again []byte
		var n int
		switch table {
		case ItemBasesTable:
			again, n = reencode[ItemBase](t, encoded)
		case ItemModsTable:
			again, n = reencode[ItemMod](t, encoded)
		case ExclusiveModsTable:
			again, n = reencode[ExclusiveMod](t, encoded)
		case ModScalabilityTable:
			again, n = reencode[ModScalability](t, encoded)
		default:
			t.Errorf("%s: no table %s", path, table)
			continue
		}
		seen[table] = true

		if n == 0 {
			t.Errorf("%s holds no rows", path)
		}
		if !bytes.Equal(again, encoded) {
			t.Errorf("%s does not re-encode to its own bytes", path)
		}
		version := filepath.Base(filepath.Dir(filepath.Dir(path)))
		if want := filepath.Join("..", "..", "data", version, filepath.FromSlash(AssetPath(table, encoded))); path != want {
			t.Errorf("%s is named for other contents; its contents are %s", path, want)
		}
	}

	for _, table := range []string{ItemBasesTable, ItemModsTable, ExclusiveModsTable, ModScalabilityTable} {
		if !seen[table] {
			t.Errorf("no asset of %s", table)
		}
	}
}

func TestEncodeRoundTrip(t *testing.T) {
	rows := []ItemBase{
		{Name: "a", ImplicitModTypes: [][]string{nil, {"x"}}, Armour: &BaseArmour{Armour: 1}},
		{Name: "b", Charm: &BaseCharm{Buff: []string{"y"}}, Req: BaseReq{Level: 2}},
	}
	encoded, err := Encode(rows)
	if err != nil {
		t.Fatal(err)
	}
	if got := decodeAll[ItemBase](t, encoded); !reflect.DeepEqual(got, rows) {
		t.Errorf("round trip: got %+v, want %+v", got, rows)
	}
}

type noCache struct{}

func (noCache) Get(string) ([]byte, error) { return nil, nil }
func (noCache) Set(string, []byte) error   { return nil }
func (noCache) Exists(string) bool         { return false }

// Load reads a published asset by its path, and a path nothing is published at is ErrAssetNotFound.
func TestLoadReadsPublishedAsset(t *testing.T) {
	testdata.ServeCommittedData(t)
	paths, err := filepath.Glob(filepath.Join("..", "..", "data", "0.5", "pob", ItemBasesTable+".*.msgpack.br"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no published %s: %v", ItemBasesTable, err)
	}
	for _, path := range paths {
		rows, err := Load[ItemBase](context.Background(), "0.5", "pob/"+filepath.Base(path), noCache{})
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) == 0 || rows[0].Name == "" {
			t.Errorf("%s loads %d rows, the first %+v", path, len(rows), rows)
		}
	}

	_, err = Load[ItemBase](context.Background(), "0.5", loader.AssetPath("pob/"+ItemBasesTable, "00000000"), noCache{})
	if !errors.Is(err, loader.ErrAssetNotFound) {
		t.Errorf("a missing asset loads with %v, want ErrAssetNotFound", err)
	}
}
