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

import "io"

// Interface represents a reusable, pooled buffer that stores an encoded log
// entry or any other serialized payload. Implementations SHOULD be backed by
// a contiguous byte slice and SHOULD return the buffer to an internal pool
// when it is no longer needed.
//
// An Interface is NOT safe for concurrent use by multiple goroutines unless the
// implementation explicitly documents otherwise. Callers MUST treat each
// Interface instance as owned by a single goroutine at a time, and MUST respect
// the lifetime rules of Bytes, Reset, and Free described below.
type Interface interface {
	io.Writer

	// Bytes returns a view of the underlying encoded data as a byte slice.
	//
	// The returned slice is valid only until Free is called; after Free, its
	// contents, length, and capacity MUST be considered invalid and MUST NOT
	// be accessed. Callers MUST NOT retain the returned slice beyond the
	// lifetime of the Interface.
	//
	// The caller MUST treat the returned slice as immutable. Modifying it in
	// place MAY corrupt the buffer pool or violate assumptions made by other
	// components that reuse the same underlying storage.
	Bytes() []byte

	// Len reports the current length of the encoded data, in bytes.
	//
	// Len MUST be equivalent to len(Bytes()) at the time of the call, but
	// MUST NOT allocate or otherwise incur additional cost beyond reading the
	// current logical length. Implementations MAY compute Len more cheaply
	// than constructing or returning a fresh slice.
	Len() int

	// Cap reports the current capacity of the underlying buffer, in bytes.
	//
	// Cap SHOULD reflect the total number of bytes that can be written without
	// requiring a reallocation. Callers MAY use Cap to reason about potential
	// growth or to make decisions about preallocation, but MUST NOT rely on
	// any particular growth strategy or upper bound.
	Cap() int

	// Reset clears the logical contents of the buffer while preserving its
	// underlying storage so that it can be efficiently reused for building a
	// new payload.
	//
	// After Reset returns, Len MUST be 0. Implementations SHOULD preserve the
	// existing capacity (i.e., Cap SHOULD remain unchanged), but MAY shrink or
	// grow the underlying storage in exceptional cases (for example, to enforce
	// maximum size limits).
	//
	// Reset MUST NOT return the buffer to any pool and MUST NOT invalidate
	// existing ownership; for releasing a buffer, callers MUST use Free.
	Reset()

	// Free releases the buffer back to its originating pool so that its
	// storage can be reused.
	//
	// After Free returns, the Interface MUST NOT be used again: all methods,
	// including Bytes, Len, Cap, Reset, and Write, are invalid to call, and
	// any previously obtained slices from Bytes become unsafe to access.
	//
	// Each Interface instance MUST be freed at most once. Double-free or use
	// after Free MAY corrupt the pool, leak memory, or cause undefined
	// behavior in callers. Implementations SHOULD document whether Free may
	// perform additional checks (such as debug assertions) in development
	// builds.
	Free()
}
