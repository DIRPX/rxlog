/*
   Copyright 2025 The DIRPX Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package sizedbufferpool

import (
	"dirpx.dev/rxlog/rxapi/buffer"
)

// Size bucket boundaries define the capacity ranges for buffer pooling.
//
// Each bucket pools buffers within a specific capacity range to reduce
// memory fragmentation and improve cache locality. Buffers are allocated
// with the minimum capacity for their bucket and will be returned to the
// same bucket when freed.
const (
	// BucketSmall pools buffers from 0 to 512 bytes.
	//
	// This bucket is optimized for short log messages with few fields,
	// typically debug or trace logs. The 512-byte threshold accommodates
	// most single-line JSON logs without additional allocations.
	BucketSmall = 512

	// BucketMedium pools buffers from 513 to 1024 bytes.
	//
	// This bucket handles typical structured logs with 5-10 fields,
	// including common metadata (timestamp, level, caller, message).
	// Most production logs fall into this category.
	BucketMedium = 1024

	// BucketLarge pools buffers from 1025 to 2048 bytes.
	//
	// This bucket accommodates logs with extensive context fields,
	// error details, or nested objects. It covers ~95% of all logs
	// in typical applications.
	BucketLarge = 2048

	// BucketXLarge pools buffers from 2049 to 4096 bytes.
	//
	// This bucket handles logs with large payloads, such as request/response
	// bodies, stack traces, or detailed audit information. Buffers in this
	// range are less common but still benefit from pooling.
	BucketXLarge = 4096

	// BucketXXLarge pools buffers from 4097 to 8192 bytes.
	//
	// This is the largest pooled bucket. Buffers exceeding this size are
	// rare (typically <1% of logs) and represent exceptional cases such as
	// crash dumps or debug snapshots. We pool them to avoid GC pressure
	// from large allocations, but we do not create buckets beyond this size
	// to limit memory overhead from empty pools.
	BucketXXLarge = 8192

	// MaxPooledSize defines the upper limit for buffer pooling.
	//
	// Buffers larger than this threshold are allocated without pooling and
	// will be garbage collected after use. This prevents the pool from
	// retaining pathologically large buffers that would waste memory.
	//
	// Callers that consistently produce buffers exceeding this size SHOULD
	// consider whether they are logging appropriately sized data or if
	// the data belongs in a separate storage system.
	MaxPooledSize = BucketXXLarge
)

// Pool is a size-bucketed buffer pool that segregates buffers by capacity
// to minimize fragmentation and improve allocation performance.
//
// Unlike a single global sync.Pool, this implementation maintains separate
// buffer.Pool instances for different size classes. When a buffer is
// requested via GetSize, the Pool selects the smallest bucket that can
// satisfy the request. Buffers automatically return to their originating
// bucket when freed via buf.Free().
//
// This design has several benefits:
//
//   - Reduced fragmentation: Buffers of similar sizes are pooled together,
//     reducing the likelihood of returning a 4KB buffer when 512 bytes
//     would suffice.
//
//   - Improved cache locality: Smaller buckets fit more buffers in CPU
//     cache, reducing cache misses during high-throughput logging.
//
//   - Predictable memory overhead: Each bucket has a known maximum size,
//     making it easier to reason about memory usage under load.
//
// Note: Buffers are NOT dynamically re-bucketed based on growth. A buffer
// allocated from the small bucket will return to the small bucket even if
// it grows to 2KB during use. This is by design to keep the implementation
// simple and compatible with the existing buffer.Pool architecture. In
// practice, this works well because most buffers do not grow significantly,
// and occasional larger buffers in smaller buckets do not cause problems.
//
// Concurrency:
//
// Pool is safe for concurrent use by multiple goroutines. All methods
// (Get, GetSize) delegate to buffer.Pool instances which are themselves
// thread-safe.
type Pool struct {
	// buckets holds one buffer.Pool per size class.
	//
	// Each bucket creates and manages buffers within its capacity range.
	// Buckets are indexed by size class (0 = small, 1 = medium, etc.).
	//
	// The buckets array is immutable after initialization and may be
	// accessed without synchronization.
	buckets [5]buffer.Pool
}

// _pool is the process-wide singleton sized buffer pool.
//
// This global instance is initialized at package load time and provides
// a convenient, zero-configuration API for callers. Applications that
// require multiple independent pools (for example, for multi-tenancy or
// different log destinations) MAY create additional Pool instances via
// New, but most use cases SHOULD use the global Get and GetSize functions
// that delegate to this singleton.
var _pool = New()

// New creates a new sized buffer pool with five size-class buckets.
//
// The returned Pool is immediately ready for use and safe for concurrent
// access. Each bucket will allocate buffers on demand when the pool is
// empty; callers do not need to pre-populate the pool.
//
// Most callers SHOULD use the package-level Get and GetSize functions,
// which operate on a shared global pool. New is provided for advanced use
// cases that require isolated pooling (for example, per-tenant pools or
// pools with custom lifecycle management).
func New() *Pool {
	p := &Pool{}

	// Initialize each bucket with a buffer.Pool that creates buffers of
	// the appropriate size class.
	//
	// Each bucket's pool wraps a generic pool.Pool[*buffer.Buffer] with
	// a factory function that allocates buffers with the minimum capacity
	// for that size class.
	//
	// Buffers obtained from a bucket will automatically return to that
	// same bucket when buf.Free() is called, because the buffer.Pool
	// reference is stored in the buffer itself.
	p.buckets[0] = newBucketPool(BucketSmall)
	p.buckets[1] = newBucketPool(BucketMedium)
	p.buckets[2] = newBucketPool(BucketLarge)
	p.buckets[3] = newBucketPool(BucketXLarge)
	p.buckets[4] = newBucketPool(BucketXXLarge)

	return p
}

// newBucketPool creates a buffer.Pool that allocates buffers with the
// specified capacity.
//
// This is a helper function used during Pool initialization to create
// the individual bucket pools. The returned buffer.Pool wraps a generic
// pool that creates *buffer.Buffer instances with the given capacity.
func newBucketPool(capacity int) buffer.Pool {
	return buffer.NewPoolWithCapacity(capacity)
}

// Get retrieves a buffer from the smallest bucket (512 bytes).
//
// This is the most common API for callers that do not know the required
// size in advance. The returned buffer will have at least 512 bytes of
// capacity and may be reused from a previous allocation or freshly
// allocated if the pool is empty.
//
// The returned buffer is owned by the caller and MUST be returned via
// buf.Free() when no longer needed. The buffer automatically knows which
// bucket it belongs to and will return itself to the correct pool when
// Free is called.
//
// Concurrency:
//
// Get is safe to call from multiple goroutines concurrently.
func (p *Pool) Get() *buffer.Buffer {
	return p.GetSize(BucketSmall)
}

// GetSize retrieves a buffer with at least the specified capacity.
//
// The Pool will select the smallest bucket that can satisfy the request.
// If size is less than or equal to BucketSmall, a buffer from the small
// bucket is returned.
//
// If size exceeds MaxPooledSize (8KB), GetSize allocates from the XXLarge
// bucket anyway. The buffer will grow to accommodate the requested size on
// first write, and when freed, it will be returned to the XXLarge pool.
// This means pathologically large buffers MAY be retained in the pool.
// Callers that consistently require buffers exceeding 8KB SHOULD consider
// whether they are logging appropriately sized data or if the data belongs
// in a separate storage system.
//
// Example:
//
//	buf := pool.GetSize(1500)  // Returns buffer from BucketLarge (2048 bytes)
//	defer buf.Free()
//
// The size parameter is a hint; the returned buffer MAY have greater
// capacity than requested. Callers MUST NOT assume the buffer has
// exactly size bytes of capacity.
//
// Important: Buffers are NOT dynamically re-bucketed. A buffer obtained
// from the small bucket will return to the small bucket even if it grows
// to 4KB during use. This is by design and does not cause problems in
// practice.
//
// Concurrency:
//
// GetSize is safe to call from multiple goroutines concurrently.
func (p *Pool) GetSize(size int) *buffer.Buffer {
	// Select bucket based on requested size.
	// We use <= comparisons to ensure that size == BucketSmall returns
	// the small bucket, not the medium bucket.
	//
	// Each bucket's buffer.Pool handles the Get operation and
	// automatically assigns the pool reference to the returned buffer,
	// so buf.Free() will work correctly.
	//
	// Note: Requests exceeding MaxPooledSize are satisfied from the
	// XXLarge bucket. The buffer will grow as needed, and may be retained
	// in the pool at its grown size.

	switch {
	case size <= BucketSmall:
		return p.buckets[0].Get()
	case size <= BucketMedium:
		return p.buckets[1].Get()
	case size <= BucketLarge:
		return p.buckets[2].Get()
	case size <= BucketXLarge:
		return p.buckets[3].Get()
	default:
		// For sizes exceeding MaxPooledSize, use the XXLarge bucket.
		// The buffer will grow on first write to accommodate the size.
		return p.buckets[4].Get()
	}
}

// Package-level convenience functions that operate on the global pool.

// Get retrieves a buffer from the global sized pool.
//
// This is the most common entry point for callers. The returned buffer
// will have at least 512 bytes of capacity and MUST be returned via
// buf.Free() when no longer needed.
//
// Example:
//
//	buf := sizedbufferpool.Get()
//	defer buf.Free()
//
//	buf.AppendString(`{"level":"info","msg":"hello"}`)
//	writer.Write(buf.Bytes())
//
// The buffer automatically knows which bucket it came from and will
// return to that bucket when Free is called. Callers do not need to
// track which pool or bucket the buffer originated from.
//
// Concurrency:
//
// Get is safe to call from multiple goroutines concurrently.
func Get() *buffer.Buffer {
	return _pool.Get()
}

// GetSize retrieves a buffer with at least the specified capacity from the
// global sized pool.
//
// See Pool.GetSize for detailed semantics. The returned buffer MUST be
// freed via buf.Free() when no longer needed.
//
// Example:
//
//	buf := sizedbufferpool.GetSize(1500)  // Gets 2KB buffer
//	defer buf.Free()
//
// Concurrency:
//
// GetSize is safe to call from multiple goroutines concurrently.
func GetSize(size int) *buffer.Buffer {
	return _pool.GetSize(size)
}
