package extractor

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/pkg/errors"
)

const psgVersion = 3

// PassiveSkillGraph is a decoded .psg file (e.g. metadata/passiveskillgraph.psg).
type PassiveSkillGraph struct {
	SkillsPerOrbit []uint8
	Roots          []uint32
	Groups         []PassiveSkillGraphGroup
	Version        uint8
	GraphType      uint8
	// Wide is PoE2's layout: u64 roots and edges carrying an orbit.
	Wide bool
}

type PassiveSkillGraphGroup struct {
	Nodes               []PassiveSkillGraphNode
	X                   float32
	Y                   float32
	GroupAssociationKey uint32
	BackgroundOverride  uint32
	IsJewelPositionRef  bool
}

type PassiveSkillGraphNode struct {
	// Out are the PassiveSkillGraphIds this node has an edge to.
	Out []uint32
	// OutOrbit is the curved-edge orbit of each Out edge in the wide layout, nil otherwise.
	OutOrbit   []int32
	ID         uint32
	Orbit      uint32
	OrbitIndex uint32
}

type psgReader struct {
	r   *bufio.Reader
	err error
}

func (p *psgReader) read(v any) {
	if p.err != nil {
		return
	}
	if err := binary.Read(p.r, binary.LittleEndian, v); err != nil {
		p.err = errors.Wrap(err, "truncated passive skill graph")
	}
}

func (p *psgReader) u8() uint8 {
	var v uint8
	p.read(&v)
	return v
}

func (p *psgReader) u32() uint32 {
	var v uint32
	p.read(&v)
	return v
}

func (p *psgReader) u64() uint64 {
	var v uint64
	p.read(&v)
	return v
}

func (p *psgReader) f32() float32 {
	var v float32
	p.read(&v)
	return v
}

// u32s reads a u32 count followed by that many u32s.
func (p *psgReader) u32s() []uint32 {
	n := p.u32()
	if p.err != nil {
		return nil
	}
	out := make([]uint32, 0, min(n, 1<<16))
	for range n {
		v := p.u32()
		if p.err != nil {
			return nil
		}
		out = append(out, v)
	}
	return out
}

// roots64 reads a u32 count followed by that many u64 node ids, refusing an id past 32 bits.
func (p *psgReader) roots64() []uint32 {
	n := p.u32()
	if p.err != nil {
		return nil
	}
	out := make([]uint32, 0, min(n, 1<<16))
	for range n {
		v := p.u64()
		if p.err != nil {
			return nil
		}
		if v > math.MaxUint32 {
			p.err = fmt.Errorf("passive skill graph root %d does not fit a node id", v)
			return nil
		}
		out = append(out, uint32(v))
	}
	return out
}

// edges2 reads a u32 count followed by that many {id u32, orbit i32} pairs.
func (p *psgReader) edges2() ([]uint32, []int32) {
	n := p.u32()
	if p.err != nil {
		return nil, nil
	}
	ids := make([]uint32, 0, min(n, 1<<16))
	orbits := make([]int32, 0, min(n, 1<<16))
	for range n {
		id := p.u32()
		var orbit int32
		p.read(&orbit)
		if p.err != nil {
			return nil, nil
		}
		ids = append(ids, id)
		orbits = append(orbits, orbit)
	}
	return ids, orbits
}

// ParsePassiveSkillGraph decodes a .psg in either layout: PoE1's, or PoE2's with u64 roots and edges
// carrying an orbit (PathOfBuilding-PoE2 src/Export/Scripts/passivetree.lua:154-205), which PoE1 uses
// from 3.30. The header does not name the layout, so the file is read in both and must decode, to its
// last byte, in exactly one.
func ParsePassiveSkillGraph(r io.Reader) (*PassiveSkillGraph, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading passive skill graph")
	}

	narrow, narrowErr := parsePassiveSkillGraph(data, false)
	wide, wideErr := parsePassiveSkillGraph(data, true)
	switch {
	case narrowErr == nil && wideErr == nil:
		return nil, errors.New("passive skill graph decodes in both layouts")
	case narrowErr == nil:
		return narrow, nil
	case wideErr == nil:
		return wide, nil
	case errors.Is(narrowErr, errUnsupportedVersion):
		return nil, narrowErr
	default:
		return nil, fmt.Errorf("passive skill graph decodes in neither layout: %w; with u64 roots: %w", narrowErr, wideErr)
	}
}

var errUnsupportedVersion = errors.New("unsupported passive skill graph version")

// parsePassiveSkillGraph decodes data in PoE1's layout, or PoE2's when wide.
func parsePassiveSkillGraph(data []byte, wide bool) (*PassiveSkillGraph, error) {
	p := &psgReader{r: bufio.NewReader(bytes.NewReader(data))}

	g := &PassiveSkillGraph{Version: p.u8(), Wide: wide}
	if p.err == nil && g.Version != psgVersion {
		return nil, fmt.Errorf("%w %d (want %d)", errUnsupportedVersion, g.Version, psgVersion)
	}

	g.GraphType = p.u8()
	orbitCount := p.u8()
	g.SkillsPerOrbit = make([]uint8, orbitCount)
	for i := range g.SkillsPerOrbit {
		g.SkillsPerOrbit[i] = p.u8()
	}

	if wide {
		g.Roots = p.roots64()
	} else {
		g.Roots = p.u32s()
	}

	groupCount := p.u32()
	for range groupCount {
		if p.err != nil {
			break
		}

		group := PassiveSkillGraphGroup{
			X:                   p.f32(),
			Y:                   p.f32(),
			GroupAssociationKey: p.u32(),
			BackgroundOverride:  p.u32(),
			IsJewelPositionRef:  p.u8() != 0,
		}

		nodeCount := p.u32()
		for range nodeCount {
			if p.err != nil {
				break
			}
			node := PassiveSkillGraphNode{
				ID:         p.u32(),
				Orbit:      p.u32(),
				OrbitIndex: p.u32(),
			}
			if wide {
				node.Out, node.OutOrbit = p.edges2()
			} else {
				node.Out = p.u32s()
			}
			group.Nodes = append(group.Nodes, node)
		}

		g.Groups = append(g.Groups, group)
	}

	if p.err != nil {
		return nil, p.err
	}

	if _, err := p.r.ReadByte(); err != io.EOF {
		if err != nil {
			return nil, errors.Wrap(err, "failed reading passive skill graph")
		}
		return nil, errors.New("trailing bytes after passive skill graph")
	}

	return g, nil
}
