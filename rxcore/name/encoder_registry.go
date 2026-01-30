/*
   Copyright 2026 The DIRPX Authors.

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

package name

import (
	"fmt"

	nameapi "dirpx.dev/rxlog/rxapi/name"
)

// registry maps symbolic encoder names to concrete name encoders.
//
// Keys are lower-case, hyphen-separated identifiers intended for configuration
// (for example, "full"). Each entry points to one of the predefined
// Encoder values above.
//
// The registry is deliberately unexported; callers SHOULD use FromString,
// Register, and MustFromString rather than accessing the map directly.
var registry = map[string]nameapi.Encoder{
	// Full format: complete logger name as-is.
	"full": FullNameEncoder,
}

// FromString looks up a name encoder by its symbolic name.
//
// The name must match one of the registered keys in the internal registry.
// Common names include:
//
//   - "full" — complete logger name as-is (e.g., "service.handler.user")
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
func FromString(name string) (nameapi.Encoder, error) {
	enc, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown name encoder: %q", name)
	}
	return enc, nil
}

// MustFromString is a convenience helper that resolves a name encoder by name
// and panics if the name is not registered.
//
// This is useful in static setup code (for example, wiring encoders from
// hard-coded configuration) where an unknown encoder name indicates a
// programmer error or a misconfigured build. It SHOULD NOT be used for
// untrusted or user-provided input, where returning an error is preferable.
//
// The panic message is the same error produced by FromString(name).
func MustFromString(name string) nameapi.Encoder {
	enc, err := FromString(name)
	if err != nil {
		panic(err)
	}
	return enc
}

// Register installs or overrides a name encoder under the given symbolic name.
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
func Register(name string, encoder nameapi.Encoder) {
	registry[name] = encoder
}
