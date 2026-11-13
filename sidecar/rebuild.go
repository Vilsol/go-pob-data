package sidecar

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"path"
	"slices"
	"sort"
	"strings"
	"unicode/utf16"
)

// Column is one column of a table's row in the reference patch's schema.
type Column struct {
	Type   string // pogo type: "string", "longid", ...
	Ref    string // referenced table, for longid
	Offset int
	Size   int
}

// Options carries what the rebuild cannot know from the bundles alone.
type Options struct {
	// Columns lays out a reference table's row by its lower-case name; nil when unknown.
	Columns func(table string) []Column
	// Check validates a non-table file's bytes (e.g. parses a .psg); nil error means it parsed.
	Check func(path string, b []byte) error
	Logf  func(format string, args ...any)
	// MovedText picks the reference UTF-16 files to look for in pieces no neighbour placed: files that moved
	// to another bundle.
	MovedText func(path string) bool
}

// BundleSource returns a decompressed bundle by name, or an error wrapping fs.ErrNotExist.
type BundleSource func(name string) ([]byte, error)

var marker = bytes.Repeat([]byte{0xbb}, 8)

const nullRow = 0xfefefefefefefefe

// refTable is one table file of the reference patch; one record covers every path that shared its bytes.
type refTable struct {
	paths  []string
	stem   string
	bundle string
	off    int64
	data   []byte
	rows   int
	width  int
	keyOff int
	keys   map[string]struct{}
	cols   []Column
	// utf32 marks a .datcl64 table: its strings are UTF-32 where .datc64's are UTF-16
	utf32 bool
}

type candidate struct {
	start, rows, width int
}

type pair struct {
	bundle int
	k      int
	cand   candidate
	ref    *refTable
	score  float64
	agree  float64
	keyed  bool
	slot   bool
}

type datTable struct {
	start  int
	end    int
	rows   int
	width  int
	ref    *refTable
	agree  float64
	keyed  bool
	slot   bool
	marker int
}

type target struct {
	name    string
	blob    []byte
	marks   []int
	cands   [][]candidate
	tables  []datTable
	entries []Entry
}

