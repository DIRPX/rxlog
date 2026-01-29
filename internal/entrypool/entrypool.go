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

// Package entrypool exposes a process-wide pool of *core.Entry instances used
// by the logging pipeline.
//
// This package intentionally provides a very small façade: callers obtain
// entries via the exported Get function and MUST return them via Put once
// they are finished. The concrete pooling strategy (backed by the pool
// package) is treated as an implementation detail and may evolve over time
// without affecting callers.
//
// Entries obtained from this package are intended to be short-lived,
// per-log-operation objects. They MUST NOT be retained beyond the scope of
// a single logging operation and MUST NOT be shared between goroutines.
package entrypool

import (
	"dirpx.dev/rxlog/rxapi/core"
	"dirpx.dev/rxlog/rxapi/pool"
)

var (
	// _pool holds the process-wide default *core.Entry pool instance.
	//
	// The pool is initialized once at package load time and is not intended
	// to be accessed directly from outside this package. External users
	// should rely on the exported Get and Put functions, which delegate to
	// this underlying pool.
	//
	// The factory function passed to pool.New is responsible for creating
	// new *core.Entry instances when the pool is empty. It MUST return
	// fully initialized entries that are safe to use immediately after
	// Get, with no additional setup required by the caller.
	//
	// The underlying Pool implementation returned by pool.New MUST be safe
	// for concurrent use by multiple goroutines. Callers are still
	// responsible for respecting the ownership rules of *core.Entry
	// instances described below.
	_pool = pool.New(func() *core.Entry {
		return &core.Entry{}
	})
)

// Get retrieves an *core.Entry from the default pool.
//
// The returned entry is expected to be in a clean, reset state suitable
// for immediate use in the logging pipeline. Callers MUST NOT assume
// that the entry is zero-allocated; they should only rely on the
// invariants guaranteed by the core.Entry type (for example, fields
// that are reset by its Reset method or constructor logic).
//
// Concurrency and ownership rules:
//   - Get is safe to call from multiple goroutines.
//   - Each entry returned by Get MUST be treated as exclusively owned
//     by the caller until it is passed to Put.
//   - Callers MUST NOT use the same entry concurrently from multiple
//     goroutines.
//   - Callers MUST NOT access an entry after it has been returned
//     to the pool via Put.
func Get() *core.Entry {
	return _pool.Get()
}

// Put returns the provided *core.Entry to the default pool.
//
// After Put is called, the entry MUST be considered invalid by the
// caller: any further reads or writes are undefined behavior, as the
// same memory may be reused by a future Get call for a different log
// operation.
//
// Callers SHOULD ensure that Put is invoked exactly once for each entry
// obtained via Get, typically using a defer pattern:
//
//	e := entrypool.Get()
//	defer entrypool.Put(e)
//
// This guarantees that entries are consistently returned even when
// the surrounding operation exits early due to errors.
func Put(e *core.Entry) {
	_pool.Put(e)
}
