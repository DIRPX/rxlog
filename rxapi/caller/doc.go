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

// Package caller defines the abstractions used by rxlog to represent,
// capture, and encode call-site information (file, line, function) for
// log entries.
//
// # High-level concepts
//
// The package revolves around three closely related concepts:
//
//   - Caller: a value type that identifies a specific location in the
//     program source (file, line, and optionally function). It also
//     carries a Defined flag that indicates whether the call-site was
//     successfully resolved.
//
//   - Resolver: an interface that knows how to obtain Caller values from
//     the current execution context, typically by inspecting the call
//     stack. It is conceptually similar to the standard library's
//     runtime.Caller function, but exposed as an interface so that
//     alternative implementations (for example, test doubles or
//     performance-tuned resolvers) can be injected.
//
//   - Encoder: a function type that encodes a Caller into a textual
//     representation, appending it to a buffer.Buffer used by the logging
//     encoder. Different encoders can present the same call-site data in
//     different formats (short/long paths, with or without function
//     names, etc.).
//
// These abstractions are deliberately decoupled:
//
//   - Capturing a call-site (Resolver) is separated from formatting it
//     (Encoder), so that the same captured Caller can be encoded in
//     multiple ways or omitted entirely, depending on configuration.
//
//   - The Caller value itself is a small, copyable struct that can be
//     stored on a core.Entry and reused later when encoding or routing
//     the log entry.
//
// # Caller semantics
//
// The Caller type represents a best-effort description of a call-site
// and typically includes the following fields:
//
//   - Defined: a boolean indicating whether the resolver was able to
//     obtain a meaningful call-site.
//
//   - PC: an optional program counter (instruction pointer) for the
//     frame. Capturing this may be disabled in some environments or
//     implementations for performance reasons.
//
//   - File: the source file path corresponding to the call-site.
//
//   - Line: the 1-based line number within File.
//
//   - Function: a human-readable name of the function that contains the
//     call-site.
//
// When Defined is false, all other fields SHOULD be treated as
// unspecified: implementations MAY leave them empty or zero-valued, and
// consumers MUST NOT rely on their contents. When Defined is true, the
// resolver guarantees only that the fields describe some best-effort
// call-site; it does not guarantee a particular formatting style:
//
//   - File may be absolute or relative, may have trimmed prefixes (for
//     example, stripping GOPATH/module roots), and may use platform-
//     specific path separators.
//
//   - Function may or may not be fully qualified (package + receiver +
//     method), and MAY follow an implementation-defined naming scheme.
//
// Consumers that need stable or normalized representations SHOULD apply
// their own post-processing or depend on specific documented behavior of
// a concrete implementation.
//
// # Resolver interface and skip semantics
//
// Resolver abstracts over the mechanism by which call-sites are
// discovered. The primary operation, Caller(skip int), returns a Caller
// describing the stack frame that is skip frames above the Resolver in
// the current goroutine's call stack.
//
// The skip parameter follows the conventional runtime.Caller contract:
//
//   - skip == 0 refers to the frame of Resolver.Caller itself;
//   - skip == 1 refers to the immediate caller of Resolver.Caller;
//   - higher values walk further up the call stack.
//
// Implementations MAY internally adjust the skip value to hide wrapper
// frames (for example, internal logging helpers). Any such adjustment
// SHOULD be documented, and callers that rely on precise frame numbers
// SHOULD treat the effective semantics as implementation-defined.
//
// If the requested frame cannot be resolved (for example, because the
// stack is too shallow or symbol information is unavailable), the
// returned Caller MUST have Defined set to false. In that case, all
// other fields SHOULD be treated as unspecified.
//
// Resolver implementations MUST be safe for concurrent use by multiple
// goroutines unless explicitly documented otherwise. They MAY cache
// symbolization data (for example, function names derived from program
// counters) to improve performance, but MUST NOT retain references to
// mutable application data outside their own internal state.
//
// # Encoder semantics and buffer interaction
//
// Encoder is a function type:
//
//	type Encoder func(dst *buffer.Buffer, caller Caller) *buffer.Buffer
//
// An Encoder receives a destination buffer and a Caller and returns the
// buffer to which it appended the encoded representation. Implementations
// MUST adhere to the following rules:
//
//   - The dst argument is owned by the caller for the duration of the
//     call. Implementations MAY treat it as scratch space and append
//     directly to it, or they MAY allocate or obtain a different buffer
//     (for example, from a pool) and return that instead.
//
//   - If an Encoder returns a different *buffer.Buffer than the dst
//     provided, the original dst MUST remain in a valid state according
//     to the buffer.Buffer contract. The Encoder MUST NOT free either
//     buffer or retain references to them beyond the end of the call.
//
//   - Encoders MUST NOT call Free on any buffer passed to them or
//     returned by them. Lifetime management is always the caller's
//     responsibility.
//
// Callers of an Encoder MUST treat the returned *buffer.Buffer as the
// authoritative destination for any subsequent writes related to that
// encoding step and MUST release it according to the buffer package
// contract (typically via Free) once it is no longer needed.
//
// When caller.Defined is false, an Encoder MAY choose to:
//
//   - append nothing,
//   - append a constant placeholder (for example, "???:0"), or
//   - append an explicit "unknown" representation.
//
// This behavior SHOULD be documented by concrete Encoder implementations.
// Generic callers MUST NOT rely on any particular textual representation
// unless the specific Encoder they use guarantees it.
//
// # Typical usage in the logging pipeline
//
// In a typical rxlog configuration, caller capture and encoding proceed
// as follows:
//
//  1. When a log entry is created, the logging infrastructure calls a
//     Resolver with an implementation-defined skip value chosen to skip
//     internal logging helpers and point at the user's call-site.
//
//  2. The resulting Caller value is stored on the core.Entry
//     representing the log entry. It MAY be reused later when encoding
//     the entry for multiple outputs or formats.
//
//  3. During encoding, a configured caller.Encoder is invoked with the
//     current buffer and the stored Caller. The Encoder appends whichever
//     subset of the Caller information is desired (for example,
//     "file:line", "short/file.go:line", or "pkg.Func file:line").
//
//  4. If caller capture is disabled for performance or policy reasons,
//     the Resolver may always return an undefined Caller, and the
//     Encoder can be configured to omit caller information entirely.
//
// This separation allows applications to adjust caller fidelity,
// formatting, and cost independently, while keeping the core.Entry and
// field encoding logic agnostic to the details of call-site resolution.
//
// # Compatibility and non-goals
//
// The caller package does not attempt to standardize how file paths or
// function names are rendered across all environments. Different
// implementations may choose different strategies (for example, trimming
// module paths, shortening directories, or omitting package names).
//
// Consumers that require stable, machine-parsable caller data SHOULD
// treat Caller as a structured value (File, Line, Function) at the
// encoding boundary and SHOULD NOT attempt to parse the textual output
// of a caller.Encoder, unless that Encoder explicitly documents a
// machine-readable format.
//
// In summary, caller provides a small set of focused abstractions for
// capturing and formatting call-site information, designed to be
// composable with the rest of the rxlog API and safe for use in
// high-throughput, concurrent logging pipelines.
package caller
