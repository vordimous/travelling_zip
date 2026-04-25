package core

import (
	"math"
	"sort"
	"testing"
)

func sampleHospitals() map[string]Hospital {
	return map[string]Hospital{
		"Alpha": {Name: "Alpha", NorthM: 3000, EastM: 4000},
		"Beta":  {Name: "Beta", NorthM: -6000, EastM: 8000},
		"Gamma": {Name: "Gamma", NorthM: 0, EastM: 10000},
	}
}

func TestNewGraphContainsNestAndAllHospitals(t *testing.T) {
	graph := NewGraph(sampleHospitals())

	got := graph.Nodes()
	sort.Strings(got)
	want := []string{NestKey, "Alpha", "Beta", "Gamma"}
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("Nodes() length = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("Nodes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestEdgeWeightFromNestIsEuclidean(t *testing.T) {
	graph := NewGraph(sampleHospitals())
	// Alpha is at (3000, 4000): distance from origin = 5000.
	got := graph.EdgeWeight(NestKey, "Alpha")
	if math.Abs(got-5000) > 1e-9 {
		t.Errorf("EdgeWeight(Nest,Alpha) = %v, want 5000", got)
	}
}

func TestEdgeWeightBetweenHospitalsIsEuclidean(t *testing.T) {
	graph := NewGraph(sampleHospitals())
	// Alpha (3000,4000) ↔ Beta (-6000,8000): sqrt(9000^2 + 4000^2) = sqrt(97_000_000)
	want := math.Sqrt(9000*9000 + 4000*4000)
	got := graph.EdgeWeight("Alpha", "Beta")
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("EdgeWeight(Alpha,Beta) = %v, want %v", got, want)
	}
}

func TestEdgeWeightIsSymmetric(t *testing.T) {
	graph := NewGraph(sampleHospitals())
	for _, pair := range [][2]string{
		{NestKey, "Alpha"},
		{"Alpha", "Beta"},
		{"Beta", "Gamma"},
		{NestKey, "Gamma"},
	} {
		ab := graph.EdgeWeight(pair[0], pair[1])
		ba := graph.EdgeWeight(pair[1], pair[0])
		if math.Abs(ab-ba) > 1e-9 {
			t.Errorf("asymmetric edge %v: %v vs %v", pair, ab, ba)
		}
	}
}

func TestEdgeWeightSelfLoopIsZero(t *testing.T) {
	graph := NewGraph(sampleHospitals())
	if got := graph.EdgeWeight("Alpha", "Alpha"); got != 0 {
		t.Errorf("EdgeWeight(Alpha,Alpha) = %v, want 0", got)
	}
	if got := graph.EdgeWeight(NestKey, NestKey); got != 0 {
		t.Errorf("EdgeWeight(Nest,Nest) = %v, want 0", got)
	}
}

func TestHasNode(t *testing.T) {
	graph := NewGraph(sampleHospitals())
	if !graph.HasNode(NestKey) {
		t.Error("HasNode(NestKey) = false, want true")
	}
	if !graph.HasNode("Alpha") {
		t.Error("HasNode(Alpha) = false, want true")
	}
	if graph.HasNode("NotAHospital") {
		t.Error("HasNode(NotAHospital) = true, want false")
	}
}

func TestEdgeWeightPanicsOnUnknownNode(t *testing.T) {
	graph := NewGraph(sampleHospitals())
	defer func() {
		if recover() == nil {
			t.Error("EdgeWeight on unknown node did not panic")
		}
	}()
	_ = graph.EdgeWeight("Alpha", "Ghost")
}
