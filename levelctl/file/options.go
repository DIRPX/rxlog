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

package filectl

import "dirpx.dev/rxlog/rxapi/level"

// Option configures a Handler.
//
// Options use the functional options pattern so that callers can tailor
// Handler behavior without exposing its fields directly.
type Option func(*Handler)

// WithPath overrides the file path used to read the log level.
//
// Passing an empty path is allowed but discouraged; if Path is empty at
// Apply time, DefaultPath will be used.
func WithPath(path string) Option {
	return func(h *Handler) {
		h.Path = path
	}
}

// WithDefaultLevel sets a fallback level that is applied when the
// configuration file is missing or effectively empty.
//
// When this option is used, Apply will set the Threshold to the provided
// default level if the file does not provide a usable value.
//
// The provided level is not validated here; callers SHOULD pass only
// valid severities (IsValid() == true).
func WithDefaultLevel(lvl level.Level) Option {
	return func(h *Handler) {
		h.DefaultLevel = lvl
		h.HasDefault = true
	}
}

// WithRequired marks the configuration file as required.
//
// When Required is true and the file is missing, unreadable, or contains
// only whitespace, Apply returns an error instead of silently falling back
// to a default or leaving the Threshold unchanged.
func WithRequired(required bool) Option {
	return func(h *Handler) {
		h.Required = required
	}
}

// WithReadFile overrides the function used to read the configuration file.
//
// This is primarily intended for tests or embedding within larger
// configuration systems. In normal usage, callers SHOULD rely on the
// default behavior, which uses os.ReadFile.
func WithReadFile(fn ReadFileFunc) Option {
	return func(h *Handler) {
		h.ReadFile = fn
	}
}
