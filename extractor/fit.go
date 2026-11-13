package extractor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"path"
	"sort"
	"strings"

	"github.com/oriath-net/pogo/dat"
)

const (
	// fitSample rows are read for the per-value checks.
	fitSample = 40
	// fitBroken is the share of rows the worst column may misread before the table is refused.
	fitBroken = 0.25
	// orderMinValues is the fewest values a column must read for its write order to count.
	orderMinValues = 8
	// maxPlausibleRows: no table has this many rows, so a key at or past it is a misread.
	maxPlausibleRows = 5_000_000
	maxPlausibleList = 100_000
	nullKey64        = 0xfefe_fefe_fefe_fefe
	nullKey32        = 0xfefe_fefe
)

type fitColumn struct {
	name   string
	typ    dat.FieldType
	offset int
}

// datLayout is a dat file cut into its fixed rows and its variable data, which starts at the 8-byte boundary
// marker that string and list offsets count from.
type datLayout struct {
	rows     []byte
	dyn      []byte
	rowCount int
	rowWidth int
	ptr      int
}

func (l *datLayout) row(i int) []byte {
	return l.rows[i*l.rowWidth : (i+1)*l.rowWidth]
}

func (l *datLayout) word(b []byte) uint64 {
	if l.ptr == 8 {
		return binary.LittleEndian.Uint64(b)
	}
	return uint64(binary.LittleEndian.Uint32(b))
}

// inDyn reports whether offset p points at a byte of the variable data past the marker.
func (l *datLayout) inDyn(p uint64) bool {
	return p >= 8 && p < uint64(len(l.dyn))
}

// stringEnd is where the string at p ends, past its terminator; false when it runs off the data.
func (l *datLayout) stringEnd(p uint64) (uint64, bool) {
	if !l.inDyn(p) {
		return 0, false
	}
	for i := p; i+1 < uint64(len(l.dyn)); i += 2 {
		if l.dyn[i] == 0 && l.dyn[i+1] == 0 {
			if i+3 < uint64(len(l.dyn)) && l.dyn[i+2] == 0 && l.dyn[i+3] == 0 {
				return i + 4, true
			}
			return i + 2, true
		}
	}
	return 0, false
}

// listSpan is the byte span of count elements of size at p; false when it overflows or runs off the data.
func (l *datLayout) listSpan(count uint64, p uint64, size int) (uint64, bool) {
	if count > maxPlausibleList || !l.inDyn(p) {
		return 0, false
	}
	end := p + count*uint64(size)
	return end, end <= uint64(len(l.dyn))
}

// elemSize is the size of one element of list type t, or 0 when the element type is unknown (void[]).
func elemSize(t dat.FieldType, ptr int) int {
	switch t {
	case dat.TypeListUint8, dat.TypeListBool:
		return 1
	case dat.TypeListUint16:
		return 2
	case dat.TypeListUint32, dat.TypeListInt32, dat.TypeListFloat32:
		return 4
	case dat.TypeListUint64, dat.TypeListInt64, dat.TypeListFloat64:
		return 8
	case dat.TypeListShortID, dat.TypeListString:
		return ptr
	case dat.TypeListLongID:
		return 2 * ptr
	}
	return 0
}

func isList(t dat.FieldType) bool {
	return strings.HasSuffix(string(t), "[]")
}

// valuePossible reports whether the fixed bytes b of a column of type t can hold a real value.
func (l *datLayout) valuePossible(t dat.FieldType, b []byte) bool {
	switch {
	case t == dat.TypeBool:
		return b[0] <= 1
	case t == dat.TypeString:
		_, ok := l.stringEnd(l.word(b))
		return ok
	case t == dat.TypeShortID:
		v := l.word(b)
		return v < maxPlausibleRows || (l.ptr == 8 && v == nullKey64) || (l.ptr == 4 && v == nullKey32)
	case t == dat.TypeLongID:
		lo := binary.LittleEndian.Uint64(b)
		if l.ptr == 4 {
			return lo < maxPlausibleRows || lo == nullKey64
		}
		hi := binary.LittleEndian.Uint64(b[8:])
		return (lo == nullKey64 && hi == nullKey64) || (hi == 0 && lo < maxPlausibleRows)
	case isList(t):
		count, p := l.word(b), l.word(b[l.ptr:])
		if count == 0 {
			return true
		}
		size := elemSize(t, l.ptr)
		if size == 0 {
			return count <= maxPlausibleList && l.inDyn(p)
		}
		end, ok := l.listSpan(count, p, size)
		if !ok || t != dat.TypeListString {
			return ok
		}
		for at := p; at < end; at += uint64(l.ptr) {
			if _, ok := l.stringEnd(l.word(l.dyn[at:])); !ok {
				return false
			}
		}
		return true
	}
	return true
}

// writeOrder follows how the game writes the variable data: each new value where the last one ended, in row
// then column order (a string list's strings just before the list), and a repeated value pointing back at its
// first copy. A column read at the wrong offset points elsewhere: into the middle of a value or off the data.
type writeOrder struct {
	seen      map[uint64]bool
	read      map[string]int
	stray     map[string]int
	writeAt   uint64
	openEnded bool
}

