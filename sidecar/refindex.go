package sidecar

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/nfisher/gstream/hash/murmur2"
)

// RefFile is one file of a bundle index: where its bytes sit in which bundle.
type RefFile struct {
	Path   string
	Bundle string
	Offset int64
	Size   int64
}

// ParseIndex reads a decompressed-on-the-fly _.index.bin (murmur2 path hashes, 3.22 onwards) into its files.
func ParseIndex(raw []byte) ([]RefFile, error) {
	d, err := DecodeBundle(raw)
	if err != nil {
		return nil, fmt.Errorf("index: %w", err)
	}
	r := reader{d: d}
	bundles := make([]string, r.u32())
	for i := range bundles {
		n := int(r.u32())
		bundles[i] = string(r.take(n))
		r.u32()
	}
	type loc struct{ b, off, size uint32 }
	fc := int(r.u32())
	files := make(map[uint64]loc, fc)
	for range fc {
		h := r.u64()
		files[h] = loc{r.u32(), r.u32(), r.u32()}
	}
	type rep struct{ off, size uint32 }
	reps := make([]rep, r.u32())
	var seed uint64
	for i := range reps {
		h := r.u64()
		if i == 0 {
			a := (h ^ (h >> 47)) * 0x5F7A0EA7E59B19BD
			seed = a ^ (a >> 47)
		}
		reps[i] = rep{r.u32(), r.u32()}
		r.u32()
	}
	if r.err != nil {
		return nil, r.err
	}
	pd, err := DecodeBundle(d[r.p:])
	if err != nil {
		return nil, fmt.Errorf("index paths: %w", err)
	}
	var out []RefFile
	for _, rp := range reps {
		if int(rp.off+rp.size) > len(pd) {
			return nil, fmt.Errorf("path block runs past the path data")
		}
		for _, name := range readPathspec(pd[rp.off : rp.off+rp.size]) {
			f, ok := files[murmur2.Hash([]byte(name), seed)]
			if !ok || int(f.b) >= len(bundles) {
				continue
			}
			out = append(out, RefFile{Path: name, Bundle: bundles[f.b], Offset: int64(f.off), Size: int64(f.size)})
		}
	}
	return out, nil
}

type reader struct {
	err error
	d   []byte
	p   int
}

func (r *reader) take(n int) []byte {
	if r.err != nil || r.p+n > len(r.d) {
		r.err = fmt.Errorf("index ends early at %d", r.p)
		return make([]byte, n)
	}
	b := r.d[r.p : r.p+n]
	r.p += n
	return b
}

func (r *reader) u32() uint32 { return binary.LittleEndian.Uint32(r.take(4)) }
func (r *reader) u64() uint64 { return binary.LittleEndian.Uint64(r.take(8)) }

func readPathspec(data []byte) []string {
	p, phase := 0, 1
	var names, out []string
	for p+4 <= len(data) {
		n := int(binary.LittleEndian.Uint32(data[p:]))
		p += 4
		if n == 0 {
			phase = 1 - phase
			continue
		}
		e := bytes.IndexByte(data[p:], 0)
		if e < 0 {
			break
		}
		s := string(data[p : p+e])
		p += e + 1
		if n-1 < len(names) {
			s = names[n-1] + s
		}
		if phase == 0 {
			names = append(names, s)
		} else {
			out = append(out, s)
		}
	}
	return out
}
