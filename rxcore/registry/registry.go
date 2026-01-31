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

package registry

import (
	"fmt"
	"reflect"
)

// Registry provides a generic type-safe registry for mapping string names
// to values of type T.
//
// Registry is designed for configuration-driven selection of encoders,
// factories, and other pluggable components. Type information is extracted
// automatically via reflection for error messages.
//
// Example usage:
//
//	var encoderRegistry = registry.New[levelapi.Encoder]()
//
//	func init() {
//	    encoderRegistry.Register("lowercase", LowercaseLevelEncoder)
//	    encoderRegistry.Register("uppercase", CapitalLevelEncoder)
//	}
//
//	encoder, err := encoderRegistry.FromString("lowercase")
//	encoder := encoderRegistry.MustFromString("uppercase")
type Registry[T any] struct {
	items    map[string]T
	typeName string
}

// New creates a new empty Registry for values of type T.
//
// The registry uses reflection to determine the type name for error messages,
// so no explicit type name parameter is required. The type name is computed
// once during initialization.
func New[T any]() *Registry[T] {
	var zero T
	t := reflect.TypeOf(zero)

	typeName := "value"
	if t != nil {
		typeName = t.String()
	}

	return &Registry[T]{
		items:    make(map[string]T),
		typeName: typeName,
	}
}

// Register adds or replaces a value in the registry under the given name.
//
// If a value is already registered under the same name, it will be replaced.
// This function does not validate the name; callers SHOULD use consistent
// naming conventions (for example, lower-case, hyphen-separated identifiers).
//
// Register is intended to be called during initialization (for example, in
// init functions) and is NOT safe for concurrent use with other Register or
// FromString calls unless the caller provides external synchronization.
func (r *Registry[T]) Register(name string, value T) {
	r.items[name] = value
}

// FromString looks up a value by its registered name.
//
// If no value is registered under the given name, FromString returns a
// zero-valued T and a non-nil error. The error message includes the type
// name extracted via reflection during registry creation.
//
// Concurrent read-only access (calling FromString after all Register calls
// have completed) is safe. If Register is called concurrently with FromString,
// the caller MUST provide external synchronization.
func (r *Registry[T]) FromString(name string) (T, error) {
	value, ok := r.items[name]
	if !ok {
		var zero T
		return zero, fmt.Errorf("unknown %s: %q", r.typeName, name)
	}
	return value, nil
}

// MustFromString is a convenience helper that looks up a value by name and
// panics if the name is not registered.
//
// This is useful in static setup code where an unknown name indicates a
// programmer error or misconfigured build. It SHOULD NOT be used for
// untrusted or user-provided input, where returning an error is preferable.
//
// The panic message is the same error produced by FromString(name).
func (r *Registry[T]) MustFromString(name string) T {
	value, err := r.FromString(name)
	if err != nil {
		panic(err)
	}
	return value
}