// visit records a value at start ending at end; known is false for a list of unknown element type, which ends
// wherever the next value starts, and ok is false when the value cannot be read at all.
func (w *writeOrder) visit(l *datLayout, column string, start, end uint64, known, ok bool) {
	if known && ok && end == start {
		// An empty list writes nothing, yet still points at the write position or an earlier value.
		if start == w.writeAt || (w.openEnded && start > w.writeAt) {
			w.seen[start] = true
		}
		return
	}

	w.read[column]++
	readable := ok && l.inDyn(start) && (!known || (end > start && l.inDyn(end-1)))
	switch {
	case readable && start >= w.writeAt:
		w.seen[start] = true
		if known {
			w.writeAt, w.openEnded = end, false
		} else {
			w.writeAt, w.openEnded = start, true
		}
	case !readable || !w.seen[start]:
		w.stray[column]++
	}
}

func (w *writeOrder) column(l *datLayout, c fitColumn, b []byte) {
	if c.typ == dat.TypeString {
		start := l.word(b)
		end, ok := l.stringEnd(start)
		w.visit(l, c.name, start, end, true, ok)
		return
	}

	count, start := l.word(b), l.word(b[l.ptr:])
	if count == 0 {
		w.visit(l, c.name, start, start, true, true)
		return
	}
	if count > maxPlausibleList {
		w.visit(l, c.name, math.MaxUint64, 0, true, false)
		return
	}
	size := elemSize(c.typ, l.ptr)
	if c.typ == dat.TypeListString {
		for k := uint64(0); k < count; k++ {
			at := start + k*uint64(l.ptr)
			if !l.inDyn(at) || at+uint64(l.ptr) > uint64(len(l.dyn)) {
				break
			}
			elem := l.word(l.dyn[at:])
			end, ok := l.stringEnd(elem)
			w.visit(l, c.name, elem, end, true, ok)
		}
	}
	if size == 0 {
		w.visit(l, c.name, start, 0, false, true)
		return
	}
	end, ok := l.listSpan(count, start, size)
	w.visit(l, c.name, start, end, true, ok)
}

func is64Bit(filename string) bool {
	ext := path.Ext(filename)
	return ext == ".dat64" || ext == ".datc64"
}

// fitColumns lays out table name's fields at the width of filename's dat variant.
func fitColumns(name string, filename string) []fitColumn {
	width := dat.Dat32
	if is64Bit(filename) {
		width = dat.Datc64
	}

	fields := tableMap[name].ToJSONFormat().Fields
	out := make([]fitColumn, 0, len(fields))
	off := 0
	for _, f := range fields {
		ft := dat.FieldType(f.Type)
		if !ft.Valid() {
			continue
		}
		out = append(out, fitColumn{name: f.Name, typ: ft, offset: off})
		off += ft.Size(width)
	}
	return out
}

// checkFit refuses a table whose schema adds up to the row width but no longer describes the rows: a shifted
// or retyped column reads string, list and key values that cannot be real, which pogo decodes silently.
func checkFit(b []byte, filename string) error {
	name := strings.Split(path.Base(filename), ".")[0]
	return fitError(name, fitShares(b, filename))
}

// fitShares is the share of rows each column misreads: sampled rows whose value cannot be real, or values
// written out of order, whichever is worse.
func fitShares(b []byte, filename string) map[string]float64 {
	name := strings.Split(path.Base(filename), ".")[0]
	share := map[string]float64{}
	rowCount := int(int32(binary.LittleEndian.Uint32(b)))
	if rowCount <= 0 {
		return share
	}

	boundary := bytes.Index(b, datBoundary)
	ptr := 4
	if is64Bit(filename) {
		ptr = 8
	}
	l := &datLayout{
		rows:     b[4:boundary],
		dyn:      b[boundary:],
		rowCount: rowCount,
		rowWidth: (boundary - 4) / rowCount,
		ptr:      ptr,
	}
	columns := fitColumns(name, filename)

	step := max(rowCount/fitSample, 1)
	sampled := 0
	bad := map[string]int{}
	for i := 0; i < rowCount && sampled < fitSample; i += step {
		sampled++
		row := l.row(i)
		for _, c := range columns {
			if !l.valuePossible(c.typ, row[c.offset:]) {
				bad[c.name]++
			}
		}
	}
	for c, n := range bad {
		share[c] = float64(n) / float64(sampled)
	}

	w := &writeOrder{seen: map[uint64]bool{}, read: map[string]int{}, stray: map[string]int{}, writeAt: 8}
	for i := range rowCount {
		row := l.row(i)
		for _, c := range columns {
			if c.typ == dat.TypeString || isList(c.typ) {
				w.column(l, c, row[c.offset:])
			}
		}
	}
	for c, n := range w.stray {
		if read := w.read[c]; read >= orderMinValues {
			share[c] = max(share[c], float64(n)/float64(read))
		}
	}

	return share
}

// fitError names the columns that misread, worst first, when the worst misreads at least fitBroken of its rows.
func fitError(name string, share map[string]float64) error {
	worst := make([]string, 0, len(share))
	for c, s := range share {
		if s > 0 {
			worst = append(worst, c)
		}
	}
	sort.Slice(worst, func(i, j int) bool {
		if share[worst[i]] != share[worst[j]] {
			return share[worst[i]] > share[worst[j]]
		}
		return worst[i] < worst[j]
	})
	if len(worst) == 0 || share[worst[0]] < fitBroken {
		return nil
	}

	parts := make([]string, 0, 4)
	for _, c := range worst[:min(len(worst), 4)] {
		parts = append(parts, fmt.Sprintf("%s (%.0f%% of rows)", c, share[c]*100))
	}
	return fmt.Errorf(
		"table %s: schema from %s no longer describes the rows, %d columns read impossible values: %s",
		name, tableSource[name], len(worst), strings.Join(parts, ", "),
	)
}
