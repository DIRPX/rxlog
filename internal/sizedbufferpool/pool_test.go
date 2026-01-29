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
	"sync"
	"testing"
)

// TestBucketSelection verifies that GetSize selects the correct bucket
// for various requested sizes.
func TestBucketSelection(t *testing.T) {
	p := New()

	tests := []struct {
		name         string
		requestSize  int
		minCapacity  int
		maxCapacity  int
	}{
		{
			name:        "small bucket (0 bytes)",
			requestSize: 0,
			minCapacity: BucketSmall,
			maxCapacity: BucketSmall,
		},
		{
			name:        "small bucket (512 bytes)",
			requestSize: 512,
			minCapacity: BucketSmall,
			maxCapacity: BucketSmall,
		},
		{
			name:        "medium bucket (513 bytes)",
			requestSize: 513,
			minCapacity: BucketMedium,
			maxCapacity: BucketMedium,
		},
		{
			name:        "medium bucket (1024 bytes)",
			requestSize: 1024,
			minCapacity: BucketMedium,
			maxCapacity: BucketMedium,
		},
		{
			name:        "large bucket (1025 bytes)",
			requestSize: 1025,
			minCapacity: BucketLarge,
			maxCapacity: BucketLarge,
		},
		{
			name:        "large bucket (2048 bytes)",
			requestSize: 2048,
			minCapacity: BucketLarge,
			maxCapacity: BucketLarge,
		},
		{
			name:        "xlarge bucket (2049 bytes)",
			requestSize: 2049,
			minCapacity: BucketXLarge,
			maxCapacity: BucketXLarge,
		},
		{
			name:        "xlarge bucket (4096 bytes)",
			requestSize: 4096,
			minCapacity: BucketXLarge,
			maxCapacity: BucketXLarge,
		},
		{
			name:        "xxlarge bucket (4097 bytes)",
			requestSize: 4097,
			minCapacity: BucketXXLarge,
			maxCapacity: BucketXXLarge,
		},
		{
			name:        "xxlarge bucket (8192 bytes)",
			requestSize: 8192,
			minCapacity: BucketXXLarge,
			maxCapacity: BucketXXLarge,
		},
		{
			name:        "exceeds max pooled size",
			requestSize: 10000,
			minCapacity: BucketXXLarge,
			maxCapacity: BucketXXLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := p.GetSize(tt.requestSize)
			defer buf.Free()

			cap := buf.Cap()
			if cap < tt.minCapacity {
				t.Errorf("capacity %d is less than minimum %d", cap, tt.minCapacity)
			}
			if cap > tt.maxCapacity {
				t.Errorf("capacity %d exceeds maximum %d", cap, tt.maxCapacity)
			}
		})
	}
}

// TestGetFreeCycle verifies that buffers can be retrieved and returned
// to the pool without leaking or corrupting state.
func TestGetFreeCycle(t *testing.T) {
	p := New()

	// Get a buffer, write some data, and return it.
	buf := p.Get()
	buf.AppendString("test data")

	if buf.Len() != 9 {
		t.Errorf("expected length 9, got %d", buf.Len())
	}

	buf.Free()

	// Get another buffer and verify it's been reset.
	buf2 := p.Get()
	defer buf2.Free()

	if buf2.Len() != 0 {
		t.Errorf("expected reset buffer with length 0, got %d", buf2.Len())
	}
}

// TestBufferGrowthAndReallocation verifies that a buffer that grows
// beyond its original bucket returns to its original bucket (not
// re-bucketed based on size). This is by design to keep the implementation
// simple.
func TestBufferGrowthAndReallocation(t *testing.T) {
	p := New()

	// Get a small buffer (512 bytes).
	buf := p.GetSize(BucketSmall)
	initialCap := buf.Cap()

	if initialCap != BucketSmall {
		t.Errorf("expected initial capacity %d, got %d", BucketSmall, initialCap)
	}

	// Write enough data to force growth into the large bucket range.
	largeData := make([]byte, BucketMedium+1)
	buf.AppendBytes(largeData)

	newCap := buf.Cap()
	if newCap <= BucketMedium {
		t.Errorf("expected capacity > %d after growth, got %d", BucketMedium, newCap)
	}

	// Return the buffer. It will go back to the small bucket despite
	// having grown to large size.
	buf.Free()

	// Request a small buffer. We might get back the grown buffer
	// (which is now in the small bucket pool despite being large).
	buf2 := p.GetSize(BucketSmall)
	defer buf2.Free()

	// buf2 should have at least BucketSmall capacity, but might be larger
	// if we got back the grown buffer.
	if buf2.Cap() < BucketSmall {
		t.Errorf("expected capacity >= %d, got %d", BucketSmall, buf2.Cap())
	}
}

