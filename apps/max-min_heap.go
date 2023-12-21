package apps

import (
	"math"

	"github.com/Jfroel/cdsf-microservice/proto/filter"
)

/*
 * Max Min Heap Interface
 */
type MaxMinHeap interface {
	Insert(item *filter.FilterItem) bool

	GetMax() *filter.FilterItem

	GetMin() *filter.FilterItem

	RemoveMax() *filter.FilterItem

	RemoveMin() *filter.FilterItem

	Clear() bool

	Size() int

	IsEmpty() bool

	IsFull() bool
}

// todo: this can be done faster with checking the most significant 1-Bit
// if the position is odd, its a max layer: 1->0b1, 4->0b100, 7->0b111
// if the position is even, its a min layer: 2->0b10, 3->0b11, 8->0b1000

// https://pkg.go.dev/encoding/binary
func isMaxLevel(i int) bool {
	// even levels (0-indexed) are max levels

	// level is given by floor(log2(n))
	level := int(math.Floor(math.Log2(float64(i))))
	return level%2 == 0
}
