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

package fields

const (
	// Error is the field key that carries the primary error value or message
	// associated with the log entry.
	//
	// This is typically the human-readable error string (for example,
	// err.Error()), although structured encoders MAY embed more detailed
	// information. Producers SHOULD use this key for the main error attached
	// to the event.
	Error = "error"

	// ErrorMessage is the field key that explicitly carries the human-readable
	// error message text, when a more structured error object is also present.
	ErrorMessage = "error_message"

	// ErrorType is the field key that records the logical type or class of
	// the error.
	//
	// Typical values include Go type names (for example, "*net.OpError",
	// "ValidationError") or application-level error codes. Producers MAY use
	// this field to enable more precise filtering and aggregation of error
	// categories than is possible with the free-form Error message alone.
	ErrorType = "error_type"

	// ErrorCode is the field key that records a stable, machine-readable
	// error code.
	//
	// This may be an application-defined code, an HTTP status code, a gRPC
	// code, or a domain-specific error identifier. Producers SHOULD prefer
	// this over parsing ErrorMessage when building alerts or metrics.
	ErrorCode = "error_code"

	// Stacktrace is the field key that carries the stack trace associated
	// with the log entry, when one is captured.
	//
	// The value is usually a multi-line string produced by a stack capturing
	// helper. Producers SHOULD reserve this field for higher-severity levels
	// (such as Error, Critical, or Fatal) to avoid excessive log volume.
	Stacktrace = "stacktrace"

	// ExceptionEscaped is the field key that indicates whether the error
	// or exception has escaped the intended handling scope (for example,
	// reached a global handler).
	ExceptionEscaped = "exception_escaped"
)