// Rebuild names the files in the target bundles. Tables are found one per 0xBB*8 marker, their starts by the
// row-count identity (u32 rows × width ends at the marker), and named by matching against the reference
// patch's tables on Id strings, width and row count. Other files are named when their reference bytes appear
// unchanged, or when they are the only file between two placed neighbours in the reference order.
func Rebuild(tgt, ref BundleSource, refFiles []RefFile, targetBundles []string, opt Options) ([]Entry, []Span, error) {
	logf := opt.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}

	refBlobs := map[string][]byte{}
	refBlob := func(name string) []byte {
		if b, ok := refBlobs[name]; ok {
			return b
		}
		b, err := ref(name)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				logf("reference bundle %s: %v", name, err)
			}
			b = nil
		}
		refBlobs[name] = b
		return b
	}

	// reference tables, one record per (bundle, offset)
	type key struct {
		bundle string
		off    int64
	}
	recs := map[key]*refTable{}
	for _, f := range refFiles {
		if !isTablePath(f.Path) {
			continue
		}
		k := key{f.Bundle, f.Offset}
		if r, ok := recs[k]; ok {
			if !isLanguagePath(f.Path) {
				r.paths = append(r.paths, f.Path)
			}
			continue
		}
		blob := refBlob(f.Bundle)
		if blob == nil || f.Offset+f.Size > int64(len(blob)) {
			continue
		}
		r := newRefTable(blob[f.Offset:f.Offset+f.Size], f)
		if r != nil {
			recs[k] = r
		}
	}
	var pool []*refTable
	for _, r := range recs {
		slices.Sort(r.paths)
		r.loadKeys(opt.Columns)
		pool = append(pool, r)
	}
	sort.Slice(pool, func(i, j int) bool { return pool[i].paths[0] < pool[j].paths[0] })
	byWidth := map[int][]*refTable{}
	byRows := slices.Clone(pool)
	sort.Slice(byRows, func(i, j int) bool { return byRows[i].rows < byRows[j].rows })
	for _, r := range pool {
		byWidth[r.width] = append(byWidth[r.width], r)
	}
	logf("reference: %d tables", len(pool))

	var targets []*target
	for _, name := range targetBundles {
		blob, err := tgt(name)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, nil, fmt.Errorf("target bundle %s: %w", name, err)
			}
			continue
		}
		t := &target{name: name, blob: blob, marks: markers(blob)}
		t.cands = make([][]candidate, len(t.marks))
		for k := range t.marks {
			lo := 0
			if k > 0 {
				lo = t.marks[k-1] + 8
			}
			t.cands[k] = startCandidates(blob, lo, t.marks[k])
		}
		targets = append(targets, t)
	}

	// score every (marker, candidate start, reference table) that could be the same table
	var pairs []pair
	for bi, t := range targets {
		for k, cs := range t.cands {
			for _, c := range cs {
				if c.rows == 0 {
					for _, r := range byWidth[0] {
						if baseBundle(r.bundle) == baseBundle(t.name) {
							pairs = append(pairs, pair{bundle: bi, k: k, cand: c, ref: r, score: 1})
						}
					}
					continue
				}
				seen := map[*refTable]bool{}
				for _, r := range byWidth[c.width] {
					seen[r] = true
					if p, ok := scorePair(t, k, c, r); ok {
						p.bundle = bi
						pairs = append(pairs, p)
					}
				}
				if c.rows < 8 {
					continue
				}
				// the same table with a changed row width: only Id strings can say so
				lo := sort.Search(len(byRows), func(i int) bool { return byRows[i].rows >= c.rows*3/4-4 })
				for i := lo; i < len(byRows) && byRows[i].rows <= c.rows*4/3+4; i++ {
					r := byRows[i]
					if seen[r] || r.keyOff < 0 || r.keyOff+8 > c.width || r.rows < 16 || c.width*2 < r.width || c.width > 2*r.width+64 {
						continue
					}
					agree := keyAgreement(t.blob, t.marks[k], c, r)
					if agree >= 0.6 {
						pairs = append(pairs, pair{bundle: bi, k: k, cand: c, ref: r, score: 6*agree + rowRatio(c.rows, r.rows), agree: agree, keyed: true})
					}
				}
			}
		}
	}
	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].score > pairs[j].score })

	type mk struct{ b, k int }
	takenMark := map[mk]bool{}
	takenRef := map[*refTable]bool{}
	assigned := map[mk]pair{}
	for _, p := range pairs {
		m := mk{p.bundle, p.k}
		if takenMark[m] || takenRef[p.ref] {
			continue
		}
		takenMark[m], takenRef[p.ref] = true, true
		assigned[m] = p
	}

	// a table whose width changed and that has no Id to match: the one reference table left between two
	// matched neighbours of the same bundle, with nearly the same row count
	inBundle := map[string][]*refTable{}
	for _, r := range pool {
		inBundle[r.bundle] = append(inBundle[r.bundle], r)
	}
	for _, rs := range inBundle {
		sort.Slice(rs, func(i, j int) bool { return rs[i].off < rs[j].off })
	}
	pos := func(r *refTable) int {
		rs := inBundle[r.bundle]
		return sort.Search(len(rs), func(i int) bool { return rs[i].off >= r.off })
	}
	for bi, t := range targets {
		for k := 1; k+1 < len(t.marks); k++ {
			if takenMark[mk{bi, k}] {
				continue
			}
			prev, ok1 := assigned[mk{bi, k - 1}]
			next, ok2 := assigned[mk{bi, k + 1}]
			if !ok1 || !ok2 || prev.ref.bundle != next.ref.bundle || pos(next.ref)-pos(prev.ref) != 2 {
				continue
			}
			r := inBundle[prev.ref.bundle][pos(prev.ref)+1]
			if takenRef[r] || r.rows < 2 {
				continue
			}
			var best *candidate
			for _, c := range t.cands[k] {
				if rowRatio(c.rows, r.rows) >= 0.9 && (best == nil || math.Abs(float64(c.rows-r.rows)) < math.Abs(float64(best.rows-r.rows))) {
					c := c
					best = &c
				}
			}
			if best == nil {
				continue
			}
			takenMark[mk{bi, k}], takenRef[r] = true, true
			assigned[mk{bi, k}] = pair{bundle: bi, k: k, cand: *best, ref: r, slot: true}
		}
	}

	// place every table: assigned starts, else the most row-like candidate
	recovered := map[string]int{} // table stem -> rows, for foreign key bounds
	for bi, t := range targets {
		t.tables = make([]datTable, len(t.marks))
		for k := range t.marks {
			dt := datTable{start: -1, marker: t.marks[k]}
			if p, ok := assigned[mk{bi, k}]; ok {
				dt.start, dt.rows, dt.width, dt.ref, dt.agree, dt.keyed, dt.slot = p.cand.start, p.cand.rows, p.cand.width, p.ref, p.agree, p.keyed, p.slot
			} else if c, ok := mostRowLike(t.cands[k]); ok {
				dt.start, dt.rows, dt.width = c.start, c.rows, c.width
			}
			t.tables[k] = dt
		}
		for k := range t.tables {
			end := len(t.blob)
			if k+1 < len(t.tables) {
				end = t.tables[k+1].start
			}
			t.tables[k].end = end
			if dt := t.tables[k]; dt.ref != nil && dt.start >= 0 && end > dt.marker {
				recovered[dt.ref.stem] = dt.rows
			}
		}
	}

	var entries []Entry
	var spans []Span
	for _, t := range targets {
		for k, dt := range t.tables {
			if dt.start < 0 || dt.end <= dt.marker {
				spans = append(spans, Span{Bundle: t.name, Offset: int64(max(dt.start, 0)), Size: 0, Note: fmt.Sprintf("table marker at %d: no start found", dt.marker)})
				continue
			}
			if dt.ref == nil {
				spans = append(spans, Span{Bundle: t.name, Offset: int64(dt.start), Size: int64(dt.end - dt.start), Note: fmt.Sprintf("table %d: rows=%d width=%d, no reference match", k, dt.rows, dt.width)})
				continue
			}
			conf, method := gradeTable(t.blob, dt, recovered, opt.Columns)
			for _, p := range dt.ref.paths {
				entries = append(entries, Entry{Path: p, Bundle: t.name, Offset: int64(dt.start), Size: int64(dt.end - dt.start), Confidence: conf, Method: method})
			}
		}
	}

	// other files: unchanged bytes, or the only file between two placed neighbours
	for _, t := range targets {
		rb := refBlob(t.name)
		if rb == nil {
			continue
		}
		known := map[int64][2]int{} // reference offset -> target [start, end)
		for _, dt := range t.tables {
			if dt.ref != nil && dt.ref.bundle == t.name && dt.start >= 0 {
				for _, f := range refFiles {
					if f.Bundle == t.name && slices.Contains(dt.ref.paths, f.Path) {
						known[f.Offset] = [2]int{dt.start, dt.end}
						break
					}
				}
			}
		}
		order := refOrder(refFiles, t.name)
		placed := map[int64]Entry{}
		for _, g := range order {
			if isTablePath(g.paths[0]) || g.size < 64 || g.off+g.size > int64(len(rb)) {
				continue
			}
			want := rb[g.off : g.off+g.size]
			if i := bytes.Index(t.blob, want); i >= 0 && bytes.Index(t.blob[i+1:], want) < 0 {
				known[g.off] = [2]int{i, i + int(g.size)}
				placed[g.off] = Entry{Bundle: t.name, Offset: int64(i), Size: g.size, Confidence: High, Method: "identical to reference"}
			}
		}
		// reference bytes found unchanged may be the start of a longer, changed file: the next file in reference
		// order must begin where they end
		for i, g := range order {
			e, ok := placed[g.off]
			if !ok {
				continue
			}
			end := int(e.Offset + e.Size)
			switch next, known2 := nextKnown(order, i, known); {
			case i+1 == len(order) && end == len(t.blob), known2 && next == end:
			case known2 || i+1 == len(order):
				delete(placed, g.off)
				delete(known, g.off)
			default:
				e.Confidence, e.Method = Medium, "identical to reference, end not confirmed by a placed neighbour"
				placed[g.off] = e
			}
		}
		for i := 0; i < len(order); i++ {
			if _, ok := known[order[i].off]; ok || isTablePath(order[i].paths[0]) {
				continue
			}
			// the run of unplaced files i..j between two placed neighbours (or the bundle's ends)
			j := i
			for j+1 < len(order) && !isTablePath(order[j+1].paths[0]) {
				if _, ok := known[order[j+1].off]; ok {
					break
				}
				j++
			}
			start, end := 0, len(t.blob)
			if i > 0 {
				prev, ok := known[order[i-1].off]
				if !ok {
					i = j
					continue
				}
				start = prev[1]
			}
			if j+1 < len(order) {
				next, ok := known[order[j+1].off]
				if !ok {
					i = j
					continue
				}
				end = next[0]
			}
			run := order[i : j+1]
			i = j
			if end <= start {
				continue
			}
			type piece struct {
				g    refGroup
				a, b int
				how  string
			}
			var pieces []piece
			if len(run) == 1 {
				pieces = append(pieces, piece{run[0], start, end, "between placed neighbours"})
			} else {
				for _, m := range matchTextRun(t.blob, start, end, rb, run) {
					pieces = append(pieces, piece{m.g, m.a, m.b, m.how})
				}
			}
			for _, pc := range pieces {
				g, a, b := pc.g, pc.a, pc.b
				how := pc.how
				conf, method := Medium, fmt.Sprintf("%s, reference size %d", how, g.size)
				if ratio := float64(b-a) / float64(max(g.size, 1)); ratio < 0.5 || ratio > 2 {
					conf, method = Low, method+", size far from reference"
				}
				if err := checkFile(g.paths[0], t.blob[a:b], opt.Check); err != nil {
					conf, method = Low, method+", check failed: "+err.Error()
				} else if opt.Check != nil && strings.HasSuffix(g.paths[0], ".psg") {
					method += ", parsed"
				}
				known[g.off] = [2]int{a, b}
				placed[g.off] = Entry{Bundle: t.name, Offset: int64(a), Size: int64(b - a), Confidence: conf, Method: method}
			}
		}
		for _, g := range order {
			e, ok := placed[g.off]
			if !ok {
				continue
			}
			for _, p := range g.paths {
				e.Path = p
				entries = append(entries, e)
			}
		}
	}
	if opt.MovedText != nil {
		entries = append(entries, movedText(targets, entries, spans, refFiles, refBlob, opt.MovedText)...)
	}
	var gaps []Span
	for _, t := range targets {
		gaps = append(gaps, uncovered(t, entries, spans)...)
	}
	return entries, append(spans, gaps...), nil
}

