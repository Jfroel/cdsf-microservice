package apps

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Jfroel/cdsf-microservice/proto/filter"
)

const CV = 1

// Tags for HeapNode: EMPTY, AVAILABLE, IID (inserting op ID)
const EMPTY = 0
const AVAILABLE = 1

type HeapNode struct {
	item *filter.FilterItem
	tag  uint64
	lk   sync.Mutex
}

func NewHeapNode(item *filter.FilterItem) *HeapNode {
	return &HeapNode{
		item: item,
		tag:  EMPTY,
	}
}

// Max Min heap definition with
// all fields are private (lowercase)
type SubtreeLkMaxMinHeap struct {
	data      []*HeapNode // underlying storage for the heap, using pointers so go can manage the memory
	capacity  int         // fixed capacity parameter, set at construction
	size      int         // current number of items in the heap
	reversed  int
	reversed2 int
	highBit   int
	heapLk    sync.Mutex

	uID atomic.Uint64
}

// public functions (Uppercase)

// ctor
func NewSubtreeLkMaxMinHeap(capacity int) *SubtreeLkMaxMinHeap {

	if capacity < 1 {
		// not allowing degenerate cases
		panic("Heap should have a minimum capacity of 1")
	}

	s := &SubtreeLkMaxMinHeap{
		capacity:  capacity,
		size:      0,
		reversed:  0,
		reversed2: 0,
		highBit:   -1,
	}

	s.uID.Store(2)

	s.heapLk.Lock()
	defer s.heapLk.Unlock()

	s.initHeap()
	return s
}

// not safe
// used to figure out what was broken in single-threaded land
//   - James Froelich, 2023
func (s *SubtreeLkMaxMinHeap) Describe() {
	s.heapLk.Lock()
	defer s.heapLk.Unlock()

	// fmt.Println(s.data)

	for i, node := range s.data {
		fmt.Printf("%v: (%v, %v) \n", i, node.item, node.tag)
		if !node.lk.TryLock() {
			fmt.Printf("%v locked, tag: %v\n", i, s.TAG(i))
		} else {
			node.lk.Unlock()
		}
	}
}

// insert item into heap, returns boolean representing success
func (s *SubtreeLkMaxMinHeap) Insert(item *filter.FilterItem) bool {

	if item == nil {
		return false
	}

	s.heapLk.Lock()

	var indexOfMin = 0

	// check if we need to do this op
	if s.size >= s.capacity {
		// fmt.Println("past cap")
		indexOfMin = s.getIndexOfMin() // min will be locked by get
		curMin := s.data[indexOfMin].item.Score
		if item.Score < curMin { // item
			// no need to change
			s.UNLOCK(indexOfMin)
			s.heapLk.Unlock()
			return true
		} else {
			// make room for the new item
			// don't decrease the size here, we don't want another inserting
			// process to think the heap is smaller than it is (since we are about to
			// add an item back into the heap)

			// adjust so

			// pick an available element from the last layer, using the distributed access pattern
			s.UNLOCK(indexOfMin)
			bottom := s.BIT_REV_CAP() // need to unlock because might return min

			// set the tag/item
			// percolate that down

			if bottom == indexOfMin {
				s.SET_TAG(bottom, EMPTY)
				s.UNLOCK(indexOfMin)
			} else {
				s.LOCK(indexOfMin)

				bottomItem := s.ITEM(bottom) // is always the last element of the heap (doesn't actually use the counter)
				// is this item
				s.SET_TAG(bottom, EMPTY)
				s.UNLOCK(bottom)

				s.SET_ITEM(indexOfMin, bottomItem)
				s.SET_TAG(indexOfMin, AVAILABLE)
				s.percolateDown(indexOfMin)
			}

			iid := s.newIID()

			s.LOCK(bottom)
			s.heapLk.Unlock()

			s.SET_ITEM(bottom, item)
			s.SET_TAG(bottom, iid)

			s.UNLOCK(bottom)

			s.percolateUpIter(bottom, iid)

			return true
		}
	} else {
		// standard insert logic

		i := s.BIT_REV_INC()

		s.LOCK(i)
		s.heapLk.Unlock()

		iid := s.newIID()

		s.SET_ITEM(i, item)
		s.SET_TAG(i, iid)

		s.UNLOCK(i)

		s.percolateUpIter(i, iid)
		return true
	}
}

