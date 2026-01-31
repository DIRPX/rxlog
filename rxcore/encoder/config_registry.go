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

	callerpkg "dirpx.dev/rxlog/rxcore/caller"
	"dirpx.dev/rxlog/rxcore/chrono/duration"
	"dirpx.dev/rxlog/rxcore/chrono/time"
	errorpkg "dirpx.dev/rxlog/rxcore/error"
	"dirpx.dev/rxlog/rxcore/level"
	namepkg "dirpx.dev/rxlog/rxcore/name"
	stackpkg "dirpx.dev/rxlog/rxcore/stack"
)

// registry maps symbolic config preset names to concrete encoder configurations.
//
// Keys are lower-case, hyphen-separated identifiers intended for configuration
// (for example, "json", "console", "development", "production"). Each entry
// points to a preconfigured Config value.
//
// The registry is deliberately unexported; callers SHOULD use FromString,
// Register, and MustFromString rather than accessing the map directly.
var registry = map[string]Config{
	// JSON format: structured JSON with lowercase levels, RFC3339 timestamps.
	"json": {
		MessageKey:        "msg",
		LevelKey:          "level",
		TimeKey:           "time",
		CallerKey:         "caller",
		StacktraceKey:     "stacktrace",
		NameKey:           "logger",
		ErrorKey:          "error",
		OperationKey:      "function",
		EncodeLevel:       level.LowercaseLevelEncoder,
		EncodeTime:        time.RFC3339TimeEncoder,
		EncodeDuration:    duration.MillisDurationEncoder,
		EncodeCaller:      callerpkg.ShortCallerEncoder,
		EncodeName:        namepkg.FullNameEncoder,
		EncodeError:       errorpkg.SimpleErrorEncoder,
		EncodeStacktrace:  stackpkg.FullStackEncoder,
		ReflectedEncoder:  nil,
		UseJSONTags:       true,
		RespectStringer:   true,
		LineEnding:        "\n",
		FieldSeparator:    "",
		SkipLineEnding:    false,
		PrettyPrint:       false,
		IndentString:      "",
		DisableHTMLEscape: false,
		ErrorHandler:      nil,
		FormatOptions:     nil,
	},

	// Console format: human-readable with capital levels, short caller.
	"console": {
		MessageKey:        "msg",
		LevelKey:          "level",
		TimeKey:           "time",
		CallerKey:         "caller",
		StacktraceKey:     "stacktrace",
		NameKey:           "logger",
		ErrorKey:          "error",
		OperationKey:      "function",
		EncodeLevel:       level.CapitalLevelEncoder,
		EncodeTime:        time.RFC3339TimeEncoder,
		EncodeDuration:    duration.StringDurationEncoder,
		EncodeCaller:      callerpkg.ShortCallerEncoder,
		EncodeName:        namepkg.FullNameEncoder,
		EncodeError:       errorpkg.SimpleErrorEncoder,
		EncodeStacktrace:  stackpkg.FullStackEncoder,
		ReflectedEncoder:  nil,
		UseJSONTags:       true,
		RespectStringer:   true,
		LineEnding:        "\n",
		FieldSeparator:    " ",
		SkipLineEnding:    false,
		PrettyPrint:       false,
		IndentString:      "",
		DisableHTMLEscape: true,
		ErrorHandler:      nil,
		FormatOptions:     nil,
	},

	// Development format: pretty-printed JSON with full paths for debugging.
	"development": {
		MessageKey:        "msg",
		LevelKey:          "level",
		TimeKey:           "time",
		CallerKey:         "caller",
		StacktraceKey:     "stacktrace",
		NameKey:           "logger",
		ErrorKey:          "error",
		OperationKey:      "function",
		EncodeLevel:       level.CapitalLevelEncoder,
		EncodeTime:        time.RFC3339TimeEncoder,
		EncodeDuration:    duration.StringDurationEncoder,
		EncodeCaller:      callerpkg.FullCallerEncoder,
		EncodeName:        namepkg.FullNameEncoder,
		EncodeError:       errorpkg.SimpleErrorEncoder,
		EncodeStacktrace:  stackpkg.FullStackEncoder,
		ReflectedEncoder:  nil,
		UseJSONTags:       true,
		RespectStringer:   true,
		LineEnding:        "\n",
		FieldSeparator:    " ",
		SkipLineEnding:    false,
		PrettyPrint:       true,
		IndentString:      "  ",
		DisableHTMLEscape: true,
		ErrorHandler:      nil,
		FormatOptions:     nil,
	},

	// Production format: compact JSON with minimal overhead.
	"production": {
		MessageKey:        "msg",
		LevelKey:          "level",
		TimeKey:           "time",
		CallerKey:         "caller",
		StacktraceKey:     "stacktrace",
		NameKey:           "logger",
		ErrorKey:          "error",
		OperationKey:      "function",
		EncodeLevel:       level.LowercaseLevelEncoder,
		EncodeTime:        time.ISO8601MillisTimeEncoder,
		EncodeDuration:    duration.MillisDurationEncoder,
		EncodeCaller:      callerpkg.ShortCallerEncoder,
		EncodeName:        namepkg.FullNameEncoder,
		EncodeError:       errorpkg.SimpleErrorEncoder,
		EncodeStacktrace:  stackpkg.FullStackEncoder,
		ReflectedEncoder:  nil,
		UseJSONTags:       true,
		RespectStringer:   true,
		LineEnding:        "\n",
		FieldSeparator:    "",
		SkipLineEnding:    false,
		PrettyPrint:       false,
		IndentString:      "",
		DisableHTMLEscape: false,
		ErrorHandler:      nil,
		FormatOptions:     nil,
	},
}

