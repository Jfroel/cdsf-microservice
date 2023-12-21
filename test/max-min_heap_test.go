package test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/Jfroel/cdsf-microservice/proto/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSizeAndCapacity(t *testing.T) {
	var tests = []struct {
		cap        int
		numInserts int
	}{
		{1, 1},
		{1, 2},
		{10, 11},
		{15, 17},
		{10, 9},
	}
	for _, tt := range tests {
		heap := heapCtor(tt.cap)

		for i := 0; i < tt.numInserts; i++ {
			heap.Insert(&filter.FilterItem{Score: rand.Float32(), Data: []byte{}})
		}

		if tt.cap < tt.numInserts {
			require.Equal(t, tt.cap, heap.Size())
		} else {
			require.Equal(t, tt.numInserts, heap.Size())
		}
	}
}

func TestEmptyAndFull(t *testing.T) {
	var tests = []struct {
		cap        int
		numInserts int
		empty      bool
		full       bool
	}{
		{10, 0, true, false},
		{10, 5, false, false},
		{10, 10, false, true},
	}
	for _, tt := range tests {
		heap := heapCtor(tt.cap)

		for i := 0; i < tt.numInserts; i++ {
			heap.Insert(&filter.FilterItem{Score: rand.Float32(), Data: []byte{}})
		}

		assert.Equal(t, tt.empty, heap.IsEmpty())
		assert.Equal(t, tt.full, heap.IsFull())
	}
}

func TestClear(t *testing.T) {
	heap := heapCtor(10)

	heap.Insert(&filter.FilterItem{Score: rand.Float32(), Data: []byte{}})

	heap.Clear()

	assert.Equal(t, true, heap.IsEmpty())
}

func TestInsertAndRemoveMax(t *testing.T) {
	show()
	var tests = []struct {
		cap        int
		numInserts int
		numRemoves int
		numRuns    int
	}{
		{10, 10, 10, 100},
		{100, 100, 50, 100},
		{100, 100, 0, 2},
		{100000, 100000, 100000, 2},
		{27, 18, 17, 2},
	}
	for _, tt := range tests {
		require.GreaterOrEqual(t, tt.cap, tt.numInserts,
			"Invalid test: number of inserts must be less than or equal to the capacity")

		require.GreaterOrEqual(t, tt.numInserts, tt.numRemoves,
			"Invalid test: number of inserts must be greater than or equal to the number of removes")

		heap := heapCtor(tt.cap)
		for i := 0; i < tt.numRuns; i++ {
			for i := 0; i < tt.numInserts; i++ {
				heap.Insert(&filter.FilterItem{Score: rand.Float32(), Data: []byte{}})
			}

			assert.Equal(t, tt.numInserts, heap.Size())

			var lastScore float32 = 2.0
			for i := 0; i < tt.numRemoves; i++ {
				curScore := heap.RemoveMax().GetScore()
				require.GreaterOrEqual(t, lastScore, curScore)
				lastScore = curScore
			}

			assert.Equal(t, tt.numInserts-tt.numRemoves, heap.Size())
			heap.Clear()
			assert.Equal(t, 0, heap.Size())
		}
	}
}

func TestConcurrentInsertAndRemoveMax(t *testing.T) {
	show()
	var tests = []struct {
		cap        int
		numInserts int
		numRemoves int
		numRuns    int
		threads    int
	}{
		{10, 10, 10, 1, 2},
		{100, 100, 0, 1, 10},
		{100000, 100000, 100000, 3, 2},
		{100000, 100000, 0, 2, 10},
		{27, 18, 17, 2, 10},
	}
	for _, tt := range tests {

		require.GreaterOrEqual(t, tt.cap, tt.numInserts,
			"Invalid test: number of inserts must be less than or equal to the capacity")

		require.GreaterOrEqual(t, tt.numInserts, tt.numRemoves,
			"Invalid test: number of inserts must be greater than or equal to the number of removes")

		heap := heapCtor(tt.cap)

		for i := 0; i < tt.numRuns; i++ {
			numJobs := tt.numInserts
			jobs := make(chan int, numJobs)
			results := make(chan int, numJobs)

			for w := 1; w <= tt.threads; w++ {
				go workerInsert(w, jobs, results, heap)
			}

			for j := 1; j <= numJobs; j++ {
				jobs <- j
			}
			close(jobs)

			for j := 1; j <= numJobs; j++ {
				<-results
			}

			assert.Equal(t, tt.numInserts, heap.Size())

			var lastScore float32 = 2.0
			for i := 0; i < tt.numRemoves; i++ {
				curScore := heap.RemoveMax().GetScore()
				require.GreaterOrEqual(t, lastScore, curScore)
				lastScore = curScore
			}

			assert.Equal(t, tt.numInserts-tt.numRemoves, heap.Size())
			heap.Clear()
			assert.Equal(t, 0, heap.Size())
		}
	}
}

