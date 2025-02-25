package neostr

import (
	"reflect"
	"testing"
)

// ========================================
// Test Data Providers
// ========================================

func testMatchKeys() *MatchKeys {
	return &MatchKeys{
		Keys: map[string][]string{
			"User":   {"name"},
			"Action": {"kind", "timestamp"},
		},
	}
}

// ========================================
// MatchKeysProvider Tests
// ========================================

func TestMatchKeysGetLabels(t *testing.T) {
	matchKeys := testMatchKeys()

	want := []string{"Action", "User"}
	labels := matchKeys.GetLabels()

	if !reflect.DeepEqual(labels, want) {
		t.Errorf("Expected MatchKeys labels %v, got %v", want, labels)
	}
}

func TestMatchKeysGetKeys(t *testing.T) {
	matchKeys := testMatchKeys()

	want := []string{"name"}
	keys, exists := matchKeys.GetKeys("User")

	if !exists {
		t.Errorf("Expected 'User' label to exist")
	}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("Expected keys %v, got %v", want, keys)
	}
}

func TestMatchKeysGetKeysNonExistent(t *testing.T) {
	matchKeys := testMatchKeys()

	keys, exists := matchKeys.GetKeys("NonExistent")

	if exists {
		t.Errorf("Expected 'NonExistent' label to not exist")
	}
	if keys != nil {
		t.Errorf("Expected nil keys for non-existent label, got %v", keys)
	}
}

// ========================================
// Node Tests
// ========================================

func TestNewNode(t *testing.T) {
	props := Properties{"name": "john", "role": "admin"}
	node := NewNode("User", props)

	if !node.Labels.Contains("User") {
		t.Errorf("Expected node to have 'User' label")
	}

	if node.Props["name"] != "john" || node.Props["role"] != "admin" {
		t.Errorf("Node properties not set correctly, got %v", node.Props)
	}
}

func TestNewNodeNilProps(t *testing.T) {
	node := NewNode("User", nil)

	if node.Props == nil {
		t.Errorf("Expected non-nil Properties when nil is provided")
	}

	if len(node.Props) != 0 {
		t.Errorf("Expected empty Properties when nil is provided, got %v",
			node.Props)
	}
}

func TestNodeMatchPropsSingleLabel(t *testing.T) {
	matchKeys := testMatchKeys()

	node := NewNode("User", Properties{"name": "john", "role": "admin"})
	label, props, err := node.MatchProps(matchKeys)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if label != "User" {
		t.Errorf("Expected label 'User', got %s", label)
	}
	if !reflect.DeepEqual(props, Properties{"name": "john"}) {
		t.Errorf("Expected match props {name: 'john'}, got %v", props)
	}
}

func TestNodeMatchPropsMultipleLabels(t *testing.T) {
	matchKeys := testMatchKeys()

	node := &Node{
		Labels: NewSet("User", "InactiveUser"),
		Props:  Properties{"name": "jane", "role": "user"},
	}
	label, props, err := node.MatchProps(matchKeys)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if label != "User" {
		t.Errorf("Expected label 'User', got %s", label)
	}
	if !reflect.DeepEqual(props, Properties{"name": "jane"}) {
		t.Errorf("Expected match props {name: 'jane'}, got %v", props)
	}

}

func TestNodeMatchPropsMissingProperty(t *testing.T) {
	matchKeys := testMatchKeys()

	node := NewNode("User", Properties{"id": 3, "role": "admin"}) // missing "name"
	_, _, err := node.MatchProps(matchKeys)

	if err == nil {
		t.Error("Expected error for missing property, got nil")
	}

}
func TestNodeMatchPropsUnrecognizedLabel(t *testing.T) {
	matchKeys := testMatchKeys()

	node := NewNode("Server", Properties{"ipv6": "abcd", "name": "server1"})
	_, _, err := node.MatchProps(matchKeys)

	if err == nil {
		t.Error("Expected error for unrecognized label, got nil")
	}
}

func TestNodeSerialize(t *testing.T) {
	props := Properties{"name": "john", "role": "admin"}
	node := NewNode("User", props)

	serialized := node.Serialize()

	if !reflect.DeepEqual(*serialized, props) {
		t.Errorf("Expected serialized node to equal its properties, got %v", *serialized)
	}
}

// ========================================
// Relationship Tests
// ========================================

func TestNewRelationship(t *testing.T) {
	// Create nodes
	startNode := NewNode("User", Properties{"name": "john"})
	endNode := NewNode("Action", Properties{"kind": "login", "timestamp": 100})

	// Test with properties
	props := Properties{"year": 2022}
	rel := NewRelationship("PERFORMED", startNode, endNode, props)

	if rel.Type != "PERFORMED" {
		t.Errorf("Expected relationship type 'PERFORMED', got %s", rel.Type)
	}
	if rel.Start != startNode {
		t.Errorf("Expected start node to match")
	}
	if rel.End != endNode {
		t.Errorf("Expected end node to match")
	}
	if rel.Props["year"] != 2022 {
		t.Errorf("Relationship properties not set correctly, got %v", rel.Props)
	}

	// Test with nil properties
	relWithNilProps := NewRelationship("PERFORMED", startNode, endNode, nil)
	if relWithNilProps.Props == nil {
		t.Errorf("Expected non-nil Properties when nil is provided")
	}
	if len(relWithNilProps.Props) != 0 {
		t.Errorf("Expected empty Properties when nil is provided, got %v", relWithNilProps.Props)
	}
}

