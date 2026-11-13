// Package sidecar names the files inside a client's bundles when the bundle index is absent (a preload ships its
// bundles before the index), using a previous patch's index as the map, and serves them through an fs.FS.
package sidecar

import (
	"encoding/binary"
	"fmt"

	"github.com/oriath-net/gooz"
)

type bundleHead struct {
	UncompressedSize             uint32
	TotalPayloadSize             uint32
	HeadPayloadSize              uint32
	FirstFileEncode              uint32
	_                            uint32
	UncompressedSize2            int64
	TotalPayloadSize2            int64
	BlockCount                   uint32
	UncompressedBlockGranularity uint32
	_                            [4]uint32
}

// DecodeBundle decompresses a whole bundle and refuses one whose blocks do not add up to its declared size.
func DecodeBundle(raw []byte) ([]byte, error) {
	var bh bundleHead
	hs := binary.Size(bh)
	if len(raw) < hs {
		return nil, fmt.Errorf("bundle of %d bytes has no head", len(raw))
	}
	if _, err := binary.Decode(raw[:hs], binary.LittleEndian, &bh); err != nil {
		return nil, err
	}
	p := hs + 4*int(bh.BlockCount)
	if bh.UncompressedBlockGranularity == 0 || p > len(raw) || bh.UncompressedSize2 < 0 {
		return nil, fmt.Errorf("bundle head is not plausible")
	}
	gran := int64(bh.UncompressedBlockGranularity)
	if want := (bh.UncompressedSize2 + gran - 1) / gran; want != int64(bh.BlockCount) {
		return nil, fmt.Errorf("%d blocks of %d for %d bytes", bh.BlockCount, gran, bh.UncompressedSize2)
	}
	out := make([]byte, bh.UncompressedSize2)
	for i := range int(bh.BlockCount) {
		sz := int(binary.LittleEndian.Uint32(raw[hs+4*i:]))
		if p+sz > len(raw) {
			return nil, fmt.Errorf("block %d runs past the bundle", i)
		}
		lo := int64(i) * gran
		hi := min(lo+gran, bh.UncompressedSize2)
		n, err := gooz.Decompress(raw[p:p+sz], out[lo:hi])
		if err != nil {
			return nil, fmt.Errorf("block %d: %w", i, err)
		}
		if int64(n) != hi-lo {
			return nil, fmt.Errorf("block %d: %d bytes, want %d", i, n, hi-lo)
		}
		p += sz
	}
	if p != len(raw) {
		return nil, fmt.Errorf("blocks end at %d of %d bytes", p, len(raw))
	}
	return out, nil
}
