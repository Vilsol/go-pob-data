package exelayout

import (
	"sort"
	"strings"

	"golang.org/x/arch/x86/x86asm"
)

// Kind is how the client fixes up a pointer column after loading the rows.
type Kind string

const (
	KindString Kind = "string" // 8-byte offset into the variable data
	KindKey    Kind = "key"    // 16-byte row index and table handle (dat-schema's foreignrow)
	KindRow    Kind = "row"    // 8-byte row index into its own table
	KindList   Kind = "list"   // 16-byte count and offset into the variable data
)

func (k Kind) width() int {
	if k == KindKey || k == KindList {
		return 16
	}
	return 8
}

type holdsKind int

const (
	hRow       holdsKind = iota + 1 // a pointer n bytes into the current row
	hItem                           // a list element of the list column at n
	hHandle                         // a table handle
	hHandlePtr                      // a pointer to a table handle
	hSlot                           // a pointer to a stack slot
)

type holds struct {
	table string
	slot  [2]int64
	n     int64
	k     holdsKind
}

type cell struct {
	element              int
	read8, write8, read4 bool
}

type target struct {
	table  string
	helper uint32
}

// fixups walks fixup functions; its caches are shared across tables.
type fixups struct {
	img     *image
	getters map[uint32]string
	helpers map[uint32]int
	sizes   map[string]int
}

func newFixups(img *image, sizes map[string]int) *fixups {
	return &fixups{img: img, getters: map[uint32]string{}, helpers: map[uint32]int{}, sizes: sizes}
}

// getter is the one table a small function names, the client's accessor for that table's handle.
func (f *fixups) getter(rva uint32) string {
	if g, ok := f.getters[rva]; ok {
		return g
	}
	res := ""
	if b, e, ok := f.img.function(rva); ok && b == rva && e-b <= 0x400 {
		set := map[string]bool{}
		for _, in := range f.img.decode(b, e) {
			if m, _, ok := mem(in); ok && in.Op == x86asm.LEA && m.Base == x86asm.RIP {
				if n := f.img.tableNamed(ripTarget(in, m)); n != "" {
					set[n] = true
				}
			}
		}
		if len(set) == 1 {
			for n := range set {
				res = n
			}
		}
	}
	f.getters[rva] = res
	return res
}

// helperRowSize is the row size a key-resolving helper bakes in: what it scales the row index (rdx) by.
func (f *fixups) helperRowSize(rva uint32) int {
	if h, ok := f.helpers[rva]; ok {
		return h
	}
	res := 0
	if b, e, ok := f.img.function(rva); ok && b == rva && e-b <= 0x200 {
		res = helperScale(f.img.decode(b, e))
	}
	f.helpers[rva] = res
	return res
}

func helperScale(insns []ins) int {
	s := newSums()
	s.regs[x86asm.RDX] = sum{1, 0}
	scaled := func(r x86asm.Reg) int64 {
		if v, ok := s.regs[r]; ok && v.a > 0 && v.b == 0 {
			return v.a
		}
		return 0
	}
	for _, in := range insns {
		var found int64
		m, _, hasMem := mem(in)
		switch {
		case in.Op == x86asm.ADD && hasMem && in.MemBytes == 8:
			if r, ok := reg(in, 0); ok {
				found = scaled(r)
			}
		case in.Op == x86asm.LEA && m.Index != 0 && m.Disp == 0:
			if _, known := s.regs[m.Base]; !known {
				found = scaled(m.Index) * int64(m.Scale)
			}
		}
		if found > 0 {
			return int(found)
		}
		s.step(in)
	}
	return 0
}