func isTablePath(p string) bool {
	return strings.HasPrefix(p, "data/") && (strings.HasSuffix(p, ".datc64") || strings.HasSuffix(p, ".datcl64"))
}

// isLanguagePath is data/<language>/x: the localised copies, which the extractor does not read.
func isLanguagePath(p string) bool {
	return strings.HasPrefix(p, "data/") && strings.Count(p, "/") > 1
}

// baseBundle drops a _N split suffix: Folders/D/data.datc64_1 continues Folders/D/data.datc64.
func baseBundle(b string) string {
	if i := strings.LastIndexByte(b, '_'); i > strings.LastIndexByte(b, '/') {
		return b[:i]
	}
	return b
}

func newRefTable(b []byte, f RefFile) *refTable {
	if len(b) < 12 || isLanguagePath(f.Path) {
		return nil
	}
	m := bytes.Index(b[4:], marker)
	if m < 0 {
		return nil
	}
	m += 4
	rows := int(binary.LittleEndian.Uint32(b))
	w := 0
	if rows > 0 {
		if (m-4)%rows != 0 {
			return nil
		}
		w = (m - 4) / rows
	}
	stem := strings.SplitN(path.Base(f.Path), ".", 2)[0]
	return &refTable{paths: []string{f.Path}, stem: stem, bundle: f.Bundle, off: f.Offset, data: b, rows: rows, width: w, keyOff: -1}
}

