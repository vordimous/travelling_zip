package core

import (
	"sort"
	"testing"
)

// These diagnostics print under `make test-verbose-backend` and are quiet
// during the regular suite. They are not assertions — they exist to make the
// graph and the routing decisions inspectable while the scheduler is being
// built leaf-by-leaf.

func loadRealHospitals(t *testing.T) map[string]Hospital {
	t.Helper()
	// `make test-verbose-backend` runs from `backend/`, so `go test` cwd is
	// `backend/core`. Walk up two and into data/.
	return LoadHospitals("../../data/hospitals.csv")
}

func newDiagScheduler(t *testing.T) *ZipScheduler {
	t.Helper()
	return NewZipScheduler(loadRealHospitals(t), DefaultConfig())
}

func TestDiagnosticGraphShape(t *testing.T) {
	hospitals := loadRealHospitals(t)
	graph := NewGraph(hospitals)

	nodes := graph.Nodes()
	sort.Strings(nodes)
	t.Logf("graph: %d nodes (Nest + %d hospitals)", len(nodes), len(nodes)-1)

	t.Log("nest→hospital distances (sorted nearest first):")
	type leg struct {
		name string
		m    float64
	}
	legs := make([]leg, 0, len(hospitals))
	for name := range hospitals {
		legs = append(legs, leg{name, graph.EdgeWeight(NestKey, name)})
	}
	sort.Slice(legs, func(i, j int) bool { return legs[i].m < legs[j].m })
	for _, l := range legs {
		h := hospitals[l.name]
		t.Logf("  %-12s (%6d, %6d)  %.0f m", l.name, h.NorthM, h.EastM, l.m)
	}

	hospitalNames := make([]string, 0, len(hospitals))
	for name := range hospitals {
		hospitalNames = append(hospitalNames, name)
	}
	sort.Strings(hospitalNames)

	t.Logf("sample cross-hospital distances:")
	for i := 0; i < 4 && i+1 < len(hospitalNames); i++ {
		a, b := hospitalNames[i], hospitalNames[len(hospitalNames)-1-i]
		t.Logf("  %s ↔ %s = %.0f m", a, b, graph.EdgeWeight(a, b))
	}
}

func TestDiagnosticBuildFlightScenarios(t *testing.T) {
	scheduler := newDiagScheduler(t)

	scenarios := []struct {
		name      string
		stopNames []string // hospital names; one Resupply order per name
	}{
		{
			"three close hospitals in suboptimal candidate order",
			[]string{"Kibilizi", "Butaro", "Gikonko"},
		},
		{
			"three orders to the same hospital — should collapse to one stop",
			[]string{"Nyanza", "Nyanza", "Nyanza"},
		},
		{
			"spread-out trio that benefits most from NN reordering",
			[]string{"Byumba", "Gitwe", "Kinihira"},
		},
		{
			"long-range trio: candidate order may exceed range, NN may rescue",
			[]string{"Bigogwe", "Butaro", "Mugonero"},
		},
	}

	for index, scenario := range scenarios {
		candidates := make([]Order, 0, len(scenario.stopNames))
		for i, name := range scenario.stopNames {
			candidates = append(candidates, Order{
				ID:           "diag-" + scenario.name + "-" + string(rune('A'+i)),
				Time:         0,
				HospitalName: name,
				Priority:     Resupply,
			})
		}

		fifoDistance := scheduler.routeDistance(scenario.stopNames)
		nnRoute := scheduler.nearestNeighborOrder(uniqueStops(scenario.stopNames))
		nnDistance := scheduler.routeDistance(nnRoute)

		flight, leftover := scheduler.buildFlight(0, candidates)

		t.Logf("scenario %d: %s", index+1, scenario.name)
		t.Logf("  candidates:        %v", scenario.stopNames)
		t.Logf("  fifo route dist:   %.0f m  (over range = %v)", fifoDistance, fifoDistance > float64(scheduler.zipMaxCumulativeRangeM))
		t.Logf("  unique stops:      %v", uniqueStops(scenario.stopNames))
		t.Logf("  nn-ordered route:  %v  dist=%.0f m  saved=%.0f m", nnRoute, nnDistance, fifoDistance-nnDistance)
		t.Logf("  buildFlight ->     stops=%v, orderIDs=%d, leftover=%d",
			flight.HospitalNames, len(flight.OrderIDs), len(leftover))
		t.Logf("  flight route dist: %.0f m", scheduler.routeDistance(flight.HospitalNames))
	}
}

func uniqueStops(stops []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range stops {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
