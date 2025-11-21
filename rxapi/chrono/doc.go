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

// Package chrono groups time-related abstractions used throughout the rxlog
// logging pipeline.
//
// The package itself does not define concrete time types or encoders. Instead,
// it serves as a namespace for a family of sub-packages that together provide
// a consistent model for working with time and duration values in log entries.
//
// # Design goals
//
// The chrono tree is designed around a few core principles:
//
//   - Separation of concerns: capturing the current time, representing it as
//     time.Time or time.Duration, and encoding it into a serialized form are
//     treated as distinct responsibilities. This makes it possible to swap
//     implementations in one area (for example, the time source) without
//     affecting others (for example, how timestamps are formatted).
//
//   - Explicit contracts: all chrono sub-packages define narrow, focused
//     interfaces with clearly documented behavior. Implementations MUST follow
//     these contracts so that higher-level components (encoders, cores, and
//     loggers) can reason about time handling in a uniform way.
//
//   - Testability and determinism: by abstracting over time sources and
//     encoders, applications can replace real wall-clock time with deterministic
//     or virtual time in tests and simulations, while still using the same
//     logging API surface.
//
//   - Compatibility with the standard library: chrono uses the standard
//     library's time.Time and time.Duration types as its primary building
//     blocks. It does not introduce alternative time representations and does
//     not attempt to hide standard time semantics.
//
// # Sub-packages
//
// The chrono namespace is organized into a small set of sub-packages, each of
// which addresses one aspect of time handling:
//
//   - chrono/clock defines the Clock interface, which abstracts over sources of
//     current time and ticker creation. It allows rxlog to be wired either to
//     the real system clock or to custom time sources.
//
//   - chrono/time defines the Encoder abstraction for serializing time.Time
//     values into buffers used by log encoders. Concrete encoder implementations
//     (for example, RFC 3339 or Unix timestamp formats) live outside this
//     API layer.
//
//   - chrono/duration is reserved for abstractions related to
//     encoding time.Duration values in a consistent and configurable way,
//     mirroring the approach used for time.Time.
//
// These sub-packages are intentionally decoupled from any particular log
// encoding format (such as JSON or text). They focus solely on the contracts
// for obtaining and converting time values, leaving the choice of concrete
// encoders and layouts to higher-level or reference implementations.
//
// # Usage in the logging pipeline
//
// Higher-level components in rxlog, such as cores and encoders, depend on the
// chrono abstractions rather than on concrete time sources or formatting
// routines:
//
//   - When a log entry is created, a Clock implementation provides the
//     timestamp to be stored on the entry.
//
//   - When the entry is encoded, a time encoder is used to serialize the
//     stored time.Time (and, where applicable, duration values) into a
//     buffer.Buffer according to the application's configuration.
//
// This indirection allows applications to:
//
//   - switch between different timestamp formats without changing core
//     logging logic;
//
//   - run tests against deterministic clocks;
//
//   - extend or replace encoding strategies while preserving the overall
//     structure and guarantees of the logging API.
//
// In summary, the chrono package defines the conceptual "time subsystem" of
// rxlog. It provides a stable set of interfaces and conventions around time
// and duration handling, while leaving concrete implementations and formats to
// dedicated sub-packages and higher-level modules.
package chrono
