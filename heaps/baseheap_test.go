package heaps

import (
	"testing"
)

func TestNewMaxHeap(t *testing.T) {
	h := NewMaxHeap[string](10)

	if h == nil {
		t.Fatal("NewMaxHeap returned nil")
	}

	if h.Len() != 0 {
		t.Errorf("Expected empty heap, got length %d", h.Len())
	}

	if !h.IsEmpty() {
		t.Error("Expected heap to be empty")
	}
}

func TestNewMinHeap(t *testing.T) {
	h := NewMinHeap[string](10)

	if h == nil {
		t.Fatal("NewMinHeap returned nil")
	}

	if h.Len() != 0 {
		t.Errorf("Expected empty heap, got length %d", h.Len())
	}

	if !h.IsEmpty() {
		t.Error("Expected heap to be empty")
	}
}

func TestMaxHeapBasicOperations(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Test PushItem and Peek
	h.PushItem("a", 5)
	h.PushItem("b", 10)
	h.PushItem("c", 3)

	if h.Len() != 3 {
		t.Errorf("Expected length 3, got %d", h.Len())
	}

	// Peek should return the maximum value
	key, value, exists := h.Peek()
	if !exists {
		t.Fatal("Expected to find element")
	}
	if key != "b" || value != 10 {
		t.Errorf("Expected (b, 10), got (%s, %d)", key, value)
	}

	// Test PopItem
	key, value, exists = h.PopItem()
	if !exists {
		t.Fatal("Expected to pop element")
	}
	if key != "b" || value != 10 {
		t.Errorf("Expected (b, 10), got (%s, %d)", key, value)
	}

	if h.Len() != 2 {
		t.Errorf("Expected length 2 after pop, got %d", h.Len())
	}
}

func TestMinHeapBasicOperations(t *testing.T) {
	h := NewMinHeap[string](5)

	// Test PushItem and Peek
	h.PushItem("a", 5)
	h.PushItem("b", 10)
	h.PushItem("c", 3)

	if h.Len() != 3 {
		t.Errorf("Expected length 3, got %d", h.Len())
	}

	// Peek should return the minimum value
	key, value, exists := h.Peek()
	if !exists {
		t.Fatal("Expected to find element")
	}
	if key != "c" || value != 3 {
		t.Errorf("Expected (c, 3), got (%s, %d)", key, value)
	}

	// Test PopItem
	key, value, exists = h.PopItem()
	if !exists {
		t.Fatal("Expected to pop element")
	}
	if key != "c" || value != 3 {
		t.Errorf("Expected (c, 3), got (%s, %d)", key, value)
	}

	if h.Len() != 2 {
		t.Errorf("Expected length 2 after pop, got %d", h.Len())
	}
}

func TestIncrement(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Add initial element
	h.PushItem("a", 5)

	// Increment existing element
	h.Increment("a")

	key, value, exists := h.Peek()
	if !exists {
		t.Fatal("Expected to find element")
	}
	if key != "a" || value != 6 {
		t.Errorf("Expected (a, 6), got (%s, %d)", key, value)
	}

	// Increment non-existing element
	h.Increment("b")

	if h.Len() != 2 {
		t.Errorf("Expected length 2, got %d", h.Len())
	}

	// Check that "b" was added with value 1
	value, exists = h.GetValue("b")
	if !exists || value != 1 {
		t.Errorf("Expected b to have value 1, got %d", value)
	}
}

func TestIncrementBy(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Add initial element
	h.PushItem("a", 5)

	// Increment by 3
	h.IncrementBy("a", 3)

	key, value, exists := h.Peek()
	if !exists {
		t.Fatal("Expected to find element")
	}
	if key != "a" || value != 8 {
		t.Errorf("Expected (a, 8), got (%s, %d)", key, value)
	}

	// Increment non-existing element by 5
	h.IncrementBy("b", 5)

	value, exists = h.GetValue("b")
	if !exists || value != 5 {
		t.Errorf("Expected b to have value 5, got %d", value)
	}
}