// Get the top ranked item, returns item and boolean representing success
func (s *SubtreeLkMaxMinHeap) GetMax() *filter.FilterItem {
	s.heapLk.Lock()
	defer s.heapLk.Unlock()

	if s.size == 0 {
		// zero-element heap
		return nil
	} else {
		// one-element or more
		s.LOCK(1)
		defer s.UNLOCK(1)
		if s.TAG(1) != EMPTY {
			return s.ITEM(1)
		} else {
			return nil
		}
	}
}

// Get the bottom ranked item, returns item and boolean representing success
func (s *SubtreeLkMaxMinHeap) GetMin() *filter.FilterItem {
	s.heapLk.Lock()
	defer s.heapLk.Unlock()

	minIndex := s.getIndexOfMin()
	defer s.UNLOCK(minIndex)
	return s.ITEM(minIndex)
}

// Remove the top ranked item, returns item and boolean representing success
func (s *SubtreeLkMaxMinHeap) RemoveMax() *filter.FilterItem {
	s.heapLk.Lock()

	if s.size == 0 {
		// zero-element heap
		s.heapLk.Unlock()
		return nil
	}

	bottom := s.BIT_REV_DEC()

	s.LOCK(bottom)
	s.heapLk.Unlock()

	retItem := s.ITEM(bottom)
	s.SET_TAG(bottom, EMPTY)
	s.UNLOCK(bottom)

	// Lock first item. Stop if it was the only item in the heap.
	s.LOCK(1)
	if s.TAG(1) == EMPTY {
		s.UNLOCK(1)
		return retItem
	}

	// Replace the top item with the item stored from
	// the bottom.

	retItem = s.SET_ITEM(1, retItem)
	s.SET_TAG(1, AVAILABLE)

	// percolate 'em lil nodes
	s.percolateDown(1)

	return retItem
}

// Remove the bottom ranked item, returns item and boolean representing success
func (s *SubtreeLkMaxMinHeap) RemoveMin() *filter.FilterItem {
	s.heapLk.Lock()

	minIndex := s.getIndexOfMin()
	if minIndex == 0 {
		s.UNLOCK(minIndex)
		s.heapLk.Unlock()
		return nil
	}

	bottom := s.BIT_REV_DEC()

	if bottom == minIndex {
		// 1, 2, 3-element heap
		// minIndex/bottom is locked
		s.heapLk.Unlock()

		retItem := s.ITEM(bottom)
		s.SET_TAG(bottom, EMPTY)
		s.UNLOCK(bottom)
		return retItem
	} else {
		s.LOCK(bottom)
		s.heapLk.Unlock()

		// set bottom node free
		bottomItem := s.ITEM(bottom)
		s.SET_TAG(bottom, EMPTY)
		s.UNLOCK(bottom)

		// replace min item with
		retItem := s.SET_ITEM(minIndex, bottomItem)
		s.SET_TAG(minIndex, AVAILABLE)
		s.percolateDown(minIndex)
		return retItem
	}
}

func (s *SubtreeLkMaxMinHeap) Clear() bool {
	s.heapLk.Lock()
	defer s.heapLk.Unlock()

	s.initHeap()

	return true
}

// get the current size of the heap
func (s *SubtreeLkMaxMinHeap) Size() int {
	s.heapLk.Lock()
	defer s.heapLk.Unlock()
	return s.size
}

func (s *SubtreeLkMaxMinHeap) IsEmpty() bool {
	s.heapLk.Lock()
	defer s.heapLk.Unlock()
	return s.size == 0
}

func (s *SubtreeLkMaxMinHeap) IsFull() bool {
	s.heapLk.Lock()
	defer s.heapLk.Unlock()
	return s.size == s.capacity
}

///////////////////////////////////
// private helper functions
///////////////////////////////////

func (s *SubtreeLkMaxMinHeap) LOCK(i int) {
	s.data[i].lk.Lock()
}

func (s *SubtreeLkMaxMinHeap) UNLOCK(i int) {
	s.data[i].lk.Unlock()
}