// loadKeys reads the table's first string column, its Id in nearly every table, as the set to match against.
func (r *refTable) loadKeys(columns func(string) []Column) {
	r.utf32 = true
	for _, p := range r.paths {
		if !strings.HasSuffix(p, ".datcl64") {
			r.utf32 = false
		}
	}
	if columns == nil {
		return
	}
	r.cols = columns(r.stem)
	off := -1
	for _, c := range columns(r.stem) {
		if c.Type == "string" {
			off = c.Offset
			break
		}
	}
	if off < 0 || off+8 > r.width {
		return
	}
	m := 4 + r.rows*r.width
	keys := make(map[string]struct{}, r.rows)
	for i := range r.rows {
		if s, ok := readString(r.data, m, 4+i*r.width+off, r.utf32); ok {
			keys[s] = struct{}{}
		}
	}
	if len(keys) >= 2 {
		r.keyOff, r.keys = off, keys
	}
}

// readString reads the string a row field at fieldAt points to, relative to the marker at varAt.
func readString(b []byte, varAt, fieldAt int, utf32 bool) (string, bool) {
	if fieldAt+8 > len(b) {
		return "", false
	}
	o := binary.LittleEndian.Uint64(b[fieldAt:])
	if o < 8 || o > uint64(len(b)-varAt) {
		return "", false
	}
	p := varAt + int(o)
	if utf32 {
		var rs []rune
		for p+4 <= len(b) && len(rs) < 512 {
			c := binary.LittleEndian.Uint32(b[p:])
			if c == 0 {
				return string(rs), true
			}
			rs = append(rs, rune(c))
			p += 4
		}
		return "", false
	}
	var u []uint16
	for p+2 <= len(b) && len(u) < 512 {
		c := binary.LittleEndian.Uint16(b[p:])
		if c == 0 {
			return string(utf16.Decode(u)), true
		}
		u = append(u, c)
		p += 2
	}
	return "", false
}

