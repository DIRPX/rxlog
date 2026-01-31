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
	timeapi "dirpx.dev/rxlog/rxapi/chrono/time"
	"dirpx.dev/rxlog/rxcore/registry"
)

// encoderRegistry holds registered time encoders keyed by symbolic names.
//
// Built-in encoders are registered during package initialization. Callers
// SHOULD use FromString, Register, and MustFromString to interact with the
// registry rather than accessing it directly.
var encoderRegistry = registry.New[timeapi.Encoder]()

func init() {
	// Register built-in time encoders with their aliases.

	// ISO 8601–style layouts (UTC by default).
	encoderRegistry.Register("iso8601", ISO8601MillisTimeEncoder)
	encoderRegistry.Register("iso8601-millis", ISO8601MillisTimeEncoder)
	encoderRegistry.Register("iso8601-sec", ISO8601SecondsTimeEncoder)
	encoderRegistry.Register("iso8601-seconds", ISO8601SecondsTimeEncoder)
	encoderRegistry.Register("iso8601-micros", ISO8601MicrosTimeEncoder)
	encoderRegistry.Register("iso8601-nanos", ISO8601NanosTimeEncoder)

	// RFC 3339 and related layouts (canonical for JSON/API payloads).
	encoderRegistry.Register("rfc3339", RFC3339TimeEncoder)
	encoderRegistry.Register("rfc3339-nano", RFC3339NanoTimeEncoder)

	// RFC1123 / RFC1123Z (HTTP-date style with textual / numeric zone).
	encoderRegistry.Register("rfc1123", RFC1123TimeEncoder)
	encoderRegistry.Register("rfc1123z", RFC1123ZTimeEncoder)

	// RFC822 / RFC822Z legacy email/date formats.
	encoderRegistry.Register("rfc822", RFC822TimeEncoder)
	encoderRegistry.Register("rfc822z", RFC822ZTimeEncoder)

	// RFC850 HTTP-date style.
	encoderRegistry.Register("rfc850", RFC850TimeEncoder)

	// Other textual layouts from the Go standard library.
	encoderRegistry.Register("ansic", ANSICTimeEncoder)
	encoderRegistry.Register("unixdate", UnixDateTimeEncoder)
	encoderRegistry.Register("kitchen", KitchenTimeEncoder)

	// Stamp family: short, locale-like representations.
	encoderRegistry.Register("stamp", StampTimeEncoder)
	encoderRegistry.Register("stamp-millis", StampMilliTimeEncoder)
	encoderRegistry.Register("stamp-micros", StampMicroTimeEncoder)
	encoderRegistry.Register("stamp-nanos", StampNanoTimeEncoder)

	// Date / time-only textual layouts.
	encoderRegistry.Register("date", DateOnlyTimeEncoder)
	encoderRegistry.Register("time", TimeOnlyTimeEncoder)
	encoderRegistry.Register("time-millis", TimeMillisOnlyTimeEncoder)

	// Numeric Unix epoch representations.
	encoderRegistry.Register("unix", UnixSecondsTimeEncoder)
	encoderRegistry.Register("unix-seconds", UnixSecondsTimeEncoder)
	encoderRegistry.Register("unix-secs", UnixSecondsTimeEncoder)
	encoderRegistry.Register("unix-millis", UnixMillisTimeEncoder)
	encoderRegistry.Register("unix-micros", UnixMicrosTimeEncoder)
	encoderRegistry.Register("unix-nanos", UnixNanosTimeEncoder)
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
	return encoderRegistry.FromString(name)
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
	return encoderRegistry.MustFromString(name)
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
	encoderRegistry.Register(name, encoder)
}
