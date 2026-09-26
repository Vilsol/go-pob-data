package treegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/testdata"
)

// TestGuards checks that game data the generator depends on going missing or renamed is an
// error, not a silently smaller tree.
func TestGuards(t *testing.T) {
	t.Run("selection rules selecting nothing", func(t *testing.T) {
		d := &gameData{}
		testza.AssertNotNil(t, buildClusterNodes(&Tree{}, d, &lookups{}))
		testza.AssertNotNil(t, buildGrantedNodes(&Tree{}, d, &lookups{}))
	})

	short := testdata.ShortGameVersion()
	if _, err := os.Stat(filepath.Join("data", short, "schema.min.json")); err != nil {
		t.Skipf("no committed schema for live version %s", short)
	}
	extractor.LoadParser(short)
	loader, err := extractor.GetBundleLoader(extractor.NewWebFS(testdata.GameVersion()), extractor.IsNewHash(short))
	testza.AssertNoError(t, err)

	t.Run("missing table", func(t *testing.T) {
		_, err := decodeTable[passiveSkillTree](loader, short, "NoSuchTable")
		testza.AssertNotNil(t, err)
	})

	t.Run("missing column", func(t *testing.T) {
		type renamed struct {
			ID string `json:"NoSuchColumn"`
		}
		_, err := decodeTable[renamed](loader, short, "PassiveSkillTrees")
		testza.AssertNotNil(t, err)
	})

	t.Run("missing psg", func(t *testing.T) {
		d := &gameData{trees: []passiveSkillTree{{ID: mainTreeID, PassiveSkillGraph: "Metadata/NoSuchGraph"}}}
		_, err := loadGraph(loader, short, d, mainTreeID)
		testza.AssertNotNil(t, err)
		_, err = loadGraph(loader, short, d, "NoSuchTree")
		testza.AssertNotNil(t, err)
	})

	t.Run("missing atlas image", func(t *testing.T) {
		r, err := newRenderer(loader, short)
		testza.AssertNoError(t, err)
		defer r.close()
		_, err = r.load(source{atlas: "NoSuchAtlasImage"})
		testza.AssertNotNil(t, err)
	})

	t.Run("missing icon", func(t *testing.T) {
		r, err := newRenderer(loader, short)
		testza.AssertNoError(t, err)
		defer r.close()
		_, err = r.load(source{file: "Art/2DArt/SkillIcons/passives/NoSuchIcon.dds"})
		testza.AssertNotNil(t, err)
	})
}
