package core

import "math"

// NestKey is the node identifier for the Nest (origin) in the delivery graph.
const NestKey = "__nest__"

// EdgeWeight is the seam between the delivery graph and the cost model used by
// the scheduler. The default implementation returns straight-line Euclidean
// distance in meters.
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

// NewGraphWithEdgeWeight builds a graph with a pluggable EdgeWeight strategy.
// Pass nil to fall back to Euclidean distance.
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
// distance in meters between two node coordinates. Distances are precomputed
// at construction so Weight is an O(1) map lookup rather than a sqrt per call.
type euclideanEdgeWeight struct {
	edges map[string]map[string]float64
}

func newEuclideanEdgeWeight(nodes map[string]Hospital) EdgeWeight {
	edges := make(map[string]map[string]float64, len(nodes))
	for fromName, fromNode := range nodes {
		row := make(map[string]float64, len(nodes))
		for toName, toNode := range nodes {
			dn := float64(fromNode.NorthM - toNode.NorthM)
			de := float64(fromNode.EastM - toNode.EastM)
			row[toName] = math.Sqrt(dn*dn + de*de)
		}
		edges[fromName] = row
	}
	return &euclideanEdgeWeight{edges: edges}
}

func (strategy *euclideanEdgeWeight) Weight(from string, to string) float64 {
	row, ok := strategy.edges[from]
	if !ok {
		panic("core.euclideanEdgeWeight: unknown node " + from)
	}
	weight, ok := row[to]
	if !ok {
		panic("core.euclideanEdgeWeight: unknown node " + to)
	}
	return weight
}