// FromString looks up an encoder configuration preset by its symbolic name.
//
// The name must match one of the registered keys in the internal registry.
// Built-in presets include:
//
//   - "json"        — structured JSON with lowercase levels, RFC3339 timestamps
//   - "console"     — human-readable with capital levels, short caller
//   - "development" — pretty-printed JSON with full paths for debugging
//   - "production"  — compact JSON with minimal overhead
//
// If no config is registered under the given name, FromString returns a
// non-nil error and a zero-valued Config.
//
// This function is intended for configuration-driven setups (for example,
// parsing encoder preset names from JSON/YAML/TOML). Callers SHOULD normalize
// user-provided values to lower-case before calling FromString.
//
// Concurrency:
//   - Concurrent read-only access (calling FromString after all Register
//     calls have completed) is safe.
//   - If Register is called concurrently with FromString, the caller MUST
//     provide external synchronization around registry mutations.
func FromString(name string) (Config, error) {
	cfg, ok := registry[name]
	if !ok {
		return Config{}, fmt.Errorf("unknown encoder config preset: %q", name)
	}
	return cfg, nil
}

// MustFromString is a convenience helper that resolves an encoder config preset
// by name and panics if the name is not registered.
//
// This is useful in static setup code (for example, wiring encoders from
// hard-coded configuration) where an unknown preset name indicates a
// programmer error or a misconfigured build. It SHOULD NOT be used for
// untrusted or user-provided input, where returning an error is preferable.
//
// The panic message is the same error produced by FromString(name).
func MustFromString(name string) Config {
	cfg, err := FromString(name)
	if err != nil {
		panic(err)
	}
	return cfg
}

// Register installs or overrides an encoder config preset under the given
// symbolic name.
//
// If a config is already registered under name, it will be replaced.
// Register does not perform any validation on name; callers SHOULD follow the
// same naming convention as the built-in presets (lower-case, hyphen-separated
// identifiers) to keep configuration consistent.
//
// Register is intended to be called during process initialization (for example,
// from init functions) to extend the set of available config presets with
// application-specific formats.
//
// Concurrency:
//   - Register mutates the shared registry map and is NOT safe to call
//     concurrently with other Register or FromString calls unless the caller
//     provides external synchronization.
//   - The recommended pattern is to perform all Register calls during
//     initialization, before any goroutine starts using FromString or
//     MustFromString.
func Register(name string, config Config) {
	registry[name] = config
}
