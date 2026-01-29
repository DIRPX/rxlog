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

// Package level defines the severity model used by the rxlog logging API.
//
// The central abstraction in this package is the Level type, which represents
// the relative importance or criticality of a log entry. Levels form a
// totally ordered set: lower values correspond to less severe events (for
// example, verbose diagnostic output), and higher values correspond to more
// severe events (for example, failures or irrecoverable conditions).
//
// In addition to the Level type itself, the package provides:
//
//   - a small, fixed set of named Level constants representing commonly used
//     severities (such as debug, info, warning, error, fatal);
//
//   - helper abstractions such as Enabler and Threshold that encode level-
//     based filtering policies; and
//
//   - the Encoder type, which describes how levels are rendered into bytes
//     for downstream encoders and outputs.
//
// # Level semantics
//
// The Level type is an integer-backed enumeration whose values are ordered
// from least to most severe. All named Level constants defined by this
// package are valid severities. Code that compares levels MUST rely only on
// their relative ordering (for example, "greater than or equal to") and MUST
// NOT depend on specific numeric values, which are considered an internal
// representation detail.
//
// A typical usage pattern is:
//
//   - configure a minimum enabled level (for example, "info");
//
//   - for each log entry, compare its Level against this threshold; and
//
//   - drop any entries whose Level is below the configured threshold.
//
// This comparison-based approach is embodied in the Enabler interface and is
// used throughout the logging pipeline to decide whether a particular event
// should be recorded.
//
// The package also defines a special Invalid sentinel value that represents an
// out-of-range or unrecognized level (for example, the result of a failed
// parse). Invalid is strictly greater than all valid levels and MUST NOT be
// used as the Level of an actual log entry. Callers and parsers MAY return
// Invalid to signal configuration or input errors, and any code that receives
// Invalid MUST treat it as an error condition rather than a normal severity.
//
// # Enabler and level-based filtering
//
// To make level-based filtering reusable and composable, this package exposes
// the Enabler interface. An Enabler encapsulates the decision of whether a
// given Level is enabled in a particular context. A simple Enabler might
// compare a Level to a fixed minimum threshold, while more advanced
// implementations could take additional factors into account.
//
// Higher-level components (such as cores, hooks, or predicates) depend on
// Enabler to answer the question "should an event at this level be logged?"
// without hard-coding any particular filtering policy. This allows:
//
//   - applications to plug in their own enabling logic;
//
//   - shared infrastructure to apply consistent level checks across multiple
//     outputs; and
//
//   - different parts of the system to use different thresholds while sharing
//     the same Level type.
//
// Implementations of Enabler MUST be safe for concurrent use by multiple
// goroutines unless explicitly documented otherwise. They SHOULD be cheap to
// evaluate, since they may be invoked on every log call, and they MUST NOT
// have side effects beyond the decision they return.
//
// The EnablerFunc helper type allows ordinary functions to satisfy the
// Enabler interface, making it easy to define custom severity filters inline
// without declaring separate types.
//
// # Thresholds and dynamic configuration
//
// In addition to read-only enabling logic, many applications need to adjust
// their effective log level at runtime (for example, via configuration reload,
// an admin HTTP endpoint, or a control file). For this purpose, the package
// defines the Threshold interface.
//
// A Threshold represents a mutable severity threshold. It combines:
//
//   - the Enabled(Level) check from Enabler; and
//
//   - methods to observe and modify the underlying minimum enabled level
//     (Level() and SetLevel(Level)).
//
// Conceptually, a Threshold models the common pattern of "minimum severity"
// filtering: events at or above the current threshold SHOULD be considered
// enabled, and events below it SHOULD be considered disabled. Typical
// implementations store the current minimum Level internally and implement
// Enabled(l) as:
//
//	l >= Level()
//
// Implementations of Threshold MUST be safe for concurrent use by multiple
// goroutines. They MUST keep Enabled as lightweight and side-effect free as
// any other Enabler. SetLevel(Level) is expected to be called much less
// frequently (for example, during reconfiguration) and MAY perform additional
// validation or bookkeeping as documented by the implementation.
//
// The level package defines only the Threshold interface. Concrete threshold
// implementations (for example, atomic thresholds used in rxcore) live in
// higher-level packages and are free to choose their own storage and
// synchronization strategies, as long as they honor the Threshold contract.
//
// # Encoding levels
//
// The Encoder type defined in this package describes how a Level is rendered
// into bytes. An Encoder is a function that appends an encoded representation
// of a Level to a buffer.Buffer and returns the buffer that should be used
// for subsequent writes:
//
//	type Encoder func(dst *buffer.Buffer, level Level) *buffer.Buffer
//
// This abstraction decouples the internal severity model from any particular
// textual or binary representation. For example, one Encoder might render
// levels as symbolic names ("DEBUG", "INFO", "WARN", "ERROR"), another might
// use numeric codes, and a third might emit structured objects with both
// code and name fields.
//
// Encoder implementations MUST obey the ownership and lifetime rules of the
// buffer package:
//
//   - The dst argument is owned by the caller for the duration of the call.
//     The Encoder MAY append directly to dst or MAY obtain and return a
//     different *buffer.Buffer if that better suits its implementation.
//
//   - The Encoder MUST return the *buffer.Buffer that now holds the encoded
//     level. If this differs from dst, the original dst MUST remain in a
//     valid state according to its own contract.
//
//   - The Encoder MUST NOT call Free on any buffer it receives or returns,
//     and it MUST NOT retain references to those buffers or to slices
//     derived from them beyond the end of the call.
//
// Within these constraints, implementations are free to choose any stable
// mapping from Level values to serialized form. Encoders SHOULD document the
// mapping they use (for example, which numeric codes or strings correspond to
// which Level values) and SHOULD keep that mapping stable over time so that
// downstream systems can reliably interpret and aggregate log entries.
//
// # Usage context
//
// The level package is intentionally small and self-contained. It does not
// perform I/O and does not depend on other parts of the logging subsystem.
// Instead, it provides a shared vocabulary and contract for representing,
// checking, and encoding log severities.
//
// In the broader rxlog pipeline:
//
//   - log entries carry a Level value indicating their severity;
//
//   - enabling decisions are made via Enabler (often embedded into cores or
//     other components), and mutable thresholds are exposed via Threshold
//     implementations in higher layers; and
//
//   - level encoders are used by higher-level encoders when serializing log
//     entries for output.
//
// By centralizing severity semantics in this package, rxlog ensures that all
// components reason about log levels in the same way, regardless of which
// concrete encoders, cores, or outputs are in use.
package level
