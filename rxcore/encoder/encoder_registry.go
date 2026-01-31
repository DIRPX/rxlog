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

package encoder

import (
	"fmt"

	"dirpx.dev/rxlog/rxapi/encoder"
	"dirpx.dev/rxlog/rxcore/registry"
)

// Factory is a function that constructs an Encoder from the provided Config.
//
// Factory implementations MUST return a fully initialized encoder that is
// ready to use for encoding log entries. The returned encoder SHOULD respect
// all configuration options in cfg, or document which options are ignored.
//
// On failure, Factory MUST return a nil encoder and a non-nil error describing
// the problem (for example, invalid configuration, missing required options,
// or resource allocation failure).
type Factory func(cfg Config) (encoder.Encoder, error)

// encoderFactoryRegistry holds registered encoder factories keyed by symbolic names.
//
// Built-in encoder implementations will register themselves during package
// initialization. Callers SHOULD use FromString, Register, and MustFromString
// to interact with the registry rather than accessing it directly.
var encoderFactoryRegistry = registry.New[Factory]()

// FromString looks up an encoder factory by its symbolic name and creates
// an encoder using the provided configuration.
//
// The name must match one of the registered keys in the internal registry.
// Common encoder types include:
//
//   - "json"    — JSON encoder (when implemented)
//   - "console" — Console encoder (when implemented)
//   - "text"    — Text encoder (when implemented)
//
// If no encoder is registered under the given name, FromString returns a
// non-nil error and a nil Encoder.
//
// If the factory function returns an error (for example, due to invalid
// configuration), FromString propagates that error to the caller.
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
func FromString(name string, cfg Config) (encoder.Encoder, error) {
	factory, err := encoderFactoryRegistry.FromString(name)
	if err != nil {
		return nil, fmt.Errorf("unknown encoder type: %q", name)
	}
	return factory(cfg)
}

// MustFromString is a convenience helper that resolves an encoder factory by
// name, creates an encoder with the provided config, and panics if the name
// is not registered or if encoder creation fails.
//
// This is useful in static setup code (for example, wiring encoders from
// hard-coded configuration) where an unknown encoder name or creation failure
// indicates a programmer error or a misconfigured build. It SHOULD NOT be
// used for untrusted or user-provided input, where returning an error is
// preferable.
//
// The panic message is the same error produced by FromString(name, cfg) or
// the factory function.
func MustFromString(name string, cfg Config) encoder.Encoder {
	enc, err := FromString(name, cfg)
	if err != nil {
		panic(err)
	}
	return enc
}

// Register installs or overrides an encoder factory under the given symbolic
// name.
//
// If a factory is already registered under name, it will be replaced.
// Register does not perform any validation on name; callers SHOULD follow the
// same naming convention as the built-in encoders (lower-case, hyphen-separated
// identifiers) to keep configuration consistent.
//
// Register is intended to be called during process initialization (for example,
// from init functions in encoder implementation packages) to make encoder
// types available for configuration-driven selection.
//
// Example:
//
//	func init() {
//	    encoder.Register("json", NewJSONEncoder)
//	    encoder.Register("console", NewConsoleEncoder)
//	}
//
// Concurrency:
//   - Register mutates the shared registry map and is NOT safe to call
//     concurrently with other Register or FromString calls unless the caller
//     provides external synchronization.
//   - The recommended pattern is to perform all Register calls during
//     initialization (for example, in init functions), before any goroutine
//     starts using FromString or MustFromString.
func Register(name string, factory Factory) {
	encoderFactoryRegistry.Register(name, factory)
}
