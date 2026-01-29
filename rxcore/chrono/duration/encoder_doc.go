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

// Package duration provides the reference set of duration encoders used by rxcore.
//
// This package builds on top of the rxapi/chrono/duration encoder abstraction
// and offers a curated collection of concrete Encoder instances, together with
// a small name-based registry and configuration helpers. It belongs to the
// rxcore implementation layer rather than the minimal rxapi contract.
//
// The primary goals of this package are:
//
//   - to provide ready-to-use encoders for the most common duration formats;
//   - to make duration encoding configurable via short, string-based names;
//   - to keep rxapi/chrono/duration focused on the low-level Encoder type and
//     delegate policy and defaults to rxcore.
//
// # Relationship to rxapi/chrono/duration
//
// The rxapi/chrono/duration package defines the core Encoder function type:
//
//   - an Encoder takes a *buffer.Buffer and a time.Duration and returns the
//     buffer that now holds the encoded duration.
//
// This rxcore package:
//
//   - implements concrete encoders that serialize duration values as:
//
//   - floating-point seconds (float64);
//
//   - integer milliseconds, microseconds, and nanoseconds (int64);
//
//   - textual representations using time.Duration.String();
//
//   - wires those encoders into a registry keyed by short, lower-case,
//     hyphen-free or hyphen-separated names such as "seconds", "millis",
//     "micros", "nanos", or "string".
//
// The underlying Encoder semantics (append-only behavior, buffer ownership,
// concurrency guarantees) are inherited directly from the rxapi layer.
//
// # Provided encoders
//
// The built-in encoders cover the most common representation choices:
//
//   - SecondsDurationEncoder:
//     serializes a time.Duration as a floating-point number of seconds
//     (float64) written in base 10.
//
//   - MillisDurationEncoder:
//     serializes a time.Duration as an integer number of milliseconds
//     elapsed since zero (int64).
//
//   - MicrosDurationEncoder:
//     serializes a time.Duration as an integer number of microseconds
//     (int64).
//
//   - NanosDurationEncoder:
//     serializes a time.Duration as an integer number of nanoseconds
//     (int64), preserving full duration precision.
//
//   - StringDurationEncoder:
//     serializes a time.Duration using its built-in String method
//     (for example, "1s", "150ms", "200µs").
//
// Callers can use these encoders directly, or select them indirectly via the
// registry based on configuration.
//
// # Registry and configuration
//
// To make duration encoder selection configuration-driven, this package
// maintains an internal registry that maps human-readable names to Encoder
// values. The registry is populated with all built-in encoders using names
// such as:
//
//   - "string";
//
//   - "seconds", "secs", "seconds-f64";
//
//   - "millis", "milliseconds", "ms";
//
//   - "micros", "microseconds", "µs";
//
//   - "nanos", "nanoseconds", "ns".
//
// The helper functions in this package allow callers to:
//
//   - resolve an encoder by name (FromString), returning an error if the name
//     is unknown;
//
//   - register additional encoders for custom formats (Register), typically
//     during process initialization; and
//
//   - resolve an encoder by name with panic-on-failure semantics
//     (MustFromString), suitable for static wiring where unknown names
//     indicate configuration or programmer errors.
//
// Applications are encouraged to perform all Register calls during startup,
// before any goroutines start using FromString or MustFromString, to avoid
// concurrency issues with registry mutation.
//
// # Usage in rxcore
//
// In a typical rxcore configuration, a higher-level encoder or logger:
//
//   - obtains a duration encoder based on a configuration value (for example,
//     "millis" or "string") via FromString or MustFromString;
//
//   - passes the resulting Encoder into a structured encoder configuration,
//     which uses it whenever a time.Duration field needs to be rendered;
//
//   - treats the resulting textual or numeric representation as part of the
//     log entry payload.
//
// Because this package lives in rxcore rather than rxapi, the set of built-in
// encoders and registry mappings can evolve over time without affecting the
// minimal API contracts. It is intended as a convenient, opinionated default
// for duration encoding in rxlog.
package duration
