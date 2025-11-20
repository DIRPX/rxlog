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

// Package buffer defines the low-level byte buffer abstraction used throughout
// the rxlog logging pipeline.
//
// The package exposes two closely related concepts:
//
//   - Interface describes the behavioral contract that all buffers used by
//     rxlog MUST satisfy. It embeds io.Writer and adds methods for inspecting
//     the underlying byte slice (Bytes, Len, Cap), resetting it (Reset),
//     and returning the buffer to its originating pool (Free).
//
//   - Buffer is the default concrete implementation of Interface. It is a
//     thin wrapper around a growable []byte that remembers the Pool that
//     created it, so that it can be safely recycled when no longer needed.
//
// A third type, Pool, provides pooled allocation and reuse of *Buffer values.
// It is a small wrapper around the generic pool.Pool[T] abstraction defined
// in dirpx.dev/rxlog/api/pool and exists so that the rest of the logging
// stack does not need to depend on generics directly.
//
// # Overview and intended usage
//
// Buffers in this package are designed as scratch space for encoders and other
// low-level components that need to construct serialized payloads, such as
// encoded log entries. A typical usage pattern is:
//
//  1. A component that needs temporary storage (for example, a log encoder)
//     acquires a *Buffer from a Pool via Get.
//
//  2. The component appends data using Write, WriteString, WriteByte and the
//     various Append* helpers (AppendString, AppendInt64, AppendTime, and so
//     on) until the encoded representation is complete.
//
//  3. The finished bytes are obtained via Bytes and written to a downstream
//     sink (for example, a writer.WriteSyncer implementation).
//
//  4. Once the encoded data has been consumed, the component releases the
//     buffer by calling Free (or, in lower-level code, returning it to the
//     Pool via Put). After this point the buffer MUST NOT be used again.
//
// This pattern enables rxlog to avoid repeated allocations for every log
// entry. Instead, a small number of buffers are recycled throughout the
// process lifetime, significantly reducing GC pressure in high-throughput
// logging scenarios.
//
// # Lifetime, ownership, and borrowed slices
//
// The core design of this package assumes an explicit ownership model:
//
//   - Ownership of a buffer flows from a Pool to a single caller and back.
//     When Pool.Get returns a *Buffer, the caller becomes its owner. When the
//     caller invokes Free or returns the buffer to the Pool via Put, ownership
//     is relinquished.
//
//   - While a caller owns a buffer, it MAY freely mutate its contents using
//     the provided methods, subject to the concurrency restrictions described
//     below.
//
//   - Once a buffer has been released (via Free or Put), the former owner MUST
//     treat it as invalid. It MUST NOT read from or write to the buffer, nor
//     access any slices previously obtained from Bytes. Doing so constitutes
//     a use-after-free bug and MAY lead to data races or corrupted output.
//
// The Bytes method returns a slice that directly aliases the buffer's internal
// storage. This slice is "borrowed": it is valid only until any of the
// following events occur:
//
//   - The buffer is modified in a way that may grow or reallocate its
//     underlying slice (for example, a subsequent Write or Append* call that
//     exceeds the current capacity).
//
//   - The buffer is Reset.
//
//   - The buffer is released via Free (or returned to its Pool via Put).
//
// Callers MUST NOT retain the slice returned by Bytes beyond the lifetime of
// the buffer and MUST treat it as immutable while it is in use. Modifying the
// slice in place MAY corrupt the buffer pool or violate assumptions made by
// other components that reuse the same underlying storage.
//
// # Concurrency model
//
// Buffers provided by this package are intentionally not safe for concurrent
// use by multiple goroutines. Unless an implementation explicitly documents
// additional guarantees, the following rules apply:
//
//   - Each Interface or *Buffer instance MUST be treated as owned by exactly
//     one goroutine at a time.
//
//   - If a buffer needs to be shared between goroutines, the caller MUST
//     provide external synchronization (for example, a mutex) around all
//     method calls and all uses of slices returned by Bytes.
//
//   - It is never safe to use a buffer concurrently with Free or Put. A
//     buffer MUST NOT be accessed from any goroutine after it has been
//     released back to its Pool.
//
// These rules are critical for correctness because Pool implementations are
// allowed to hand out the same *Buffer instance to multiple callers over time.
// Violating them MAY manifest as subtle data races, corrupted encoded output,
// or panics in unrelated parts of the logging pipeline.
//
// # Relationship between Interface, Buffer, and Pool
//
// Interface defines the minimal set of operations that encoders and other
// subsystems rely on. Code that does not care about the concrete buffer
// implementation SHOULD depend only on Interface, so that alternative buffer
// types (for example, debug-only or metrics-instrumented buffers) can be
// substituted without changing callers.
//
// Buffer is rxlog's default implementation of Interface. It is optimized for
// the common case where callers:
//
//   - append data in a mostly linear fashion,
//   - occasionally reset the buffer for reuse, and
//   - return it to a pool once the encoded representation has been consumed.
//
// Internally, Buffer maintains:
//
//   - a byte slice that holds the current contents, and
//   - a reference to the Pool that created it, so that Free can return the
//     buffer to the correct pool.
//
// Pool is the mechanism by which *Buffer instances are allocated and reused.
// NewPool constructs a Pool that creates freshly allocated Buffer values with a
// default initial capacity defined by the Size constant. Get acquires a buffer
// from the pool and resets it to an empty state while preserving any reusable
// capacity. Put releases a buffer back to the pool so that it can be handed
// out again by subsequent Get calls.
//
// Most application code SHOULD rely on higher-level components (such as
// encoders or cores) to manage buffer pools and lifetimes. Direct interaction
// with Pool is typically reserved for infrastructure code within rxlog itself
// or for advanced users who need fine-grained control over allocation
// behavior.
//
// # Compatibility and non-goals
//
// The Buffer type intentionally mirrors some aspects of standard library types
// such as bytes.Buffer and bufio.Writer:
//
//   - It implements io.Writer and io.ByteWriter.
//   - Its Write and WriteString methods always report a full write on success
//     and return nil errors.
//
// However, Buffer is not a drop-in replacement for bytes.Buffer and does not
// attempt to match its full API surface. In particular:
//
//   - Buffer instances are expected to be short-lived and frequently recycled;
//     they are not designed to be long-lived in-memory accumulators.
//
//   - Callers MUST respect the explicit lifetime rules around Bytes, Reset and
//     Free, which are stricter than those for bytes.Buffer.
//
//   - The semantics of Size and Pool are tuned for the logging use case and
//     MAY change across rxlog versions, provided that the documented behavior
//     of Interface and Buffer is preserved.
//
// In summary, the buffer package provides the foundational building blocks for
// efficient encoding in rxlog: a well-specified buffer Interface, a default
// Buffer implementation, and a Pool for safe reuse. Correct use of these
// abstractions is essential for achieving both high throughput and predictable
// memory behavior in applications that emit large volumes of logs.
package buffer
