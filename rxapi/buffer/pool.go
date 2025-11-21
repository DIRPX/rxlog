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

package buffer

import (
	"dirpx.dev/rxlog/rxapi/pool"
)

// Pool manages the reuse of *Buffer instances across log entries and other
// serialized payloads.
//
// This concrete type is a thin, type-safe wrapper around a generic
// [pool.Pool[*Buffer]]. It is responsible for handing out buffers via Get
// and accepting them back via Put (directly or indirectly through Buffer.Free,
// depending on how Buffer is implemented).
//
// The primary purpose of Pool is to reduce allocations and GC pressure by
// recycling the underlying byte slices used by Buffer. Callers MUST respect
// the ownership and lifecycle contract: every *Buffer obtained from Get
// SHOULD eventually be returned to the pool, either by calling Put explicitly
// or by invoking whatever higher-level release mechanism Buffer exposes
// (for example, a Free method).
//
// Unless explicitly documented otherwise, Pool SHOULD be safe for concurrent
// use by multiple goroutines and MAY be shared as a process-wide or
// logger-wide singleton.
type Pool struct {
	p *pool.Pool[*Buffer]
}

// NewPool constructs a new Pool that allocates *Buffer instances with a
// predefined initial capacity.
//
// Internally, NewPool configures the underlying generic pool to create
// Buffer values whose internal byte slice starts empty (len == 0) and has
// a default capacity of Size. The exact value of Size is an implementation
// detail, but callers MAY assume that buffers obtained from this Pool are
// immediately ready for writing without further initialization.
func NewPool() Pool {
	return Pool{
		p: pool.New(func() *Buffer {
			return &Buffer{
				data: make([]byte, 0, Size),
			}
		}),
	}
}

// Get acquires a *Buffer from the pool, creating a new one if none are
// currently available.
//
// Before the buffer is returned to the caller, Get:
//
//   - resets its logical contents (via Reset), so that it appears empty
//     while preserving any previously allocated capacity, and
//   - associates the buffer with this Pool (buf.pool = p), enabling the
//     buffer to return itself to the correct pool if it provides a
//     higher-level lifetime API such as Free.
//
// The caller gains exclusive ownership of the returned *Buffer until it is
// released back to the pool. Once a buffer has been returned (via Put or a
// higher-level mechanism), the caller MUST NOT read from or write to it.
func (p Pool) Get() *Buffer {
	buf := p.p.Get()
	buf.Reset()
	buf.pool = p
	return buf
}

// Put returns the given *Buffer to the underlying pool so that it can be
// reused by future Get calls.
//
// After Put returns, the caller MUST treat buf as no longer owned: it MUST NOT
// access, mutate, or rely on any of its fields or its internal byte slice,
// since the buffer may be handed out to another caller at any time.
//
// In typical usage, application code SHOULD prefer to use the buffer’s own
// release method (for example, Buffer.Free) if such a method is provided,
// and call Put only from within that implementation or other low-level
// infrastructure code that manages the Pool directly.
func (p Pool) Put(buf *Buffer) {
	p.p.Put(buf)
}
