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

package encoder

import (
	"dirpx.dev/rxlog/api/buffer"
	"dirpx.dev/rxlog/api/core"
	"dirpx.dev/rxlog/api/encoder/base"
	"dirpx.dev/rxlog/api/field"
)

// Encoder is a format-agnostic interface for encoding complete log entries
// into a serialized representation (for example, JSON, text, or a binary
// protocol).
//
// Encoder extends ObjectEncoder, so it can accumulate structured fields on
// the encoder instance before producing a final encoded entry. Typical usage
// patterns build up static or per-logger context on the Encoder via
// ObjectEncoder methods, and then call Encode for each log event.
//
// Implementations of ObjectEncoder methods (such as AddString, AddInt, etc.)
// MAY freely mutate the receiver and are generally NOT safe for concurrent
// use by multiple goroutines. Callers MUST treat an Encoder instance as
// single-owner while they are building up context on it (for example, during
// logger construction or configuration).
//
// By contrast, Encode MUST be safe to call concurrently on the same
// Encoder value from multiple goroutines. Encode MUST NOT mutate
// any shared internal state of the Encoder in a way that affects other
// goroutines, except where such mutation is explicitly documented as
// concurrency-safe (for example, shared immutable caches). Implementations
// MAY internally clone or snapshot state on each Encode call to satisfy
// this requirement.
type Encoder interface {
	base.ObjectEncoder

	// Encode serializes the provided Entry and fields, together with any
	// context already accumulated on the Encoder, into the supplied destination
	// Buffer and returns the (possibly different) Buffer that holds the result.
	//
	// The dst argument is an existing buffer that the encoder MAY append to and
	// return, or it MAY be ignored in favor of another *buffer.Buffer (for
	// example, one obtained from an internal pool). Callers MUST treat the
	// returned buffer as the authoritative destination after the call and MUST
	// NOT assume that dst and the returned buffer are the same instance.
	//
	// Implementations MUST:
	//
	//   - include all relevant context already held by the Encoder (for example,
	//     fields added via ObjectEncoder methods prior to the call);
	//   - include the supplied entry and fields arguments in the encoded output;
	//   - apply their documented rules for omitting "empty" fields (for example,
	//     zero-valued timestamps or empty strings, if that is the chosen policy);
	//   - return a *buffer.Buffer whose contents remain valid until the caller
	//     releases it according to the Buffer contract (typically via Free).
	//
	// Encode MUST be safe to call concurrently with other Encode invocations on
	// the same Encoder instance. To satisfy this requirement, implementations
	// MUST NOT mutate shared state in ways observable across goroutines and MAY
	// internally clone or snapshot per-entry state as needed.
	//
	// Implementations MUST NOT call Free on dst or on the returned buffer; buffer
	// lifetime management is the caller’s responsibility.
	//
	// On failure, Encode MUST return a non-nil error and MAY return a nil buffer.
	// If a non-nil buffer is returned alongside an error, the caller is still
	// responsible for eventually releasing that buffer if the underlying Buffer
	// type requires explicit Free calls.
	Encode(dst *buffer.Buffer, entry core.Entry, fields []field.Field) (*buffer.Buffer, error)
}
