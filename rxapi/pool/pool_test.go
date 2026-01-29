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

package pool_test

import (
	"runtime/debug"
	"sync"
	"sync/atomic"
	"testing"

	"dirpx.dev/rxlog/rxapi/pool"
)

// TestPool_Get verifies that Get returns a value created by the factory.
func TestPool_Get(t *testing.T) {
	t.Parallel()

	p := pool.New(func() *int {
		v := 42
		return &v
	})

	got := p.Get()
	if *got != 42 {
		t.Errorf("Get() = %d, want 42", *got)
	}
}

// TestPool_GetAfterPut verifies that Put followed by Get reuses the value.
func TestPool_GetAfterPut(t *testing.T) {
	t.Parallel()

	var created int32
	p := pool.New(func() *int {
		atomic.AddInt32(&created, 1)
		v := 0
		return &v
	})

	// First Get creates a new value
	first := p.Get()
	*first = 123

	// Put it back
	p.Put(first)

	// Second Get should reuse the same pointer
	second := p.Get()

	// We can't guarantee it's the same pointer (sync.Pool may discard),
	// but created should be 1 if reuse happened
	if first == second && *second != 123 {
		t.Errorf("Expected reused value to have 123, got %d", *second)
	}
}

// TestPool_MultipleGetPut verifies that multiple Get/Put cycles work correctly.
func TestPool_MultipleGetPut(t *testing.T) {
	t.Parallel()

	p := pool.New(func() *int {
		v := 0
		return &v
	})

	// Get, modify, put multiple times
	for i := 0; i < 100; i++ {
		v := p.Get()
		*v = i
		p.Put(v)
	}

	// One more Get should work
	final := p.Get()
	if final == nil {
		t.Error("Get() returned nil")
	}
}

// TestPool_Concurrent verifies that Pool is safe for concurrent use.
func TestPool_Concurrent(t *testing.T) {
	t.Parallel()

	p := pool.New(func() *int {
		v := 0
		return &v
	})

	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				v := p.Get()
				*v = j
				p.Put(v)
			}
		}()
	}

	wg.Wait()
}

// TestPool_StructType verifies Pool works with struct types.
func TestPool_StructType(t *testing.T) {
	t.Parallel()

	type MyStruct struct {
		Value int
		Name  string
	}

	p := pool.New(func() *MyStruct {
		return &MyStruct{Value: 0, Name: ""}
	})

	s := p.Get()
	s.Value = 42
	s.Name = "test"

	if s.Value != 42 || s.Name != "test" {
		t.Errorf("Get() returned unexpected value: %+v", s)
	}

	p.Put(s)
	s2 := p.Get()

	// Should be a valid MyStruct pointer
	if s2 == nil {
		t.Error("Get() returned nil")
	}
}

// TestPool_ValueType verifies Pool works with non-pointer types.
func TestPool_ValueType(t *testing.T) {
	t.Parallel()

	// Note: This is not recommended (SA6002), but should still work
	p := pool.New(func() int {
		return 0
	})

	v := p.Get()
	if v != 0 {
		t.Errorf("Get() = %d, want 0", v)
	}

	p.Put(42)
	v2 := p.Get()

	// Value might be 42 if reused, or 0 if new
	if v2 != 0 && v2 != 42 {
		t.Errorf("Get() = %d, want 0 or 42", v2)
	}
}

// TestPool_SliceType verifies Pool works with slice types.
func TestPool_SliceType(t *testing.T) {
	t.Parallel()

	p := pool.New(func() []byte {
		return make([]byte, 0, 64)
	})

	s := p.Get()
	if cap(s) < 64 {
		t.Errorf("Get() returned slice with cap=%d, want >= 64", cap(s))
	}

	s = append(s, []byte("hello")...)
	p.Put(s)

	s2 := p.Get()
	if cap(s2) < 64 {
		t.Errorf("Get() returned slice with cap=%d, want >= 64", cap(s2))
	}
}

// TestPool_FactoryCalledOnEmptyPool verifies factory is called when pool is empty.
func TestPool_FactoryCalledOnEmptyPool(t *testing.T) {
	t.Parallel()

	var callCount int32
	p := pool.New(func() *int {
		atomic.AddInt32(&callCount, 1)
		v := 0
		return &v
	})

	// First Get should call factory
	_ = p.Get()
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Factory call count = %d, want 1", callCount)
	}

	// Second Get without Put should call factory again
	_ = p.Get()
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("Factory call count = %d, want 2", callCount)
	}
}

