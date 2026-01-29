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

package clictl

import "dirpx.dev/rxlog/rxapi/level"

// Option configures a Handler.
//
// Options use the functional options pattern so that callers can tailor
// Handler behavior without exposing its fields directly.
type Option func(*Handler)

// WithFlagName overrides the long flag name used to read the log level.
//
// For example, WithFlagName("verbosity") will cause Handler to recognize
// "--verbosity=info" and "--verbosity info". If name is left empty,
// DefaultFlagName is used as a fallback.
func WithFlagName(name string) Option {
	return func(h *Handler) {
		h.FlagName = name
	}
}

// WithShortFlag configures an optional one-character short flag name.
//
// For example, WithShortFlag("L") will cause Handler to recognize
// "-L=info" and "-L info". Passing an empty string disables short-flag
// support.
func WithShortFlag(short string) Option {
	return func(h *Handler) {
		h.ShortFlag = short
	}
}

// WithDefaultLevel sets a fallback level that is applied when no CLI flag
// provides a usable value (either because the flag is absent or because
// its value is effectively empty).
//
// When this option is used, HasDefault is set to true.
func WithDefaultLevel(lvl level.Level) Option {
	return func(h *Handler) {
		h.DefaultLevel = lvl
		h.HasDefault = true
	}
}

// WithRequired marks the CLI flag as required.
//
// When Required is true and no non-empty value is found for the configured
// flag(s), Apply returns an error instead of silently falling back to a
// default or leaving the threshold unchanged.
func WithRequired(required bool) Option {
	return func(h *Handler) {
		h.Required = required
	}
}

// WithArgs overrides the argument list scanned by Apply.
//
// This is primarily intended for tests or for applications that want to
// control explicitly which subset of arguments is considered. In normal
// usage, callers MAY omit this option and let Handler use os.Args[1:].
func WithArgs(args []string) Option {
	return func(h *Handler) {
		// Store a copy to avoid accidental mutation by the caller after
		// the handler has been constructed.
		if args == nil {
			h.Args = nil
			return
		}
		dst := make([]string, len(args))
		copy(dst, args)
		h.Args = dst
	}
}
