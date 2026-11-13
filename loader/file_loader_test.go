package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/testdata"
)

func TestFileLoader(t *testing.T) {
	testdata.ServeCommittedData(t)
	ctx := context.Background()
	version := testdata.LatestDataVersion()

	ot, err := LoadFileBundle(ctx, version, ".ot", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, bytes.Contains(ot["Metadata/Characters/Character.ot"], []byte("\tlife_per_level = 12")), "decoded to UTF-8")

	txt, err := LoadFileBundle(ctx, version, ".txt", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, bytes.HasPrefix(txt["art/uiimages1.txt"], []byte(`"Art/2DArt/UIImages/`)), "no byte order mark")

	stats, err := LoadFile(ctx, version, "api/trade/data/stats.json", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, json.Valid(stats))

	_, err = LoadFile(ctx, version, "Metadata/Characters/Character.ot", FakeCache{})
	testza.AssertErrorIs(t, err, ErrAssetNotFound, "bundled files are not published one at a time")
}
