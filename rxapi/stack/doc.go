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

// Package stack defines abstractions for capturing and encoding goroutine
// stack traces for inclusion in log entries.
//
// In many logging scenarios, it is useful to attach a stack trace to selected
// events (for example, when reporting errors or panics). This package provides
// two focused abstractions to support that use case:
//
//   - Collector: an interface that knows how to capture a stack trace from the
//     current goroutine and format it as a single string.
//
//   - Encoder: a function type that serializes a previously captured stack
//     trace string into a buffer.Buffer used by higher-level log encoders.
//
// These abstractions separate the concerns of:
//
//   - discovering which stack frames to include and how to arrange them, and
//   - deciding how the resulting textual stack trace should be rendered into
//     the final log output.
//
// # Collector semantics
//
// A Collector is responsible for building stack trace strings. Conceptually,
// it provides a method of the form:
//
//	Stack(skip, depth int) string
//
// The skip and depth parameters control which part of the current goroutine's
// call stack is captured:
//
//   - skip selects how many frames at the top of the stack should be omitted.
//     Implementations generally follow runtime.Caller-style semantics, where
//     skip == 0 refers to the frame of the Stack method itself, skip == 1
//     refers to its caller, and so on. Callers typically choose a skip value
//     that hides internal logging wrappers and points at the user-facing call
//     site where the log event originated.
//
//   - depth limits how many frames are included in the resulting trace.
//     Implementations SHOULD interpret positive depth values as an upper
//     bound on the number of frames to capture. If depth is zero or negative,
//     implementations MAY treat this as "no explicit limit" and capture as
//     many frames as are reasonably available, subject to their own internal
//     safety limits.
//
// The Stack method returns a single string that represents the captured stack
// trace. That string is expected to be:
//
//   - multi-line and human-oriented (for example, one frame per line with
//     function name, file path, and line number); and
//
//   - stable in structure for a given Collector implementation, so that
//     downstream tools can parse, search, or display stack traces
//     consistently over time.
//
// If a Collector cannot obtain a stack trace for some reason, it MAY return an
// empty string or a sentinel representation. The specific behavior SHOULD be
// documented by the implementation. The interface itself does not distinguish
// between "no stack available" and "empty stack".
//
// Collector implementations MUST be safe for concurrent use by multiple
// goroutines unless explicitly documented otherwise. They MAY cache auxiliary
// data (for example, symbolization results), but MUST ensure that such
// caching does not introduce data races or leak information between callers.
//
// # Encoder semantics and buffer interaction
//
// An Encoder is responsible for taking a stack trace string (typically
// produced by a Collector) and appending a serialized representation of that
// string to a buffer.Buffer. The Encoder type is:
//
//	type Encoder func(dst *buffer.Buffer, stack string) *buffer.Buffer
//
// Encoder implementations MUST obey the ownership and lifetime rules of the
// buffer package:
//
//   - The dst argument is owned by the caller for the duration of the call.
//     The Encoder MAY append directly to dst or MAY obtain and return a
//     different *buffer.Buffer if that better suits its implementation.
//
//   - The Encoder MUST return the *buffer.Buffer that now holds the encoded
//     stack trace. If this differs from dst, the original dst MUST remain in
//     a valid state according to its own contract.
//
//   - The Encoder MUST NOT call Free on any buffer it receives or returns.
//     Buffer lifetime management is always the caller's responsibility.
//
//   - The Encoder MUST NOT retain references to dst, to any other buffer it
//     creates or uses, or to slices derived from those buffers beyond the end
//     of the call. Doing so would violate the buffer ownership model and can
//     lead to data races or corruption when buffers are reused.
//
// The input stack string is typically a preformatted, multi-line trace built
// by a Collector. Encoders MAY transform or normalize this string (for
// example, trimming whitespace, reformatting frame lines, or applying
// redaction), but SHOULD document any such transformations and SHOULD keep
// the resulting representation stable over time so that downstream consumers
// can reliably parse, index, or display stack traces.
//
// # Usage context
//
// The stack package is intentionally small and self-contained. It does not
// perform I/O and does not depend on other parts of the logging subsystem
// beyond the buffer package. Typical usage patterns include:
//
//   - capturing stack traces at the point where an error or panic is
//     detected, using a configured Collector;
//
//   - storing the resulting stack trace string on a log entry or error
//     wrapper type; and
//
//   - invoking a stack.Encoder during log encoding to render that stack trace
//     into the serialized log output.
//
// By centralizing stack capture and encoding contracts in this package,
// rxlog allows applications to plug in different collection and formatting
// strategies while keeping the rest of the logging pipeline agnostic to the
// details of how stack traces are produced and represented.
package stack
