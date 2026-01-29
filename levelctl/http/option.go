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

package httpctl

// Option configures an Handler.
//
// Options use the functional options pattern so that callers can tailor
// handler behavior without exporting the struct fields directly.
type Option func(*Handler)

// WithParamName overrides the default parameter name ("level") used when
// reading the desired level from query parameters and JSON bodies.
//
// Passing an empty name is allowed but discouraged; if ParamName is empty,
// only plain-text bodies will be used as a source for mutation requests.
func WithParamName(name string) Option {
	return func(h *Handler) {
		h.ParamName = name
	}
}

// WithMaxBodyBytes overrides the default body size limit.
//
// If n <= 0, the handler falls back to its internal default (4096 bytes).
func WithMaxBodyBytes(n int64) Option {
	return func(h *Handler) {
		h.MaxBodyBytes = n
	}
}

// WithReadOnly forbids runtime level changes via this handler.
//
// When enabled, GET/HEAD remain functional, but POST/PUT/PATCH will
// respond with HTTP 405 Method Not Allowed.
func WithReadOnly(readOnly bool) Option {
	return func(h *Handler) {
		h.ReadOnly = readOnly
	}
}
