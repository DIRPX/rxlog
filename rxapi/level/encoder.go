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

package level

import (
	"dirpx.dev/rxlog/rxapi/buffer"
)

// Encoder encodes a logging Level into the provided Buffer and returns the
// (possibly updated) Buffer pointer.
//
// An Encoder receives a destination buffer dst and a Level value, and appends
// an encoded representation of that level to dst. Implementations SHOULD
// write into the existing buffer when possible to avoid unnecessary
// allocations, but MAY obtain and return a different *buffer.Buffer (for
// example, from an internal pool) if they need to grow or replace the
// underlying storage.
//
// The pointer returned from Encoder MUST be treated as the authoritative
// destination after the call. Callers MUST:
//
//   - use the returned *buffer.Buffer for any subsequent writes, and
//   - eventually release that buffer according to the Buffer contract
//     (typically via Free).
//
// Implementations MUST NOT call Free on either dst or the returned buffer;
// lifetime management is always the caller’s responsibility. If a different
// buffer is returned than the one passed in, the original dst MUST remain in
// a valid state according to its own contract, and the Encoder MUST NOT
// retain references to either buffer beyond the duration of the call.
//
// Implementations SHOULD document the chosen level encoding format (for
// example, symbolic names such as "TRACE", "DEBUG", "INFO", "WARN", "ERROR",
// "FATAL", or numeric codes) and SHOULD keep this mapping stable over time so
// that downstream systems can reliably interpret, filter, and aggregate log
// entries by severity.
type Encoder func(dst *buffer.Buffer, level Level) *buffer.Buffer
