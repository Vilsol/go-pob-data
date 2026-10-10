package loader

import (
	"bytes"
	"context"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/testdata"
)

func TestFileLoader(t *testing.T) {
	testdata.ServeCommittedData(t)
	ctx := context.Background()
	version := testdata.LatestDataVersion()

	character, err := LoadFile(ctx, version, "Metadata/Characters/Character.ot", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, bytes.Contains(character, []byte("\tlife_per_level = 12")), "decoded to UTF-8")

	images, err := LoadFile(ctx, version, "art/uiimages1.txt", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, bytes.HasPrefix(images, []byte(`"Art/2DArt/UIImages/`)), "no byte order mark")

	_, err = LoadFile(ctx, "3.28", "Metadata/Characters/Character.ot", FakeCache{})
	testza.AssertErrorIs(t, err, ErrAssetNotFound)
}
