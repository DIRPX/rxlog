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

// Package rxapi defines the public, implementation-neutral contracts for the
// rxlog logging stack.
//
// This module is the stable boundary between application code, third-party
// integrations, and concrete implementations provided by other modules
// (such as rxcore). It contains only interfaces, value types, and small
// helper abstractions that describe how the logging pipeline is wired and
// how its components interact; it does not contain concrete loggers,
// encoders, or writers.
//
// # Design goals
//
// The rxapi namespace is organized around a few core principles:
//
//   - Clear separation of API and implementation: the rxapi module defines
//     contracts (interfaces, data structures, and invariants). Concrete
//     implementations live in separate modules and packages that depend on
//     rxapi, not the other way around.
//
//   - Minimal but complete surface: each package in this namespace exposes
//     a small, focused set of abstractions (for example, "core", "encoder",
//     "field") that together are sufficient to describe the entire logging
//     pipeline without prescribing a particular backend or format.
//
//   - Orthogonal components: abstractions are designed to compose cleanly.
//     For example, cores depend on encoders and writers; encoders depend on
//     buffers, fields, and error handling; predicate and hook packages
//     express optional behavior that can be layered on top.
//
//   - Concurrency awareness: contracts explicitly state which components
//     MUST be safe for concurrent use (for example, cores, encoders,
//     handlers) and which are confined to a single goroutine at a time
//     (for example, individual object and array encoder instances).
//
//   - Compatibility over time: types and interfaces in this module are
//     intended to evolve slowly. Additive changes are preferred; breaking
//     changes require a deliberate versioning strategy.
//
// # Subpackages
//
// The rxapi namespace is split into subpackages, each responsible for a
// specific aspect of the logging model. At a high level:
//
//   - buffer: abstractions and contracts for reusable, growable byte
//     buffers used by encoders and writers.
//
//   - chrono: time-related abstractions used when working with timestamps
//     and durations in log entries.
//
//   - core: the core logging primitives, including Entry, Core, and
//     CheckedEntry, which define how log events flow through the pipeline.
//
//   - encoder and its subpackages: interfaces and helpers for turning
//     entries and fields into serialized byte streams, along with
//     reusable building blocks for object/array encoders and extension
//     points for custom encoding logic.
//
//   - error: contracts for handling per-field encoding failures and
//     for serializing error values into buffers.
//
//   - field and its subpackages: the structured attribute model, including
//     a compact field representation and the type discriminator used to
//     interpret field storage.
//
//   - hook: extensibility points for observing and influencing the
//     lifecycle of log entries before and after they are written.
//
//   - level: the severity model used by the logging API to classify log
//     entries and enable or disable them based on thresholds.
//
//   - name: abstractions for logical names associated with loggers,
//     cores, components, or subsystems.
//
//   - predicate: composable boolean conditions used to express routing,
//     sampling, and filtering policies for log events.
//
//   - writer: low-level sink abstractions for delivering serialized log
//     entries to files, network connections, or other destinations.
//
//   - caller and stack: abstractions for representing call-site
//     information and stack traces attached to log entries.
//
//   - reflect: contracts for reflection-based encoding of arbitrary Go
//     values as a fallback when type-specific encoders are not available.
//
//   - context and its subpackages: helpers for turning values stored in
//     context.Context into structured fields attached to log entries.
//
//   - pool: a small, typed abstraction over sync.Pool for reusing
//     heap-allocated values such as buffers in performance-sensitive
//     paths.
//
// These subpackages are loosely coupled but share a common set of
// expectations around ownership, lifetime, error handling, and concurrency.
// They are designed to be usable independently where appropriate—for
// example, an application might depend only on field and encoder contracts
// when integrating a custom backend.
//
// # Relationship to implementations
//
// The rxapi module deliberately does not provide concrete encoders, cores,
// or writers. Instead, it is intended to be imported by:
//
//   - application code that wants to remain agnostic to specific logging
//     backends while still depending on well-defined contracts;
//
//   - infrastructure libraries that implement cores, encoders, writers,
//     hooks, and error handlers on top of rxapi abstractions; and
//
//   - reference implementations that live in separate modules (for
//     example, rxcore) and are free to evolve as long as they continue to
//     satisfy the contracts published here.
//
// By centralizing the logging contracts in this module, rxlog enables
// applications, libraries, and backends to interoperate without tight
// coupling, while keeping the semantics of logging explicit and
// well-documented.
package rxapi
