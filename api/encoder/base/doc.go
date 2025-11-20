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

// Package base provides reusable building blocks for implementing
// encoder.ObjectEncoder and encoder.ArrayEncoder on top of buffer.Buffer.
//
// The types in this package are intended to be embedded or wrapped by
// concrete encoders (for example, JSON or text encoders). They factor out
// common concerns such as:
//
//   - managing the relationship between an encoder and its backing
//     buffer.Buffer;
//
//   - enforcing the state machine of object and array encoders (for
//     example, preventing further writes after an object has been closed);
//
//   - routing field.Field values to the appropriate AddXXX methods
//     without using reflection;
//
//   - integrating with error-handling policies defined by the error
//     package; and
//
//   - providing a small set of helpers for encoding primitive values in a
//     consistent way.
//
// The base package is format-agnostic: it does not mandate a particular
// on-the-wire representation (such as JSON or text). Instead, it provides
// infrastructure that concrete encoders can use to implement their own
// serialization formats.
//
// # Design goals
//
// The design of this package is guided by a few goals:
//
//   - Encoders should be easy to implement correctly: most of the
//     repetitive bookkeeping and state management should live here, so
//     that format-specific code can focus on how to represent values.
//
//   - Encoders should be efficient: base types work directly with
//     buffer.Buffer and field.Field to avoid reflection and minimize
//     allocations in hot paths.
//
//   - Semantics should be explicit: object and array encoders expose a
//     clear state machine, and error-handling behavior is centralized
//     via the error.Handler abstraction rather than scattered across
//     individual AddXXX methods.
//
//   - The base layer should be reusable across multiple concrete encoders,
//     even when they differ in details such as key quoting, delimiter
//     syntax, or representation of null values.
//
// # Object and array encoding
//
// Concrete types in this package implement the encoder.ObjectEncoder and
// encoder.ArrayEncoder interfaces defined in the parent encoder package.
// They typically maintain:
//
//   - a reference to a buffer.Buffer, which holds the serialized
//     representation being constructed;
//
//   - a small amount of encoder-local state (for example, whether the
//     current object is empty, whether a comma is needed before the next
//     element, or whether the encoder has been closed); and
//
//   - references to supporting components such as error.Handler or
//     reflection-based encoders for "any" values.
//
// The expected usage pattern is:
//
//  1. A higher-level encoder obtains or resets a base object or array
//     encoder for the current log entry.
//
//  2. It delegates AddXXX calls to the base encoder, which:
//
//     - updates its internal state,
//     - appends to the backing buffer, and
//     - consults the configured error.Handler if encoding of a value
//     fails.
//
//  3. Once all fields or elements have been added, the higher-level
//     encoder finalizes the object or array (for example, closes any
//     delimiters) and hands the buffer off to a writer.
//
// Base encoders are not safe for concurrent use by multiple goroutines.
// Each instance MUST be used by at most one goroutine at a time, and it
// MUST NOT be reused for a new log entry until it has been fully reset
// according to the documentation of the concrete type.
//
// # Field routing and error handling
//
// One of the responsibilities of this package is to bridge between the
// field.Field representation and the encoder.ObjectEncoder /
// encoder.ArrayEncoder methods.
//
// Rather than performing reflection at encoding time, concrete object
// encoders typically expose a method that accepts field.Field values.
// That method:
//
//   - switches on the field.Type discriminator,
//   - reads the appropriate storage slot(s) (Integer, String, Interface),
//   - calls the corresponding AddXXX method on the underlying encoder, and
//   - handles any errors produced by those AddXXX calls.
//
// The base package provides helpers and patterns for implementing this
// routing consistently and for integrating with error.Handler:
//
//   - When an AddXXX method returns an error, the base encoder forwards
//     that error to the configured error.Handler along with the logical
//     field key and value.
//
//   - If the handler returns a non-nil error, the encoder aborts encoding
//     for the current entry and propagates that error to its caller.
//
//   - If the handler returns nil, the failing field is treated as skipped
//     and encoding of the remaining fields continues.
//
// This centralizes error-handling policy and keeps the AddXXX methods
// themselves focused on the mechanics of writing values to the buffer.
//
// # Buffer interaction and lifetime
//
// Base encoders are tightly coupled to buffer.Buffer. They MUST obey the
// buffer package's ownership and lifetime rules:
//
//   - Each encoder instance operates on a buffer that is owned by the
//     caller for the duration of the encoding operation. The encoder MAY
//     append to this buffer but MUST NOT call Free on it.
//
//   - If a base encoder needs to switch to a different buffer (for example,
//     to implement a particular formatting strategy), it MUST return that
//     new buffer to its caller and leave the original buffer in a valid
//     state. The encoder MUST NOT retain references to any buffer or to
//     slices derived from it beyond the lifetime of the encoding
//     operation.
//
//   - Higher-level components that manage buffer pools are responsible for
//     returning buffers to their pools once encoding is complete.
//
// The base package does not manage buffer pools directly; it assumes that
// encoders are wired with buffers by infrastructure code that understands
// the pooling and lifetime rules for the application.
//
// # Format-specific behavior
//
// While this package provides shared infrastructure, it does not impose a
// particular serialization format. Concrete encoders built on top of base
// types are expected to define and document:
//
//   - how keys and values are rendered (for example, as JSON strings,
//     numbers, or nested objects),
//
//   - how special values such as nil, NaN, or infinities are handled,
//
//   - how time and duration values are mapped to their on-the-wire
//     representation, and
//
//   - which characters need to be escaped or quoted.
//
// Base types are designed to be flexible enough to support these different
// strategies without prescribing them.
//
// # Usage context
//
// The encoder/base package is intended for use by infrastructure and
// encoder authors, not by most application code. Typical usage looks like:
//
//   - A library implements a JSON encoder by embedding or composing the
//     base object and array encoders and customizing how primitives are
//     rendered.
//
//   - A console or text encoder reuses the same base machinery but chooses
//     a different representation for keys, values, and delimiters.
//
//   - Test or debug encoders use base types to ensure consistent handling of
//     field.Type and error.Handler while experimenting with alternative
//     formats.
//
// By centralizing common encoder mechanics in this package, rxlog reduces
// duplication, keeps encoder behavior consistent across formats, and makes
// it easier to evolve the encoding contracts over time.
package base
