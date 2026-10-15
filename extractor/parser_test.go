package extractor

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/oriath-net/pogo/dat"
)

func useWidthFixture(t *testing.T) {
	t.Helper()

	prevMap, prevSource, prevParser := tableMap, tableSource, parser
	t.Cleanup(func() {
		tableMap, tableSource, parser = prevMap, prevSource, prevParser
	})

	tableMap = map[string]Table{
		"Fixture": {Name: "Fixture", Columns: []Column{{Type: TypeI32}, {Type: TypeBool}}},
	}
	tableSource = map[string]string{"Fixture": "fixture-schema"}
	parser = dat.InitParser("3.29", &schemaFS{})
}

// datFile is rowCount, then fixed zero bytes, then the boundary.
func datFile(rowCount int32, fixed int) []byte {
	b := binary.LittleEndian.AppendUint32(nil, uint32(rowCount))
	b = append(b, make([]byte, fixed)...)
	return append(b, datBoundary...)
}

func TestParseDatWidthFits(t *testing.T) {
	useWidthFixture(t)

	rows, err := ParseDat(bytes.NewReader(datFile(3, 15)), "Fixture.datc64")
	testza.AssertNoError(t, err)
	testza.AssertLen(t, rows, 3)
}

func TestParseDatWidthMismatch(t *testing.T) {
	useWidthFixture(t)

	for _, tc := range []struct {
		name  string
		width string
		file  []byte
	}{
		{"long", "data row width 6", datFile(3, 18)},
		{"short", "data row width 4", datFile(3, 12)},
		{"fractional", "data row width 5 (11 fixed bytes / 2 rows)", datFile(2, 11)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDat(bytes.NewReader(tc.file), "Fixture.datc64")
			testza.AssertNotNil(t, err)
			for _, want := range []string{"table Fixture", "schema row width 5", tc.width, "fixture-schema"} {
				testza.AssertTrue(t, strings.Contains(err.Error(), want), err.Error())
			}
		})
	}
}

func TestParseDatUnknownTable(t *testing.T) {
	useWidthFixture(t)

	_, err := ParseDat(bytes.NewReader(datFile(1, 5)), "NoSuchTable.datc64")
	testza.AssertNotNil(t, err)
	testza.AssertTrue(t, strings.Contains(err.Error(), "schema row width 0"), err.Error())
}