func (s *SubtreeLkMaxMinHeap) ITEM(i int) *filter.FilterItem {
	return s.data[i].item
}

// returns old item
func (s *SubtreeLkMaxMinHeap) SET_ITEM(i int, item *filter.FilterItem) *filter.FilterItem {
	ret := s.data[i].item
	s.data[i].item = item
	return ret
}

func (s *SubtreeLkMaxMinHeap) TAG(i int) uint64 {
	return s.data[i].tag
}

func (s *SubtreeLkMaxMinHeap) SET_TAG(i int, tag uint64) {
	s.data[i].tag = tag
}

func (s *SubtreeLkMaxMinHeap) newIID() uint64 {
	n := s.uID.Add(1)
	if n == 0 {
		s.uID.Store(2)
		n = 2
	}
	return n
}

func (s *SubtreeLkMaxMinHeap) BIT_REV_INC() int {
	s.size++

	var bit int

	for true {
		for bit = s.highBit - 1; bit >= 0; bit-- {
			s.reversed = s.reversed ^ (1 << bit)
			if s.reversed&(1<<bit) != 0 {
				// some 0-bit is flipped into 1-bit
				break
			}
		}

		if bit < 0 {
			// case: 11...111, so no bit of 0 to flip
			s.reversed = s.size
			s.reversed2 = s.size
			s.highBit++
			break
		}

		if s.reversed <= s.capacity {
			break
		}

	}
	return s.reversed
}

func (s *SubtreeLkMaxMinHeap) BIT_REV_DEC() int {
	s.size--

	oldRev := s.reversed

	var bit int
	for true {
		for bit = s.highBit - 1; bit >= 0; bit-- {
			s.reversed = s.reversed ^ (1 << bit)
			if s.reversed&(1<<bit) == 0 {
				// some 0-bit is flipped into 1-bit
				break
			}
		}

		if bit < 0 {
			// case: 11...111, so no bit of 0 to flip
			s.reversed = s.size
			s.reversed2 = s.size
			s.highBit--
			break
		}

		if s.reversed <= s.capacity {
			break
		}
	}

	return oldRev
}

func (s *SubtreeLkMaxMinHeap) BIT_REV_CAP() int {

	var bit int

	for true {
		for bit = s.highBit - 1; bit >= 0; bit-- {
			s.reversed2 = s.reversed2 ^ (1 << bit)
			if s.reversed2&(1<<bit) == 0 {
				// some 0-bit is flipped into 1-bit
				break
			}
		}

		// wrap around
		if bit < 0 {
			// case: 11...111, so no bit of 0 to flip
			s.reversed2 = s.size
			// s.highBit--
		}

		if s.reversed2 <= s.capacity {
			s.LOCK(s.reversed2)
			if s.TAG(s.reversed2) == AVAILABLE {
				break // retrun the locked item
			}
			s.UNLOCK(s.reversed2)
		}
	}

	return s.reversed2
}

func (s *SubtreeLkMaxMinHeap) initHeap() {
	// extra node above the root for index adjustment and special cases
	// extra node at the end for overfill safety
	s.data = make([]*HeapNode, s.capacity+2, s.capacity+2)

	s.size = 0
	s.highBit = -1
	s.reversed = 0
	for i := 0; i < s.capacity+2; i++ {
		s.data[i] = NewHeapNode(nil)
	}
}

// precon: heap lock is held
// return min node with lock help
func (s *SubtreeLkMaxMinHeap) getIndexOfMin() int {
	if s.size == 0 {
		// zero-element heap
		s.LOCK(0)
		return 0
	} else if s.size == 1 {
		// one-element
		s.LOCK(1)
		return 1
	} else if s.size == 2 {
		// two-element
		s.LOCK(2)
		return 2
	} else {
		// 3+ elements
		s.LOCK(2)
		s.LOCK(3)
		if s.smaller(2, 3) { // less comp, returns a < b
			s.UNLOCK(3)
			return 2
		} else {
			s.UNLOCK(2)
			return 3
		}
	}
}

func (s *SubtreeLkMaxMinHeap) percolateDown(i int) {
	if isMaxLevel(i) {
		s.percolateDownMax(i)
	} else {
		s.percolateDownMin(i)
	}
}

