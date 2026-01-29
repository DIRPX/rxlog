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

// Package checkedentrypool exposes a process-wide pool of *core.CheckedEntry
// instances used by the logging pipeline.
//
// CheckedEntry typically represents an Entry that has already passed level
// checks and any additional predicates, and is ready to be written via one
// or more cores. Pooling these objects helps reduce allocations on hot
// logging paths, especially when a large number of log sites are evaluated.
//
// This package intentionally provides a very small façade: callers obtain
// checked entries via the exported Get function and MUST return them via
// Put once they are finished. The concrete pooling strategy (backed by the
// pool package) is treated as an implementation detail and may evolve over
// time without affecting callers.
//
// CheckedEntry values obtained from this package are intended to be
// short-lived, per-log-operation objects. They MUST NOT be retained beyond
// the scope of a single logging operation and MUST NOT be shared between
// goroutines.
package checkedentrypool

import (
	"dirpx.dev/rxlog/rxapi/core"
	"dirpx.dev/rxlog/rxapi/pool"
)

var (
	// _pool holds the process-wide default *core.CheckedEntry pool instance.
	//
	// The pool is initialized once at package load time and is not intended
	// to be accessed directly from outside this package. External users
	// should rely on the exported Get and Put functions, which delegate to
	// this underlying pool.
	//
	// The factory function passed to pool.New is responsible for creating
	// new *core.CheckedEntry instances when the pool is empty. It MUST
	// return fully initialized entries that are safe to use immediately
	// after Get, with no additional setup required by the caller.
	//
	// The underlying Pool implementation returned by pool.New MUST be safe
	// for concurrent use by multiple goroutines. Callers are still
	// responsible for respecting the ownership rules of *core.CheckedEntry
	// instances described below.
	_pool = pool.New(func() *core.CheckedEntry {
		return &core.CheckedEntry{}
	})
)

// Get retrieves a *core.CheckedEntry from the default pool.
//
// The returned value is expected to be in a clean, reset state suitable
// for immediate use in the logging pipeline. Callers MUST NOT assume that
// any previous state (such as attached cores, context, or fields) is
// preserved; they should only rely on the invariants guaranteed by the
// core.CheckedEntry type and the reset behavior defined by its
// implementation.
//
// Concurrency and ownership rules:
//   - Get is safe to call from multiple goroutines.
//   - Each CheckedEntry returned by Get MUST be treated as exclusively
//     owned by the caller until it is passed to Put.
//   - Callers MUST NOT use the same CheckedEntry concurrently from
//     multiple goroutines.
//   - Callers MUST NOT access a CheckedEntry after it has been returned
//     to the pool via Put.
func Get() *core.CheckedEntry {
	return _pool.Get()
}

// Put returns the provided *core.CheckedEntry to the default pool.
//
// After Put is called, the CheckedEntry MUST be considered invalid by the
// caller: any further reads or writes are undefined behavior, as the same
// memory may be reused by a future Get call for a different log operation.
//
// The exact reset semantics (for example, whether the CheckedEntry is
// cleared by the caller before Put or by the pool implementation itself)
// are defined by the core.CheckedEntry and pool packages. Callers MUST NOT
// rely on any state being preserved after Put.
//
// Callers SHOULD ensure that Put is invoked exactly once for each
// CheckedEntry obtained via Get, typically using a defer pattern:
//
//	ce := checkedentrypool.Get()
//	defer checkedentrypool.Put(ce)
//
// This guarantees that entries are consistently returned even when the
// surrounding operation exits early due to errors.
func Put(e *core.CheckedEntry) {
	_pool.Put(e)
}