func TestInsertAndRemoveMin(t *testing.T) {
	var tests = []struct {
		cap        int
		numInserts int
		numRemoves int
		numRuns    int
	}{
		{10, 10, 10, 100},
		{100, 100, 50, 100},
		{100, 100, 0, 100},
		{100000, 100000, 100000, 2},
		{27, 18, 17, 2},
	}
	for _, tt := range tests {
		require.GreaterOrEqual(t, tt.cap, tt.numInserts,
			"Invalid test: number of inserts must be less than or equal to the capacity")

		require.GreaterOrEqual(t, tt.numInserts, tt.numRemoves,
			"Invalid test: number of inserts must be greater than or equal to the number of removes")

		heap := heapCtor(tt.cap)
		for i := 0; i < tt.numRuns; i++ {
			for i := 0; i < tt.numInserts; i++ {
				heap.Insert(&filter.FilterItem{Score: rand.Float32(), Data: []byte{}})
			}

			assert.Equal(t, tt.numInserts, heap.Size())

			var lastScore float32 = -1.0
			for i := 0; i < tt.numRemoves; i++ {
				curScore := heap.RemoveMin().GetScore()
				require.LessOrEqual(t, lastScore, curScore)
				lastScore = curScore
			}

			assert.Equal(t, tt.numInserts-tt.numRemoves, heap.Size())
			heap.Clear()
			assert.Equal(t, 0, heap.Size())
		}
	}
}

func TestGetMaxAndMin(t *testing.T) {
	var tests = []struct {
		cap      int
		scores   []float32
		max, min float32
	}{
		{10, []float32{1.0}, 1.0, 1.0},
		{10, []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}, 9.0, 1.0},
		{10, []float32{1.0, 2.0, 9.0, 4.0, 5.0, 3.0, 7.0, 8.0, 6.0}, 9.0, 1.0},
		{8, []float32{1.0, 2.0, 9.0, 4.0, 5.0, 3.0, 7.0, 8.0, 6.0}, 9.0, 2.0},
		{1, []float32{1.0, 2.0}, 2.0, 2.0},
	}
	for _, tt := range tests {

		heap := heapCtor(tt.cap)

		for _, score := range tt.scores {
			heap.Insert(&filter.FilterItem{Score: score, Data: []byte{}})
		}

		assert.Equal(t, tt.max, heap.GetMax().GetScore())
		assert.Equal(t, tt.min, heap.GetMin().GetScore())
	}
}

func TestInsertPastCapacity(t *testing.T) {
	show()
	var tests = []struct {
		cap        int
		numInserts int
		numRuns    int
		threads    int
	}{
		{3, 8, 2, 8},
		{10, 100, 3, 10},
		{100, 150, 3, 10},
		{int(math.Pow(2, 17)) - 1, int(math.Pow(2, 18)), 1, 8},
	}
	for _, tt := range tests {

		heap := heapCtor(tt.cap)
		for i := 0; i < tt.numRuns; i++ {
			numJobs := tt.numInserts
			jobs := make(chan int, numJobs)
			results := make(chan int, numJobs)

			for w := 1; w <= tt.threads; w++ {
				go workerInsert(w, jobs, results, heap)
			}
			for j := 1; j <= numJobs; j++ {
				jobs <- j
			}
			close(jobs)

			for j := 1; j <= numJobs; j++ {
				<-results
			}

			var lastScore float32 = heap.RemoveMax().GetScore()
			for !heap.IsEmpty() {
				curScore := heap.RemoveMax().GetScore()
				require.GreaterOrEqual(t, lastScore, curScore)
				lastScore = curScore
			}
			heap.Clear()
			require.Equal(t, 0, heap.Size())
		}
	}
}
