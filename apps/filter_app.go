package apps

import (
	"log"

	"github.com/Jfroel/cdsf-microservice/proto/filter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
 * App Interface
 *
 * This interface serves as a shim layer wrapper
 * around the different max-min heap implementaions
 * to provide standardized error handling
 */

type ConcurrentDataStreamFilter interface {
	Insert(item *filter.FilterItem) error

	GetMax() (*filter.FilterItem, error)

	GetMin() (*filter.FilterItem, error)

	RemoveMax() (*filter.FilterItem, error)

	RemoveMin() (*filter.FilterItem, error)

	GetSize() int

	Clear() error
}

// The app is just a wrapper around any MaxMinHeap implementation
type CDSFApp struct {
	heap MaxMinHeap
}

// Change the Heap constructor to change the used implementaion
func NewCDSFApp(lkType string, capacity int) *CDSFApp {
	var heap MaxMinHeap
	switch lkType {
	case "coarseRW":
		log.Println("locking policy: coarse grain RW")
		heap = NewCoarseRWMaxMinHeap(capacity)
	case "subtree":
		log.Println("locking policy: subtree")
		panic("subtree locking not yet implemented")
	default:
		panic("bad arg to CDSF constructor")
	}
	log.Println("filter max capacity: ", capacity)
	return &CDSFApp{
		heap: heap,
	}
}

func (s *CDSFApp) Insert(item *filter.FilterItem) error {

	ok := s.heap.Insert(item)
	if !ok {
		return status.Errorf(codes.Internal, "Filter failed to insert item")
	}

	return status.Errorf(codes.OK, "Item Inserted")
}

func (s *CDSFApp) GetMax() (*filter.FilterItem, error) {
	if s.heap.IsEmpty() {
		return nil, status.Errorf(codes.Internal, "Filter is empty")
	}

	item := s.heap.GetMax()
	if item == nil {
		return nil, status.Errorf(codes.Internal, "Filter failed to retrieve max item")
	}

	return item, status.Errorf(codes.OK, "Max item retrieved")
}

func (s *CDSFApp) GetMin() (*filter.FilterItem, error) {
	if s.heap.IsEmpty() {
		return nil, status.Errorf(codes.Internal, "Filter is empty")
	}

	item := s.heap.GetMin()
	if item == nil {
		return nil, status.Errorf(codes.Internal, "Filter failed to retrieve min item")
	}

	return item, status.Errorf(codes.OK, "Min item retrieved")
}

func (s *CDSFApp) RemoveMax() (*filter.FilterItem, error) {
	if s.heap.IsEmpty() {
		return nil, status.Errorf(codes.Internal, "Filter is empty")
	}

	item := s.heap.RemoveMax()
	if item == nil {
		return nil, status.Errorf(codes.Internal, "Filter failed to remove max item")
	}

	return item, status.Errorf(codes.OK, "Max item removed")
}

func (s *CDSFApp) RemoveMin() (*filter.FilterItem, error) {
	if s.heap.IsEmpty() {
		return nil, status.Errorf(codes.Internal, "Filter is empty")
	}

	item := s.heap.RemoveMin()
	if item == nil {
		return nil, status.Errorf(codes.Internal, "Filter failed to remove min item")
	}

	return item, status.Errorf(codes.OK, "Min item removed")
}

func (s *CDSFApp) GetSize() int {
	return s.heap.Size()
}

func (s *CDSFApp) Clear() error {
	if s.heap.Clear() {
		return status.Errorf(codes.OK, "Filtered cleared")
	} else {
		return status.Errorf(codes.OK, "Filter failed to clear")
	}
}
