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

// Package reflect defines an abstraction for encoding arbitrary Go values
// using reflection.
//
// This package is intended as a "last resort" encoding mechanism for values
// whose concrete types are not known at compile time or for which no
// type-specific encoding is available. It complements, rather than replaces,
// the more efficient, strongly typed field and encoder APIs used elsewhere
// in the logging stack.
//
// The package is named reflect because it conceptually parallels the
// standard library's reflect package, but it is unrelated to that package
// at the code level. Applications MAY choose to import this package under
// an alias (for example, rxreflect) to avoid confusion with reflect from
// the standard library.
//
// # Overview
//
// The central concept in this package is:
//
//   - Encoder: a function that serializes an arbitrary Go value into a
//     buffer.Buffer, typically by inspecting its dynamic type via the
//     reflect package and applying a generic encoding strategy.
//
// The Encoder type is:
//
//	type Encoder func(dst *buffer.Buffer, v interface{}) (*buffer.Buffer, error)
//
// Implementations are responsible for inspecting v's dynamic type and
// appending a serialized representation of that value to the provided
// buffer. They do not perform I/O and do not manage buffer lifetimes.
// Instead, they focus solely on reflective inspection and encoding.
//
// # Encoder semantics and buffer interaction
//
// An Encoder implementation MUST obey the following rules:
//
//   - The dst argument is owned by the caller for the duration of the call.
//     The Encoder MAY append directly to dst or MAY obtain and use a
//     different *buffer.Buffer value (for example, from a pool) if that is
//     more efficient or necessary for its internal logic.
//
//   - The Encoder MUST return the *buffer.Buffer that now holds the encoded
//     representation. In the simple case, this will be the same value as dst.
//     If a different buffer is returned, the original dst MUST remain in a
//     valid state according to the buffer package contract.
//
//   - The Encoder MUST NOT call Free on any buffer it receives or returns.
//     Buffer lifetime management is always the responsibility of the caller.
//
//   - The Encoder MUST NOT retain references to dst, to any other buffer it
//     creates or uses, or to slices derived from those buffers beyond the end
//     of the call. Doing so would violate the ownership model of the buffer
//     package and may lead to data races or corruption when buffers are
//     reused.
//
// # Error handling and partial encodings
//
// Encoding arbitrary values via reflection can fail for various reasons,
// including unsupported types, cycles in the object graph, or internal
// encoder limitations. To make such failures visible, the Encoder type
// returns an error alongside the buffer.
//
// On error:
//
//   - Implementations MAY have appended a partially encoded representation
//     of v to the returned buffer. Callers MUST NOT assume that the buffer
//     is empty or unchanged when an error is non-nil.
//
//   - Callers that require all-or-nothing semantics (for example, a single
//     field that must either be fully encoded or omitted entirely) MUST
//     enforce atomicity at a higher level. A common pattern is to encode
//     into a temporary buffer and discard it if the Encoder returns an
//     error, copying or merging the result only when encoding succeeds.
//
//   - Encoders SHOULD return errors that allow callers to distinguish
//     between programmer errors (such as unsupported types) and transient
//     failures, where applicable.
//
// # Type mapping and stability
//
// A reflection-based encoder must decide how to map Go types to the chosen
// serialized representation. While this package does not prescribe a specific
// mapping, implementations SHOULD:
//
//   - document how major Go kinds (structs, maps, slices, arrays, pointers,
//     interfaces, basic scalar types, and nil) are represented;
//
//   - specify how exported versus unexported struct fields are treated;
//
//   - describe how cycles and shared references are handled (for example,
//     by rejecting cyclic structures or by using depth limits); and
//
//   - keep the mapping stable over time, so that downstream systems can rely
//     on the structure of the encoded output for parsing, indexing, and
//     aggregation.
//
// Applications that rely on the precise structure of reflect-encoded values
// SHOULD depend on the documented behavior of a specific Encoder
// implementation, rather than assuming any particular encoding strategy from
// the interface alone.
//
// # Performance characteristics and usage guidelines
//
// Reflection-based encoding is inherently more expensive than encoding values
// whose type is known at compile time. It typically involves:
//
//   - inspecting v's dynamic type using the reflect package;
//
//   - performing type switches or lookup tables for kinds; and
//
//   - traversing composite values (structs, maps, slices, etc.) recursively.
//
// As a result, reflect.Encoder implementations are expected to be:
//
//   - slower than type-specific encoders that operate on concrete types,
//   - more allocation-heavy, due to intermediate values and metadata, and
//   - more complex in behavior, especially when handling arbitrary user types.
//
// Callers SHOULD therefore reserve reflection-based encoding for cases where:
//
//   - the concrete type of v is not known at compile time;
//
//   - a generic "catch-all" encoder is needed as a fallback; or
//
//   - the flexibility of reflection outweighs the performance cost for the
//     specific use case.
//
// In hot paths or high-throughput logging scenarios, type-specific field and
// encoder APIs SHOULD be preferred wherever possible.
//
// # Usage context
//
// The reflect package is intentionally small and self-contained. It does not
// perform I/O and does not depend on other parts of the logging subsystem
// beyond the buffer package. Typical usage patterns include:
//
//   - higher-level encoders that, when encountering a value of unknown or
//     unsupported type, delegate to a reflect.Encoder as a fallback;
//
//   - application code that needs to log arbitrary user-provided values
//     without defining explicit field constructors for each type;
//
//   - debugging or diagnostic tools that prioritize breadth of coverage over
//     optimal performance.
//
// In summary, the reflect package provides a focused abstraction for
// reflection-based encoding of arbitrary Go values. It defines the function
// type, error semantics, and buffer interaction rules, while leaving the
// choice of concrete encoding strategy and type mapping to implementations.
package reflect
