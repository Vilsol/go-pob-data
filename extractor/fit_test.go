package extractor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/MarvinJWendt/testza"
)

type fitRow struct {
	id    string
	name  string
	tags  []string
	key   uint64
	value int32
}

// fitDat writes rows as the game would: Id, Value, Name, Key, Tags in a datc64 row, and the variable data
// in row then column order, a string list's strings just before the list.
func fitDat(rows []fitRow) []byte {
	var fixed bytes.Buffer
	dyn := bytes.NewBuffer(append([]byte(nil), datBoundary...))
	u64 := func(b *bytes.Buffer, v uint64) { _ = binary.Write(b, binary.LittleEndian, v) }
	str := func(s string) uint64 {
		at := uint64(dyn.Len())
		for _, c := range utf16.Encode([]rune(s)) {
			_ = binary.Write(dyn, binary.LittleEndian, c)
		}
		dyn.Write([]byte{0, 0, 0, 0})
		return at
	}

	for _, r := range rows {
		u64(&fixed, str(r.id))
		_ = binary.Write(&fixed, binary.LittleEndian, r.value)
		u64(&fixed, str(r.name))
		u64(&fixed, r.key)
		u64(&fixed, 0)
		offsets := make([]uint64, len(r.tags))
		for i, t := range r.tags {
			offsets[i] = str(t)
		}
		u64(&fixed, uint64(len(r.tags)))
		u64(&fixed, uint64(dyn.Len()))
		for _, o := range offsets {
			u64(dyn, o)
		}
	}

	out := binary.LittleEndian.AppendUint32(nil, uint32(len(rows)))
	out = append(out, fixed.Bytes()...)
	return append(out, dyn.Bytes()...)
}

func fitRows(n int) []fitRow {
	rows := make([]fitRow, n)
	for i := range rows {
		rows[i] = fitRow{
			id:    fmt.Sprintf("Row%d", i),
			value: int32(i * 7),
			name:  fmt.Sprintf("Name of row %d", i),
			key:   uint64(i % 5),
			tags:  []string{fmt.Sprintf("tag%d", i%3), fmt.Sprintf("other%d", i)},
		}
	}
	return rows
}

func fitColumnOf(name string, typ Type, array bool) Column {
	return Column{Name: &name, Type: typ, Array: array}
}

func withFitSchema(t *testing.T, columns ...Column) {
	t.Helper()
	prev, ok := tableMap["FitTest"]
	if tableMap == nil {
		tableMap = map[string]Table{}
	}
	tableMap["FitTest"] = Table{Name: "FitTest", Columns: columns}
	t.Cleanup(func() {
		if ok {
			tableMap["FitTest"] = prev
		} else {
			delete(tableMap, "FitTest")
		}
	})
}

var (
	fitID    = fitColumnOf("Id", TypeString, false)
	fitValue = fitColumnOf("Value", TypeI32, false)
	fitName  = fitColumnOf("Name", TypeString, false)
	fitKey   = fitColumnOf("Key", TypeForeignRow, false)
	fitTags  = fitColumnOf("Tags", TypeString, true)
)

func TestFitMatchingSchema(t *testing.T) {
	withFitSchema(t, fitID, fitValue, fitName, fitKey, fitTags)
	b := fitDat(fitRows(100))
	testza.AssertNoError(t, checkRowWidth(b, "FitTest.datc64"))
	testza.AssertNoError(t, checkFit(b, "FitTest.datc64"))
}

func TestFitRepeatedStrings(t *testing.T) {
	withFitSchema(t, fitID, fitValue, fitName, fitKey, fitTags)
	rows := fitRows(50)
	b := fitDat(rows)

	// Point every row's Name at the first row's: a repeated value refers back to its first copy.
	first := binary.LittleEndian.Uint64(b[4+12:])
	for i := range rows {
		binary.LittleEndian.PutUint64(b[4+i*52+12:], first)
	}
	testza.AssertNoError(t, checkFit(b, "FitTest.datc64"))
}

func TestFitShiftedColumns(t *testing.T) {
	// Value moved to the front: the row width is unchanged, every column before it reads 4 bytes off.
	withFitSchema(t, fitValue, fitID, fitName, fitKey, fitTags)
	b := fitDat(fitRows(100))
	testza.AssertNoError(t, checkRowWidth(b, "FitTest.datc64"))

	err := checkFit(b, "FitTest.datc64")
	testza.AssertNotNil(t, err)
	testza.AssertContains(t, err.Error(), "table FitTest")
	testza.AssertContains(t, err.Error(), "Id (")
}

func TestFitRetypedKey(t *testing.T) {
	// A foreign row read as two strings: its row index and zero high half are no string offsets.
	withFitSchema(t, fitID, fitValue, fitName, fitColumnOf("KeyA", TypeString, false), fitColumnOf("KeyB", TypeString, false), fitTags)
	b := fitDat(fitRows(100))
	testza.AssertNoError(t, checkRowWidth(b, "FitTest.datc64"))

	err := checkFit(b, "FitTest.datc64")
	testza.AssertNotNil(t, err)
	testza.AssertTrue(t, strings.Contains(err.Error(), "KeyA (") && strings.Contains(err.Error(), "KeyB ("))
}

func TestFitFewBadRows(t *testing.T) {
	withFitSchema(t, fitID, fitValue, fitName, fitKey, fitTags)
	rows := fitRows(100)
	b := fitDat(rows)

	// A key no table could hold in a tenth of the rows is noise, not a moved column.
	for i := 0; i < len(rows); i += 10 {
		binary.LittleEndian.PutUint64(b[4+i*52+20:], maxPlausibleRows)
	}
	testza.AssertNoError(t, checkFit(b, "FitTest.datc64"))
	testza.AssertGreater(t, fitShares(b, "FitTest.datc64")["Key"], 0.0)
}

func TestFitStrayOffsets(t *testing.T) {
	withFitSchema(t, fitID, fitValue, fitName, fitKey, fitTags)
	rows := fitRows(100)
	b := fitDat(rows)

	// Name points two bytes into its row's Id: still a readable string, but one nothing ever wrote there.
	for i := range rows {
		id := binary.LittleEndian.Uint64(b[4+i*52:])
		binary.LittleEndian.PutUint64(b[4+i*52+12:], id+2)
	}
	err := checkFit(b, "FitTest.datc64")
	testza.AssertNotNil(t, err)
	testza.AssertContains(t, err.Error(), "Name (")
}
