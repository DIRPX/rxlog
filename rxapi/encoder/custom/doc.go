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

// Package custom defines extension points for registering and invoking
// application-specific encoders.
//
// While the core encoder API focuses on built-in types (primitive values,
// time, duration, errors, structured fields, etc.), many applications need
// to control how their own domain types are rendered in logs. The custom
// package provides a small set of abstractions for:
//
//   - identifying custom encoder slots in a configuration-stable way,
//
//   - associating those identifiers with concrete encoder implementations,
//     and
//
//   - resolving and invoking custom encoders from higher-level encoders
//     without introducing ad hoc reflection or hard-coded type switches.
//
// # Design goals
//
// The custom extension points are designed around a few principles:
//
//   - Separation of concerns: the decision to use a custom encoder is made
//     at the field or configuration level, while the actual encoding logic
//     lives in dedicated encoder implementations.
//
//   - Configuration stability: custom encoders can be referred to by stable
//     identifiers (for example, names or tokens) that are safe to embed in
//     configuration files or wiring code.
//
//   - Type safety at the edges: while custom encoders often work with
//     interface{} values, their relationship to particular Go types is
//     documented and enforced by the registration and lookup layer.
//
//   - Concurrency safety: registration and lookup mechanisms MUST be safe
//     for use by multiple goroutines once the logging stack is wired.
//
// # Abstractions
//
// This package typically introduces three kinds of abstractions:
//
//   - an identifier type that represents a "custom encoder slot" (for example,
//     a string name or small struct used as a registry key),
//
//   - an encoder function or interface that describes how a custom-encoded
//     value is transformed into bytes on top of buffer.Buffer, and
//
//   - a registry or resolver API that maps identifiers (and, where relevant,
//     Go types) to concrete encoder implementations.
//
// The exact type names and signatures are documented alongside their
// definitions in this package. At a high level:
//
//   - Custom encoder functions are expected to follow the same buffer
//     ownership and lifetime rules as other encoders in the rxlog API:
//     they MAY append directly to a provided buffer, MAY return a different
//     buffer, MUST NOT call Free on any buffer, and MUST NOT retain
//     references to buffers or slices beyond the end of a call.
//
//   - Custom encoder registries are expected to be populated during process
//     initialization or wiring. After registration is complete, lookups MUST
//     be safe for concurrent use by the encoding pipeline.
//
// # Registration and lookup
//
// A central piece of this package is the ability to register custom encoder
// implementations under stable identifiers and to resolve them when needed.
//
// Typical patterns include:
//
//   - a registration API that accepts an identifier and a custom encoder,
//     establishing a mapping that encoders can consult later;
//
//   - a lookup API that, given an identifier (and optionally a Go type),
//     returns the corresponding encoder or reports that none is registered;
//
//   - well-defined behavior for duplicate registrations (for example, panic,
//     overwrite, or ignore), which MUST be documented by the implementation.
//
// Registration is normally performed at startup:
//
//   - in init functions in application code or libraries, or
//   - in explicit wiring code that assembles the logging stack.
//
// Once registration is complete, the registry is treated as read-mostly or
// effectively immutable; the encoding pipeline assumes that resolvers do not
// change behavior in unexpected ways at runtime.
//
// # Integration with field and encoder packages
//
// The custom package does not define field.Field or the core encoder
// interfaces; instead, it complements them:
//
//   - field helpers may allow callers to mark certain values as requiring a
//     custom encoder, for example by associating a custom identifier with a
//     field or by wrapping values in a marker type;
//
//   - object or array encoders may consult a custom encoder registry when
//     they encounter such marked values, dispatching to the registered
//     implementation instead of falling back to built-in or reflection-based
//     encoding;
//
//   - error handlers and reflective encoders can be combined with custom
//     encoders to provide sensible fallbacks when no custom implementation
//     is available for a given value.
//
// This separation allows applications to:
//
//   - introduce custom encoders incrementally, without modifying the core
//     encoding logic;
//
//   - share custom encoding logic across multiple encoders or outputs;
//
//   - change or extend custom encoders via configuration or wiring without
//     changing application call sites.
//
// # Concurrency and error handling
//
// Custom encoders and registries MUST be safe for concurrent use by multiple
// goroutines once the logging stack is initialized. In particular:
//
//   - encoder implementations MUST NOT mutate global state in ways that are
//     not explicitly synchronized;
//
//   - registries MUST protect any internal maps or tables they maintain with
//     appropriate synchronization or initialization-time freezing;
//
//   - when a custom encoder fails (for example, due to an unsupported value),
//     it SHOULD return a well-defined error that higher-level encoders can
//     route through the error.Handler policy configured for the logging
//     pipeline.
//
// The custom package itself does not prescribe a specific error-handling
// strategy; it defines how errors produced by custom encoders are surfaced
// so that the surrounding encoder infrastructure can decide whether to abort,
// skip, or replace failing fields.
//
// # Usage context
//
// The encoder/custom package is primarily intended for encoder authors and
// infrastructure code configuring the logging stack. Typical usage patterns
// include:
//
//   - registering encoders for domain-specific types (for example, IDs,
//     money, version objects, or rich error types),
//
//   - wiring registry-backed resolvers into object encoders so that fields
//     annotated as "custom" dispatch to the appropriate implementation, and
//
//   - defining shared custom encoders that can be reused across multiple
//     services or binaries.
//
// Application code generally interacts with this package indirectly through
// higher-level field or encoder helpers rather than calling the registry
// APIs directly. By centralizing custom encoder contracts and registration
// here, rxlog keeps the extension mechanism explicit, testable, and
// configuration-friendly.
package custom
