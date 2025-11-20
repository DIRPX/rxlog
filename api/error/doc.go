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

// Package error defines abstractions for handling and encoding errors produced
// by the rxlog encoding pipeline.
//
// This package is intentionally small and focused. It provides:
//
//   - Handler: an interface that describes how an encoder should react when
//     encoding of an individual field fails, and
//
//   - Encoder: a function type for serializing error values into
//     buffer.Buffer instances used by higher-level encoders.
//
// The package is named error because it deals with error handling and error
// serialization within the logging subsystem. To avoid confusion with the
// built-in error type and the standard library's errors package, callers
// typically import this package under an alias, for example:
//
//	import (
//	    aerr "dirpx.dev/rxlog/api/error"
//	)
//
// # Handler semantics
//
// A Handler defines a policy for dealing with per-field encoding failures.
// When an encoder attempts to serialize a field and encounters an error, it
// MUST call the configured Handler rather than deciding on its own whether
// to abort or continue.
//
// Conceptually, HandleError has the following shape:
//
//	type Handler interface {
//	    HandleError(key string, value interface{}, err error) error
//	}
//
// The parameters have the following meaning:
//
//   - key is the logical field name the encoder was attempting to emit.
//
//   - value is the Go value associated with that field. Depending on where
//     the failure occurred, this may be the original application value, a
//     normalized representation, or nil if the failure happened before the
//     value was fully obtained or inspected.
//
//   - err is the non-nil error raised by the encoder or by underlying
//     marshaling logic.
//
// A Handler implementation MUST obey this contract:
//
//   - If HandleError returns a non-nil error, encoding of the current log
//     entry MUST stop immediately and that error MUST be propagated to the
//     caller of the encoding operation.
//
//   - If HandleError returns nil, the failing field MUST be treated as
//     skipped and encoding of the remaining fields MUST continue. The
//     original error is considered handled by the policy.
//
//   - Implementations MUST NOT mutate encoder-internal state in undocumented
//     ways and MUST be safe for concurrent use by multiple goroutines when
//     the associated encoder is used concurrently.
//
// Typical policies built on top of Handler include:
//
//   - treat any failure as fatal and abort encoding of the entry;
//
//   - silently skip failing fields while preserving the rest of the entry;
//
//   - log or record encoding failures to a separate sink while continuing
//     to encode remaining fields;
//
//   - substitute placeholder values for fields that cannot be encoded
//     successfully.
//
// Concrete policies can be implemented in separate packages or modules by
// providing types that satisfy the Handler interface.
//
// # Error encoding
//
// In addition to Handler, this package defines Encoder, a function type for
// serializing error values into buffers used by encoders:
//
//	type Encoder func(dst *buffer.Buffer, err error) *buffer.Buffer
//
// An error.Encoder is responsible only for converting an error value into
// bytes according to some format and appending those bytes to the provided
// buffer. It does not perform I/O and does not manage buffer lifetimes.
//
// Encoder implementations MUST obey the ownership and lifetime rules of the
// buffer package:
//
//   - The dst argument is owned by the caller for the duration of the call.
//     The Encoder MAY append directly to dst or MAY obtain and return a
//     different *buffer.Buffer if that is more efficient or necessary for
//     its internal implementation.
//
//   - The Encoder MUST return the *buffer.Buffer that now holds the encoded
//     representation. If a different buffer pointer is returned than the one
//     passed in, the original dst MUST remain in a valid state according to
//     its own contract, and the Encoder MUST NOT retain references to either
//     buffer beyond the duration of the call.
//
//   - The Encoder MUST NOT call Free on any buffer it receives or returns.
//     Lifetime management is always the caller’s responsibility.
//
// The err parameter MAY be nil, and implementations MUST handle this case
// explicitly. Common strategies include:
//
//   - appending nothing at all for a nil error;
//
//   - appending a well-defined placeholder such as "<nil>"; or
//
//   - using a structured representation that distinguishes between absent
//     and present errors.
//
// Whatever representation is chosen for both nil and non-nil errors SHOULD
// be documented and SHOULD remain stable over time so that downstream
// systems can reliably interpret, parse, or aggregate error information.
//
// # Usage context
//
// The error package sits at the boundary between low-level encoders and
// higher-level logging components:
//
//   - Encoders call a configured Handler whenever a field fails to encode,
//     delegating policy decisions about aborting or skipping fields.
//
//   - Encoders that include error-valued fields (for example, fields
//     representing failures, wrapped errors, or diagnostics) call a
//     configured error.Encoder to serialize those values into buffers.
//
// The package itself does not implement any particular policies or formats;
// it defines the contracts that such policies and formats must follow. This
// separation allows applications and reference implementations to provide
// their own Handler and Encoder implementations without changing the core
// logging API.
package error
