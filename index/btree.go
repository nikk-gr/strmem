package index

import (
	"errors"
	"sync"

	"github.com/google/btree"
)

type SearchMethod uint8

const (
	// EQ Equal
	EQ SearchMethod = iota
	// GT Gather
	GT
	// GTE Gather or equal
	GTE
	// LT Lighter
	LT
	// LTE Lighter or equal
	LTE
)

// indexNode is a set of index of the base array with the same indexed data
type indexNode[T btree.Ordered] struct {
	index []int
	data  T
}

// BTree is a balanced tree index for the cache data array
type BTree[T btree.Ordered, A any] struct {
	dataPtr  *[]A
	rw       sync.RWMutex
	tree     *btree.BTreeG[indexNode[T]]
	getField func(cache *A) T
}

// NewBTree make a balanced tree index for the cache data array
// data is an array of any type data
// field is a function that returns the field that should be indexed
func NewBTree[T btree.Ordered, A any](
	data *[]A,
	field func(cache *A) T,
) *BTree[T, A] {
	ind := BTree[T, A]{
		dataPtr:  data,
		getField: field,
	}
	ind.Rebuild()
	return &ind
}

// Rebuild removes the old index and builds new
func (i *BTree[T, A]) Rebuild() {
	i.rw.Lock()
	defer i.rw.Unlock()
	i.tree = btree.NewG(4, func(a, b indexNode[T]) bool {
		return a.data < b.data
	})

	var (
		tmpINode indexNode[T]
		ok       bool
		tmpData  T
	)
	for j := range *i.dataPtr {
		tmpData = i.getField(&(*i.dataPtr)[j])
		tmpINode, ok = i.tree.Get(indexNode[T]{
			data: tmpData,
		})
		if ok {
			tmpINode.index = append(tmpINode.index, j)
		} else {
			tmpINode = indexNode[T]{
				index: []int{j},
				data:  tmpData,
			}
		}
		tmpINode, ok =
			i.tree.ReplaceOrInsert(tmpINode)
	}
}

// Get returns the slice of data array indexes that match selected key
func (i *BTree[T, A]) Get(key T) []int {
	i.rw.RLock()
	defer i.rw.RUnlock()
	iNode, ok := i.tree.Get(indexNode[T]{
		data: key,
	})
	if !ok {
		return nil
	}
	return iNode.index
}

// Put returns the slice of data array indexes that match selected key
func (i *BTree[T, A]) Put(item *A, index int) {
	i.rw.Lock()
	defer i.rw.Unlock()
	var (
		tmpINode indexNode[T]
		ok       bool
		tmpData  T
	)
	tmpData = i.getField(item)
	tmpINode, ok = i.tree.Get(indexNode[T]{
		data: tmpData,
	})
	if ok {
		tmpINode.index = append(tmpINode.index, index)
	} else {
		tmpINode = indexNode[T]{
			index: []int{index},
			data:  tmpData,
		}
	}
	tmpINode, ok =
		i.tree.ReplaceOrInsert(tmpINode)
}

func (i *BTree[T, A]) Rm(item *A, index int) {
	tmpData := i.getField(item)
	key := i.Get(tmpData)
	if key == nil {
		i.Rebuild()
	}

	if len(key) > 1 {
		key = rmFromArr(key, index)
		i.rw.Lock()
		defer i.rw.Unlock()
		i.tree.ReplaceOrInsert(indexNode[T]{
			index: key,
			data:  tmpData,
		})
		return
	}
	i.rw.Lock()
	defer i.rw.Unlock()
	i.tree.Delete(indexNode[T]{
		index: []int{index},
		data:  tmpData,
	})
}

func (i *BTree[T, A]) Find(key T, method SearchMethod) []int {
	i.rw.RLock()
	defer i.rw.RUnlock()
	if method == EQ {
		return i.Get(key)
	}

	iNode := indexNode[T]{
		data: key,
	}
	var data []int
	saver := func(in indexNode[T]) bool {
		data = append(data, in.index...)
		return true
	}
	switch method {
	case GT:
		i.tree.DescendGreaterThan(iNode, saver)
	case GTE:
		i.tree.AscendGreaterOrEqual(iNode, saver)
	case LT:
		i.tree.AscendLessThan(iNode, saver)
	case LTE:
		i.tree.DescendLessOrEqual(iNode, saver)
	default:
		panic(errors.New("invalid search method"))
	}
	return data
}

func (i *BTree[T, A]) GetRange(from, to T, includeFrom, includeTo bool) []int {
	i.rw.RLock()
	defer i.rw.RUnlock()
	if to == from {
		if includeFrom && includeTo {
			return i.Get(from)
		}
		return nil
	}

	if from > to {
		to, from = from, to
	}

	var data []int
	saver := func(in indexNode[T]) bool {
		if !includeFrom && in.data == from {
			return true
		}
		if !includeTo && in.data == to {
			return false
		}
		if in.data > to {
			return false
		}
		data = append(data, in.index...)
		return true
	}

	i.tree.DescendGreaterThan(indexNode[T]{
		data: from,
	}, saver)
	return data
}

func rmFromArr[T btree.Ordered](arr []T, val T) []T {
	var shortener int
	for i := range arr {
		if arr[i-shortener] == val {

			arr[i-shortener] = arr[len(arr)-1]
			arr = arr[:len(arr)-1]
			shortener++
		}
	}
	return arr
}