func markers(b []byte) []int {
	var out []int
	for i := bytes.Index(b, marker); i >= 0; {
		out = append(out, i)
		j := i
		for j < len(b) && b[j] == 0xbb {
			j++
		}
		n := bytes.Index(b[j:], marker)
		if n < 0 {
			break
		}
		i = j + n
	}
	return out
}

// startCandidates lists every s in [lo, hi-4] where u32(s) rows of a whole width (at most 8192) end at hi.
func startCandidates(b []byte, lo, hi int) []candidate {
	var out []candidate
	for s := hi - 4; s >= lo; s-- {
		n := int(binary.LittleEndian.Uint32(b[s:]))
		f := hi - s - 4
		if n == 0 {
			if f == 0 {
				out = append(out, candidate{s, 0, 0})
			}
			continue
		}
		if f > 0 && f%n == 0 && f/n <= 8192 {
			out = append(out, candidate{s, n, f / n})
		}
	}
	return out
}

func rowRatio(a, b int) float64 {
	if a == 0 && b == 0 {
		return 1
	}
	return float64(min(a, b)) / float64(max(a, b))
}

func keyAgreement(b []byte, varAt int, c candidate, r *refTable) float64 {
	if r.keys == nil || c.rows == 0 {
		return 0
	}
	n := min(c.rows, 256)
	hits := 0
	for i := range n {
		row := i * c.rows / n
		if s, ok := readString(b, varAt, c.start+4+row*c.width+r.keyOff, r.utf32); ok {
			if _, in := r.keys[s]; in {
				hits++
			}
		}
	}
	// rows added since the reference cannot match; judge the share of the rows that could
	possible := float64(n) * min(1, float64(len(r.keys))/float64(c.rows))
	return min(1, float64(hits)/possible)
}

