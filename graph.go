// This module defines types and functions for working with Neo4j graph
// entities.

package neostr

import (
	"fmt"
	"sort"
	"strings"
)

// ========================================
// Types
// ========================================

// Properties represents a map of node or relationship props.
type Properties map[string]any

// ========================================
// Match Key Provider
// ========================================

// MatchKeysProvider defines methods for querying a mapping of node labels and
// the property keys used to match nodes with them.
type MatchKeysProvider interface {
	// GetLabels returns the array of node labels in the mapping.
	GetLabels() []string

	// GetKeys returns the node property keys used to match nodes with the
	// given label and a boolean indicating the success of the lookup.
	GetKeys(label string) ([]string, bool)
}

// MatchKeys is a simple implementation of the MatchKeysProvider interface.
type MatchKeys struct {
	Keys map[string][]string
}

func (p *MatchKeys) GetLabels() []string {
	labels := []string{}
	for l := range p.Keys {
		labels = append(labels, l)
	}
	sort.Strings(labels)
	return labels
}

func (p *MatchKeys) GetKeys(label string) ([]string, bool) {
	if keys, exists := p.Keys[label]; exists {
		return keys, exists
	} else {
		return nil, exists
	}
}

// ========================================
// Nodes
// ========================================

// Node represents a Neo4j node entity, encapsulating its labels and
// properties.
type Node struct {
	// Set of labels on the node.
	Labels Set[string]
	// Mapping of properties on the node.
	Props Properties
}

// NewNode creates a new node with the given label and properties.
func NewNode(label string, props Properties) *Node {
	if props == nil {
		props = make(Properties)
	}
	return &Node{
		Labels: NewSet(label),
		Props:  props,
	}
}

// MatchProps returns the node label and the property values to match it in the
// database.
func (n *Node) MatchProps(
	matchProvider MatchKeysProvider) (string, Properties, error) {

	// Iterate over each label on the node, checking whether each has match
	// keys associated with it.
	labels := n.Labels.ToArray()
	sort.Strings(labels)
	for _, label := range labels {
		if keys, exists := matchProvider.GetKeys(label); exists {
			props := make(Properties)

			// Get the property values associated with each match key.
			for _, key := range keys {
				if value, exists := n.Props[key]; exists {
					props[key] = value
				} else {

					// If any match property values are missing, return an
					// error.
					return label, nil,
						fmt.Errorf(
							"missing property %s for label %s", key, label)
				}
			}

			// Return the label and match properties
			return label, props, nil
		}
	}

	// If none of the node labels have defined match keys, return an error.
	return "", nil, fmt.Errorf("no recognized label found in %v", n.Labels)
}

type SerializedNode = Properties

func (n *Node) Serialize() *SerializedNode {
	return &n.Props
}

// ========================================
// Relationships
// ========================================

// Relationship represents a Neo4j relationship between two nodes, including
// its type and properties.
type Relationship struct {
	// The relationship type.
	Type string
	// The start node for the relationship.
	Start *Node
	// The end node for the relationship.
	End *Node
	// Mapping of properties on the relationship
	Props Properties
}

// NewRelationship creates a new relationship with the given type, start node,
// end node, and properties
func NewRelationship(
	rtype string, start *Node, end *Node, props Properties) *Relationship {

	if props == nil {
		props = make(Properties)
	}
	return &Relationship{
		Type:  rtype,
		Start: start,
		End:   end,
		Props: props,
	}
}

type SerializedRel = map[string]Properties

func (r *Relationship) Serialize() *SerializedRel {
	srel := make(map[string]Properties)
	srel["props"] = r.Props
	srel["start"] = r.Start.Props
	srel["end"] = r.End.Props
	return &srel
}

// ========================================
// Simple Subgraph
// ========================================

// Subgraph represents a simple collection of nodes and relationships.
type Subgraph struct {
	// The nodes in the subgraph.
	nodes []*Node
	// The relationships in the subgraph.
	rels []*Relationship
}

// NewSubgraph creates an empty subgraph.
func NewSubgraph() *Subgraph {
	return &Subgraph{
		nodes: []*Node{},
		rels:  []*Relationship{},
	}
}

// AddNode adds a node to the subgraph
func (s *Subgraph) AddNode(node *Node) {
	s.nodes = append(s.nodes, node)
}

// AddRel adds a relationship to the subgraph.
func (s *Subgraph) AddRel(rel *Relationship) {
	s.rels = append(s.rels, rel)
}

// ========================================
// Structured Subgraph
// ========================================

// StructuredSubgraph is a structured collection of nodes and relationships for
// the purpose of conducting batch operations.
type StructuredSubgraph struct {
	// A map of grouped nodes, sorted by their label combinations.
	nodes map[string][]*Node
	// A map of grouped relationships, sorted by their type and related node
	// labels.
	rels map[string][]*Relationship
	// Provides node property keys used to match nodes with given labels in the
	// database.
	matchProvider MatchKeysProvider
}

