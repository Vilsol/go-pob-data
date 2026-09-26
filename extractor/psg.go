package extractor

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

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
	Out        []uint32
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

func ParsePassiveSkillGraph(r io.Reader) (*PassiveSkillGraph, error) {
	p := &psgReader{r: bufio.NewReader(r)}

	g := &PassiveSkillGraph{Version: p.u8()}
	if p.err == nil && g.Version != psgVersion {
		return nil, fmt.Errorf("unsupported passive skill graph version %d (want %d)", g.Version, psgVersion)
	}

	g.GraphType = p.u8()
	orbitCount := p.u8()
	g.SkillsPerOrbit = make([]uint8, orbitCount)
	for i := range g.SkillsPerOrbit {
		g.SkillsPerOrbit[i] = p.u8()
	}

	g.Roots = p.u32s()

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
			group.Nodes = append(group.Nodes, PassiveSkillGraphNode{
				ID:         p.u32(),
				Orbit:      p.u32(),
				OrbitIndex: p.u32(),
				Out:        p.u32s(),
			})
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
