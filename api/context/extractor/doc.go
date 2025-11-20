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

// Package extractor defines abstractions for deriving structured log fields
// from a context.Context.
//
// In many applications, request-scoped or operation-scoped metadata (such as
// trace IDs, user IDs, tenant information, or feature flags) is stored in a
// context.Context. This package provides a small, focused API for extracting
// such metadata and turning it into []field.Field values that can be attached
// to log entries.
//
// The central abstractions are:
//
//   - Extractor: an interface implemented by types that know how to inspect a
//     context.Context and produce zero or more field.Field values.
//
//   - Func: an adapter that allows ordinary functions with the appropriate
//     signature to be used wherever an Extractor is expected.
//
// The package may also expose helper functions for combining multiple
// Extractor implementations into a single composite Extractor, so that
// applications can assemble their context-handling policy from reusable
// building blocks.
//
// # Extractor semantics
//
// An Extractor is conceptually a pure function from context.Context to
// []field.Field:
//
//   - It MUST NOT modify the context.Context it receives.
//
//   - It SHOULD be free of observable side effects other than the fields it
//     returns; in particular, it SHOULD NOT perform blocking I/O or long-
//     running operations.
//
//   - It MUST be safe for concurrent use by multiple goroutines, since the
//     same Extractor instance may be reused across many log calls in a
//     multi-goroutine application.
//
// The typical Extract method has the following shape:
//
//	Extract(ctx context.Context) []field.Field
//
// Implementations interpret the context according to their own conventions:
// they may look up values by well-known context keys, unwrap framework-specific
// carrier types, or derive fields from nested structures stored in the context.
//
// The returned slice:
//
//   - MAY be nil to indicate that no fields were extracted from this context.
//
//   - MAY be reused by the Extractor between calls, as long as callers treat
//     the slice as read-only and do not retain it beyond the point where the
//     logging subsystem has consumed it.
//
//   - MUST be treated as immutable by callers; they MUST NOT modify the slice
//     itself or any Field values contained in it unless they have made their
//     own copy.
//
// # Ordering and key conflicts
//
// This package does not impose a global ordering on the fields returned by an
// Extractor; however, implementations SHOULD choose a deterministic ordering
// for a given context input, to make logs easier to read and to help
// downstream systems that rely on field ordering.
//
// When multiple Extractors are used together, or when context-derived fields
// are combined with explicitly provided fields, key collisions may occur:
// different producers may emit fields with the same Key but different values.
//
// The extractor package does not define a global collision policy. Instead:
//
//   - Individual Extractors SHOULD avoid introducing keys that are likely to
//     conflict with application-level fields, for example by using namespaces
//     or prefixes where appropriate.
//
//   - Higher-level components that merge field lists MUST define their own
//     policy (for example, "explicit fields override context fields" or vice
//     versa) and apply it consistently when combining slices of fields.
//
// # Error handling
//
// The Extractor abstraction does not expose an error return. Extractors
// therefore need to handle failures internally. Common strategies include:
//
//   - returning no fields when necessary information is missing or malformed;
//
//   - emitting diagnostic fields that capture the fact that extraction failed
//     (for example, a field indicating a parse error for a particular context
//     key);
//
//   - logging internal errors through a separate, out-of-band mechanism,
//     while keeping Extract itself side-effect-free with respect to the
//     primary logging pipeline.
//
// Extractors SHOULD avoid panicking in normal operation; a panic typically
// indicates a programmer error in the extraction logic rather than a normal
// runtime failure.
//
// # Func adapter and composition
//
// The Func type is a convenience adapter that allows ordinary functions to
// satisfy the Extractor interface. This makes it easy to define small,
// focused extraction steps as standalone functions and then use them wherever
// an Extractor is expected.
//
// The package may also provide helper functions for combining multiple
// Extractor values into a single composite Extractor. Such helpers typically:
//
//   - invoke each underlying Extractor in a defined order,
//   - concatenate the returned field slices (skipping nil results), and
//   - return the combined slice to the caller.
//
// Composition helpers MUST preserve the Extractor semantics described above:
// they MUST be safe for concurrent use and MUST NOT modify the context.
//
// # Usage context
//
// The extractor package is intentionally small and self-contained. It does
// not perform I/O and does not depend on higher-level logging concepts such
// as cores or writers. Instead, it provides:
//
//   - a single abstraction (Extractor) for turning context.Context values
//     into []field.Field, and
//
//   - adapters and helpers that make it easy to compose context-handling
//     logic from smaller pieces.
//
// In the broader rxlog pipeline, Extractor values are typically used by
// components that construct log entries from a context, such as logger
// helpers that accept a context.Context parameter. These components:
//
//   - pass the application-provided context to one or more Extractors,
//
//   - merge the resulting fields with explicit fields supplied by the caller,
//     and
//
//   - attach the combined set of fields to the log entry before encoding it.
//
// By centralizing the contract for context-derived attributes in this
// package, rxlog keeps context handling explicit, testable, and easy to
// customize for different environments and frameworks.
package extractor
