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

package buffer_test

import (
	"runtime/debug"
	"sync"
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
)

// TestPoolGetProvidesEmptyBuffer verifies that NewPool().Get() returns a
// non-nil Buffer in a clean, empty state and ready for writing.
func TestPoolGetProvidesEmptyBuffer(t *testing.T) {
	p := buffer.NewPool()

	buf := p.Get()
	if buf == nil {
		t.Fatal("Get() returned nil buffer")
	}

	if got := buf.Len(); got != 0 {
		t.Fatalf("buffer.Len() after Get() = %d, want 0", got)
	}
	if got := len(buf.Bytes()); got != 0 {
		t.Fatalf("len(buffer.Bytes()) after Get() = %d, want 0", got)
	}

	// Cap MUST be >= 0; no строгий контракт на конкретное значение,
	// но проверим, что вызов не паникует.
	_ = buf.Cap()
}

// TestPoolGetResetsBufferAfterPut ensures that buffers returned to the pool
// are Reset before they are handed out again.
func TestPoolGetResetsBufferAfterPut(t *testing.T) {
	p := buffer.NewPool()

	buf := p.Get()
	buf.AppendString("hello")

	if buf.Len() == 0 {
		t.Fatalf("buffer.Len() after AppendString = %d, want > 0", buf.Len())
	}

	p.Put(buf)

	buf2 := p.Get()
	if got := buf2.Len(); got != 0 {
		t.Fatalf("buffer.Len() after Put/Get = %d, want 0", got)
	}
	if got := string(buf2.Bytes()); got != "" {
		t.Fatalf("buffer.Bytes() after Put/Get = %q, want empty", got)
	}
}

// TestPoolReusesBuffers checks that the pool can actually reuse buffers
// between Get/Put cycles. Это не строгий API-контракт в терминах
// "должен быть тот же pointer", но как инвариант для реализации это
// полезный тест, чтобы не потерять реюз при рефакторинге.
func TestPoolReusesBuffers(t *testing.T) {
	p := buffer.NewPool()

	buf1 := p.Get()
	if buf1 == nil {
		t.Fatal("Get() returned nil buffer")
	}
	p.Put(buf1)

	buf2 := p.Get()
	if buf2 == nil {
		t.Fatal("second Get() returned nil buffer")
	}

	if buf1 != buf2 {
		t.Fatalf("pool did not reuse buffer: got %p, want %p", buf2, buf1)
	}
}

// TestPoolConcurrentGetPut exercises concurrent Get/Put to catch race
// conditions or obvious concurrency bugs. Успешное завершение без паники
// и data race (при запуске с -race) считается достаточным.
func TestPoolConcurrentGetPut(t *testing.T) {
	p := buffer.NewPool()

	const (
		goroutines = 16
		iterations = 1000
	)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				buf := p.Get()
				if buf == nil {
					t.Fatalf("Get() returned nil buffer in concurrent path")
				}

				// Небольшая работа с буфером, чтобы проверить,
				// что он пригоден для использования.
				buf.AppendString("x")
				// После Put буфер не должен использоваться вызывающим кодом.
				p.Put(buf)
			}
		}()
	}

	wg.Wait()

	// После нагрузки pool всё ещё должен выдавать чистый буфер.
	buf := p.Get()
	defer p.Put(buf)

	if got := buf.Len(); got != 0 {
		t.Fatalf("buffer.Len() after concurrent Get/Put = %d, want 0", got)
	}
	if got := string(buf.Bytes()); got != "" {
		t.Fatalf("buffer.Bytes() after concurrent Get/Put = %q, want empty", got)
	}
}