func (s *SubtreeLkMaxMinHeap) percolateDownMax(i int) {
	m := s.largestChildOrGrandchild(i) // m is zero if no children or grandchildren
	// m (and its parent if gc) are locked by function above
	if m != 0 {
		if m > i*2+1 {
			parentOfM := m / 2
			// m is a grandchild of i
			if s.larger(m, i) { // h[m] > h[i]
				s.swap(m, i)
				if s.smaller(m, parentOfM) {
					s.swap(m, parentOfM)
				}
				s.UNLOCK(parentOfM)
				s.UNLOCK(i)
				s.percolateDown(m)
			} else {
				s.UNLOCK(m)
				s.UNLOCK(parentOfM)
				s.UNLOCK(i)
			}
		} else if s.larger(m, i) {
			s.swap(m, i)
			s.UNLOCK(m)
			s.UNLOCK(i)
		} else {
			s.UNLOCK(m)
			s.UNLOCK(i)
		}
	} else {
		s.UNLOCK(i)
	}
}

func (s *SubtreeLkMaxMinHeap) percolateDownMin(i int) {
	m := s.smallestChildOrGrandchild(i) // m is zero if no children or grandchildren
	// m (and its parent if gc) are locked by function above
	if m != 0 {
		if m > i*2+1 {
			// m is a grandchild of i
			parentOfM := m / 2
			if s.smaller(m, i) { // h[m] < h[i]
				s.swap(m, i)
				if s.larger(m, parentOfM) {
					s.swap(m, parentOfM)
				}
				s.UNLOCK(parentOfM)
				s.UNLOCK(i)
				s.percolateDown(m)
			} else {
				s.UNLOCK(m)
				s.UNLOCK(parentOfM)
				s.UNLOCK(i)
			}
		} else if s.smaller(m, i) {
			s.swap(m, i)
			s.UNLOCK(m)
			s.UNLOCK(i)
		} else {
			s.UNLOCK(m)
			s.UNLOCK(i)
		}
	} else {
		s.UNLOCK(i)
	}
}

func (s *SubtreeLkMaxMinHeap) compareAndSwap(gp, p, i int) int {
	// proceed as normal
	if !isMaxLevel(i) {
		// min level
		if s.larger(i, p) {
			s.swap(i, p)
			return p
		} else {
			if gp > 0 {
				if s.smaller(i, gp) {
					s.swap(i, gp)
					return gp
				} else {
					s.SET_TAG(i, AVAILABLE)
					return 0
				}
			} else {
				// no gp, stop
				s.SET_TAG(i, AVAILABLE)
				return 0
			}
		}
	} else {
		// max level
		if s.smaller(i, p) {
			s.swap(i, p)
			return p
		} else {
			if gp > 0 {
				if s.larger(i, gp) {
					s.swap(i, gp)
					return gp
				} else {
					s.SET_TAG(i, AVAILABLE)
					return 0
				}
			} else {
				// no gp, stop
				s.SET_TAG(i, AVAILABLE)
				return 0
			}
		}
	}
}

// the job of this method is to determine if the new element should be considered
// as a min or a max. Once that is decided, it only needs to be compared to other
// min or max levels respectively

// the paper completely unlocks the nodes percolating up so that they may be intercepted
// by a currently happening remove opperation. However, when the it unlocks the last node
// it tags it with its pid

// if the insert resumes and the node you left on doesn't have the matching pid,
// you know the node was moved by a remove op. Percolate up then tries to chase after the
// node until it finds the tagged node. You just need to be sure that you do find it, so
// you can ensure its in the correct place and remove the tag

// in the case of the max-min heap percolate down could move the percolate up item
// up two levels or one. This means we probably need to remove the logic where
// we only compare in the max levels or min levels.

