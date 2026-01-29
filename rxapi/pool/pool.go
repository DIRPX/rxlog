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

package pool

import "sync"

// Pool is a thin, strongly-typed wrapper around [sync.Pool] that provides
// generic object pooling for values of type T.
//
// The primary goal of Pool is to reduce allocations and GC pressure by
// reusing objects across call sites, while avoiding the unsafe type casting
// that comes with using sync.Pool directly.
//
// IMPORTANT: Because this implementation stores values in an underlying
// sync.Pool without additional wrapping, T SHOULD be a pointer type in
// performance-sensitive code. Staticcheck rule SA6002 (see
// https://staticcheck.io/docs/checks/#SA6002) will NOT automatically flag
// misuse here, so it is the caller’s responsibility to ensure that:
//
//   - pooled values are safe to reuse,
//   - values that carry large internal buffers are typically pooled as
//     pointers (e.g. *bytes.Buffer, *MyStruct),
//   - values that should not escape or be shared MUST NOT be pooled.
type Pool[T any] struct {
	pool sync.Pool
}

// New constructs a new Pool for values of type T.
//
// The fn argument is used to create new instances of T when the underlying
// pool is empty. It MUST return a value that is ready for immediate use by
// callers (for example, a zeroed or reset instance).
//
// Callers SHOULD choose fn such that the returned value is in a valid,
// reusable state (for example, Reset called on buffers, zeroed fields on
// structs). If T is a pointer type, fn typically allocates the target:
//
//	New(func() *MyType { return &MyType{} })
func New[T any](fn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return fn()
			},
		},
	}
}

// Get retrieves a value of type T from the pool.
//
// If the pool currently holds an object, that object is returned; otherwise,
// Get calls the New function supplied at construction time to create a new
// instance. The returned value is owned exclusively by the caller until it
// is returned to the pool via Put.
//
// Callers MUST NOT assume any particular initial state beyond what fn
// guarantees. If T has internal mutable state, callers SHOULD reset it to a
// known-good state before reuse when necessary.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns x to the pool so that it can be reused by future Get calls.
//
// After Put returns, the caller MUST treat x as no longer owned: it MUST NOT
// read from, write to, or otherwise rely on x’s state, since it may be handed
// out to other callers via subsequent Get invocations.
//
// Values placed into the pool SHOULD be in a reusable state. For types with
// internal buffers or references, callers SHOULD clear or Reset them before
// calling Put, to avoid leaking sensitive data or unexpected state across
// users of the pool.
func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}
