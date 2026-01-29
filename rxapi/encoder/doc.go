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

// Package encoder defines the core encoding abstractions used by rxlog to
// turn log entries into serialized byte streams.
//
// At this layer, logging is expressed in terms of three closely related
// concepts:
//
//   - Encoder: an interface that takes a core.Entry and a set of structured
//     fields and produces a serialized representation in a buffer.Buffer.
//
//   - ObjectEncoder: an interface for encoding key–value pairs into a
//     structured object representation.
//
//   - ArrayEncoder: an interface for encoding ordered sequences of values
//     into a structured array representation.
//
// The encoder package itself is format-agnostic: it does not mandate JSON,
// text, or any other specific on-the-wire format. Instead, it defines the
// contracts that concrete encoders must satisfy so that the rest of the
// logging pipeline (cores, hooks, writers) can treat them uniformly.
//
// # Entry encoders
//
// The top-level Encoder interface is responsible for serializing a single
// log event. Conceptually, it has the following shape:
//
//	type Encoder interface {
//	    // Clone and other methods omitted.
//	    Encode(dst *buffer.Buffer, entry core.Entry, fields []field.Field) (*buffer.Buffer, error)
//	}
//
// Implementations of Encoder MUST obey the following rules:
//
//   - The dst argument is a buffer.Buffer that is owned by the caller for
//     the duration of the call. The Encoder MAY append directly to dst or
//     MAY obtain and return a different *buffer.Buffer if that is more
//     efficient or necessary for its internal strategy (for example, when
//     switching to a larger backing array).
//
//   - The Encoder MUST return the *buffer.Buffer that now holds the encoded
//     representation of the entry. Callers MUST treat the returned buffer as
//     the authoritative destination for any subsequent writes related to
//     this encoding step and MUST release it according to the buffer
//     package’s contract (typically via Free) once it is no longer needed.
//
//   - The Encoder MUST NOT call Free on dst or on the returned buffer.
//     Buffer lifetime management is always the responsibility of the code
//     that owns the buffer (for example, a core that manages a buffer pool).
//
//   - The Encoder MUST NOT retain references to dst, to any other buffer it
//     creates or uses, or to slices derived from those buffers beyond the
//     end of the Encode call. Doing so would violate the ownership model of
//     the buffer package and can lead to data races or corruption when
//     buffers are reused.
//
//   - Encode MUST be safe for concurrent use by multiple goroutines. In
//     particular, it MUST be valid to call Encode on the same Encoder
//     instance from different goroutines at the same time. To satisfy this
//     requirement, implementations MUST NOT mutate shared state in ways that
//     are observable across goroutines and MAY internally clone or snapshot
//     per-entry state as needed.
//
// On failure, Encode MUST return a non-nil error and MAY return a nil
// buffer. If a non-nil buffer is returned alongside an error, the caller is
// still responsible for eventually releasing that buffer if the underlying
// Buffer type requires explicit Free calls.
//
// # Object encoders
//
// ObjectEncoder describes the operations required to encode a structured
// object as a collection of named fields. It is the primary target for
// field.Field.AddTo and is used by higher-level encoders to build up the
// representation of a log entry.
//
// While the exact method set is defined in this package, it typically
// includes:
//
//   - AddX(key string, value T) error methods for scalar types (booleans,
//     numbers, strings, binary data, time, duration, etc.),
//
//   - AddArray / AddObject methods that accept array and object marshaler
//     interfaces for nested structures, and
//
//   - AddReflected or similar methods that delegate to reflection-based
//     encoders for "any" values when no more specific representation is
//     available.
//
// Implementations of ObjectEncoder MUST follow these rules:
//
//   - Methods that append to the underlying buffer MUST respect the same
//     ownership and lifetime guarantees as Encoder.Encode: they MUST NOT
//     call Free and MUST NOT retain references to buffers or derived slices
//     beyond the lifetime of the encoder instance.
//
//   - Methods MAY return errors when encoding fails (for example, because a
//     nested marshaler returns an error or because a value cannot be
//     represented in the chosen format). Such errors are typically routed
//     through an error.Handler configured by the calling encoder.
//
//   - ObjectEncoder instances are not safe for concurrent use by multiple
//     goroutines. Each instance MUST be used by at most one goroutine at a
//     time and MUST be reset or discarded before being reused for a new
//     entry.
//
// ObjectEncoder is format-neutral: it does not prescribe how keys or values
// are rendered (for example, as JSON strings or text fragments). Those
// details are defined by concrete implementations that embed or wrap base
// encoders from subpackages such as encoder/base.
//
// # Array encoders
//
// ArrayEncoder describes the operations required to encode a structured
// array as an ordered sequence of elements. It is typically used to encode
// array-valued fields and nested arrays inside objects.
//
// As with ObjectEncoder, the exact method set is defined in this package
// and typically includes:
//
//   - AppendX(value T) error methods for scalar types,
//
//   - AppendArray / AppendObject methods for nested arrays and objects, and
//
//   - AppendReflected or similar methods for reflection-based encoding of
//     arbitrary values.
//
// ArrayEncoder implementations MUST obey the same buffer ownership and
// concurrency rules as ObjectEncoder:
//
//   - They operate on a buffer owned by their caller and MUST NOT call Free
//     on it.
//
//   - They MUST NOT retain references to buffers or derived slices beyond
//     the lifetime of the encoding operation.
//
//   - They are not safe for concurrent use; each instance MUST be confined
//     to a single goroutine at a time.
//
// # Relationship to subpackages
//
// The encoder package defines only interfaces and high-level contracts. It
// does not provide concrete encoders itself. Instead:
//
//   - encoder/base provides reusable building blocks for implementing
//     ObjectEncoder and ArrayEncoder on top of buffer.Buffer, including
//     common state management and field routing helpers.
//
//   - encoder/encoders provides small helpers for encoding common
//     interfaces (such as error and fmt.Stringer) into object encoders,
//     handling tricky cases like typed nil values and panics.
//
//   - encoder/custom provides mechanisms for registering and resolving
//     application-specific encoders under stable identifiers.
//
// These subpackages are optional building blocks; applications are free to
// implement encoders directly against the interfaces defined here if they
// prefer.
//
// # Usage in the logging pipeline
//
// In the broader rxlog pipeline:
//
//   - A core constructs a core.Entry and a slice of field.Field values for
//     each log event.
//
//   - It obtains a buffer.Buffer from a pool and calls Encoder.Encode with
//     the entry and fields. Encode uses an ObjectEncoder implementation to
//     materialize fields, potentially delegating to ArrayEncoder instances
//     for nested arrays.
//
//   - Once encoding is complete, the core hands the populated buffer to a
//     writer.WriteSyncer for delivery to an output sink and eventually
//     returns the buffer to its pool.
//
// By centralizing encoding contracts in this package and reusing them across
// all formats and backends, rxlog ensures that:
//
//   - structured fields can be encoded without reflection,
//   - buffer ownership and lifetime rules are explicit, and
//   - concrete encoders can evolve independently of higher-level logging
//     logic.
package encoder
