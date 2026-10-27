package extractor

import (
	"bytes"
	"encoding/binary"
	"io"
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
	g, err := ParsePassiveSkillGraph(bytes.NewReader(minimalPSG(t)), PoE1)
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

// minimalPSG2 is minimalPSG in PoE2's layout: u64 roots and edges carrying an orbit.
func minimalPSG2(t *testing.T) []byte {
	t.Helper()
	return psgBytes(t,
		uint8(3), uint8(0), uint8(2), []uint8{1, 12},
		uint32(1), uint64(7),
		uint32(1),
		float32(-1.5), float32(2.25), uint32(0), uint32(2), uint8(1),
		uint32(2),
		uint32(7), uint32(0), uint32(0), uint32(1), uint32(9), int32(-4),
		uint32(9), uint32(1), uint32(5), uint32(0),
	)
}

func TestParsePassiveSkillGraphPoE2Minimal(t *testing.T) {
	g, err := ParsePassiveSkillGraph(bytes.NewReader(minimalPSG2(t)), PoE2)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, &PassiveSkillGraph{
		Version:        3,
		SkillsPerOrbit: []uint8{1, 12},
		Roots:          []uint32{7},
		Groups: []PassiveSkillGraphGroup{{
			X:                  -1.5,
			Y:                  2.25,
			BackgroundOverride: 2,
			IsJewelPositionRef: true,
			Nodes: []PassiveSkillGraphNode{
				{ID: 7, Out: []uint32{9}, OutOrbit: []int32{-4}},
				{ID: 9, Orbit: 1, OrbitIndex: 5, Out: []uint32{}, OutOrbit: []int32{}},
			},
		}},
	}, g)
}

func TestParsePassiveSkillGraphLayoutsDisagree(t *testing.T) {
	_, err := ParsePassiveSkillGraph(bytes.NewReader(minimalPSG2(t)), PoE1)
	testza.AssertNotNil(t, err)
	_, err = ParsePassiveSkillGraph(bytes.NewReader(minimalPSG(t)), PoE2)
	testza.AssertNotNil(t, err)
}

func TestParsePassiveSkillGraphPoE2WideRoot(t *testing.T) {
	data := minimalPSG2(t)
	data[20] = 1 // high word of the root
	_, err := ParsePassiveSkillGraph(bytes.NewReader(data), PoE2)
	testza.AssertNotNil(t, err)
}

func TestParsePassiveSkillGraphPoE2Truncated(t *testing.T) {
	data := minimalPSG2(t)
	for _, n := range []int{0, 1, 5, 17, 20, len(data) - 5, len(data) - 1} {
		_, err := ParsePassiveSkillGraph(bytes.NewReader(data[:n]), PoE2)
		testza.AssertNotNil(t, err, n)
	}
}

func TestParsePassiveSkillGraphUnknownVersion(t *testing.T) {
	data := minimalPSG(t)
	data[0] = 4
	_, err := ParsePassiveSkillGraph(bytes.NewReader(data), PoE1)
	testza.AssertNotNil(t, err)
}

func TestParsePassiveSkillGraphTruncated(t *testing.T) {
	data := minimalPSG(t)
	for _, n := range []int{0, 1, 5, 20, len(data) - 1} {
		_, err := ParsePassiveSkillGraph(bytes.NewReader(data[:n]), PoE1)
		testza.AssertNotNil(t, err, n)
	}
}

func TestParsePassiveSkillGraphTrailingBytes(t *testing.T) {
	_, err := ParsePassiveSkillGraph(bytes.NewReader(append(minimalPSG(t), 0)), PoE1)
	testza.AssertNotNil(t, err)
}

// encodePSG writes g back in game's layout, for the round trip against the game file.
func encodePSG(t *testing.T, g *PassiveSkillGraph, game Game) []byte {
	t.Helper()
	fields := []any{g.Version, g.GraphType, uint8(len(g.SkillsPerOrbit)), g.SkillsPerOrbit, uint32(len(g.Roots))}
	for _, r := range g.Roots {
		if game == PoE2 {
			fields = append(fields, uint64(r))
		} else {
			fields = append(fields, r)
		}
	}
	fields = append(fields, uint32(len(g.Groups)))
	for _, group := range g.Groups {
		jewel := uint8(0)
		if group.IsJewelPositionRef {
			jewel = 1
		}
		fields = append(fields, group.X, group.Y, group.GroupAssociationKey, group.BackgroundOverride, jewel, uint32(len(group.Nodes)))
		for _, node := range group.Nodes {
			fields = append(fields, node.ID, node.Orbit, node.OrbitIndex, uint32(len(node.Out)))
			for i, out := range node.Out {
				fields = append(fields, out)
				if game == PoE2 {
					fields = append(fields, node.OutOrbit[i])
				}
			}
		}
	}
	return psgBytes(t, fields...)
}

func TestParseLivePassiveSkillGraph(t *testing.T) {
	requireLiveSchema(t)
	testLivePassiveSkillGraph(t, PoE1, testdata.GameVersion())
}

func TestParseLivePassiveSkillGraphPoE2(t *testing.T) {
	patch, err := LatestPatch(PoE2)
	if err != nil {
		t.Skipf("no live poe2 patch: %v", err)
	}
	testLivePassiveSkillGraph(t, PoE2, patch)
}

func testLivePassiveSkillGraph(t *testing.T, game Game, patch string) {
	t.Helper()

	loader, err := GetBundleLoader(NewWebFS(game, patch), true)
	testza.AssertNoError(t, err)

	f, err := loader.Open("metadata/passiveskillgraph.psg")
	testza.AssertNoError(t, err)
	raw, err := io.ReadAll(f)
	testza.AssertNoError(t, err)

	g, err := ParsePassiveSkillGraph(bytes.NewReader(raw), game)
	testza.AssertNoError(t, err)
	testza.AssertGreater(t, len(g.Roots), 0)
	testza.AssertGreater(t, len(g.Groups), 0)
	testza.AssertEqual(t, raw, encodePSG(t, g, game))

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
