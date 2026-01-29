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

import (
	"fmt"
	"os"
	"strings"

	"dirpx.dev/rxlog/rxapi/level"
)

// DefaultVarName is the default environment variable name used to
// configure the log level when no explicit VarName is provided.
//
// Applications MAY choose a different naming convention (for example,
// "APP_LOG_LEVEL") by using WithVarName.
const DefaultVarName = "RXLOG_LEVEL"

// Handler configures a mutable log level threshold from an environment
// variable.
//
// It is intended primarily for startup-time configuration, but Apply MAY be
// called multiple times (for example, after reloading process environment)
// to re-apply external settings.
//
// Typical usage:
//
//	atomic := rxclvl.NewAtomicLevel()
//	atomic.SetLevel(level.Info) // compile-time default
//
//	handler := envctl.NewHandler(&atomic,
//	    envctl.WithVarName("MYAPP_LOG_LEVEL"),
//	)
//
//	// At process initialization.
//	if err := handler.Apply(); err != nil {
//	    // Decide whether to fail fast or log-and-continue
//	}
//
// Semantics overview:
//
//   - If the configured environment variable is unset and no default level
//     is configured, Apply leaves the Threshold unchanged and returns nil.
//   - If the environment variable is unset but a default level is configured,
//     Apply sets the Threshold to that default level.
//   - If the environment variable is set to a non-empty string, Apply parses
//     it via level.Parse and updates the Threshold on success; on parse error,
//     Apply returns a non-nil error.
//   - If Required is true and the environment variable is unset (or empty
//     after trimming whitespace), Apply returns an error instead of silently
//     ignoring the missing value.
//
// Concurrency:
//
// Handler itself does not maintain internal mutable state beyond its
// configuration. It delegates all updates to the provided level.Threshold,
// which MUST be safe for concurrent use if Apply is called from multiple
// goroutines.
type Handler struct {
	// Level is the mutable log level threshold being controlled.
	//
	// This value MUST be non-nil. If it is nil at Apply time, Apply returns
	// an error.
	Level level.Threshold

	// VarName is the name of the environment variable carrying the desired
	// log level string (for example, "RXLOG_LEVEL" or "MYAPP_LOG_LEVEL").
	//
	// If empty, DefaultVarName is used.
	VarName string

	// DefaultLevel is the fallback level used when the environment variable
	// is unset or empty.
	//
	// If HasDefault is false, DefaultLevel is ignored and Apply will leave
	// the Threshold unchanged when the environment variable is missing.
	DefaultLevel level.Level

	// HasDefault indicates whether DefaultLevel is meaningful.
	//
	// This extra flag avoids treating the zero value of level.Level as an
	// implicit default.
	HasDefault bool

	// Required controls behavior when the environment variable is missing
	// or empty.
	//
	// If Required is true and the environment variable is unset or contains
	// only whitespace, Apply returns an error.
	//
	// If Required is false (the default), a missing/empty environment
	// variable is treated as "no override" unless a DefaultLevel is
	// configured.
	Required bool

	// Lookup is the function used to read environment variables.
	//
	// If nil, os.Lookup is used. This indirection exists primarily to
	// make Handler easy to test without mutating process-wide environment.
	Lookup func(string) (string, bool)
}

// NewHandler constructs a Handler bound to the provided mutable
// level threshold.
//
// The returned handler is ready for use with Apply or MustApply. Callers
// MAY further adjust its exported fields or use Option helpers before
// first use.
//
// If th is nil, NewHandler still returns a handler, but Apply will fail
// with an error until Level is set to a non-nil Threshold.
func NewHandler(th level.Threshold, opts ...Option) *Handler {
	h := &Handler{
		Level:   th,
		VarName: DefaultVarName,
		Lookup:  os.LookupEnv,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.VarName == "" {
		h.VarName = DefaultVarName
	}
	if h.Lookup == nil {
		h.Lookup = os.LookupEnv
	}

	return h
}

// Apply reads the configured environment variable, interprets its value
// as a log level, and updates the underlying Threshold accordingly.
//
// The method first verifies that the Handler has a non-nil Level. If Level
// is nil, Apply immediately returns an error, since there is no target to
// update.
//
// Next, Apply determines the effective environment variable name. If VarName
// is non-empty, it is used as-is; otherwise, DefaultVarName is used. The
// variable is then looked up using the configured Lookup function, or
// os.LookupEnv if no custom lookup is provided. The resulting value is
// trimmed of leading and trailing whitespace.
//
// If the environment variable is not set at all, or if its trimmed value is
// an empty string, the behavior depends on the Handler's configuration:
//
//   - If Required is true, Apply returns an error indicating that the
//     variable is missing.
//   - If Required is false and HasDefault is true, Apply sets the Threshold
//     to DefaultLevel and returns nil.
//   - If Required is false and HasDefault is false, Apply leaves the
//     Threshold unchanged and returns nil.
//
// If the environment variable is present and its trimmed value is non-empty,
// Apply attempts to parse it as a log level using level.Parse. If parsing
// fails, Apply returns an error describing the invalid value. If parsing
// succeeds, the parsed level is written into the underlying Threshold via
// SetLevel, and Apply returns nil.
//
// Apply is safe to call multiple times; each invocation re-evaluates the
// current environment and may adjust the Threshold accordingly.
func (h *Handler) Apply() error {
	if h.Level == nil {
		return fmt.Errorf("env handler: Level Threshold is nil")
	}

	name := h.VarName
	if name == "" {
		name = DefaultVarName
	}

	lookup := h.Lookup
	if lookup == nil {
		lookup = os.LookupEnv
	}

	raw, ok := lookup(name)
	value := strings.TrimSpace(raw)

	if !ok || value == "" {
		// environment variable is absent or effectively empty.
		if h.Required {
			return fmt.Errorf("env handler: required environment variable %q is not set", name)
		}
		if h.HasDefault {
			h.Level.SetLevel(h.DefaultLevel)
		}
		// No override; either default was applied or Threshold is left unchanged.
		return nil
	}

	parsed, err := level.Parse(value)
	if err != nil {
		return fmt.Errorf("env handler: invalid level %q for %s: %w", value, name, err)
	}

	h.Level.SetLevel(parsed)
	return nil
}

// MustApply is a convenience helper that calls Apply and panics if it
// returns a non-nil error.
//
// It is intended for use in process initialization paths where an
// invalid environment configuration should prevent the application from
// starting successfully.
//
// Example:
//
//	handler := envctl.NewHandler(&atomic)
//	handler.MustApply()
func (h *Handler) MustApply() {
	if err := h.Apply(); err != nil {
		panic(err)
	}
}