func TestRelationshipSerialize(t *testing.T) {
	startNode := NewNode("User", Properties{"name": "john"})
	endNode := NewNode("Action", Properties{"kind": "login", "timestamp": 100})
	props := Properties{"year": 2022}

	rel := NewRelationship("PERFORMED", startNode, endNode, props)

	serialized := rel.Serialize()

	expectedSerialized := &SerializedRel{
		"props": props,
		"start": startNode.Props,
		"end":   endNode.Props,
	}

	if !reflect.DeepEqual(*serialized, *expectedSerialized) {
		t.Errorf("Expected serialized relationship\n%v\n, got\n%v", *expectedSerialized, *serialized)
	}
}

// ========================================
// Subgraph Tests
// ========================================

func TestNewSubgraph(t *testing.T) {
	subgraph := NewSubgraph()

	if subgraph.nodes == nil {
		t.Error("Expected nodes slice to be initialized")
	}
	if subgraph.rels == nil {
		t.Error("Expected relationships slice to be initialized")
	}
	if len(subgraph.nodes) != 0 || len(subgraph.rels) != 0 {
		t.Error("Expected empty slices for new subgraph")
	}
}

func TestSubgraphAddNode(t *testing.T) {
	subgraph := NewSubgraph()
	node := NewNode("User", nil)

	subgraph.AddNode(node)

	if len(subgraph.nodes) != 1 {
		t.Errorf("Expected 1 node in subgraph, got %d", len(subgraph.nodes))
	}
	if subgraph.nodes[0] != node {
		t.Error("Node not added correctly to subgraph")
	}
}

func TestSubgraphAddRel(t *testing.T) {
	subgraph := NewSubgraph()
	startNode := NewNode("User", nil)
	endNode := NewNode("Action", nil)
	rel := NewRelationship("PERFORMED", startNode, endNode, nil)

	subgraph.AddRel(rel)

	if len(subgraph.rels) != 1 {
		t.Errorf("Expected 1 relationship in subgraph, got %d", len(subgraph.rels))
	}
	if subgraph.rels[0] != rel {
		t.Error("Relationship not added correctly to subgraph")
	}
}

// ========================================
// StructuredSubgraph Tests
// ========================================

func TestNewStructuredSubgraph(t *testing.T) {
	matchKeys := &MatchKeys{
		Keys: map[string][]string{
			"User": {"name"},
		},
	}

	subgraph := NewStructuredSubgraph(matchKeys)

	if subgraph.nodes == nil {
		t.Error("Expected nodes map to be initialized")
	}
	if subgraph.rels == nil {
		t.Error("Expected relationships map to be initialized")
	}
	if subgraph.matchProvider != matchKeys {
		t.Error("Expected matchProvider to be set correctly")
	}
}

func TestStructuredSubgraphAddNode(t *testing.T) {
	matchKeys := testMatchKeys()

	subgraph := NewStructuredSubgraph(matchKeys)
	node := NewNode("User", Properties{"name": "john"})

	subgraph.AddNode(node)

	// Expected sort key: "User:User"
	sortKey := "User:User"

	if len(subgraph.nodes) != 1 {
		t.Errorf("Expected 1 node group in subgraph, got %d", len(subgraph.nodes))
	}

	nodes, exists := subgraph.nodes[sortKey]
	if !exists {
		t.Errorf("Expected node group with key %s to exist", sortKey)
	}
	if len(nodes) != 1 || nodes[0] != node {
		t.Error("Node not added correctly to structured subgraph")
	}
}

func TestStructuredSubgraphAddNodePanics(t *testing.T) {
	matchKeys := testMatchKeys()

	subgraph := NewStructuredSubgraph(matchKeys)
	// Test node with missing match property
	nodeMissingProp := NewNode("User", Properties{"id": 2})

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected AddNode to panic with missing property")
		}
	}()

	subgraph.AddNode(nodeMissingProp) // Should panic
}

func TestStructuredSubgraphAddRel(t *testing.T) {
	matchKeys := testMatchKeys()

	subgraph := NewStructuredSubgraph(matchKeys)
	startNode := NewNode("User", Properties{"name": "john"})
	endNode := NewNode("Action", Properties{"kind": "login", "timestamp": 100})
	rel := NewRelationship("PERFORMED", startNode, endNode, nil)

	subgraph.AddRel(rel)

	// Expected sort key: "PERFORMED,User,Action"
	sortKey := "PERFORMED,User,Action"

	if len(subgraph.rels) != 1 {
		t.Errorf("Expected 1 relationship group in subgraph, got %d", len(subgraph.rels))
	}

	rels, exists := subgraph.rels[sortKey]
	if !exists {
		t.Errorf("Expected relationship group with key %s to exist", sortKey)
	}
	if len(rels) != 1 || rels[0] != rel {
		t.Error("Relationship not added correctly to structured subgraph")
	}
}