func TestUpdate(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Add element
	h.PushItem("a", 5)

	// Update existing element
	updated := h.Update("a", 10)
	if !updated {
		t.Error("Expected update to succeed")
	}

	key, value, exists := h.Peek()
	if !exists {
		t.Fatal("Expected to find element")
	}
	if key != "a" || value != 10 {
		t.Errorf("Expected (a, 10), got (%s, %d)", key, value)
	}

	// Update non-existing element
	updated = h.Update("b", 15)
	if updated {
		t.Error("Expected update to fail for non-existing element")
	}

	if h.Len() != 1 {
		t.Errorf("Expected length 1, got %d", h.Len())
	}
}

func TestRemove(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Add elements
	h.PushItem("a", 5)
	h.PushItem("b", 10)
	h.PushItem("c", 3)

	// Remove existing element
	removed := h.Remove("b")
	if !removed {
		t.Error("Expected remove to succeed")
	}

	if h.Len() != 2 {
		t.Errorf("Expected length 2, got %d", h.Len())
	}

	// Check that "b" is gone
	if h.Contains("b") {
		t.Error("Expected b to be removed")
	}

	// Remove non-existing element
	removed = h.Remove("d")
	if removed {
		t.Error("Expected remove to fail for non-existing element")
	}

	if h.Len() != 2 {
		t.Errorf("Expected length 2, got %d", h.Len())
	}
}

func TestContains(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Add element
	h.PushItem("a", 5)

	// Check existing element
	if !h.Contains("a") {
		t.Error("Expected to find element 'a'")
	}

	// Check non-existing element
	if h.Contains("b") {
		t.Error("Expected not to find element 'b'")
	}
}

func TestGetValue(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Add element
	h.PushItem("a", 5)

	// Get existing element
	value, exists := h.GetValue("a")
	if !exists {
		t.Error("Expected to find element 'a'")
	}
	if value != 5 {
		t.Errorf("Expected value 5, got %d", value)
	}

	// Get non-existing element
	value, exists = h.GetValue("b")
	if exists {
		t.Error("Expected not to find element 'b'")
	}
	if value != 0 {
		t.Errorf("Expected value 0, got %d", value)
	}
}

func TestSizeAndIsEmpty(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Test empty heap
	if h.Size() != 0 {
		t.Errorf("Expected size 0, got %d", h.Size())
	}
	if !h.IsEmpty() {
		t.Error("Expected heap to be empty")
	}

	// Add element
	h.PushItem("a", 5)

	if h.Size() != 1 {
		t.Errorf("Expected size 1, got %d", h.Size())
	}
	if h.IsEmpty() {
		t.Error("Expected heap to not be empty")
	}
}

func TestClear(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Add elements
	h.PushItem("a", 5)
	h.PushItem("b", 10)

	if h.Len() != 2 {
		t.Errorf("Expected length 2, got %d", h.Len())
	}

	// Clear heap
	h.Clear()

	if h.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", h.Len())
	}
	if !h.IsEmpty() {
		t.Error("Expected heap to be empty after clear")
	}
}

func TestTopN(t *testing.T) {
	h := NewMaxHeap[string](10)

	// Add elements
	h.PushItem("a", 5)
	h.PushItem("b", 10)
	h.PushItem("c", 3)
	h.PushItem("d", 8)
	h.PushItem("e", 1)

	// Test TopN with n < heap size
	result := h.TopN(3)
	if len(result) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(result))
	}

	// Check that elements are sorted in descending order (max heap)
	if result[0].Value != 10 || result[0].Key != "b" {
		t.Errorf("Expected first element (b, 10), got (%s, %d)", result[0].Key, result[0].Value)
	}
	if result[1].Value != 8 || result[1].Key != "d" {
		t.Errorf("Expected second element (d, 8), got (%s, %d)", result[1].Key, result[1].Value)
	}
	if result[2].Value != 5 || result[2].Key != "a" {
		t.Errorf("Expected third element (a, 5), got (%s, %d)", result[2].Key, result[2].Value)
	}

	// Test TopN with n > heap size
	result = h.TopN(10)
	if len(result) != 5 {
		t.Errorf("Expected 5 elements, got %d", len(result))
	}

	// Test TopN with n = 0
	result = h.TopN(0)
	if result != nil {
		t.Error("Expected nil result for n = 0")
	}

	// Test TopN with negative n
	result = h.TopN(-1)
	if result != nil {
		t.Error("Expected nil result for negative n")
	}
}

