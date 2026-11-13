package exelayout

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/arch/x86/x86asm"
)

type section struct {
	data      []byte
	rva, size uint32
	code      bool
}

// function is one RUNTIME_FUNCTION of the exception directory.
type function struct{ begin, end, unwind uint32 }

type image struct {
	sections []section
	funcs    []function
}

type ins struct {
	x86asm.Inst
	pc  uint32
	bad bool
}

func parseImage(r io.ReaderAt) (*image, error) {
	f, err := pe.NewFile(r)
	if err != nil {
		return nil, errors.Wrap(err, "not a PE file")
	}
	defer f.Close()
	oh, ok := f.OptionalHeader.(*pe.OptionalHeader64)
	if !ok {
		return nil, errors.New("not a PE32+ image")
	}
	img := &image{}
	for _, s := range f.Sections {
		d, err := s.Data()
		if err != nil {
			return nil, errors.Wrapf(err, "reading section %s", s.Name)
		}
		if uint32(len(d)) > s.VirtualSize {
			d = d[:s.VirtualSize]
		}
		img.sections = append(img.sections, section{
			data: d,
			rva:  s.VirtualAddress,
			size: s.VirtualSize,
			code: s.Characteristics&pe.IMAGE_SCN_MEM_EXECUTE != 0,
		})
	}
	exc := oh.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_EXCEPTION]
	pd := img.at(exc.VirtualAddress, int(exc.Size))
	for i := 0; i+12 <= len(pd); i += 12 {
		img.funcs = append(img.funcs, function{le32(pd[i:]), le32(pd[i+4:]), le32(pd[i+8:])})
	}
	sort.Slice(img.funcs, func(a, b int) bool { return img.funcs[a].begin < img.funcs[b].begin })
	return img, nil
}

func le32(b []byte) uint32 { return binary.LittleEndian.Uint32(b) }

func (img *image) at(rva uint32, n int) []byte {
	for _, s := range img.sections {
		if rva >= s.rva && rva-s.rva < uint32(len(s.data)) {
			off := rva - s.rva
			end := min(int(off)+n, len(s.data))
			return s.data[off:end]
		}
	}
	return nil
}

// function returns the bounds of the function containing rva, following chained unwind info to the primary entry.
func (img *image) function(rva uint32) (uint32, uint32, bool) {
	i := sort.Search(len(img.funcs), func(i int) bool { return img.funcs[i].begin > rva }) - 1
	if i < 0 || rva >= img.funcs[i].end {
		return 0, 0, false
	}
	for range 8 {
		u := img.at(img.funcs[i].unwind, 4)
		// UNW_FLAG_CHAININFO
		if len(u) < 4 || (u[0]>>3)&4 == 0 {
			break
		}
		codes := int(u[2])
		p := img.at(img.funcs[i].unwind+4+uint32((codes+1)&^1)*2, 4)
		if len(p) < 4 {
			break
		}
		parent := le32(p)
		j := sort.Search(len(img.funcs), func(k int) bool { return img.funcs[k].begin >= parent })
		if j == len(img.funcs) || img.funcs[j].begin != parent {
			break
		}
		i = j
	}
	return img.funcs[i].begin, img.funcs[i].end, true
}

func (img *image) decode(begin, end uint32) []ins {
	return decodeAt(img.at(begin, int(min(end-begin, 1<<20))), begin)
}

// decodeAt decodes code as if loaded at pc; an undecodable byte becomes one bad instruction.
func decodeAt(code []byte, pc uint32) []ins {
	var out []ins
	for off := 0; off < len(code); {
		in, err := x86asm.Decode(code[off:], 64)
		if err != nil {
			out = append(out, ins{pc: pc + uint32(off), bad: true})
			off++
			continue
		}
		out = append(out, ins{Inst: in, pc: pc + uint32(off)})
		off += in.Len
	}
	return out
}

// wide reads a NUL-terminated printable-ASCII UTF-16LE string at rva.
func (img *image) wide(rva uint32) string {
	b := img.at(rva, 520)
	var s []byte
	for i := 0; i+1 < len(b); i += 2 {
		u := binary.LittleEndian.Uint16(b[i:])
		if u == 0 {
			return string(s)
		}
		if u < 0x20 || u >= 0x7f {
			return ""
		}
		s = append(s, byte(u))
	}
	return ""
}

// tableNamed is the table a wide "Data/.../<Table>.dat" string at rva names.
func (img *image) tableNamed(rva uint32) string {
	return tableFromPath(img.wide(rva))
}

func tableFromPath(p string) string {
	if len(p) < 9 || !strings.HasPrefix(p, "Data/") || !strings.HasSuffix(p, ".dat") {
		return ""
	}
	p = p[5 : len(p)-4]
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		p = p[i+1:]
	}
	return p
}

type candidate struct {
	name string
	rva  uint32
}

var rowSuffix = []byte{'_', 0, 'R', 0, 'o', 0, 'w', 0, 0, 0}

// rowNames finds every UTF-16 "<Table>_Row" string in data sections and returns each one's possible starts, longest first.
func (img *image) rowNames() [][]candidate {
	var out [][]candidate
	for _, s := range img.sections {
		if !s.code {
			out = append(out, rowNamesIn(s.data, s.rva)...)
		}
	}
	return out
}

func rowNamesIn(d []byte, rva uint32) [][]candidate {
	ident := func(at int) bool {
		c := d[at]
		return (c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') && d[at+1] == 0
	}
	var out [][]candidate
	for at := 0; ; {
		k := bytes.Index(d[at:], rowSuffix)
		if k < 0 {
			break
		}
		at += k
		start := at
		for start >= 2 && ident(start-2) {
			start -= 2
		}
		var cs []candidate
		for st := start; st < at; st += 2 {
			n := make([]byte, 0, (at-st)/2)
			for i := st; i < at; i += 2 {
				n = append(n, d[i])
			}
			cs = append(cs, candidate{string(n), rva + uint32(st)})
		}
		if len(cs) > 0 {
			out = append(out, cs)
		}
		at += len(rowSuffix)
	}
	return out
}

// references finds every rel32 in code that resolves to one of targets, keyed by target.
func (img *image) references(targets map[uint32]bool) map[uint32][]uint32 {
	out := map[uint32][]uint32{}
	if len(targets) == 0 {
		return out
	}
	var lo, hi uint32 = ^uint32(0), 0
	for t := range targets {
		lo, hi = min(lo, t), max(hi, t)
	}
	for _, s := range img.sections {
		if !s.code {
			continue
		}
		for i := 0; i+4 <= len(s.data); i++ {
			t := int64(s.rva) + int64(i) + 4 + int64(int32(le32(s.data[i:])))
			if t >= int64(lo) && t <= int64(hi) && targets[uint32(t)] {
				out[uint32(t)] = append(out[uint32(t)], s.rva+uint32(i))
			}
		}
	}
	return out
}

func ripTarget(in ins, m x86asm.Mem) uint32 {
	return uint32(int64(in.pc) + int64(in.Len) + m.Disp)
}

func branchTarget(in ins) (uint32, bool) {
	if r, ok := in.Args[0].(x86asm.Rel); ok {
		return uint32(int64(in.pc) + int64(in.Len) + int64(r)), true
	}
	return 0, false
}
