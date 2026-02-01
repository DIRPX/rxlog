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
	callerpkg "dirpx.dev/rxlog/rxcore/caller"
	"dirpx.dev/rxlog/rxcore/chrono/duration"
	"dirpx.dev/rxlog/rxcore/chrono/time"
	errorpkg "dirpx.dev/rxlog/rxcore/error"
	"dirpx.dev/rxlog/rxcore/field/fields"
	"dirpx.dev/rxlog/rxcore/level"
	namepkg "dirpx.dev/rxlog/rxcore/name"
	"dirpx.dev/rxlog/rxcore/registry"
	stackpkg "dirpx.dev/rxlog/rxcore/stack"
)

// encoderConfigRegistry holds registered encoder config presets keyed by symbolic names.
//
// Built-in presets are registered during package initialization. Callers
// SHOULD use ConfigFromString, RegisterConfig, and MustConfigFromString to
// interact with the registry rather than accessing it directly.
var encoderConfigRegistry = registry.New[Config]()

func init() {
	// Register built-in encoder config presets.

	// JSON format: structured JSON with lowercase levels, RFC3339 timestamps.
	encoderConfigRegistry.Register("json", Config{
		MessageKey:        fields.Message,
		LevelKey:          fields.Level,
		TimeKey:           fields.Timestamp,
		CallerKey:         fields.Caller,
		StacktraceKey:     fields.Stacktrace,
		NameKey:           fields.Logger,
		ErrorKey:          fields.Error,
		OperationKey:      fields.Function,
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
	})

	// Console format: human-readable with capital levels, short caller.
	encoderConfigRegistry.Register("console", Config{
		MessageKey:        fields.Message,
		LevelKey:          fields.Level,
		TimeKey:           fields.Timestamp,
		CallerKey:         fields.Caller,
		StacktraceKey:     fields.Stacktrace,
		NameKey:           fields.Logger,
		ErrorKey:          fields.Error,
		OperationKey:      fields.Function,
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
	})

	// Development format: pretty-printed JSON with full paths for debugging.
	encoderConfigRegistry.Register("development", Config{
		MessageKey:        fields.Message,
		LevelKey:          fields.Level,
		TimeKey:           fields.Timestamp,
		CallerKey:         fields.Caller,
		StacktraceKey:     fields.Stacktrace,
		NameKey:           fields.Logger,
		ErrorKey:          fields.Error,
		OperationKey:      fields.Function,
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
	})

	// Production format: compact JSON with minimal overhead.
	encoderConfigRegistry.Register("production", Config{
		MessageKey:        fields.Message,
		LevelKey:          fields.Level,
		TimeKey:           fields.Timestamp,
		CallerKey:         fields.Caller,
		StacktraceKey:     fields.Stacktrace,
		NameKey:           fields.Logger,
		ErrorKey:          fields.Error,
		OperationKey:      fields.Function,
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
	})
}

// ConfigFromString looks up an encoder configuration preset by its symbolic name.
//
// The name must match one of the registered keys in the internal registry.
// Built-in presets include:
//
//   - "json"        — structured JSON with lowercase levels, RFC3339 timestamps
//   - "console"     — human-readable with capital levels, short caller
//   - "development" — pretty-printed JSON with full paths for debugging
//   - "production"  — compact JSON with minimal overhead
//
// If no config is registered under the given name, ConfigFromString returns a
// non-nil error and a zero-valued Config.
//
// This function is intended for configuration-driven setups (for example,
// parsing encoder preset names from JSON/YAML/TOML). Callers SHOULD normalize
// user-provided values to lower-case before calling ConfigFromString.
//
// Concurrency:
//   - Concurrent read-only access (calling ConfigFromString after all RegisterConfig
//     calls have completed) is safe.
//   - If RegisterConfig is called concurrently with ConfigFromString, the caller MUST
//     provide external synchronization around registry mutations.
func ConfigFromString(name string) (Config, error) {
	return encoderConfigRegistry.FromString(name)
}

// MustConfigFromString is a convenience helper that resolves an encoder config preset
// by name and panics if the name is not registered.
//
// This is useful in static setup code (for example, wiring encoders from
// hard-coded configuration) where an unknown preset name indicates a
// programmer error or a misconfigured build. It SHOULD NOT be used for
// untrusted or user-provided input, where returning an error is preferable.
//
// The panic message is the same error produced by ConfigFromString(name).
func MustConfigFromString(name string) Config {
	return encoderConfigRegistry.MustFromString(name)
}

// RegisterConfig installs or overrides an encoder config preset under the given
// symbolic name.
//
// If a config is already registered under name, it will be replaced.
// RegisterConfig does not perform any validation on name; callers SHOULD follow the
// same naming convention as the built-in presets (lower-case, hyphen-separated
// identifiers) to keep configuration consistent.
//
// RegisterConfig is intended to be called during process initialization (for example,
// from init functions) to extend the set of available config presets with
// application-specific formats.
//
// Concurrency:
//   - RegisterConfig mutates the shared registry map and is NOT safe to call
//     concurrently with other RegisterConfig or ConfigFromString calls unless the caller
//     provides external synchronization.
//   - The recommended pattern is to perform all RegisterConfig calls during
//     initialization, before any goroutine starts using ConfigFromString or
//     MustConfigFromString.
func RegisterConfig(name string, config Config) {
	encoderConfigRegistry.Register(name, config)
}
