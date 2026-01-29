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

// Package timeenc provides the reference set of time encoders used by rxcore.
//
// This package builds on top of the rxapi/chrono/time encoder abstraction and
// offers a curated collection of concrete Encoder instances, configuration-
// friendly lookup by name, and small convenience helpers. It is part of the
// rxcore implementation layer, not the minimal rxapi contract.
//
// The goals of this package are:
//
//   - to provide a ready-to-use set of time encoders covering the most common
//     textual and numeric timestamp formats;
//
//   - to expose a stable, string-keyed registry so that encoders can be
//     selected from configuration (for example, JSON/YAML/TOML);
//
//   - to keep rxapi/chrono/time focused on the low-level Encoder type and
//     leave policy and defaults to rxcore.
//
// # Relationship to rxapi/chrono/time
//
// The rxapi/chrono/time package defines the core Encoder function type:
//
//   - an Encoder takes a *buffer.Buffer and a stdtime.Time and returns the buffer
//     that now holds the encoded timestamp.
//
// This rxcore package:
//
//   - implements concrete encoders for specific layouts (ISO 8601 variants,
//     RFC3339/RFC3339Nano, RFC1123/RFC822/RFC850, ANSIC/UnixDate, and the
//     Stamp/Kitchen family);
//
//   - provides numeric Unix epoch encoders (seconds, milliseconds, microseconds,
//     nanoseconds) that write integer timestamps;
//
//   - wires those encoders into a registry keyed by short, lower-case,
//     hyphen-separated names such as "rfc3339", "iso8601-millis",
//     "unix-millis", "date", or "time-millis".
//
// The underlying Encoder semantics (buffer ownership, append-only behavior,
// and concurrency guarantees) are inherited directly from the rxapi layer.
//
// # Layouts and the layout subpackage
//
// The encoder package uses layout strings compatible with stdtime.Time.Format to
// implement its textual encoders. To avoid scattering literal layout strings
// throughout the codebase, common layouts are centralized in:
//
//   - layout.go in this package, which defines the core set of layouts used by
//     the built-in encoders; and
//
//   - the nested layout subpackage, which exposes these layout constants as a
//     reusable, importable module for other code.
//
// Consumers that need a custom layout can either:
//
//   - use the helper that wraps an arbitrary layout into an Encoder, or
//   - define their own layout string and register the resulting encoder under
//     an application-specific name.
//
// # Registry and configuration
//
// To make encoder selection configuration-driven, this package maintains an
// internal registry that maps human-readable names to Encoder values. The
// registry is populated with all built-in encoders, including:
//
//   - ISO 8601–style encoders (with seconds, millis, micros, nanos);
//
//   - RFC3339/RFC3339Nano and other RFC-style layouts (RFC1123, RFC1123Z,
//     RFC822, RFC822Z, RFC850);
//
//   - ANSIC, UnixDate, Kitchen, and the Stamp family;
//
//   - date-only and time-only layouts;
//
//   - numeric Unix epoch encoders (seconds/millis/micros/nanos).
//
// The helper functions provided in registry.go allow callers to:
//
//   - resolve encoders by name (FromString), returning an error on unknown
//     names;
//
//   - register additional encoders for custom formats (Register); and
//
//   - perform name resolution that panics on failure (MustFromString), which
//     is suitable for static initialization but not for untrusted input.
//
// Applications are encouraged to perform all Register calls during process
// initialization and to treat FromString errors as configuration issues.
//
// # Shortcuts and integration
//
// The shortcuts.go file contains small convenience helpers that make it easier
// to wire time encoders into higher-level rxcore components (for example,
// default encoder configurations or logger presets). These shortcuts do not
// introduce new semantics; they simply package common combinations of
// encoders, layouts, and policies.
//
// In typical usage, callers:
//
//   - pick a symbolic encoder name (for example, "rfc3339-nano" or
//     "iso8601-millis") from configuration;
//
//   - resolve it via FromString or MustFromString;
//
//   - pass the resulting Encoder into the higher-level encoder configuration
//     used by rxcore.
//
// Because this package lives in rxcore rather than rxapi, its set of built-in
// encoders and registry mappings can evolve over time without affecting the
// minimal API contracts. It is intended as a convenient, opinionated default
// rather than a hard requirement for all rxlog integrations.
package time
