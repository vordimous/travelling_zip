package core

import "math"

// NestKey is the node identifier for the Nest (origin) in the delivery graph.
const NestKey = "__nest__"

// EdgeWeight is the seam between the delivery graph and the cost model used by
// the scheduler. The default implementation returns straight-line Euclidean
// distance in meters; alternative implementations (Step 2a / C3) may layer in
// real-world considerations such as terrain or wind.
//
// Implementations must be symmetric (Weight(a,b) == Weight(b,a)) and should
// return 0 when from == to.
type EdgeWeight interface {
	Weight(from string, to string) float64
}

// Graph is the delivery network: the Nest plus every hospital, fully connected.
// Edge cost is delegated to a pluggable EdgeWeight strategy; the default is
// Euclidean distance.
type Graph struct {
	nodes      map[string]Hospital
	edgeWeight EdgeWeight
}

// NewGraph builds a graph containing the Nest at (0,0) and every hospital from
// the given map, using Euclidean distance for edge weights.
func NewGraph(hospitals map[string]Hospital) *Graph {
	return NewGraphWithEdgeWeight(hospitals, nil)
}

// NewGraphWithEdgeWeight builds a graph and lets the caller supply an
// alternative EdgeWeight implementation. Pass nil to fall back to the default
// Euclidean strategy.
func NewGraphWithEdgeWeight(hospitals map[string]Hospital, edgeWeight EdgeWeight) *Graph {
	nodes := make(map[string]Hospital, len(hospitals)+1)
	nodes[NestKey] = Hospital{Name: NestKey}
	for name, hospital := range hospitals {
		nodes[name] = hospital
	}
	if edgeWeight == nil {
		edgeWeight = newEuclideanEdgeWeight(nodes)
	}
	return &Graph{nodes: nodes, edgeWeight: edgeWeight}
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

// EdgeWeight returns the cost of the edge between two nodes under the graph's
// configured strategy. Returns 0 when from == to. Panics if either node is
// unknown — the graph contract is that callers ask only about nodes they got
// from Nodes().
func (graph *Graph) EdgeWeight(from string, to string) float64 {
	return graph.edgeWeight.Weight(from, to)
}

// euclideanEdgeWeight is the default EdgeWeight strategy: straight-line
// distance in meters between two node coordinates.
type euclideanEdgeWeight struct {
	nodes map[string]Hospital
}

func newEuclideanEdgeWeight(nodes map[string]Hospital) EdgeWeight {
	return &euclideanEdgeWeight{nodes: nodes}
}

func (strategy *euclideanEdgeWeight) Weight(from string, to string) float64 {
	a, ok := strategy.nodes[from]
	if !ok {
		panic("core.euclideanEdgeWeight: unknown node " + from)
	}
	b, ok := strategy.nodes[to]
	if !ok {
		panic("core.euclideanEdgeWeight: unknown node " + to)
	}
	dn := float64(a.NorthM - b.NorthM)
	de := float64(a.EastM - b.EastM)
	return math.Sqrt(dn*dn + de*de)
}
