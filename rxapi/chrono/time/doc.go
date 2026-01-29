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

// Package timeenc defines the time encoder abstraction used by rxlog to
// serialize time.Time values into byte buffers.
//
// This package is intentionally small and narrowly focused. It does not know
// where time values come from, how they are obtained, or which concrete
// formats a particular application will use. Its sole responsibility is to
// define how a timestamp that is already available as a time.Time should be
// converted into bytes.
//
// Higher-level configuration (for example, choosing specific layouts or
// predefined encoders) is provided by other files in this package and by the
// nested layout subpackage.
//
// # Overview
//
// The central concept in this package is:
//
//   - Encoder: a function that appends a serialized representation of a
//     time.Time value to a buffer.Buffer and returns the buffer that should
//     be used for subsequent writes.
//
// The Encoder type is intentionally minimal:
//
//	type Encoder func(dst *buffer.Buffer, t time.Time) *buffer.Buffer
//
// Implementations are responsible only for taking a time.Time value and
// appending some representation of that value to the provided buffer. They do
// not perform I/O, do not decide how timestamps are obtained, and do not manage
// buffer lifetimes.
//
// # Encoder semantics and buffer interaction
//
// An Encoder implementation MUST obey the following rules:
//
//   - The dst argument is owned by the caller for the duration of the call.
//     The Encoder MAY append directly to dst, or it MAY obtain and use a
//     different *buffer.Buffer value (for example, from a pool) if that is
//     more efficient for the chosen encoding strategy.
//
//   - The Encoder MUST return the *buffer.Buffer that now holds the encoded
//     timestamp. In the simple case this will be the same value as dst. If a
//     different buffer is returned, the original dst MUST remain in a valid
//     state according to the buffer package contract.
//
//   - The Encoder MUST NOT call Free on any buffer it receives or returns.
//     Buffer lifetime management is always the responsibility of the caller,
//     typically via buffer.Buffer.Free or the corresponding pool.
//
//   - The Encoder MUST NOT retain references to dst, to any other buffer it
//     creates or uses, or to slices derived from those buffers beyond the end
//     of the call. Doing so would violate the ownership model of the buffer
//     package and can lead to data races or corruption when buffers are
//     subsequently reused.
//
// Within these constraints, implementations are free to choose whatever
// internal strategy is most appropriate for the desired format and performance
// characteristics.
//
// # Formats and configuration
//
// This package does not prescribe any particular timestamp format. Different
// Encoder implementations may choose, for example:
//
//   - human-readable layouts (such as variants of RFC 3339 or ISO 8601),
//   - numeric epochs (such as Unix seconds, milliseconds, or nanoseconds),
//   - custom binary encodings,
//   - any other representation that can be appended to a buffer.Buffer.
//
// The choice of format is entirely up to the code that constructs and wires
// Encoders. This package only defines the function type and its behavioral
// contract. Common layouts and predefined encoders are provided by encoders.go
// in this package and by the layout subpackage.
//
// # Time zones and monotonic component
//
// Encoders operate on fully constructed time.Time values and do not decide
// how those values are created:
//
//   - The time zone or location associated with the timestamp (for example,
//     UTC versus local time) is determined by the code that constructs the
//     time.Time value prior to calling the Encoder.
//
//   - If an application needs to rely on the monotonic component of
//     time.Time for duration measurements, it SHOULD perform those
//     calculations before invoking an Encoder. Encoders typically serialize
//     only the wall-clock portion of time.Time, because timestamp formats
//     rarely preserve the monotonic component.
//
// # Usage context
//
// In the broader rxlog logging pipeline, Encoder values from this package are
// typically used by higher-level encoders that build complete log entries:
//
//   - Some component obtains a time.Time value by whatever means are
//     appropriate for the application.
//
//   - That component invokes a configured time encoder with the current buffer
//     and the time.Time value to append a timestamp field.
//
//   - The same buffer is then used to encode other fields and eventually
//     written to an output sink by other parts of the system.
//
// Because Encoder is just a pure function from (buffer, time.Time) to buffer,
// the same Encoder can be reused across multiple callers, cores, and outputs,
// as long as all of them adhere to the buffer ownership and lifetime rules.
//
// In summary, the encoder package provides a small but critical abstraction for
// timestamp serialization in rxlog. It defines how time.Time values are
// converted into bytes, while leaving the choice of concrete formats, layouts,
// and time sources to code outside this package.
package time
