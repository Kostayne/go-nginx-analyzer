package heaps

import (
	"container/heap"
)

type Item[K comparable] struct {
	Key   K
	Value uint
	index int
}

type BaseHeap[K comparable] struct {
	items []*Item[K]
	less  func(a, b uint) bool
}

func NewMaxHeap[K comparable](maxLen uint) *BaseHeap[K] {
	lessFn := func(a, b uint) bool { return a > b }

	h := &BaseHeap[K]{
		items: make([]*Item[K], 0, maxLen),
		less:  lessFn,
	}

	heap.Init(h)
	return h
}

func NewMinHeap[K comparable](maxLen uint) *BaseHeap[K] {
	lessFn := func(a, b uint) bool { return a < b }

	h := &BaseHeap[K]{
		items: make([]*Item[K], 0, maxLen),
		less:  lessFn,
	}

	heap.Init(h)
	return h
}

func (h *BaseHeap[K]) Len() int {
	return len(h.items)
}

func (h *BaseHeap[K]) Less(i, j int) bool {
	return h.less(h.items[i].Value, h.items[j].Value)
}

func (h *BaseHeap[K]) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.items[i].index = i
	h.items[j].index = j
}

func (h *BaseHeap[K]) Push(x any) {
	item := x.(*Item[K])
	item.index = len(h.items)
	h.items = append(h.items, item)
}

func (h *BaseHeap[K]) Pop() any {
	old := h.items
	n := len(old)

	item := old[n-1]
	old[n-1] = nil

	item.index = -1
	h.items = old[0 : n-1]

	return item
}

func (h *BaseHeap[K]) PushItem(key K, value uint) {
	heap.Push(h, &Item[K]{Key: key, Value: value})
}

// Returns element with the most (biggest / smallest) value & removes it from heap
func (h *BaseHeap[K]) PopItem() (K, uint, bool) {
	if h.Len() == 0 {
		var zeroKey K
		return zeroKey, 0, false
	}

	item := heap.Pop(h).(*Item[K])
	return item.Key, item.Value, true
}

// Returns element with the most (biggest / smallest) value
func (h *BaseHeap[K]) Peek() (K, uint, bool) {
	if h.Len() == 0 {
		var zeroKey K
		return zeroKey, 0, false
	}

	item := h.items[0]
	return item.Key, item.Value, true
}

// Returns element with the most (biggest / smallest) value
func (h *BaseHeap[K]) Increment(key K) {
	for _, item := range h.items {
		if item.Key == key {
			item.Value++
			heap.Fix(h, item.index)
			return
		}
	}

	// If key is not found - add new element
	h.PushItem(key, 1)
}

func (h *BaseHeap[K]) IncrementBy(key K, delta uint) {
	for _, item := range h.items {
		if item.Key == key {
			item.Value += delta
			heap.Fix(h, item.index)
			return
		}
	}

	// If key is not found - add new element
	h.PushItem(key, delta)
}

// Returns true if key was found
func (h *BaseHeap[K]) Update(key K, newValue uint) bool {
	for _, item := range h.items {
		if item.Key == key {
			oldValue := item.Value
			item.Value = newValue

			if newValue != oldValue {
				heap.Fix(h, item.index)
			}

			return true
		}
	}

	return false
}

func (h *BaseHeap[K]) Remove(key K) bool {
	for i, item := range h.items {
		if item.Key == key {
			heap.Remove(h, i)
			return true
		}
	}

	return false
}

func (h *BaseHeap[K]) Contains(key K) bool {
	for _, item := range h.items {
		if item.Key == key {
			return true
		}
	}

	return false
}

func (h *BaseHeap[K]) GetValue(key K) (uint, bool) {
	for _, item := range h.items {
		if item.Key == key {
			return item.Value, true
		}
	}

	return 0, false
}

func (h *BaseHeap[K]) Size() int {
	return len(h.items)
}

func (h *BaseHeap[K]) IsEmpty() bool {
	return h.Len() == 0
}

func (h *BaseHeap[K]) Clear() {
	h.items = make([]*Item[K], 0)
}

func (h *BaseHeap[K]) TopN(n int) []Item[K] {
	if n <= 0 || h.IsEmpty() {
		return nil
	}

	if n > h.Len() {
		n = h.Len()
	}

	temp := &BaseHeap[K]{
		items: make([]*Item[K], 0, h.Len()),
		less:  h.less,
	}

	for i, item := range h.items {
		temp.items = append(temp.items, &Item[K]{
			Key:   item.Key,
			Value: item.Value,
			index: i,
		})
	}

	heap.Init(temp)

	result := make([]Item[K], n)
	for i := 0; i < n; i++ {
		item := heap.Pop(temp).(*Item[K])
		result[i] = *item
	}

	return result
}