// TestOversizedBuffer verifies that buffers exceeding MaxPooledSize
// are allocated from the XXLarge bucket and will grow as needed.
func TestOversizedBuffer(t *testing.T) {
	p := New()

	// Request a buffer larger than MaxPooledSize.
	// It will come from the XXLarge bucket with 8KB capacity.
	buf := p.GetSize(MaxPooledSize + 1000)

	// Initial capacity should be from XXLarge bucket (8KB).
	if buf.Cap() != BucketXXLarge {
		t.Errorf("expected capacity %d, got %d", BucketXXLarge, buf.Cap())
	}

	// Write enough data to force growth beyond MaxPooledSize.
	largeData := make([]byte, MaxPooledSize+1000)
	buf.AppendBytes(largeData)

	// Buffer should have grown to accommodate the data.
	if buf.Cap() < MaxPooledSize+1000 {
		t.Errorf("expected capacity >= %d after growth, got %d",
			MaxPooledSize+1000, buf.Cap())
	}

	// Return it to the pool. It will be retained in the XXLarge bucket
	// at its grown size.
	buf.Free()

	// Request another XXLarge buffer. We might get back the grown buffer.
	buf2 := p.GetSize(BucketXXLarge)
	defer buf2.Free()

	// Should have at least XXLarge capacity.
	if buf2.Cap() < BucketXXLarge {
		t.Errorf("expected capacity >= %d, got %d", BucketXXLarge, buf2.Cap())
	}
}

// TestConcurrentGetFree verifies that the pool is safe for concurrent
// use by multiple goroutines.
func TestConcurrentGetFree(t *testing.T) {
	p := New()

	const numGoroutines = 100
	const numIterations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				// Alternate between different size requests to exercise
				// multiple buckets.
				size := BucketSmall
				if j%2 == 0 {
					size = BucketMedium
				}
				if j%3 == 0 {
					size = BucketLarge
				}

				buf := p.GetSize(size)
				buf.AppendString("concurrent test")
				buf.Free()
			}
		}(i)
	}

	wg.Wait()
}

// TestGlobalPoolFunctions verifies that the package-level Get and GetSize
// functions correctly delegate to the global pool.
func TestGlobalPoolFunctions(t *testing.T) {
	buf := Get()
	if buf == nil {
		t.Fatal("Get() returned nil")
	}
	buf.AppendString("global pool test")
	buf.Free()

	buf2 := GetSize(BucketLarge)
	if buf2 == nil {
		t.Fatal("GetSize() returned nil")
	}
	if buf2.Cap() < BucketLarge {
		t.Errorf("expected capacity >= %d, got %d", BucketLarge, buf2.Cap())
	}
	buf2.Free()
}

// BenchmarkGetFree measures the performance of the Get/Free cycle
// for various buffer sizes.
func BenchmarkGetFree(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"Small_512B", BucketSmall},
		{"Medium_1KB", BucketMedium},
		{"Large_2KB", BucketLarge},
		{"XLarge_4KB", BucketXLarge},
		{"XXLarge_8KB", BucketXXLarge},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			p := New()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				buf := p.GetSize(bm.size)
				buf.AppendString("benchmark test data")
				buf.Free()
			}
		})
	}
}

// BenchmarkGetFreeParallel measures the concurrent performance of the
// sized buffer pool under high contention.
func BenchmarkGetFreeParallel(b *testing.B) {
	p := New()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := p.Get()
			buf.AppendString("parallel benchmark")
			buf.Free()
		}
	})
}
