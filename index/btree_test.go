package index

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtree(t *testing.T) {
	// Make base data array
	type (
		Entity struct {
			I   int
			Key uint32
		}
		Cache struct {
			data  *[]Entity
			index *BTree[uint32, Entity]
		}
	)

	init := func() (c Cache) {
		c.data = &[]Entity{
			{0, 6},
			{1, 1},
			{2, 1},
			{3, 5},
			{4, 6},
			{5, 7},
			{6, 8},
			{7, 8},
			{8, 10},
			{9, 10},
		}
		c.index = NewBTree(c.data, func(e *Entity) uint32 {
			return e.Key
		})
		return c
	}

	// Get by index
	t.Run("Get by key 1", func(t *testing.T) {
		cache := init()
		expectation := []int{1, 2}
		actual := cache.index.Get(1)
		sort.Ints(actual)
		assert.Equal(t, expectation, actual)
	})

	t.Run("Get by key 2", func(t *testing.T) {
		cache := init()
		expectation := []int{1, 2}
		actual := cache.index.Find(1, EQ)
		sort.Ints(actual)
		assert.Equal(t, expectation, actual)
	})
	t.Run("Get gather", func(t *testing.T) {
		cache := init()
		expectation := []int{5, 6, 7, 8, 9}
		actual := cache.index.Find(6, GT)
		sort.Ints(actual)
		assert.Equal(t, expectation, actual)
	})
	t.Run("Get gather or equal", func(t *testing.T) {
		cache := init()
		expectation := []int{0, 4, 5, 6, 7, 8, 9}
		actual := cache.index.Find(6, GTE)
		sort.Ints(actual)
		assert.Equal(t, expectation, actual)
	})
	t.Run("Get lighter", func(t *testing.T) {
		cache := init()
		expectation := []int{1, 2, 3}
		actual := cache.index.Find(6, LT)
		sort.Ints(actual)
		assert.Equal(t, expectation, actual)
	})
	t.Run("Get lighter or equal", func(t *testing.T) {
		cache := init()
		expectation := []int{0, 4, 5, 6, 7, 8, 9}
		actual := cache.index.Find(6, GTE)
		sort.Ints(actual)
		assert.Equal(t, expectation, actual)
	})
	t.Run("Add uniq val and get", func(t *testing.T) {
		cache := init()
		*cache.data = append(*cache.data, Entity{10, 20})
		cache.index.Put(&(*cache.data)[len(*cache.data)-1], len(*cache.data)-1)
		actual := cache.index.Get(20)
		expectation := []int{10}
		sort.Ints(actual)
		assert.Equal(t, expectation, actual)
	})
	t.Run("Add non uniq val and get", func(t *testing.T) {
		cache := init()
		*cache.data = append(*cache.data, Entity{10, 1})
		cache.index.Put(&(*cache.data)[len(*cache.data)-1], len(*cache.data)-1)
		actual := cache.index.Get(1)
		expectation := []int{1, 2, 10}
		sort.Ints(actual)
		assert.Equal(t, expectation, actual)
	})
	t.Run("Remove val and get", func(t *testing.T) {
		cache := init()
		indexToBeRemoved := 5
		// Remove val from index
		cache.index.Rm(&(*cache.data)[indexToBeRemoved], indexToBeRemoved)
		// To remove the element from the data array, we should replace it by the last element
		// of the array and shorten the data array by one item
		// So we should also remove the last element of the array from index
		cache.index.Rm(&(*cache.data)[len(*cache.data)-1], len(*cache.data)-1)
		// replace deleted element by the last element
		(*cache.data)[indexToBeRemoved] = (*cache.data)[len(*cache.data)-1]
		// shorten the data array
		*cache.data = (*cache.data)[:len(*cache.data)-1]
		// Put the element that was moved from the end of the data array
		// to deleted element cell to the index
		cache.index.Put(&(*cache.data)[indexToBeRemoved], indexToBeRemoved)
		// Check if it's work
		actual1 := cache.index.Get(7)
		var expectation1 []int = nil
		assert.Equal(t, expectation1, actual1, "value wasn't deleted")
		actual2 := cache.index.Get(10)
		sort.Ints(actual2)
		expectation2 := []int{5, 8}
		assert.Equal(t, expectation2, actual2, "index of the replaced value are wrong")
	})
}

func TestRmFromArr(t *testing.T) {
	tests := []struct {
		name        string
		array       []int
		valueToBeRm int
		expectation []int
	}{
		{
			name:        "rm one from the middle",
			array:       []int{1, 2, 3, 4},
			valueToBeRm: 2,
			expectation: []int{1, 4, 3},
		},
		{
			name:        "rm many from the middle",
			array:       []int{1, 2, 2, 4},
			valueToBeRm: 2,
			expectation: []int{1, 4},
		},
		{
			name:        "rm nothing from the middle",
			array:       []int{1, 2, 3, 4},
			valueToBeRm: 5,
			expectation: []int{1, 2, 3, 4},
		},
		{
			name:        "rm first element",
			array:       []int{1, 2, 3},
			valueToBeRm: 1,
			expectation: []int{3, 2},
		},
		{
			name:        "rm last element",
			array:       []int{1, 2, 3},
			valueToBeRm: 3,
			expectation: []int{1, 2},
		},
		{
			name:        "rm the only element",
			array:       []int{1},
			valueToBeRm: 1,
			expectation: []int{},
		},
		{
			name:        "rm the only elements",
			array:       []int{1, 1, 1},
			valueToBeRm: 1,
			expectation: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := rmFromArr(tt.array, tt.valueToBeRm)
			assert.Equal(t, tt.expectation, actual)
		})
	}
}