func TestStructuredSubgraphAddRelPanics(t *testing.T) {
	matchKeys := testMatchKeys()

	subgraph := NewStructuredSubgraph(matchKeys)
	// Test relationship with start node missing match property
	startNodeMissingProp := NewNode("User", Properties{"id": 1})
	endNode := NewNode("Action", Properties{"kind": "login", "timestamp": 100})
	relWithInvalidStart := NewRelationship("PERFORMED", startNodeMissingProp, endNode, nil)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected AddRel to panic with invalid start node")
		}
	}()

	subgraph.AddRel(relWithInvalidStart) // Should panic
}

func TestStructuredSubgraphGetters(t *testing.T) {
	matchKeys := testMatchKeys()

	subgraph := NewStructuredSubgraph(matchKeys)

	user1 := NewNode("User", Properties{"name": "john"})
	action1 := NewNode("Action", Properties{"kind": "login", "timestamp": 100})
	action2 := NewNode("Action", Properties{"kind": "login", "timestamp": 200})

	rel1 := NewRelationship("PERFORMED", user1, action1, nil)
	rel2 := NewRelationship("PERFORMED", user1, action2, nil)

	subgraph.AddNode(user1)
	subgraph.AddNode(action1)
	subgraph.AddNode(action2)
	subgraph.AddRel(rel1)
	subgraph.AddRel(rel2)

	// Test GetNodes
	personNodes := subgraph.GetNodes("User:User")
	if len(personNodes) != 1 {
		t.Errorf("Expected 1 User node, got %d", len(personNodes))
	}

	// Test GetRels
	createdRels := subgraph.GetRels("PERFORMED,User,Action")
	if len(createdRels) != 2 {
		t.Errorf("Expected 2 PERFORMED relationships, got %d", len(createdRels))
	}

	// Test NodeCount
	if count := subgraph.NodeCount(); count != 3 {
		t.Errorf("Expected NodeCount() = 3, got %d", count)
	}

	// Test RelCount
	if count := subgraph.RelCount(); count != 2 {
		t.Errorf("Expected RelCount() = 2, got %d", count)
	}

	// Test NodeKeys
	nodeKeys := subgraph.NodeKeys()
	if len(nodeKeys) != 2 { // User:User and Action:Action
		t.Errorf("Expected 2 node keys, got %d: %v", len(nodeKeys), nodeKeys)
	}

	// Test RelKeys
	relKeys := subgraph.RelKeys()
	if len(relKeys) != 1 { // PERFORMED,User,Action
		t.Errorf("Expected 2 relationship keys, got %d: %v", len(relKeys), relKeys)
	}
}

// ========================================
// Sort Key Tests
// ========================================

func TestCreateNodeSortKey(t *testing.T) {
	key := createNodeSortKey("User", []string{"User", "InactiveUser"})
	expected := "User:InactiveUser,User"

	if key != expected {
		t.Errorf("Expected node sort key '%s', got '%s'", expected, key)
	}
}

func TestCreateRelSortKey(t *testing.T) {
	key := createRelSortKey("PERFORMED", "User", "Action")
	expected := "PERFORMED,User,Action"

	if key != expected {
		t.Errorf("Expected relationship sort key '%s', got '%s'", expected, key)
	}
}

func TestDeserializeNodeKey(t *testing.T) {
	matchLabel, labels := DeserializeNodeKey("User:InactiveUser,User")

	if matchLabel != "User" {
		t.Errorf("Expected match label 'User', got '%s'", matchLabel)
	}

	expectedLabels := []string{"InactiveUser", "User"}
	if !reflect.DeepEqual(labels, expectedLabels) {
		t.Errorf("Expected labels %v, got %v", expectedLabels, labels)
	}
}

func TestDeserializeNodeKeyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected DeserializeNodeKey to panic with invalid key")
		}
	}()

	DeserializeNodeKey("InvalidKey") // Should panic
}

func TestDeserializeRelKey(t *testing.T) {
	rtype, startLabel, endLabel := DeserializeRelKey("PERFORMED,User,Action")

	if rtype != "PERFORMED" {
		t.Errorf("Expected relationship type 'PERFORMED', got '%s'", rtype)
	}
	if startLabel != "User" {
		t.Errorf("Expected start label 'User', got '%s'", startLabel)
	}
	if endLabel != "Action" {
		t.Errorf("Expected end label 'Action', got '%s'", endLabel)
	}
}

func TestDeserializeRelKeyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected DeserializeRelKey to panic with invalid key")
		}
	}()

	DeserializeRelKey("InvalidKey") // Should panic
}
