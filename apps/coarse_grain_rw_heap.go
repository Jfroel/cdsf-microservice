package apps

import (
	"sync"

	"github.com/Jfroel/cdsf-microservice/proto/filter"
)

// Max Min heap definition with
// all fields are private (lowercase)
type CoarseRWMaxMinHeap struct {
	data     []*filter.FilterItem // underlying storage for the heap, using pointers so go can manage the memory
	capacity int                  // fixed capacity parameter, set at construction
	size     int                  // current number of items in the heap
	rwLk     sync.RWMutex
}

// public functions (Uppercase)

// ctor
func NewCoarseRWMaxMinHeap(capacity int) *CoarseRWMaxMinHeap {
	return &CoarseRWMaxMinHeap{
		data:     make([]*filter.FilterItem, 1, capacity+1), // initialize the array to the size param
		capacity: capacity,
		size:     0,
	}
}

func (s *CoarseRWMaxMinHeap) Describe() {}

// insert item into heap, returns boolean representing success
func (s *CoarseRWMaxMinHeap) Insert(item *filter.FilterItem) bool {
	s.rwLk.Lock()
	defer s.rwLk.Unlock()

	if item == nil {
		return false
	}

	if s.size >= s.capacity {
		indexOfMin := s.getIndexOfMin()
		curMin := s.data[indexOfMin]
		if curMin.GetScore() < item.GetScore() { // curMin < item
			// make room for inserting the bigger item
			// pick out min, remove it, reorder heap
			toRemove := s.getIndexOfMin()
			// fmt.Println("removing: ", toRemove)
			// fmt.Println("replacing with: ", s.size)
			s.data[toRemove] = s.data[s.size]
			s.data = s.data[:s.size]
			s.size--
			s.percolateDown(toRemove)
		} else {
			// don't insert this item
			return true
		}
	}

	s.data = append(s.data, item)
	s.size++

	s.percolateUp(s.size)

	return true
}

// Get the top ranked item, returns item and boolean representing success
func (s *CoarseRWMaxMinHeap) GetMax() *filter.FilterItem {
	s.rwLk.RLock()
	defer s.rwLk.RUnlock()

	if s.size == 0 {
		// zero-element heap
		return nil
	} else {
		// one-element or more
		return s.data[1]
	}
}

// Get the bottom ranked item, returns item and boolean representing success
func (s *CoarseRWMaxMinHeap) GetMin() *filter.FilterItem {
	s.rwLk.RLock()
	defer s.rwLk.RUnlock()

	return s.data[s.getIndexOfMin()]
}

// Remove the top ranked item, returns item and boolean representing success
func (s *CoarseRWMaxMinHeap) RemoveMax() *filter.FilterItem {
	s.rwLk.Lock()
	defer s.rwLk.Unlock()

	// _, 1, 2, 3 -> len() = 4
	if s.size == 0 {
		// zero-element heap
		return nil
	}

	retItem := s.data[1]
	if s.size == 1 {
		// one-element heap
		s.data = s.data[:1]
	} else {
		// take item from end of the heap and add to top, then percolate down
		s.data[1] = s.data[len(s.data)-1]
		s.data = s.data[:len(s.data)-1]

		// percolate 'em lil nodes
		s.percolateDown(1)
	}

	s.size--

	return retItem
}

// Remove the bottom ranked item, returns item and boolean representing success
func (s *CoarseRWMaxMinHeap) RemoveMin() *filter.FilterItem {
	s.rwLk.Lock()
	defer s.rwLk.Unlock()

	if s.size == 0 {
		// zero-element heap
		return nil
	}

	var retItem *filter.FilterItem
	if s.size == 1 {
		// one-element heap
		retItem = s.data[1]
		s.data = s.data[:1]
	} else if s.size == 2 {
		retItem = s.data[2]
		s.data = s.data[:2]
	} else {
		// three or more elements
		toRemove := 2
		if s.smaller(3, 2) { // less comp, returns a < b
			toRemove = 3
		}

		retItem = s.data[toRemove]
		// take item from end of the heap and add to top, then percolate down
		s.data[toRemove] = s.data[len(s.data)-1]
		s.data = s.data[:len(s.data)-1]

		// percolate 'em lil nodes
		s.percolateDown(toRemove)
	}

	s.size--

	return retItem
}

func (s *CoarseRWMaxMinHeap) Clear() bool {
	s.rwLk.Lock()
	defer s.rwLk.Unlock()
	s.data = make([]*filter.FilterItem, 1, s.capacity+1)
	s.size = 0
	return true
}

// get the current size of the heap
func (s *CoarseRWMaxMinHeap) Size() int {
	s.rwLk.RLock()
	defer s.rwLk.RUnlock()
	return s.size
}

func (s *CoarseRWMaxMinHeap) IsEmpty() bool {
	s.rwLk.RLock()
	defer s.rwLk.RUnlock()
	return s.size == 0
}