// stringsInRange is the share of sampled string fields, laid out by r's columns, that point inside the
// variable section, which ends no later than the next marker.
func stringsInRange(t *target, k int, c candidate, r *refTable) float64 {
	limit := uint64(len(t.blob) - t.marks[k])
	if k+1 < len(t.marks) {
		limit = uint64(t.marks[k+1] - t.marks[k])
	}
	n := min(c.rows, 64)
	ok, all := 0, 0
	for i := range n {
		row := c.start + 4 + (i*c.rows/n)*c.width
		for _, col := range r.cols {
			if col.Type != "string" || col.Offset+8 > c.width {
				continue
			}
			o := binary.LittleEndian.Uint64(t.blob[row+col.Offset:])
			all++
			if o >= 8 && o < limit {
				ok++
			}
		}
	}
	if all == 0 {
		return 1
	}
	return float64(ok) / float64(all)
}

func scorePair(t *target, k int, c candidate, r *refTable) (pair, bool) {
	ratio := rowRatio(c.rows, r.rows)
	same := 0.0
	if baseBundle(r.bundle) == baseBundle(t.name) {
		same = 1
	}
	if stringsInRange(t, k, c, r) < 0.95 {
		return pair{}, false
	}
	if r.keys != nil {
		agree := keyAgreement(t.blob, t.marks[k], c, r)
		if agree < 0.6 || ratio < 0.33 {
			return pair{}, false
		}
		return pair{k: k, cand: c, ref: r, score: 10*agree + 2*ratio + same, agree: agree, keyed: true}, true
	}
	if same == 0 || ratio < 0.5 {
		return pair{}, false
	}
	return pair{k: k, cand: c, ref: r, score: 2*ratio + same}, true
}

// mostRowLike picks the candidate with the most rows, then the earliest: a false start inside the table's own
// rows claims fewer rows, and one inside the previous table's variable data rarely matches the identity.
func mostRowLike(cs []candidate) (candidate, bool) {
	best, found := candidate{}, false
	for _, c := range cs {
		if !found || c.rows > best.rows || (c.rows == best.rows && c.start < best.start) {
			best, found = c, true
		}
	}
	return best, found
}

func gradeTable(b []byte, dt datTable, recovered map[string]int, columns func(string) []Column) (Confidence, string) {
	r := dt.ref
	span := b[dt.start:dt.end]
	method := fmt.Sprintf("table rows=%d/%d width=%d/%d", dt.rows, r.rows, dt.width, r.width)
	if dt.keyed {
		method += fmt.Sprintf(" ids=%.2f", dt.agree)
	}
	if bytes.Equal(span, r.data) {
		return High, method + " identical to reference"
	}
	var conf Confidence
	switch {
	case dt.slot && (dt.width*2 < r.width || dt.width > 2*r.width+64):
		conf, method = Low, method+" only table left between matched neighbours, width far from reference"
	case dt.slot && dt.rows == r.rows:
		conf, method = Medium, method+" only table left between matched neighbours, same rows"
	case dt.slot:
		conf, method = Low, method+" only table left between matched neighbours"
	case dt.keyed && dt.width == r.width && dt.agree >= 0.9:
		conf = High
	case dt.keyed && dt.agree >= 0.9:
		conf, method = Medium, method+" width changed"
	case dt.keyed && dt.width == r.width:
		conf = Medium
	case dt.keyed:
		conf, method = Low, method+" width changed"
	case dt.width == r.width && rowRatio(dt.rows, r.rows) >= 0.8:
		conf, method = Medium, method+" shape and bundle only"
	default:
		conf, method = Low, method+" shape only"
	}
	if dt.width != r.width || columns == nil {
		return conf, method
	}
	cols := columns(r.stem)
	if cols == nil {
		return conf, method
	}
	badStr, badFK, checked := 0, 0, 0
	varLen := uint64(dt.end - dt.marker)
	for i := range dt.rows {
		row := dt.start + 4 + i*dt.width
		for _, c := range cols {
			if c.Offset+c.Size > dt.width {
				continue
			}
			switch c.Type {
			case "string":
				o := binary.LittleEndian.Uint64(b[row+c.Offset:])
				if o < 8 || o >= varLen {
					badStr++
				}
			case "longid":
				n, ok := recovered[strings.ToLower(c.Ref)]
				if !ok {
					continue
				}
				id := binary.LittleEndian.Uint64(b[row+c.Offset:])
				checked++
				if id != nullRow && id >= uint64(n) {
					badFK++
				}
			}
		}
	}
	method += fmt.Sprintf(" strings-out-of-range=%d fk-out-of-range=%d/%d", badStr, badFK, checked)
	if (badStr > 0 || badFK > 0) && conf > Low {
		conf--
	}
	return conf, method
}

