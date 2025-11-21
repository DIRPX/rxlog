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

// Package pool provides a small, typed abstraction over sync.Pool for
// reusing heap-allocated values in high-throughput parts of the rxlog
// logging stack.
//
// The central type in this package is Pool[T], a generic wrapper around
// sync.Pool that offers:
//
//   - type-safe access to pooled values without manual type assertions,
//   - a clear ownership model for values obtained from and returned to the
//     pool, and
//   - a documented set of behavioral guarantees aligned with the semantics
//     of sync.Pool.
//
// # Design goals
//
// The pool package is intended to be used for amortizing allocations of
// short-lived, reusable objects such as buffers, encoders, or temporary
// scratch structures. It is not a general-purpose resource manager, nor
// is it suitable for representing scarce external resources (such as
// network connections or file descriptors).
//
// The main design goals are:
//
//   - Type safety: callers work with Pool[T] for a specific T and do not
//     need to perform type assertions when calling Get or Put.
//
//   - Explicit contracts: the lifetime and ownership rules for pooled
//     values are spelled out in one place and can be referenced by other
//     packages that build on top of Pool[T].
//
//   - Alignment with sync.Pool: the behavior of Pool[T] is defined in
//     terms of the standard library's sync.Pool, so existing knowledge
//     about sync.Pool carries over.
//
// # High-level semantics
//
// Conceptually, a Pool[T] owns a set of values of type T that may be
// handed out to callers on demand:
//
//   - Get returns a value of type T for exclusive use by the caller.
//
//   - Put returns a value of type T to the pool so that it MAY be reused
//     by future Get calls.
//
// A Pool[T] is safe for concurrent use by multiple goroutines. Multiple
// goroutines MAY call Get and Put on the same Pool[T] instance without
// external synchronization. However, this does not change the ownership
// rules for the values themselves:
//
//   - At any point in time, each value obtained from the pool MUST be
//     treated as owned by exactly one goroutine (or one higher-level
//     component) until it is returned via Put.
//
//   - After a value has been returned to the pool via Put, the former
//     owner MUST no longer use it: it MUST NOT read from, write to, or
//     otherwise rely on that value, since it may be handed out to another
//     caller at any time.
//
// # Factory functions and value initialization
//
// A Pool[T] is typically constructed with a factory function that knows
// how to create new values of type T in a valid initial state. When the
// underlying pool is empty or when the implementation chooses not to
// reuse an existing value, it calls this factory to supply a new value.
//
// The factory is responsible for:
//
//   - constructing values of type T that are safe for immediate use by
//     callers, and
//
//   - ensuring any invariants required by the broader application are
//     satisfied (for example, pre-allocating internal buffers or setting
//     default configuration).
//
// Callers MUST NOT assume that Get will always return a value that was
// previously passed to Put. As with sync.Pool, values may be dropped at
// any time (for example, under memory pressure or across GC cycles), and
// the pool is allowed to allocate new values as needed.
//
// # Ownership, reuse, and state reset
//
// The Pool[T] abstraction does not automatically reset or sanitize values.
// Instead, it relies on a simple convention:
//
//   - The code that calls Put is responsible for ensuring that the value
//     being returned is in a reusable state. For types that expose a Reset
//     or Clear method, callers SHOULD invoke that method before calling
//     Put.
//
//   - Callers SHOULD avoid returning values that contain references to
//     sensitive data (such as secrets or user content) unless they have
//     been appropriately cleared, to prevent accidental data leakage across
//     users of the pool.
//
//   - Callers MAY choose to discard a value instead of returning it to the
//     pool if they determine that it is not worth reusing (for example,
//     because it has grown too large). In that case Put SHOULD simply not
//     be called for that value.
//
// Because the pool MAY drop values at any time, callers MUST NOT rely on
// Put as a mechanism for enforcing upper bounds on memory usage or for
// keeping a fixed number of live objects. Pool[T] is a best-effort cache,
// not a strict resource pool.
//
// # Relationship to sync.Pool
//
// Internally, Pool[T] is implemented in terms of the standard library's
// sync.Pool. This implies that:
//
//   - The lifetime of pooled values is tied to the garbage collector and
//     may be affected by GC cycles.
//
//   - The pool may opportunistically discard values, so reuse is not
//     guaranteed even when Put is called.
//
//   - The concurrency guarantees of sync.Pool apply: it is safe for
//     multiple goroutines to call Get and Put concurrently.
//
// The pool package does not attempt to change these underlying semantics.
// It merely wraps sync.Pool to provide type safety and to centralize the
// documentation of ownership and reuse rules for values of type T.
//
// # Usage context
//
// In the broader rxlog codebase, Pool[T] is used by low-level packages
// that need efficient reuse of heap-allocated objects (for example,
// buffer pools). Most application code will interact only with those
// higher-level abstractions and will not need to manipulate Pool[T]
// directly.
//
// Advanced users MAY choose to construct and manage their own Pool[T]
// instances when they have specific allocation patterns that benefit from
// pooling. When doing so, they MUST adhere to the ownership, reuse, and
// state-reset rules described above to avoid subtle data races and
// cross-request contamination.
//
// In summary, the pool package provides a small, generic, and well-
// specified wrapper around sync.Pool. It enables safe, type-aware reuse
// of heap-allocated values in performance-sensitive code, while leaving
// policy decisions about when and what to pool in the hands of callers.
package pool