// NewStructuredSubgraph creates an empty structured subgraph with the given
// match keys provider.
func NewStructuredSubgraph(matchProvider MatchKeysProvider) *StructuredSubgraph {
	return &StructuredSubgraph{
		nodes:         make(map[string][]*Node),
		rels:          make(map[string][]*Relationship),
		matchProvider: matchProvider,
	}
}

// AddNode sorts a node into the subgraph.
func (s *StructuredSubgraph) AddNode(node *Node) {

	// Verify that the node has defined match property values.
	matchLabel, _, err := node.MatchProps(s.matchProvider)
	if err != nil {
		panic(fmt.Errorf("invalid node: %s", err))
	}

	// Determine the node's sort key.
	sortKey := createNodeSortKey(matchLabel, node.Labels.ToArray())

	if _, exists := s.nodes[sortKey]; !exists {
		s.nodes[sortKey] = []*Node{}
	}

	// Add the node to the subgraph.
	s.nodes[sortKey] = append(s.nodes[sortKey], node)
}

// AddRel sorts a relationship into the subgraph.
func (s *StructuredSubgraph) AddRel(rel *Relationship) {

	// Verify that the start node has defined match property values.
	startLabel, _, err := rel.Start.MatchProps(s.matchProvider)
	if err != nil {
		panic(fmt.Errorf("invalid start node: %s", err))
	}

	// Verify that the end node has defined match property values.
	endLabel, _, err := rel.End.MatchProps(s.matchProvider)
	if err != nil {
		panic(fmt.Errorf("invalid end node: %s", err))
	}

	// Determine the relationship's sort key.
	sortKey := createRelSortKey(rel.Type, startLabel, endLabel)

	if _, exists := s.rels[sortKey]; !exists {
		s.rels[sortKey] = []*Relationship{}
	}

	// Add the relationship to the subgraph.
	s.rels[sortKey] = append(s.rels[sortKey], rel)
}

// GetNodes returns the nodes grouped under the given sort key.
func (s *StructuredSubgraph) GetNodes(nodeKey string) []*Node {
	return s.nodes[nodeKey]
}

// GetRels returns the rels grouped under the given sort key.
func (s *StructuredSubgraph) GetRels(relKey string) []*Relationship {
	return s.rels[relKey]
}

// NodeCount returns the number of nodes in the subgraph.
func (s *StructuredSubgraph) NodeCount() int {
	count := 0
	for l := range s.nodes {
		count += len(s.nodes[l])
	}
	return count
}

// RelCount returns the number of relationships in the subgraph.
func (s *StructuredSubgraph) RelCount() int {
	count := 0
	for t := range s.rels {
		count += len(s.rels[t])
	}
	return count
}

// NodeKeys returns the list of node sort keys in the subgraph.
func (s *StructuredSubgraph) NodeKeys() []string {
	keys := []string{}
	for l := range s.nodes {
		keys = append(keys, l)
	}
	return keys
}

// RelKeys returns the list of relationship sort keys in the subgraph.
func (s *StructuredSubgraph) RelKeys() []string {
	keys := []string{}
	for t := range s.rels {
		keys = append(keys, t)
	}
	return keys
}

// createNodeSortKey returns the serialized node labels for sorting.
func createNodeSortKey(matchLabel string, labels []string) string {
	sort.Strings(labels)
	serializedLabels := strings.Join(labels, ",")
	return fmt.Sprintf("%s:%s", matchLabel, serializedLabels)
}

// createRelSortKey returns the serialized relationship type and start/end node
// labels for sorting.
func createRelSortKey(
	rtype string, startLabel string, endLabel string) string {
	return strings.Join([]string{rtype, startLabel, endLabel}, ",")
}

// DeserializeNodeKey returns the list of node labels from the serialized sort
// key.
func DeserializeNodeKey(sortKey string) (string, []string) {
	parts := strings.Split(sortKey, ":")
	if len(parts) != 2 {
		panic(fmt.Sprintf("invalid node sort key: %s", sortKey))
	}
	matchLabel, serializedLabels := parts[0], parts[1]
	labels := strings.Split(serializedLabels, ",")
	return matchLabel, labels
}

// DeserializeRelKey returns the relationship type, start node label, and end
// node label from the serialized sort key. Panics if the sort key is invalid.
func DeserializeRelKey(sortKey string) (string, string, string) {
	parts := strings.Split(sortKey, ",")
	if len(parts) != 3 {
		panic(fmt.Sprintf("invalid relationship sort key: %s", sortKey))
	}
	rtype, startLabel, endLabel := parts[0], parts[1], parts[2]
	return rtype, startLabel, endLabel
}