type refGroup struct {
	off, size int64
	paths     []string
}

func refOrder(files []RefFile, bundle string) []refGroup {
	by := map[int64]*refGroup{}
	for _, f := range files {
		if f.Bundle != bundle || isLanguagePath(f.Path) {
			continue
		}
		g, ok := by[f.Offset]
		if !ok {
			g = &refGroup{off: f.Offset, size: f.Size}
			by[f.Offset] = g
		}
		g.paths = append(g.paths, f.Path)
	}
	var out []refGroup
	for _, g := range by {
		slices.Sort(g.paths)
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].off < out[j].off })
	return out
}

// movedText names UTF-16 pieces of unidentified spans after the unplaced reference file they share most lines with.
func movedText(targets []*target, entries []Entry, spans []Span, refFiles []RefFile, refBlob func(string) []byte, pick func(string) bool) []Entry {
	placed := map[string]bool{}
	for _, e := range entries {
		placed[e.Path] = true
	}
	type item struct {
		f     RefFile
		lines map[string]bool
	}
	var pool []item
	for _, f := range refFiles {
		if !pick(f.Path) || placed[f.Path] || isLanguagePath(f.Path) {
			continue
		}
		blob := refBlob(f.Bundle)
		if blob == nil || f.Offset+f.Size > int64(len(blob)) || !bytes.HasPrefix(blob[f.Offset:], []byte{0xff, 0xfe}) {
			continue
		}
		pool = append(pool, item{f, lineSet(blob[f.Offset : f.Offset+f.Size])})
	}
	used := map[int]bool{}
	var out []Entry
	for _, t := range targets {
		for _, g := range uncovered(t, entries, spans) {
			a, end := int(g.Offset), int(g.Offset+g.Size)
			var cuts []int
			for p := a; p+2 <= end; p += 2 {
				if t.blob[p] == 0xff && t.blob[p+1] == 0xfe {
					cuts = append(cuts, p)
				}
			}
			if len(cuts) == 0 || cuts[0] != a {
				continue
			}
			cuts = append(cuts, end)
			for i := 0; i+1 < len(cuts); i++ {
				lines := lineSet(t.blob[cuts[i]:cuts[i+1]])
				best, bestSim := -1, 0.5
				for j, it := range pool {
					if used[j] {
						continue
					}
					shared := 0
					for l := range lines {
						if it.lines[l] {
							shared++
						}
					}
					if sim := float64(shared) / float64(max(1, min(len(lines), len(it.lines)))); sim >= bestSim {
						best, bestSim = j, sim
					}
				}
				if best < 0 {
					continue
				}
				used[best] = true
				f := pool[best].f
				out = append(out, Entry{Path: f.Path, Bundle: t.name, Offset: int64(cuts[i]), Size: int64(cuts[i+1] - cuts[i]), Confidence: Medium,
					Method: fmt.Sprintf("moved from %s, split at UTF-16 byte order marks, %.2f of lines shared with reference size %d", f.Bundle, bestSim, f.Size)})
			}
		}
	}
	return out
}

// nextKnown is the target start of the file after order[i], when it has been placed.
func nextKnown(order []refGroup, i int, known map[int64][2]int) (int, bool) {
	if i+1 >= len(order) {
		return 0, false
	}
	v, ok := known[order[i+1].off]
	return v[0], ok
}

