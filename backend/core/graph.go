package core

import "math"

// NestKey is the node identifier for the Nest (origin) in the delivery graph.
const NestKey = "__nest__"

// Graph is the delivery network: the Nest plus every hospital, fully connected,
// with edge weights equal to straight-line Euclidean distance in meters.
//
// Step 2a will swap the edge-weight calculation for a pluggable strategy; for
// Step 1 the weight is pure geometry.
type Graph struct {
	nodes map[string]Hospital
}

// NewGraph builds a graph containing the Nest at (0,0) and every hospital from
// the given map. The input map is not mutated.
func NewGraph(hospitals map[string]Hospital) *Graph {
	nodes := make(map[string]Hospital, len(hospitals)+1)
	nodes[NestKey] = Hospital{Name: NestKey}
	for name, hospital := range hospitals {
		nodes[name] = hospital
	}
	return &Graph{nodes: nodes}
}

// Nodes returns the node identifiers in the graph (Nest + hospitals).
func (graph *Graph) Nodes() []string {
	names := make([]string, 0, len(graph.nodes))
	for name := range graph.nodes {
		names = append(names, name)
	}
	return names
}

// HasNode reports whether the named node exists in the graph.
func (graph *Graph) HasNode(name string) bool {
	_, ok := graph.nodes[name]
	return ok
}

// EdgeWeight returns the Euclidean distance in meters between two nodes.
// Returns 0 when from == to. Panics if either node is unknown — the graph
// contract is that callers ask only about nodes they got from Nodes().
func (graph *Graph) EdgeWeight(from string, to string) float64 {
	a, ok := graph.nodes[from]
	if !ok {
		panic("core.Graph.EdgeWeight: unknown node " + from)
	}
	b, ok := graph.nodes[to]
	if !ok {
		panic("core.Graph.EdgeWeight: unknown node " + to)
	}
	dn := float64(a.NorthM - b.NorthM)
	de := float64(a.EastM - b.EastM)
	return math.Sqrt(dn*dn + de*de)
}