func (s *CoarseRWMaxMinHeap) IsFull() bool {
	s.rwLk.RLock()
	defer s.rwLk.RUnlock()
	return s.size == s.capacity
}

///////////////////////////////////
// private helper functions
///////////////////////////////////

// helpers expect the locks to already be held

func (s *CoarseRWMaxMinHeap) getIndexOfMin() int {
	if s.size == 0 {
		// zero-element heap
		return 0
	} else if s.size == 1 {
		// one-element
		return 1
	} else if s.size == 2 {
		// one-element
		return 2
	} else {
		// three or more elements
		if s.smaller(2, 3) { // less comp, returns a < b
			return 2
		} else {
			return 3
		}
	}
}

func (s *CoarseRWMaxMinHeap) percolateDown(i int) {
	if isMaxLevel(i) {
		s.percolateDownMax(i)
	} else {
		s.percolateDownMin(i)
	}
}

func (s *CoarseRWMaxMinHeap) percolateDownMax(i int) {
	if s.hasChildren(i) {
		m := s.largestChildOrGrandchild(i) // m is zero if no children or grandchildren
		if m > i*2+1 {
			// m is a grandchild of i
			if s.smaller(i, m) { // h[m] > h[i]
				s.swap(m, i)
				parentOfM := m / 2
				if s.smaller(m, parentOfM) {
					s.swap(m, parentOfM)
				}
				s.percolateDown(m)
			}
		} else if s.smaller(i, m) {
			s.swap(m, i)
		}
	}
}

func (s *CoarseRWMaxMinHeap) percolateDownMin(i int) {
	if s.hasChildren(i) {
		m := s.smallestChildOrGrandchild(i) // m is zero if no children or grandchildren
		if m > i*2+1 {
			// m is a grandchild of i
			if s.smaller(m, i) { // h[m] < h[i]
				s.swap(m, i)
				parentOfM := m / 2
				if s.smaller(parentOfM, m) {
					s.swap(m, parentOfM)
				}
				s.percolateDown(m)
			}
		} else if s.smaller(m, i) {
			s.swap(m, i)
		}
	}
}

func (s *CoarseRWMaxMinHeap) percolateUp(i int) {
	// like bubbles!!!???!?!?!?!
	if i == 1 {
		return
	}

	parentOfI := i / 2

	if !isMaxLevel(i) {
		// min level
		if s.smaller(parentOfI, i) {
			s.swap(i, parentOfI)
			s.percolateUpMax(parentOfI)
		} else {
			s.percolateUpMin(i)
		}
	} else {
		// max level
		if s.smaller(i, parentOfI) {
			s.swap(i, parentOfI)
			s.percolateUpMin(parentOfI)
		} else {
			s.percolateUpMax(i)
		}
	}

}

func (s *CoarseRWMaxMinHeap) percolateUpMax(i int) {
	gp := i / 4
	if gp > 0 && s.smaller(gp, i) {
		s.swap(i, gp)
		s.percolateUpMax(gp)
	}
}

func (s *CoarseRWMaxMinHeap) percolateUpMin(i int) {
	gp := i / 4
	if gp > 0 && s.smaller(i, gp) {
		s.swap(i, gp)
		s.percolateUpMin(gp)
	}
}

func (s *CoarseRWMaxMinHeap) swap(i, j int) {
	if i >= len(s.data) || j >= len(s.data) {
		return
	}

	temp := s.data[i]
	s.data[i] = s.data[j]
	s.data[j] = temp
}

func (s *CoarseRWMaxMinHeap) hasChildren(i int) bool {
	return 2*i < len(s.data)
}

func (s *CoarseRWMaxMinHeap) largestChildOrGrandchild(i int) int {
	minIndex := 0

	// check the children
	for j := 2 * i; j < len(s.data) && j < (2*i+2); j++ {
		if minIndex == 0 {
			minIndex = j
		} else if s.smaller(minIndex, j) { // less
			minIndex = j
		}
	}

	// check the grandchildren
	for j := 4 * i; j < len(s.data) && j < (4*i+4); j++ {
		if minIndex == 0 {
			minIndex = j
		} else if s.smaller(minIndex, j) { // less
			minIndex = j
		}
	}

	return minIndex
}

func (s *CoarseRWMaxMinHeap) smallestChildOrGrandchild(i int) int {
	minIndex := 0

	// check the children
	for j := 2 * i; j < len(s.data) && j < (2*i+2); j++ {
		if minIndex == 0 {
			minIndex = j
		} else if s.smaller(j, minIndex) { // less
			minIndex = j
		}
	}

	// check the grandchild
	for j := 4 * i; j < len(s.data) && j < (4*i+4); j++ {
		if minIndex == 0 {
			minIndex = j
		} else if s.smaller(j, minIndex) { // less
			minIndex = j
		}
	}

	return minIndex
}

func (s *CoarseRWMaxMinHeap) smaller(a, b int) bool {
	return s.data[a].GetScore() < s.data[b].GetScore()
}
