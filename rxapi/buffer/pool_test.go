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