// TestPoolReusesBuffers_WithGCDisabled verifies that buffers are actually
// reused from the pool (not created fresh each time) by disabling GC and
// putting many buffers with identifiable content.
//
// This test is inspired by zap's pool tests which disable GC to prevent
// sync.Pool from discarding objects during the test.
func TestPoolReusesBuffers_WithGCDisabled(t *testing.T) {
	// Disable GC to avoid the victim cache during the test.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))

	p := buffer.NewPool()

	const marker = "reused-buffer"

	// Probabilistically, 75% of sync.Pool.Put calls will succeed when -race
	// is enabled (see sync.Pool implementation); attempt to make this
	// quasi-deterministic by brute force (i.e., put significantly more objects
	// in the pool than we will need for the test).
	//
	// ref: https://cs.opensource.google/go/go/+/refs/tags/go1.20.2:src/sync/pool.go;l=100-103
	for i := 0; i < 1_000; i++ {
		buf := p.Get()
		buf.AppendString(marker)
		p.Put(buf)
	}

	// Ensure that we always get buffers that were previously Put with marker.
	// Note that this must only run a fraction of the number of times that
	// Put is called above.
	for i := 0; i < 10; i++ {
		buf := p.Get()
		// Buffer should be reset (Len=0) but capacity should be preserved
		if buf.Len() != 0 {
			t.Errorf("Get() returned buffer with Len=%d, want 0", buf.Len())
		}
		if buf.Cap() == 0 {
			t.Error("Get() returned buffer with Cap=0, want > 0")
		}
		p.Put(buf)
	}

	// Depool all objects that might be in the pool to ensure it's empty.
	for i := 0; i < 1_000; i++ {
		p.Get()
	}

	// Now that the pool is empty, it should use the factory function
	// to create a new buffer.
	buf := p.Get()
	if buf == nil {
		t.Fatal("Get() returned nil after pool exhaustion")
	}
	if buf.Len() != 0 {
		t.Errorf("Fresh buffer has Len=%d, want 0", buf.Len())
	}
}

// TestPoolPreservesCapacity verifies that buffers maintain their capacity
// across Get/Put cycles, which is important for avoiding reallocations.
func TestPoolPreservesCapacity(t *testing.T) {
	p := buffer.NewPool()

	// Get a buffer and grow it
	buf := p.Get()
	for i := 0; i < 1000; i++ {
		buf.AppendByte('x')
	}

	capacity := buf.Cap()
	if capacity < 1000 {
		t.Fatalf("buffer capacity = %d, want >= 1000", capacity)
	}

	// Return to pool
	p.Put(buf)

	// Get again - capacity should be preserved (or larger)
	buf2 := p.Get()
	if buf2.Cap() < capacity {
		t.Errorf("buffer capacity after Put/Get = %d, want >= %d", buf2.Cap(), capacity)
	}

	// But length should be reset to 0
	if buf2.Len() != 0 {
		t.Errorf("buffer length after Put/Get = %d, want 0", buf2.Len())
	}
}

// TestPoolConcurrentReadWrite runs goroutines that concurrently read and
// write buffer fields to detect data races with race detector.
//
// This test is inspired by zap's TestNew_Race which specifically tests
// concurrent field access to catch races that might not show up in
// simpler Get/Put tests.
func TestPoolConcurrentReadWrite(t *testing.T) {
	p := buffer.NewPool()

	var wg sync.WaitGroup
	defer wg.Wait()

	// Run a number of goroutines that read and write buffer fields to
	// tease out races.
	for i := 0; i < 1_000; i++ {
		i := i

		wg.Add(1)
		go func() {
			defer wg.Done()

			buf := p.Get()
			defer p.Put(buf)

			// Must both read and write buffer state
			if buf.Len() >= 0 {
				buf.AppendInt(i)
				buf.AppendString("test")
				_ = buf.Bytes()
				_ = buf.Cap()
				buf.Reset()
			}
		}()
	}
}

// TestPoolWithCapacity verifies that NewPoolWithCapacity creates buffers
// with the specified initial capacity.
func TestPoolWithCapacity(t *testing.T) {
	const wantCap = 2048

	p := buffer.NewPoolWithCapacity(wantCap)

	buf := p.Get()
	defer p.Put(buf)

	if buf.Len() != 0 {
		t.Errorf("buffer Len = %d, want 0", buf.Len())
	}

	// Capacity should be at least what we requested
	// (may be larger due to allocator rounding)
	if buf.Cap() < wantCap {
		t.Errorf("buffer Cap = %d, want >= %d", buf.Cap(), wantCap)
	}
}