type textMatch struct {
	g    refGroup
	a, b int
	how  string
}

// matchTextRun cuts [start, end) at the byte order marks that begin UTF-16 files and names each piece after
// the reference file in run it shares the most lines with (at least half of either's lines).
func matchTextRun(b []byte, start, end int, ref []byte, run []refGroup) []textMatch {
	bom := []byte{0xff, 0xfe}
	for _, g := range run {
		if g.off+2 > int64(len(ref)) || !bytes.HasPrefix(ref[g.off:], bom) {
			return nil
		}
	}
	var cuts []int
	for p := start; p+2 <= end; p += 2 {
		if b[p] == 0xff && b[p+1] == 0xfe {
			cuts = append(cuts, p)
		}
	}
	if len(cuts) == 0 || cuts[0] != start {
		return nil
	}
	cuts = append(cuts, end)
	refLines := make([]map[string]bool, len(run))
	for i, g := range run {
		refLines[i] = lineSet(ref[g.off : g.off+g.size])
	}
	type cand struct {
		piece, ref int
		sim        float64
	}
	var cs []cand
	for pi := 0; pi+1 < len(cuts); pi++ {
		lines := lineSet(b[cuts[pi]:cuts[pi+1]])
		for ri, rl := range refLines {
			shared := 0
			for l := range lines {
				if rl[l] {
					shared++
				}
			}
			sim := float64(shared) / float64(max(1, min(len(lines), len(rl))))
			if sim >= 0.5 {
				cs = append(cs, cand{pi, ri, sim})
			}
		}
	}
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].sim > cs[j].sim })
	usedP, usedR := map[int]bool{}, map[int]bool{}
	var out []textMatch
	for _, c := range cs {
		if usedP[c.piece] || usedR[c.ref] {
			continue
		}
		usedP[c.piece], usedR[c.ref] = true, true
		how := fmt.Sprintf("split at UTF-16 byte order marks, %.2f of lines shared with reference", c.sim)
		if len(cuts)-1 == len(run) {
			how += ", piece count equals reference"
		}
		out = append(out, textMatch{run[c.ref], cuts[c.piece], cuts[c.piece+1], how})
	}
	return out
}

// lineSet is the distinct non-blank lines of a UTF-16 text, at most 20000 of them.
func lineSet(b []byte) map[string]bool {
	u := make([]uint16, 0, len(b)/2)
	for p := 0; p+2 <= len(b); p += 2 {
		u = append(u, binary.LittleEndian.Uint16(b[p:]))
	}
	out := map[string]bool{}
	for _, l := range strings.Split(string(utf16.Decode(u)), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out[l] = true
			if len(out) >= 20000 {
				break
			}
		}
	}
	return out
}

func checkFile(p string, b []byte, check func(string, []byte) error) error {
	if strings.HasSuffix(p, ".txt") && !bytes.HasPrefix(b, []byte{0xff, 0xfe}) {
		return fmt.Errorf("no UTF-16 byte order mark")
	}
	if check != nil {
		return check(p, b)
	}
	return nil
}

func uncovered(t *target, entries []Entry, spans []Span) []Span {
	type iv struct{ a, b int64 }
	var ivs []iv
	for _, e := range entries {
		if e.Bundle == t.name {
			ivs = append(ivs, iv{e.Offset, e.Offset + e.Size})
		}
	}
	for _, s := range spans {
		if s.Bundle == t.name {
			ivs = append(ivs, iv{s.Offset, s.Offset + s.Size})
		}
	}
	sort.Slice(ivs, func(i, j int) bool { return ivs[i].a < ivs[j].a })
	var out []Span
	at := int64(0)
	for _, v := range ivs {
		if v.a > at {
			out = append(out, Span{Bundle: t.name, Offset: at, Size: v.a - at, Note: "not identified"})
		}
		at = max(at, v.b)
	}
	if n := int64(len(t.blob)); n > at {
		out = append(out, Span{Bundle: t.name, Offset: at, Size: n - at, Note: "not identified"})
	}
	return out
}
