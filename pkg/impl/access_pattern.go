package impl

import (
	"fmt"

	"github.com/gavinvolpe/nexus/pkg/types"
)

// AccessNode implements types.IAccessNode
type AccessNode struct {
	Base
	NodeType     string   `json:"node_type"`
	Requirements []string `json:"requirements"`
	Capabilities []string `json:"capabilities"`
}

func NewAccessNode(nodeType string) *AccessNode {
	return &AccessNode{
		Base:         NewBase(),
		NodeType:     nodeType,
		Requirements: make([]string, 0),
		Capabilities: make([]string, 0),
	}
}

func (n *AccessNode) GetType() string           { return n.NodeType }
func (n *AccessNode) GetRequirements() []string { return n.Requirements }
func (n *AccessNode) GetCapabilities() []string { return n.Capabilities }
func (n *AccessNode) ValidateNode() error {
	if n.NodeType == "" {
		return fmt.Errorf("node type cannot be empty")
	}
	return nil
}

// AccessEdge implements types.IAccessEdge
type AccessEdge struct {
	Base
	FromNode    types.IAccessNode   `json:"from_node"`
	ToNode      types.IAccessNode   `json:"to_node"`
	EdgeWeight  float64             `json:"weight"`
	Transform   *Transformation     `json:"transformation"`
	Constraints []types.IConstraint `json:"constraints"`
}

func NewAccessEdge(from, to types.IAccessNode, weight float64) *AccessEdge {
	return &AccessEdge{
		Base:        NewBase(),
		FromNode:    from,
		ToNode:      to,
		EdgeWeight:  weight,
		Constraints: make([]types.IConstraint, 0),
	}
}

func (e *AccessEdge) From() types.IAccessNode                  { return e.FromNode }
func (e *AccessEdge) To() types.IAccessNode                    { return e.ToNode }
func (e *AccessEdge) Weight() float64                          { return e.EdgeWeight }
func (e *AccessEdge) GetTransformation() types.ITransformation { return e.Transform }
func (e *AccessEdge) GetConstraints() []types.IConstraint      { return e.Constraints }
func (e *AccessEdge) ValidateEdge() error {
	if e.FromNode == nil || e.ToNode == nil {
		return fmt.Errorf("edge must have both from and to nodes")
	}
	return nil
}

// Transformation implements types.ITransformation
type Transformation struct {
	Base
	TransformType string                                 `json:"transform_type"`
	Cost          float64                                `json:"cost"`
	Function      func(interface{}) (interface{}, error) `json:"-"`
}

func NewTransformation(transformType string, cost float64, fn func(interface{}) (interface{}, error)) *Transformation {
	return &Transformation{
		Base:          NewBase(),
		TransformType: transformType,
		Cost:          cost,
		Function:      fn,
	}
}

func (t *Transformation) Transform(input interface{}) (interface{}, error) {
	if t.Function == nil {
		return nil, fmt.Errorf("transformation function not set")
	}
	return t.Function(input)
}

func (t *Transformation) GetType() string  { return t.TransformType }
func (t *Transformation) GetCost() float64 { return t.Cost }
func (t *Transformation) ValidateTransform() error {
	if t.TransformType == "" {
		return fmt.Errorf("transform type cannot be empty")
	}
	if t.Function == nil {
		return fmt.Errorf("transform function cannot be nil")
	}
	return nil
}

// KnowledgeGraph implements types.IKnowledgeGraph
type KnowledgeGraph struct {
	Base
	Nodes map[string]types.IAccessNode   `json:"nodes"`
	Edges map[string][]types.IAccessEdge `json:"edges"`
}

func NewKnowledgeGraph() *KnowledgeGraph {
	return &KnowledgeGraph{
		Base:  NewBase(),
		Nodes: make(map[string]types.IAccessNode),
		Edges: make(map[string][]types.IAccessEdge),
	}
}

func (g *KnowledgeGraph) AddNode(node types.IAccessNode) error {
	if err := node.ValidateNode(); err != nil {
		return fmt.Errorf("invalid node: %w", err)
	}
	g.Nodes[node.ID()] = node
	return nil
}

func (g *KnowledgeGraph) AddEdge(edge types.IAccessEdge) error {
	if err := edge.ValidateEdge(); err != nil {
		return fmt.Errorf("invalid edge: %w", err)
	}
	fromID := edge.From().ID()
	g.Edges[fromID] = append(g.Edges[fromID], edge)
	return nil
}

func (g *KnowledgeGraph) RemoveNode(nodeID string) error {
	if _, exists := g.Nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}
	delete(g.Nodes, nodeID)
	delete(g.Edges, nodeID)
	return nil
}

func (g *KnowledgeGraph) RemoveEdge(fromID, toID string) error {
	edges, exists := g.Edges[fromID]
	if !exists {
		return fmt.Errorf("no edges found for node: %s", fromID)
	}

	for i, edge := range edges {
		if edge.To().ID() == toID {
			g.Edges[fromID] = append(edges[:i], edges[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("edge not found from %s to %s", fromID, toID)
}

func (g *KnowledgeGraph) FindPath(from, to types.IAccessNode) ([]types.IAccessEdge, error) {
	// Simple BFS implementation - can be extended with more sophisticated algorithms
	visited := make(map[string]bool)
	queue := [][]types.IAccessEdge{{}}
	current := from.ID()

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		if current == to.ID() {
			return path, nil
		}

		if edges, exists := g.Edges[current]; exists {
			for _, edge := range edges {
				if !visited[edge.To().ID()] {
					visited[edge.To().ID()] = true
					newPath := make([]types.IAccessEdge, len(path))
					copy(newPath, path)
					newPath = append(newPath, edge)
					queue = append(queue, newPath)
					current = edge.To().ID()
				}
			}
		}
	}

	return nil, fmt.Errorf("no path found from %s to %s", from.ID(), to.ID())
}

func (g *KnowledgeGraph) ValidateGraph() error {
	for _, node := range g.Nodes {
		if err := node.ValidateNode(); err != nil {
			return fmt.Errorf("invalid node %s: %w", node.ID(), err)
		}
	}

	for _, edges := range g.Edges {
		for _, edge := range edges {
			if err := edge.ValidateEdge(); err != nil {
				return fmt.Errorf("invalid edge from %s to %s: %w",
					edge.From().ID(), edge.To().ID(), err)
			}
		}
	}

	return nil
}
