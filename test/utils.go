package test

import (
	"fmt"
	"math/rand"

	"flag"
	"runtime"
	"time"

	"github.com/Jfroel/cdsf-microservice/apps"
	"github.com/Jfroel/cdsf-microservice/proto/filter"
)

var (
	HEAP = flag.Int("heap", 0, "proxy server port")
)

func heapCtor(cap int) apps.MaxMinHeap {
	if *HEAP == 0 {
		return apps.NewCoarseRWMaxMinHeap(cap)
	} else {
		panic("subtree locking heap not yet implemented")
	}
}

func show() {
	if *HEAP == 0 {
		fmt.Println("RW lock")
	} else {
		fmt.Println("Subtree lock")
	}
}

func workerInsert(id int, jobs <-chan int, results chan<- int, heap apps.MaxMinHeap) {
	for range jobs {
		heap.Insert(&filter.FilterItem{Score: rand.Float32(), Data: []byte{}})
		results <- 0
	}
}

func workerRemoveMax(id int, jobs <-chan int, results chan<- int, heap apps.MaxMinHeap) {
	for range jobs {
		heap.RemoveMax()
		results <- 0
	}
}

func workerInsertTimed(id int, jobs <-chan int, results chan<- int64, heap apps.MaxMinHeap) {
	for range jobs {
		start := time.Now().UnixNano()
		heap.Insert(&filter.FilterItem{Score: rand.Float32(), Data: []byte{}})
		end := time.Now().UnixNano()
		results <- (end - start)
	}
}

func workerRemoveMaxTimed(id int, jobs <-chan int, results chan<- int64, heap apps.MaxMinHeap) {
	runtime.LockOSThread()
	for range jobs {
		start := time.Now().UnixNano()
		heap.RemoveMax()
		end := time.Now().UnixNano()

		results <- (end - start)
	}
	runtime.UnlockOSThread()
}
