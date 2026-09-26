package extractor

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/testdata"
)

func psgBytes(t *testing.T, fields ...any) []byte {
	t.Helper()
	var buf bytes.Buffer
	for _, f := range fields {
		testza.AssertNoError(t, binary.Write(&buf, binary.LittleEndian, f))
	}
	return buf.Bytes()
}

// minimalPSG has two orbits, one root and one group holding nodes 7 -> 9.
func minimalPSG(t *testing.T) []byte {
	t.Helper()
	return psgBytes(t,
		uint8(3), uint8(1), uint8(2), []uint8{1, 6},
		uint32(1), uint32(7),
		uint32(1),
		float32(-1.5), float32(2.25), uint32(0), uint32(2), uint8(1),
		uint32(2),
		uint32(7), uint32(0), uint32(0), uint32(1), uint32(9),
		uint32(9), uint32(1), uint32(5), uint32(0),
	)
}

func TestParsePassiveSkillGraphMinimal(t *testing.T) {
	g, err := ParsePassiveSkillGraph(bytes.NewReader(minimalPSG(t)))
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, &PassiveSkillGraph{
		Version:        3,
		GraphType:      1,
		SkillsPerOrbit: []uint8{1, 6},
		Roots:          []uint32{7},
		Groups: []PassiveSkillGraphGroup{{
			X:                  -1.5,
			Y:                  2.25,
			BackgroundOverride: 2,
			IsJewelPositionRef: true,
			Nodes: []PassiveSkillGraphNode{
				{ID: 7, Out: []uint32{9}},
				{ID: 9, Orbit: 1, OrbitIndex: 5, Out: []uint32{}},
			},
		}},
	}, g)
}

func TestParsePassiveSkillGraphUnknownVersion(t *testing.T) {
	data := minimalPSG(t)
	data[0] = 4
	_, err := ParsePassiveSkillGraph(bytes.NewReader(data))
	testza.AssertNotNil(t, err)
}

func TestParsePassiveSkillGraphTruncated(t *testing.T) {
	data := minimalPSG(t)
	for _, n := range []int{0, 1, 5, 20, len(data) - 1} {
		_, err := ParsePassiveSkillGraph(bytes.NewReader(data[:n]))
		testza.AssertNotNil(t, err, n)
	}
}

func TestParsePassiveSkillGraphTrailingBytes(t *testing.T) {
	_, err := ParsePassiveSkillGraph(bytes.NewReader(append(minimalPSG(t), 0)))
	testza.AssertNotNil(t, err)
}

func TestParseLivePassiveSkillGraph(t *testing.T) {
	requireLiveSchema(t)

	loader, err := GetBundleLoader(NewWebFS(testdata.GameVersion()), true)
	testza.AssertNoError(t, err)

	f, err := loader.Open("metadata/passiveskillgraph.psg")
	testza.AssertNoError(t, err)

	g, err := ParsePassiveSkillGraph(f)
	testza.AssertNoError(t, err)
	testza.AssertGreater(t, len(g.Roots), 0)
	testza.AssertGreater(t, len(g.Groups), 0)

	ids := make(map[uint32]bool)
	for _, group := range g.Groups {
		for _, node := range group.Nodes {
			ids[node.ID] = true
		}
	}
	testza.AssertGreater(t, len(ids), 0)

	for _, root := range g.Roots {
		testza.AssertTrue(t, ids[root], root)
	}
	for _, group := range g.Groups {
		for _, node := range group.Nodes {
			for _, out := range node.Out {
				testza.AssertTrue(t, ids[out], out)
			}
		}
	}
}