func (s *SubtreeLkMaxMinHeap) percolateUpIter(i int, iid uint64) {
	// Inv: nothing's locked
	for i > 1 {

		oldI := i
		gp := i / 4
		p := i / 2

		if gp > 0 {
			// lock gp, check for 3 cases
			// GP is some other PID -> spin on tag
			// GP is our PID -> move to gp
			// GP is Available -> proceed as normal

			s.LOCK(gp)

			if s.TAG(gp) == iid {
				s.UNLOCK(gp)
				i = gp
				continue
			} else if s.TAG(gp) > AVAILABLE { // A = 1, E = 0
				s.UNLOCK(gp)
				continue
			}
		}

		// gp is available or empty

		// lock p, check for 3 cases
		// P is some other PID -> spin on pid
		// P is our PID -> move to p
		// P is Available -> proceed as normal
		// P is empty -> i is now at root

		s.LOCK(p)
		if s.TAG(p) == iid {
			s.UNLOCK(p)
			if gp > 0 {
				s.UNLOCK(gp)
			}
			i = p
			continue
		} else if s.TAG(p) > AVAILABLE { // other inserts iid
			s.UNLOCK(p)
			if gp > 0 {
				s.UNLOCK(gp)
			}
			i = p
			continue
		}

		// fall through, parent is empty or available

		// lock i, check for 3 cases
		// i is our PID -> proceed as normal
		// i is some other PID -> fuck, need to chase
		// i is Available -> fuck, need to chase

		s.LOCK(i)

		if s.TAG(i) == iid {
			i = s.compareAndSwap(gp, p, i)
		} else if s.TAG(p) == EMPTY {
			i = 0
		} else {
			// chase
			i = p
		}

		s.UNLOCK(oldI)
		s.UNLOCK(p)
		if gp > 0 {
			s.UNLOCK(gp)
		}
	}
	// made it to the root, "untag"
	if i == 1 {
		s.LOCK(i)
		if s.TAG(i) == iid {
			s.SET_TAG(i, AVAILABLE)
		}
		s.UNLOCK(i)
	}
}

func (s *SubtreeLkMaxMinHeap) swap(i, j int) {
	tmp_tag := s.data[i].tag
	tmp_item := s.data[i].item

	s.data[i].tag = s.data[j].tag
	s.data[i].item = s.data[j].item

	s.data[j].tag = tmp_tag
	s.data[j].item = tmp_item
}

func (s *SubtreeLkMaxMinHeap) largestChildOrGrandchild(i int) int {
	return s.xChildOrGrandchild(i, s.larger)
}

func (s *SubtreeLkMaxMinHeap) smallestChildOrGrandchild(i int) int {
	return s.xChildOrGrandchild(i, s.smaller)
}

// returned index stays locked
// if returned index is a gc, it's parent stays locked
func (s *SubtreeLkMaxMinHeap) xChildOrGrandchild(i int, comp func(int, int) bool) int {
	l := 2 * i
	r := 2*i + 1
	ll := 4 * i
	lr := 4*i + 1
	rl := 4*i + 2
	rr := 4*i + 3

	all := []int{l, r, ll, lr, rl, rr}
	indexes := make([]int, 0)

	for _, index := range all {
		if index <= s.capacity { // should this include the extra element?
			indexes = append(indexes, index)
		}
	}

	if len(indexes) == 0 {
		return 0
	}

	for _, index := range indexes {
		s.LOCK(index)
	}

	if s.TAG(l) == EMPTY && s.TAG(r) == EMPTY {
		// no valid children / grand children
		for _, index := range indexes {
			s.UNLOCK(index)
		}
		return 0
	}

	// l child is known to be valid
	ret := l

	// find the largest child / gc
	for _, index := range indexes {
		if s.TAG(index) != EMPTY && comp(index, ret) {
			ret = index
		}
	}

	if ret == l || ret == r { // child is the swap val
		// unlock the others
		for _, index := range indexes {
			if index != ret {
				s.UNLOCK(index)
			}
		}
	} else { // gchild is swap val, need to keep parent locked
		if ret == ll || ret == lr {
			for _, index := range indexes {
				if index != ret && index != l {
					s.UNLOCK(index)
				}
			}
		} else {
			for _, index := range indexes {
				if index != ret && index != r {
					s.UNLOCK(index)
				}
			}
		}
	}

	// largest is still locked
	return ret
}

func (s *SubtreeLkMaxMinHeap) larger(a, b int) bool {
	return s.data[a].item.GetScore() > s.data[b].item.GetScore()
}

func (s *SubtreeLkMaxMinHeap) smaller(a, b int) bool {
	return s.data[a].item.GetScore() < s.data[b].item.GetScore()
}
