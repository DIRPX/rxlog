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

package custom

import (
	"dirpx.dev/rxlog/api/buffer"
	"dirpx.dev/rxlog/api/encoder/base"
)

// TypeEncoder defines the function signature for user-provided value encoders
// that integrate with the TypeEncoder Registry.
//
// A TypeEncoder is responsible for encoding a single logical field identified
// by key and carrying the value val, appending its serialized representation
// to dst using the provided ObjectEncoder enc. It is typically registered
// for one or more Go types so that the registry can route matching values
// through this custom encoder instead of using a generic/reflection-based
// path.
//
// Parameters:
//
//   - dst is the destination buffer that already contains any previously
//     encoded data. Implementations SHOULD append to dst when possible to
//     minimize allocations, but MAY return a different *buffer.Buffer (for
//     example, one obtained from a pool) if they need to grow or replace the
//     underlying storage. The returned buffer MUST be treated as the new
//     authoritative destination by the caller.
//
//   - key is the field name under which val MUST be encoded. The TypeEncoder
//     MUST NOT change the key semantics; it is responsible for encoding the
//     value for exactly this key.
//
//   - val is the value to be encoded. It MAY be nil, and it MAY have any
//     dynamic type, though in practice registries usually dispatch only
//     values of specific, known types to a given TypeEncoder. Implementations
//     SHOULD document which types they accept and how nil is handled.
//
//   - enc is the ObjectEncoder used to actually write the field. The
//     TypeEncoder MUST treat enc as a single-use helper for the duration of
//     the call and MUST NOT retain it or use it after the function returns.
//     It SHOULD use enc’s Add* methods (AddString, AddInt, AddObject, etc.)
//     to encode key/val into the buffer.
//
// Return values:
//
//   - The returned *buffer.Buffer is the buffer that now holds the encoded
//     data. It MAY or MAY NOT be the same instance as dst. Callers MUST use
//     the returned buffer for all subsequent writes and MUST eventually
//     release it according to the Buffer contract (typically via Free).
//
//   - The error indicates whether encoding succeeded. On success, the error
//     MUST be nil. On failure, the error MUST be non-nil. If strict atomicity
//     is required (i.e., “all-or-nothing” encoding for this field), the
//     TypeEncoder SHOULD avoid leaving a partially encoded key/value pair in
//     the buffer when returning a non-nil error. Higher-level code MAY still
//     choose to log error details separately (for example, as a "<key>Error"
//     companion field).
//
// Implementations MUST NOT call Free on dst or on the returned buffer;
// lifetime management is always the caller’s responsibility. TypeEncoders MAY
// be invoked concurrently by the registry from multiple goroutines; any
// shared state they use MUST be safe for concurrent access.
type TypeEncoder func(dst *buffer.Buffer, key string, val interface{}, enc base.ObjectEncoder) (*buffer.Buffer, error)
