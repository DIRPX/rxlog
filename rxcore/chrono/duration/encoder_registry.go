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

package duration

import (
	durationapi "dirpx.dev/rxlog/rxapi/chrono/duration"
	"dirpx.dev/rxlog/rxcore/registry"
)

// encoderRegistry holds registered duration encoders keyed by symbolic names.
//
// Built-in encoders are registered during package initialization. Callers
// SHOULD use FromString, Register, and MustFromString to interact with the
// registry rather than accessing it directly.
var encoderRegistry = registry.New[durationapi.Encoder]()

func init() {
	// Register built-in duration encoders with their aliases.

	// String representation (Go's native duration format).
	encoderRegistry.Register("string", StringDurationEncoder)

	// Floating-point seconds.
	encoderRegistry.Register("seconds", SecondsDurationEncoder)
	encoderRegistry.Register("secs", SecondsDurationEncoder)
	encoderRegistry.Register("seconds-f64", SecondsDurationEncoder)
	encoderRegistry.Register("secs-float64", SecondsDurationEncoder)

	// Integer milliseconds.
	encoderRegistry.Register("millis", MillisDurationEncoder)
	encoderRegistry.Register("milliseconds", MillisDurationEncoder)
	encoderRegistry.Register("ms", MillisDurationEncoder)
	encoderRegistry.Register("millis-int64", MillisDurationEncoder)
	encoderRegistry.Register("ms-int64", MillisDurationEncoder)

	// Integer microseconds.
	encoderRegistry.Register("micros", MicrosDurationEncoder)
	encoderRegistry.Register("microseconds", MicrosDurationEncoder)
	encoderRegistry.Register("µs", MicrosDurationEncoder) // for completeness; use with care in configs
	encoderRegistry.Register("micros-int64", MicrosDurationEncoder)
	encoderRegistry.Register("microsecs-int64", MicrosDurationEncoder)

	// Integer nanoseconds.
	encoderRegistry.Register("nanos", NanosDurationEncoder)
	encoderRegistry.Register("nanoseconds", NanosDurationEncoder)
	encoderRegistry.Register("ns", NanosDurationEncoder)
	encoderRegistry.Register("nanos-int64", NanosDurationEncoder)
	encoderRegistry.Register("nanosecs-int64", NanosDurationEncoder)
}

// FromString looks up a duration encoder by its symbolic name.
//
// The name must match one of the registered keys in the internal registry.
// Common names include:
//
//   - "string"           — time.Duration.String() (e.g. "150ms")
//   - "seconds" / "secs" — floating-point seconds (float64)
//   - "millis" / "ms"    — integer milliseconds (int64)
//   - "micros" / "µs"    — integer microseconds (int64)
//   - "nanos" / "ns"     — integer nanoseconds (int64)
//
// If no encoder is registered under the given name, FromString returns a
// non-nil error and a nil Encoder.
//
// This function is intended for configuration-driven setups (for example,
// parsing encoder names from JSON/YAML/TOML). Callers SHOULD normalize
// user-provided values to lower-case before calling FromString.
//
// Concurrency:
//   - Concurrent read-only access (calling FromString after all Register
//     calls have completed) is safe.
//   - If Register is called concurrently with FromString, the caller MUST
//     provide external synchronization around registry mutations.
func FromString(name string) (durationapi.Encoder, error) {
	return encoderRegistry.FromString(name)
}

// MustFromString is a convenience helper that resolves a duration encoder by
// name and panics if the name is not registered.
//
// This is useful in static setup code (for example, wiring encoders from
// hard-coded configuration) where an unknown encoder name indicates a
// programmer error or a misconfigured build. It SHOULD NOT be used for
// untrusted or user-provided input, where returning an error is preferable.
//
// The panic message is the same error produced by FromString(name).
func MustFromString(name string) durationapi.Encoder {
	return encoderRegistry.MustFromString(name)
}

// Register installs or overrides a duration encoder under the given symbolic name.
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
//     concurrently with other Register or FromString calls unless the caller
//     provides external synchronization.
//   - The recommended pattern is to perform all Register calls during
//     initialization, before any goroutine starts using FromString or
//     MustFromString.
func Register(name string, encoder durationapi.Encoder) {
	encoderRegistry.Register(name, encoder)
}
