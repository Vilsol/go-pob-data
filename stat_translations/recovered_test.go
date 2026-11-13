package st

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/MarvinJWendt/testza"
)

// TestRecoveredFiles parses every listed stat description file present under GOPOB_330_SCRATCH, a directory
// holding game files at their game paths (metadata/statdescriptions/*.txt or *.csd).
func TestRecoveredFiles(t *testing.T) {
	dir := os.Getenv("GOPOB_330_SCRATCH")
	if dir == "" {
		t.Skip("GOPOB_330_SCRATCH is not set")
	}
	loader := os.DirFS(dir)
	parsed := 0
	for _, name := range Files {
		p := NewTranslationParser(loader)
		err := p.ParseFile("metadata/statdescriptions/" + name + ".txt")
		if errors.Is(err, fs.ErrNotExist) {
			t.Logf("absent: %s", name)
			continue
		}
		testza.AssertNoError(t, err, name)
		if err == nil {
			testza.AssertGreater(t, len(p.ordered)+len(p.noDescription), 0, name)
			_, err = NewDescriber(loader, "metadata/statdescriptions/"+name+".txt")
			testza.AssertNoError(t, err, name+" with its includes")
			parsed++
		}
	}
	t.Logf("parsed %d of %d listed files", parsed, len(Files))
	testza.AssertGreater(t, parsed, 0)
}