// rowLoop finds the register that steps by rowSize through the rows, and the instruction range of that loop.
func rowLoop(insns []ins, rowSize int) (x86asm.Reg, int, int, bool) {
	size := int64(rowSize)
	steps := func(in ins) (x86asm.Reg, bool) {
		r, ok1 := reg(in, 0)
		by, ok2 := imm(in, 1)
		if !ok1 || !ok2 || isStack(r) {
			return 0, false
		}
		if (in.Op == x86asm.ADD && by == size) || (in.Op == x86asm.SUB && by == -size) {
			return r, true
		}
		return 0, false
	}
	uses := func(r x86asm.Reg) int {
		n := 0
		for _, in := range insns {
			if m, _, ok := mem(in); ok && m.Base == r && m.Index == 0 && m.Disp >= 0 && m.Disp < size {
				n++
			}
		}
		return n
	}
	best, bestUses, found := x86asm.Reg(0), -1, false
	seen := map[x86asm.Reg]bool{}
	for _, in := range insns {
		if r, ok := steps(in); ok && !seen[r] {
			seen[r] = true
			if u := uses(r); u > bestUses || (u == bestUses && r > best) {
				best, bestUses, found = r, u, true
			}
		}
	}
	if !found {
		return 0, 0, 0, false
	}
	end := -1
	for i := len(insns) - 1; i >= 0; i-- {
		if r, ok := steps(insns[i]); ok && r == best {
			end = i
			break
		}
	}
	begin := 0
	for i := end - 1; i >= 0; i-- {
		if r, ok := reg(insns[i], 0); ok && insns[i].Op == x86asm.MOV && r == best && insns[i].Args[1] != nil {
			begin = i + 1
			break
		}
	}
	return best, begin, end, true
}

// pointers walks a fixup function and classifies every row offset it touches by its read/write pattern.
// A key or list whose target table's handle the loop passes along gets that target; it is proven when the
// helper resolving it scales by the target's own row size.
func (f *fixups) pointers(insns []ins, rowSize int) ([]Pointer, bool) {
	row, begin, end, ok := rowLoop(insns, rowSize)
	if !ok {
		return nil, false
	}
	size := int64(rowSize)
	cells := map[int64]*cell{}
	getCell := func(at int64) *cell {
		if c, ok := cells[at]; ok {
			return c
		}
		c := &cell{}
		cells[at] = c
		return c
	}
	regs := map[x86asm.Reg]holds{}
	slots := map[[2]int64]string{}
	targets := map[int64]target{}
	pending := ""

	for i, in := range insns {
		if in.bad {
			continue
		}
		inLoop := i >= begin && i < end
		within := func(base x86asm.Reg, disp int64) (int64, bool) {
			var from int64
			if base == row {
				from = 0
			} else if h, ok := regs[base]; ok && h.k == hRow {
				from = h.n
			} else {
				return 0, false
			}
			return from + disp, inLoop && from+disp >= 0 && from+disp < size
		}
		stored := ""
		if r, ok := reg(in, 1); ok && in.Op == x86asm.MOV {
			if h, ok := regs[r]; ok && h.k == hHandle {
				stored = h.table
			}
		}

		if a, ok := memAccess(in); ok && a.m.Index == 0 && (a.reads || a.writes) {
			if a.writes && isStack(a.m.Base) {
				for part := int64(0); part < max(a.width, 8); part += 8 {
					delete(slots, [2]int64{int64(a.m.Base), a.m.Disp + part})
				}
			}
			if at, ok := within(a.m.Base, a.m.Disp); ok {
				switch a.width {
				case 8, 16, 32:
					for part := int64(0); part < a.width; part += 8 {
						c := getCell(at + part)
						c.read8 = c.read8 || a.reads
						c.write8 = c.write8 || a.writes
					}
				case 4:
					c := getCell(at)
					c.read4 = c.read4 || a.reads
				}
				if a.writes && a.width == 8 && stored != "" {
					targets[at-8] = target{table: stored}
				}
			} else if h, ok := regs[a.m.Base]; ok && inLoop && a.writes && a.width == 8 && a.m.Disp == 8 && h.k == hItem && stored != "" {
				targets[h.n] = target{table: stored}
			}
		}

		dst, hasDst := reg(in, 0)
		m, _, hasMem := mem(in)
		plain := hasMem && m.Index == 0
		var set *holds
		switch {
		case in.Op == x86asm.LEA && hasMem && m.Base == x86asm.RIP:
			if t := f.img.tableNamed(ripTarget(in, m)); t != "" {
				pending = t
			}
		case in.Op == x86asm.LEA && plain:
			if at, ok := within(m.Base, m.Disp); ok {
				set = &holds{k: hRow, n: at}
			} else if isStack(m.Base) {
				set = &holds{k: hSlot, slot: [2]int64{int64(m.Base), m.Disp}}
			}
		case in.Op == x86asm.MOV && hasDst && !hasMem:
			if from, ok := reg(in, 1); ok {
				if inLoop && from == row {
					set = &holds{k: hRow}
				} else if h, ok := regs[from]; ok {
					set = &h
				}
			}
		case in.Op == x86asm.MOV && hasDst && plain:
			if h, ok := regs[m.Base]; ok && h.k == hHandlePtr && m.Disp == 0 {
				set = &holds{k: hHandle, table: h.table}
			} else if t, ok := slots[[2]int64{int64(m.Base), m.Disp}]; ok {
				set = &holds{k: hHandle, table: t}
			}
		case in.Op == x86asm.MOV && plain && isStack(m.Base):
			if stored != "" {
				slots[[2]int64{int64(m.Base), m.Disp}] = stored
			}
		case in.Op == x86asm.ADD && hasDst && plain:
			if at, ok := within(m.Base, m.Disp); ok {
				set = &holds{k: hItem, n: at - 8}
				if i > 0 {
					p := insns[i-1]
					if r, ok := reg(p, 0); ok && p.Op == x86asm.SHL && r == dst {
						if s, ok := imm(p, 1); ok && s >= 0 && s < 8 {
							getCell(at).element = 1 << s
						}
					}
				}
			}
		case in.Op == x86asm.CALL:
			first, hf := regs[x86asm.RCX]
			second, hs := regs[x86asm.RDX]
			third, ht := regs[x86asm.R8]
			callee, direct := branchTarget(in)
			if inLoop && hf && (first.k == hRow || first.k == hItem) && ht && third.k == hSlot {
				if t, ok := slots[third.slot]; ok {
					c := uint32(0)
					if direct {
						c = callee
					}
					targets[first.n] = target{t, c}
				}
			}
			for _, r := range volatile {
				delete(regs, r)
			}
			named := ""
			if direct {
				named = f.getter(callee)
			}
			switch {
			case named != "":
				if hf && first.k == hSlot {
					slots[first.slot] = named
				}
				regs[x86asm.RAX] = holds{k: hHandlePtr, table: named}
			case hf && first.k == hSlot && hs && second.k == hSlot:
				if pending != "" {
					slots[first.slot] = pending
					pending = ""
				}
			}
			continue
		}
		for _, r := range written(in) {
			delete(regs, r)
		}
		if hasDst && set != nil {
			regs[dst] = *set
		}
	}

	return f.classify(cells, targets), true
}

