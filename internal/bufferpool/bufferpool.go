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

// Package bufferpool exposes a process-wide buffer pool façade used by the
// logging subsystem.
//
// This package intentionally provides a very small surface area: callers are
// expected to obtain buffers through the exported Get function and to return
// them using the lifecycle mechanism defined by the buffer package (for
// example, via a dedicated Release/Free method on Buffer or a higher-level
// encoder/core API).
//
// The concrete pooling strategy (backed by buffer.NewPool) is treated as an
// implementation detail and may evolve over time without affecting callers.
package bufferpool

import "dirpx.dev/rxlog/rxapi/buffer"

var (
	// _pool holds the process-wide default buffer pool instance.
	//
	// This value is initialized once at package load time and is not intended
	// to be accessed directly from outside this package. All external users
	// should rely on the exported Get function, which delegates to this pool.
	//
	// The underlying implementation returned by buffer.NewPool MUST be safe
	// for concurrent use by multiple goroutines. Callers of this package are
	// still responsible for respecting Buffer ownership rules: a Buffer
	// obtained from the pool MUST NOT be used concurrently from multiple
	// goroutines and MUST NOT be retained after it has been released /
	// returned according to the buffer package contract.
	_pool = buffer.NewPool()
)

// Get retrieves a Buffer instance from the default pool.
//
// The returned Buffer is expected to be in a reset/empty state and ready
// for immediate use by the caller. The exact lifecycle of the Buffer
// (how it is returned to the pool) is governed by the buffer package:
// callers MUST follow the buffer.Buffer contract (for example, by calling
// a dedicated Release/Free method or by passing the Buffer back to an API
// that guarantees returning it to the pool).
//
// Concurrency rules:
//   - Get itself is safe to call from multiple goroutines.
//   - Each Buffer returned by Get MUST be used by at most one goroutine
//     at a time.
//   - Callers MUST NOT access a Buffer after it has been returned to
//     the pool, as future operations may reuse the same memory for a
//     different logical Buffer instance.
func Get() *buffer.Buffer {
	return _pool.Get()
}
