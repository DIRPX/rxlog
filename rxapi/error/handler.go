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

package handler

// Handler defines how encoding errors are handled during log entry encoding.
//
// When an encoder encounters an error (for example, a reflection failure,
// an unsupported type, or an I/O error), it consults the Handler to decide
// how to proceed. This allows applications to centralize policy around
// encoding failures, such as logging them, emitting metrics, or applying
// fallback behavior.
//
// A Handler MAY choose to stop encoding by returning a non-nil error, or it
// MAY decide to ignore the failing field and continue encoding the remaining
// data by returning nil. Implementations SHOULD be side-effect free with
// respect to the encoder’s internal state, and MUST be safe for concurrent
// use when the encoder is used concurrently (which is typical in logging
// libraries).
type Handler interface {
	// HandleError is invoked whenever encoding fails for a specific field.
	//
	// The key identifies the logical field name for which encoding failed.
	// The value carries the data that could not be encoded and MAY be nil if
	// the failure occurred before the value was fully obtained or inspected.
	// The err parameter is the error raised by the encoder or underlying
	// marshaling logic and MUST be non-nil.
	//
	// If HandleError returns a non-nil error, encoding MUST stop immediately
	// and that error MUST be propagated to the caller of the encoding
	// operation. If HandleError returns nil, the failing field MUST be treated
	// as skipped and encoding of the remaining fields MUST continue.
	//
	// Implementations MUST NOT mutate encoder-internal state in undocumented
	// ways and MUST be safe for concurrent use by multiple goroutines if the
	// associated encoder is used concurrently.
	HandleError(key string, value interface{}, err error) error
}