func (f *fixups) classify(cells map[int64]*cell, targets map[int64]target) []Pointer {
	type resolved struct {
		table  string
		proven bool
	}
	final := map[int64]resolved{}
	for at, t := range targets {
		baked := 0
		if t.helper != 0 {
			baked = f.helperRowSize(t.helper)
		}
		proven := baked != 0 && baked == f.sizes[strings.ToLower(t.table)]
		if baked == 0 || proven {
			final[at] = resolved{t.table, proven}
		}
	}
	keys := make([]int64, 0, len(cells))
	for k := range cells {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(a, b int) bool { return keys[a] < keys[b] })
	var out []Pointer
	taken := map[int64]bool{}
	for _, at := range keys {
		if taken[at] {
			continue
		}
		k, ok := classifyCell(*cells[at], cells[at+8])
		if !ok {
			continue
		}
		if k.width() == 16 {
			taken[at+8] = true
		}
		p := Pointer{Offset: int(at), Kind: k}
		if k == KindList {
			p.Element = cells[at+8].element
		}
		if t, ok := final[at]; ok && (k == KindKey || k == KindList) {
			p.Target, p.Proven = t.table, t.proven
		}
		out = append(out, p)
	}
	return out
}

// classifyCell names the pointer kind of the cell at an offset given the cell 8 bytes after it.
func classifyCell(c cell, next *cell) (Kind, bool) {
	n := cell{}
	if next != nil {
		n = *next
	}
	switch {
	case (c.read8 || c.read4) && c.write8 && n.write8 && !n.read8 && !n.read4:
		return KindKey, true
	case c.read8 && !c.write8 && n.read8 && n.write8:
		return KindList, true
	case c.read8 && c.write8:
		return KindString, true
	case c.read4 && c.write8:
		return KindRow, true
	}
	return "", false
}
