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

package envctl

import "dirpx.dev/rxlog/rxapi/level"

// Option configures an Handler.
//
// Options use the functional options pattern so that callers can tailor
// Handler behavior without exposing its fields directly.
type Option func(*Handler)

// WithVarName overrides the default environment variable name used
// to read the log level.
//
// Passing an empty name is allowed but discouraged; if VarName is empty,
// Apply falls back to DefaultVarName.
func WithVarName(name string) Option {
	return func(h *Handler) {
		h.VarName = name
	}
}

// WithDefaultLevel sets a fallback level that is applied when the
// environment variable is missing or empty.
//
// When this option is used, Apply will set the Threshold to the provided
// default level if the environment variable does not provide a value.
//
// The provided level is not validated here; callers SHOULD pass only
// valid severities.
func WithDefaultLevel(lvl level.Level) Option {
	return func(h *Handler) {
		h.DefaultLevel = lvl
		h.HasDefault = true
	}
}

// WithRequired marks the environment variable as required.
//
// When Required is true and the environment variable is unset or contains
// only whitespace, Apply returns an error instead of silently falling back
// to a default or leaving the Threshold unchanged.
func WithRequired(required bool) Option {
	return func(h *Handler) {
		h.Required = required
	}
}

// WithLookup overrides the function used to read environment variables.
//
// This is primarily intended for tests or embedding within larger
// configuration systems. In normal usage, callers SHOULD rely on the
// default behavior, which uses os.Lookup.
func WithLookup(fn func(string) (string, bool)) Option {
	return func(h *Handler) {
		h.Lookup = fn
	}
}
