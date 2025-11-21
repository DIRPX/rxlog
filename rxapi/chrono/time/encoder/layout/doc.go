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

// Package layout defines named time formatting layouts used across the rxlog
// ecosystem.
//
// The goal of this package is to centralize commonly used time formats under
// stable, descriptive identifiers so that configuration remains
// self-documenting and unambiguous. Instead of scattering raw layout strings
// like "2006-01-02T15:04:05Z07:00" throughout the codebase or configuration,
// callers SHOULD refer to the symbolic names exposed by this package.
//
// # Design and usage
//
// All constants in this package are plain Go time layouts and MAY be used
// anywhere a layout string is required:
//
//   - directly with time.Format / time.Parse,
//   - passed to helper constructors such as EncoderOfLayout,
//   - stored in configuration files or flags (for example, as default values).
//
// Layout names are intentionally verbose and explicit. For example,
// ISO8601Seconds and ISO8601Millis encode both the style (ISO-like) and the
// precision (seconds, milliseconds, microseconds, nanoseconds). This allows
// operators and readers of configuration to understand the format without
// memorizing the underlying reference string.
//
// # Stability guarantees
//
// The semantic meaning of each exported constant (for example, “ISO-like
// layout with second precision and colon offset”) MUST remain stable over time.
// The concrete layout string associated with a given name SHOULD NOT change in
// a way that would break existing logs, stored data, or consumers that parse
// those logs.
//
// New layouts MAY be added over time, but existing ones SHOULD be considered
// part of the public logging contract and changed only with great care.
//
// # Layout families
//
// The package groups layouts into several broad families:
//
//   - ISO 8601–inspired layouts
//     These follow widely used ISO-style patterns (for example,
//     ISO8601Seconds, ISO8601Millis, ISO8601Micros, ISO8601Nanos). They are
//     convenient for structured logging and interoperable text formats.
//
//   - Date / time only layouts
//     These format only the calendar date or only the clock time (for example,
//     Date, Time, TimeMillis) and are useful for human-oriented outputs,
//     console logs, or derived metrics.
//
//   - Re-exports of standard library layouts
//     Well-known layouts from the standard library (RFC3339, RFC1123, ANSIC,
//     UnixDate, Kitchen, Stamp*, etc.) are re-exported under the same names.
//     This allows callers to configure formats via this package without
//     importing time directly.
//
// # Interoperability with encoders
//
// The layouts defined here are typically consumed by higher-level encoder
// helpers such as EncoderOfLayout and the predefined time encoders in other
// rxlog packages. Those helpers:
//
//   - take a layout from this package,
//   - optionally apply a specific time.Location (for example, UTC or Local),
//   - write formatted timestamps directly into a buffer without allocating
//     intermediate strings when possible.
//
// Callers configuring time encoders SHOULD prefer these named layouts over
// ad-hoc literals to keep configuration concise, consistent, and grep-friendly.
//
// # Extensibility
//
// This package intentionally does not attempt to cover every possible time
// format. Applications MAY define their own custom layouts alongside these
// constants. When doing so, they SHOULD follow the same naming principles
// (explicit, precise, and stable) to maintain clarity in configuration and
// documentation.
package layout