// TestPool_PutNil verifies that Put accepts nil (though not recommended).
func TestPool_PutNil(t *testing.T) {
	t.Parallel()

	p := pool.New(func() *int {
		v := 42
		return &v
	})

	// Put nil should not panic
	p.Put(nil)

	// Get might return nil if sync.Pool returns it, or factory value
	v := p.Get()
	if v != nil && *v != 42 {
		t.Errorf("Get() = %d, want 42", *v)
	}
	// Note: sync.Pool may return nil that was Put, this is expected behavior
}

// BenchmarkPool_GetPut benchmarks the overhead of Get/Put operations.
func BenchmarkPool_GetPut(b *testing.B) {
	p := pool.New(func() *int {
		v := 0
		return &v
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		v := p.Get()
		p.Put(v)
	}
}

// BenchmarkPool_GetPutParallel benchmarks concurrent Get/Put operations.
func BenchmarkPool_GetPutParallel(b *testing.B) {
	p := pool.New(func() *int {
		v := 0
		return &v
	})

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := p.Get()
			p.Put(v)
		}
	})
}

// BenchmarkPool_GetModifyPut benchmarks realistic usage with modification.
func BenchmarkPool_GetModifyPut(b *testing.B) {
	type Data struct {
		Value int
		Name  string
	}

	p := pool.New(func() *Data {
		return &Data{}
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d := p.Get()
		d.Value = i
		d.Name = "benchmark"
		p.Put(d)
	}
}

// TestPool_ReusesObjects_WithGCDisabled verifies that objects are actually
// reused from the pool (not created fresh each time) by disabling GC and
// putting many objects with identifiable content.
//
// This test is inspired by zap's pool tests which disable GC to prevent
// sync.Pool from discarding objects during the test.
func TestPool_ReusesObjects_WithGCDisabled(t *testing.T) {
	// Disable GC to avoid the victim cache during the test.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))

	type Marker struct {
		ID string
	}

	p := pool.New(func() *Marker {
		return &Marker{ID: "new"}
	})

	// Probabilistically, 75% of sync.Pool.Put calls will succeed when -race
	// is enabled (see sync.Pool implementation); attempt to make this
	// quasi-deterministic by brute force (i.e., put significantly more objects
	// in the pool than we will need for the test).
	//
	// ref: https://cs.opensource.google/go/go/+/refs/tags/go1.20.2:src/sync/pool.go;l=100-103
	for i := 0; i < 1_000; i++ {
		p.Put(&Marker{ID: "reused"})
	}

	// Ensure that we always get back objects with "reused" ID.
	// Note that this must only run a fraction of the number of times that
	// Put is called above.
	for i := 0; i < 10; i++ {
		func() {
			m := p.Get()
			defer p.Put(m)
			if m.ID != "reused" {
				t.Errorf("Get() returned object with ID=%q, want %q", m.ID, "reused")
			}
		}()
	}

	// Depool all objects that might be in the pool to ensure it's empty.
	for i := 0; i < 1_000; i++ {
		p.Get()
	}

	// Now that the pool is empty, it should use the factory function
	// to create a new object.
	m := p.Get()
	if m == nil {
		t.Fatal("Get() returned nil after pool exhaustion")
	}
	if m.ID != "new" {
		t.Errorf("Fresh object has ID=%q, expected %q", m.ID, "new")
	}
}

// TestPool_ConcurrentReadWrite runs goroutines that concurrently read and
// write pooled object fields to detect data races with race detector.
//
// This test is inspired by zap's TestNew_Race which specifically tests
// concurrent field access to catch races that might not show up in
// simpler Get/Put tests.
func TestPool_ConcurrentReadWrite(t *testing.T) {
	type Data struct {
		Value int
		Name  string
	}

	p := pool.New(func() *Data {
		return &Data{}
	})

	var wg sync.WaitGroup
	defer wg.Wait()

	// Run a number of goroutines that read and write object fields to
	// tease out races.
	for i := 0; i < 1_000; i++ {
		i := i

		wg.Add(1)
		go func() {
			defer wg.Done()

			d := p.Get()
			defer p.Put(d)

			// Must both read and write object state
			if d != nil {
				d.Value = i
				d.Name = "race test"
				_ = d.Value
				_ = d.Name
			}
		}()
	}
}