func TestTopNMinHeap(t *testing.T) {
	h := NewMinHeap[string](10)

	// Add elements
	h.PushItem("a", 5)
	h.PushItem("b", 10)
	h.PushItem("c", 3)
	h.PushItem("d", 8)
	h.PushItem("e", 1)

	// Test TopN with n < heap size
	result := h.TopN(3)
	if len(result) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(result))
	}

	// Check that elements are sorted in ascending order (min heap)
	if result[0].Value != 1 || result[0].Key != "e" {
		t.Errorf("Expected first element (e, 1), got (%s, %d)", result[0].Key, result[0].Value)
	}
	if result[1].Value != 3 || result[1].Key != "c" {
		t.Errorf("Expected second element (c, 3), got (%s, %d)", result[1].Key, result[1].Value)
	}
	if result[2].Value != 5 || result[2].Key != "a" {
		t.Errorf("Expected third element (a, 5), got (%s, %d)", result[2].Key, result[2].Value)
	}
}

func TestEmptyHeapOperations(t *testing.T) {
	h := NewMaxHeap[string](5)

	// Test PopItem on empty heap
	key, value, exists := h.PopItem()
	if exists {
		t.Error("Expected PopItem to return false for empty heap")
	}
	if key != "" || value != 0 {
		t.Errorf("Expected zero values, got (%s, %d)", key, value)
	}

	// Test Peek on empty heap
	key, value, exists = h.Peek()
	if exists {
		t.Error("Expected Peek to return false for empty heap")
	}
	if key != "" || value != 0 {
		t.Errorf("Expected zero values, got (%s, %d)", key, value)
	}

	// Test TopN on empty heap
	result := h.TopN(5)
	if result != nil {
		t.Error("Expected nil result for empty heap")
	}
}

func TestHeapOrdering(t *testing.T) {
	h := NewMaxHeap[int](10)

	// Add elements in random order
	h.PushItem(1, 5)
	h.PushItem(2, 10)
	h.PushItem(3, 3)
	h.PushItem(4, 8)
	h.PushItem(5, 1)

	// Pop all elements and verify they come out in descending order
	expected := []uint{10, 8, 5, 3, 1}
	for i, exp := range expected {
		_, value, exists := h.PopItem()
		if !exists {
			t.Fatalf("Expected to pop element %d", i)
		}
		if value != exp {
			t.Errorf("Expected value %d at position %d, got %d", exp, i, value)
		}
	}

	if !h.IsEmpty() {
		t.Error("Expected heap to be empty after popping all elements")
	}
}

func TestHeapWithDuplicateKeys(t *testing.T) {
	h := NewMaxHeap[string](10)

	// Add elements with same key but different values
	h.PushItem("a", 5)
	h.PushItem("a", 10) // This should create a separate entry

	if h.Len() != 2 {
		t.Errorf("Expected length 2, got %d", h.Len())
	}

	// Both entries should exist
	if !h.Contains("a") {
		t.Error("Expected to find element 'a'")
	}

	// Get all values for key "a"
	values := make([]uint, 0)
	for _, item := range h.items {
		if item.Key == "a" {
			values = append(values, item.Value)
		}
	}

	if len(values) != 2 {
		t.Errorf("Expected 2 values for key 'a', got %d", len(values))
	}
}

func BenchmarkPushItem(b *testing.B) {
	h := NewMaxHeap[int](uint(b.N))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.PushItem(i, uint(i))
	}
}

func BenchmarkPopItem(b *testing.B) {
	h := NewMaxHeap[int](uint(b.N))

	// Pre-fill heap
	for i := 0; i < b.N; i++ {
		h.PushItem(i, uint(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.PopItem()
	}
}

func BenchmarkTopN(b *testing.B) {
	h := NewMaxHeap[int](1000)

	// Pre-fill heap
	for i := 0; i < 1000; i++ {
		h.PushItem(i, uint(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.TopN(10)
	}
}
