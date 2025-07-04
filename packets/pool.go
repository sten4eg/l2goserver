package packets

import (
	"sort"
	"sync"
	"sync/atomic"
)

const (
	minBitSize = 6  // 2^6=64 corresponds to the CPU cache line size
	steps      = 20 // Number of steps for buffer sizes

	minSize = 1 << minBitSize // Minimum buffer size
	// Maximum size is calculated dynamically during calibration

	calibrateCallsThreshold = 42000 // Threshold of calls to trigger calibration
	maxPercentile           = 0.95  // Percentile to determine the maximum size
)

// pools represents a pool of byte buffers.
//
// Different pools can be used for different types of buffers.
// Properly defined buffer types with their own pools help reduce
// memory usage.
type pools struct {
	calls [steps]uint64 // Call counters for each buffer size

	calibrating uint64 // Flag indicating calibration in progress (0 or 1)

	defaultSize uint64 // Default size for new buffers
	maxSize     uint64 // Maximum allowed size to return to the pool

	pool sync.Pool // Main pool for storing buffers
}

var pool pools // Global instance of the pool

// GetBuffer returns an empty buffer from the pool.
//
// After use, the buffer should be returned to the pool via PutBuffer.
// This reduces memory allocations for buffer management.
func GetBuffer() *Buffer { return pool.get() }

// get returns a new buffer with length 0.
//
// The buffer can be returned to the pool via Put to reduce GC pressure.
func (p *pools) get() *Buffer {
	v := p.pool.Get()
	if v != nil {
		return v.(*Buffer)
	}
	return &Buffer{
		b: make([]byte, 0, atomic.LoadUint64(&p.defaultSize)),
	}
}

// Put returns a buffer to the pool.
//
// The buffer must not be used after being returned to the pool
// to avoid race conditions.
func Put(b *Buffer) { pool.put(b) }

// put returns a buffer to the pool.
//
// After returning to the pool, the buffer must not be accessed.
func (p *pools) put(b *Buffer) {
	// Determine the size index in the calls array
	idx := index(len(b.b))

	// If call threshold is exceeded, trigger calibration
	if atomic.AddUint64(&p.calls[idx], 1) > calibrateCallsThreshold {
		p.calibrate()
	}

	// Check the maximum allowed size
	maxSize := int(atomic.LoadUint64(&p.maxSize))
	if maxSize == 0 || cap(b.b) <= maxSize {
		b.Reset()
		p.pool.Put(b)
	}
}

// calibrate adjusts optimal buffer sizes based on usage statistics.
func (p *pools) calibrate() {
	// Acquire calibration flag
	if !atomic.CompareAndSwapUint64(&p.calibrating, 0, 1) {
		return
	}

	// Collect usage statistics
	a := make(callSizes, 0, steps)
	var callsSum uint64
	for i := uint64(0); i < steps; i++ {
		calls := atomic.SwapUint64(&p.calls[i], 0)
		callsSum += calls
		a = append(a, callSize{
			calls: calls,
			size:  minSize << i,
		})
	}

	// Sort sizes by descending frequency of use
	sort.Sort(a)

	// Compute default and maximum buffer sizes
	defaultSize := a[0].size
	maxSize := defaultSize

	// Determine size covering 95% of the requests
	maxSum := uint64(float64(callsSum) * maxPercentile)
	callsSum = 0
	for i := 0; i < steps; i++ {
		if callsSum > maxSum {
			break
		}
		callsSum += a[i].calls
		size := a[i].size
		if size > maxSize {
			maxSize = size
		}
	}

	// Update pool settings
	atomic.StoreUint64(&p.defaultSize, defaultSize)
	atomic.StoreUint64(&p.maxSize, maxSize)

	// Reset calibration flag
	atomic.StoreUint64(&p.calibrating, 0)
}

// callSize represents the relation between call count and buffer size
type callSize struct {
	calls uint64 // Number of requests
	size  uint64 // Buffer size
}

// callSizes implements sorting interface for a slice of callSize
type callSizes []callSize

func (ci callSizes) Len() int           { return len(ci) }
func (ci callSizes) Less(i, j int) bool { return ci[i].calls > ci[j].calls }
func (ci callSizes) Swap(i, j int)      { ci[i], ci[j] = ci[j], ci[i] }

// index computes the index in the calls array for a given buffer size
func index(n int) int {
	n--
	n >>= minBitSize
	idx := 0
	for n > 0 {
		n >>= 1
		idx++
	}
	// Limit the maximum index
	if idx >= steps {
		idx = steps - 1
	}
	return idx
}
