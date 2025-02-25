package neostr

import (
	"fmt"
	"sort"
	"testing"
)

// ========================================
// Set Tests
// ========================================

func TestSetString(t *testing.T) {
	set := NewSet("apple", "banana", "carrot")
	want := "Set{apple banana carrot}"
	formatted := fmt.Sprintf("%v", set)
	if formatted != want {
		t.Errorf("String formatted Set = %s; want %s", formatted, want)
	}
}

func TestNewSet(t *testing.T) {
	want := [3]int{1, 2, 3}
	set := NewSet(1, 2, 3)
	for _, item := range want {
		if !set.Contains(item) {
			t.Errorf("Expected item %d in set %v", item, set)
		}
	}
}

func TestSetAdd(t *testing.T) {
	set := NewSet[int]()
	item := 5
	set.Add(5)
	if !set.Contains(item) {
		t.Errorf("Failed to add item %d to set %v", item, set)
	}
}

func TestSetRemove(t *testing.T) {
	set := NewSet(1, 2, 3)
	item := 2
	set.Remove(2)
	if set.Contains(item) {
		t.Errorf("Failed to remove item %d from set %v", item, set)
	}
}

func TestContainsNonExistingItem(t *testing.T) {
	set := NewSet[int]()
	item := 5
	if set.Contains(item) {
		t.Errorf("Expected set %v not to contain nonexisting item %d", set, item)
	}
}

func TestToArray(t *testing.T) {
	set := NewSet(1, 2, 3)
	want := [3]int{1, 2, 3}
	array := set.ToArray()
	sortedArray := array[:]
	sort.Ints(sortedArray)
	for index, item := range sortedArray {
		if item != want[index] {
			t.Errorf("Expected ToArray() to have the same elements as %v; got %v", want, array)
		}
	}
}

func TestSetFromArray(t *testing.T) {
	set := NewSet(1, 2, 3)
	array := set.ToArray()
	newSet := NewSet(array...)
	for item := range set.inner {
		if !newSet.Contains(item) {
			t.Errorf("Expected newSet %v to have the same as the original %v", newSet, set)
		}
	}
}

// ========================================
// Flatten Tests
// ========================================

func TestFlattenMapping(t *testing.T) {
	mapping := map[int][]int{
		1: {5, 6},
		2: {7, 8},
	}
	want := [4]int{5, 6, 7, 8}
	flattened := Flatten(mapping)
	sortedFlattened := flattened[:]
	sort.Ints(sortedFlattened)
	for index, item := range sortedFlattened {
		if item != want[index] {
			t.Errorf("Expected flattened map to have the same elements as %v; got %v", want, flattened)
		}
	}
}

func TestFlattenEmptyMapping(t *testing.T) {
	mapping := map[int][]int{}
	flattened := Flatten(mapping)
	if len(flattened) != 0 {
		t.Errorf("Expected Flatten of an empty map to be an empty slice; got %v", flattened)
	}
}
