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

// Package encoders provides reusable helpers for encoding common Go
// interfaces into object encoders built on top of buffer.Buffer.
//
// The functions in this package are small, focused adapters. They take
// values such as error or fmt.Stringer, call the appropriate method
// (Error or String), and encode the resulting string into a base.ObjectEncoder
// as a single field. Along the way they handle tricky cases:
//
//   - nil and typed nil values,
//   - panics raised by Error or String methods, and
//   - interaction with the underlying buffer.Buffer.
//
// The goal is to centralize this behavior so that callers do not have to
// re-implement the same panic handling and nil semantics at each call site.
//
// # Buffer and encoder interaction
//
// All helpers in this package follow the same signature pattern:
//
//	func EncodeX(
//	    dst *buffer.Buffer,
//	    key string,
//	    v   <interface type>,
//	    enc base.ObjectEncoder,
//	) (*buffer.Buffer, error)
//
// where:
//
//   - dst is the current destination buffer.Buffer that holds the encoded
//     representation built so far;
//
//   - key is the logical field name under which the value should be encoded;
//
//   - v is the value to encode (for example, an error or a value implementing
//     fmt.Stringer); and
//
//   - enc is a base.ObjectEncoder used to append the field to the buffer.
//
// Callers MUST treat the returned *buffer.Buffer as the new authoritative
// destination for subsequent writes related to the same encoding operation.
// It MAY differ from dst if the underlying storage had to grow. Helpers in
// this package:
//
//   - MUST NOT call Free on dst or on the returned buffer;
//
//   - MUST NOT retain references to enc, dst, the returned buffer, or any
//     slices derived from them beyond the duration of the call; and
//
//   - MUST be safe to use from a single goroutine at a time per buffer, in
//     accordance with the conventions of buffer.Buffer and base encoders.
//
// # Error handling and panics
//
// Encoding values via interface methods can trigger panics, for example:
//
//   - when the receiver is a typed nil pointer and the method is called on
//     it; or
//
//   - when the method itself panics due to internal logic errors.
//
// Helpers in this package adopt the following conventions:
//
//   - They wrap calls to Error or String in a deferred recover block.
//
//   - If a panic occurs and the underlying value is a nil pointer, they
//     treat this as a valid "<nil>" representation:
//
//   - they encode the literal "<nil>" as the field value, and
//
//   - they return a nil error from the helper.
//
//   - If a panic occurs for any other reason, they recover the panic and
//     return a non-nil error describing it. In this case they do NOT write
//     a value for the field; it is the caller's responsibility to react,
//     for example by emitting a companion "<Key>Error" field or routing the
//     error through an error.Handler policy.
//
//   - If no panic occurs, they follow the "happy path" behavior described
//     for each helper below.
//
// These conventions ensure that panics in Error or String do not crash the
// logging pipeline and that callers have a clear way to observe and react
// to such failures.
//
// # EncodeError
//
// EncodeError encodes an error value into an object encoder as a single
// string field:
//
//   - On the happy path, if err is non-nil, it calls err.Error() and encodes
//     the result via enc.AddString.
//
//   - If err is nil (including typed nil errors), it encodes the literal
//     "<nil>" as the field value.
//
//   - If calling Error or encoding triggers a panic and the underlying
//     error value is a nil pointer, EncodeError recovers the panic and
//     encodes "<nil>" as the field value, returning a nil error.
//
//   - If a panic occurs for any other reason, EncodeError recovers it and
//     returns a non-nil error describing the panic without writing a value
//     for the field.
//
// Callers that require a best-effort log of failing error encodings can
// combine EncodeError with an error handling policy (for example, by
// emitting a companion "<Key>Error" field when the helper returns a
// non-nil error).
//
// # EncodeStringer
//
// EncodeStringer encodes a value implementing fmt.Stringer as a single
// string field:
//
//   - The stringer parameter is expected to be either nil or to hold a
//     value that implements fmt.Stringer.
//
//   - On the happy path, if stringer is non-nil, EncodeStringer calls
//     String() and encodes the result via enc.AddString.
//
//   - If stringer is nil, it encodes the literal "<nil>" as the field
//     value without calling String().
//
//   - If calling String or encoding triggers a panic and the underlying
//     value is a nil pointer, EncodeStringer recovers the panic and encodes
//     "<nil>" as the field value, returning a nil error.
//
//   - If a panic occurs for any other reason, EncodeStringer recovers it
//     and returns a non-nil error describing the panic, without writing a
//     value for the field.
//
// As with EncodeError, callers are responsible for deciding how to react
// to the returned error value (for example, by recording diagnostics or
// by surfacing the failure via an additional field).
//
// # Usage context
//
// The encoders package is intended for use by encoder implementations and
// low-level logging infrastructure, not by most application code. Typical
// usage includes:
//
//   - object encoders that need to support error-valued or fmt.Stringer
//     fields while handling typed nils and panics uniformly;
//
//   - helper functions that build field.Field values from errors or
//     fmt.Stringer values and want to reuse this package's semantics; and
//
//   - diagnostic or test encoders that want consistent behavior around
//     "<nil>" representations and panic recovery.
//
// By centralizing this logic, rxlog keeps encoder implementations simpler
// and ensures that common corner cases around error and Stringer encoding
// are handled consistently across formats and backends.
package encoders
