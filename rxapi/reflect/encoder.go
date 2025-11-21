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

package reflect

import (
	"dirpx.dev/rxlog/rxapi/buffer"
)

// Encoder encodes an arbitrary Go value into the provided Buffer using
// reflection-based semantics and reports success or failure via an error.
//
// An Encoder receives a destination buffer dst and an arbitrary value v, and
// appends an encoded representation of v to dst. Implementations SHOULD write
// into the existing buffer when possible to minimize allocations, but MAY
// obtain and return a different *buffer.Buffer (for example, from an internal
// pool) if they need to grow or replace the underlying storage.
//
// The pointer returned from Encoder MUST be treated as the authoritative
// destination after the call. Callers MUST:
//
//   - use the returned *buffer.Buffer for any subsequent writes related to
//     this encoding step, and
//   - eventually release that buffer according to the Buffer contract
//     (typically via Free), regardless of whether the error result is nil
//     or non-nil, as long as the buffer itself is non-nil.
//
// Implementations MUST NOT call Free on either the incoming dst or the
// returned buffer; lifetime management is always the caller’s responsibility.
// If a different buffer pointer is returned than the one passed in, the
// original dst MUST remain in a valid state according to its own contract,
// and the Encoder MUST NOT retain references to either buffer beyond the
// duration of the call.
//
// Because this Encoder operates on interface{} values, implementations will
// typically rely on reflection and therefore MAY be significantly slower and
// more allocation-heavy than type-specific encoders. Callers SHOULD reserve
// this Encoder for cases where the concrete type of v is not known at
// compile time or where generic, reflection-based handling is required.
//
// On error, implementations MAY have appended a partially encoded
// representation of v to the returned buffer. Callers that require strict
// atomicity MUST enforce it at a higher level (for example, by encoding into
// a temporary buffer and discarding it when an error is returned).
//
// Implementations SHOULD document how specific Go types are mapped to the
// encoded representation (for example, how structs, maps, slices, pointers,
// interfaces, and nil values are handled) and SHOULD keep this mapping stable
// over time so that downstream systems can reliably interpret the output.
type Encoder func(dst *buffer.Buffer, v interface{}) (*buffer.Buffer, error)
