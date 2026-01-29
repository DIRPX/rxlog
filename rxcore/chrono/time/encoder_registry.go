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

package time

import (
	"fmt"

	timeapi "dirpx.dev/rxlog/rxapi/chrono/time"
)

// registry maps symbolic encoder names to concrete time encoders.
//
// Keys are lower-case, hyphen-separated identifiers intended for configuration
// (for example, JSON/YAML/TOML). Each entry points to one of the predefined
// Encoder values above.
//
// The registry is deliberately unexported; higher-level packages SHOULD expose
// a stable lookup helper (for example, LookupEncoder(name string) (Encoder, bool))
// rather than relying on map access directly.
var registry = map[string]timeapi.Encoder{
	// ISO 8601–style layouts (UTC by default).

	// Canonical ISO 8601 with millisecond precision (zap-style).
	"iso8601":        ISO8601MillisTimeEncoder,
	"iso8601-millis": ISO8601MillisTimeEncoder,

	// ISO 8601 with second precision.
	"iso8601-sec":     ISO8601SecondsTimeEncoder,
	"iso8601-seconds": ISO8601SecondsTimeEncoder,

	// ISO 8601 with microsecond precision.
	"iso8601-micros": ISO8601MicrosTimeEncoder,

	// ISO 8601 with nanosecond precision.
	"iso8601-nanos": ISO8601NanosTimeEncoder,

	// RFC 3339 and related layouts (canonical for JSON/API payloads).

	// RFC3339 with second precision.
	"rfc3339": RFC3339TimeEncoder,

	// RFC3339Nano with nanosecond precision.
	"rfc3339-nano": RFC3339NanoTimeEncoder,

	// RFC1123 / RFC1123Z (HTTP-date style with textual / numeric zone).
	"rfc1123":  RFC1123TimeEncoder,
	"rfc1123z": RFC1123ZTimeEncoder,

	// RFC822 / RFC822Z legacy email/date formats.
	"rfc822":  RFC822TimeEncoder,
	"rfc822z": RFC822ZTimeEncoder,

	// RFC850 HTTP-date style.
	"rfc850": RFC850TimeEncoder,

	// Other textual layouts from the Go standard library.

	// ANSIC layout, using the local time zone.
	"ansic": ANSICTimeEncoder,

	// UnixDate layout, using the local time zone.
	"unixdate": UnixDateTimeEncoder,

	// Kitchen 12-hour clock layout.
	"kitchen": KitchenTimeEncoder,

	// Stamp family: short, locale-like representations.
	"stamp":        StampTimeEncoder,
	"stamp-millis": StampMilliTimeEncoder,
	"stamp-micros": StampMicroTimeEncoder,
	"stamp-nanos":  StampNanoTimeEncoder,

	// Date / time-only textual layouts.

	// Calendar date only (YYYY-MM-DD) in UTC.
	"date": DateOnlyTimeEncoder,

	// Clock time only (HH:MM:SS) in UTC.
	"time": TimeOnlyTimeEncoder,

	// Clock time with millisecond precision in UTC.
	"time-millis": TimeMillisOnlyTimeEncoder,

	// Numeric Unix epoch representations.

	// Seconds since Unix epoch (1970-01-01T00:00:00Z).
	"unix":         UnixSecondsTimeEncoder,
	"unix-seconds": UnixSecondsTimeEncoder,
	"unix-secs":    UnixSecondsTimeEncoder,

	// Milliseconds since Unix epoch.
	"unix-millis": UnixMillisTimeEncoder,

	// Microseconds since Unix epoch.
	"unix-micros": UnixMicrosTimeEncoder,

	// Nanoseconds since Unix epoch.
	"unix-nanos": UnixNanosTimeEncoder,
}

// FromString looks up a time encoder by its symbolic name.
//
// The name must match one of the registered keys in the internal registry,
// which are conventionally lower-case, hyphen-separated identifiers such as
// "rfc3339", "iso8601-millis", or "unix-millis".
//
// If no encoder is registered under the given name, FromString returns a
// non-nil error and a nil Encoder.
//
// This function is intended for configuration-driven setups (for example,
// parsing encoder names from JSON/YAML/TOML). Callers SHOULD validate or
// normalize user-provided values (for example, to lower-case) before calling
// FromString.
//
// Concurrency:
//   - Concurrent read-only access (calling FromString after all Register calls
//     have completed) is safe.
//   - If Register is called concurrently with FromString, the caller MUST
//     provide external synchronization (for example, a mutex) around registry
//     mutations.
func FromString(name string) (timeapi.Encoder, error) {
	enc, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown time encoder: %q", name)
	}
	return enc, nil
}

// MustFromString is a convenience helper that resolves a time encoder by name
// and panics if the name is not registered.
//
// This is useful in static setup code (for example, wiring encoders from
// hard-coded configuration) where an unknown encoder name indicates a
// programmer error or a misconfigured build. It SHOULD NOT be used for
// untrusted or user-provided input, where returning an error is preferable.
//
// The panic message is the same error produced by FromString(name).
func MustFromString(name string) timeapi.Encoder {
	enc, err := FromString(name)
	if err != nil {
		panic(err)
	}
	return enc
}

// Register installs or overrides a time encoder under the given symbolic name.
//
// If an encoder is already registered under name, it will be replaced.
// Register does not perform any validation on name; callers SHOULD follow the
// same naming convention as the built-in encoders (lower-case, hyphen-separated
// identifiers) to keep configuration consistent.
//
// Register is intended to be called during process initialization (for example,
// from init functions) to extend the set of available encoders with
// application-specific formats.
//
// Concurrency:
//   - Register mutates the shared registry map and is NOT safe to call
//     concurrently with other Register calls or FromString unless the caller
//     provides external synchronization.
//   - The recommended pattern is to perform all Register calls during
//     initialization, before any goroutine starts using FromString or
//     MustFromString.
func Register(name string, encoder timeapi.Encoder) {
	registry[name] = encoder
}
