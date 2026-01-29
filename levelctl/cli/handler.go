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

import (
	"fmt"
	"io"
	"os"
	"strings"

	"dirpx.dev/rxlog/rxapi/level"
	"github.com/spf13/pflag"
)

// DefaultFlagName is the fallback CLI flag name used when Handler.FlagName
// is empty.
//
// By default, the handler looks for a flag named "--log-level". Applications
// MAY override this via WithFlagName to align with their own flag naming
// conventions.
const DefaultFlagName = "log-level"

// Handler configures a mutable log level threshold from command-line flags.
//
// It is intended for scenarios where the minimum enabled log level of a
// process should be configurable via CLI arguments (for example,
// "--log-level=debug"), without requiring additional configuration
// mechanisms.
//
// Semantics overview:
//
//   - The handler scans the argument list for a specific long flag
//     ("--<FlagName>") and, optionally, a short flag ("-<ShortFlag>").
//   - It supports both "--flag=value" and "--flag value" forms, as well as
//     "-f=value" and "-f value" for the short variant.
//   - When a value is found, it is trimmed and parsed via level.Parse; on
//     success, the resulting level is written into the underlying threshold.
//
// Handler does not register flags with the standard library's flag package;
// instead, it performs a minimal, self-contained scan of the argument list
// provided to it (or os.Args[1:] by default).
type Handler struct {
	// Level is the mutable log level threshold being controlled.
	//
	// This value MUST be non-nil. If it is nil at Apply time, Apply returns
	// an error.
	Level level.Threshold

	// FlagName is the long flag name used to carry the log level value.
	//
	// For example, if FlagName is "log-level", the handler will look for:
	//
	//   --log-level=info
	//   --log-level info
	//
	// If empty, DefaultFlagName is used.
	FlagName string

	// ShortFlag is the optional one-character short flag name.
	//
	// For example, if ShortFlag is "L", the handler will additionally look
	// for:
	//
	//   -L=info
	//   -L info
	//
	// If empty, no short flag is recognized.
	ShortFlag string

	// DefaultLevel is the fallback level used when no CLI flag provides
	// a value or the value is effectively empty.
	//
	// If HasDefault is false, DefaultLevel is ignored and Apply will leave
	// the Threshold unchanged when no flag provides a usable value (unless
	// Required is true).
	DefaultLevel level.Level

	// HasDefault indicates whether DefaultLevel is meaningful.
	//
	// This extra flag avoids treating the zero value of level.Level as an
	// implicit default.
	HasDefault bool

	// Required controls behavior when the CLI flag is absent or provides
	// an empty value.
	//
	// If Required is true and no non-empty value is found, Apply returns
	// an error.
	//
	// If Required is false (the default), a missing/empty flag is treated
	// as "no override" unless a DefaultLevel is configured.
	Required bool

	// Args is the argument list scanned by Apply.
	//
	// If nil, Apply uses os.Args[1:] (the usual "arguments after the
	// program name"). If non-nil, Apply scans exactly this slice and
	// ignores os.Args.
	Args []string
}

// NewHandler constructs a Handler bound to the provided mutable level
// threshold.
//
// The returned handler is ready for use with Apply or MustApply. Callers
// MAY further adjust its exported fields or use Option helpers before
// first use.
//
// Example:
//
//	atomic := rxclvl.NewAtomicLevel()
//	atomic.SetLevel(level.Info) // compile-time default
//
//	h := clictl.NewHandler(&atomic,
//	    clictl.WithFlagName("log-level"),
//	    clictl.WithShortFlag("L"),
//	    clictl.WithDefaultLevel(level.Info),
//	)
//
//	if err := h.Apply(); err != nil {
//	    // Handle CLI configuration error
//	}
//
// If th is nil, NewHandler still returns a handler, but Apply will fail
// with an error until Level is set to a non-nil Threshold.
func NewHandler(th level.Threshold, opts ...Option) *Handler {
	h := &Handler{
		Level:    th,
		FlagName: DefaultFlagName,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.FlagName == "" {
		h.FlagName = DefaultFlagName
	}

	return h
}

// Apply scans the argument list using a pflag.FlagSet, interprets the
// configured CLI flag(s) as a log level, and updates the underlying Threshold.
//
// The method first verifies that the Handler has a non-nil Level. If Level
// is nil, Apply immediately returns an error, because there is no target to
// update.
//
// Next, Apply determines which arguments to inspect. If Args is non-nil, it
// uses that slice as-is. Otherwise, it falls back to os.Args[1:], that is,
// the process arguments excluding the program name.
//
// The handler then builds a dedicated pflag.FlagSet configured with
// ContinueOnError error handling and suppressed output. On this FlagSet it
// registers a string flag corresponding to FlagName (and optionally a short
// flag corresponding to ShortFlag). The flag is declared with an empty string
// as its default value so that a non-empty value after parsing reliably
// indicates that the user has provided an override.
//
// To avoid interfering with other application flags, the FlagSet is configured
// with ParseErrorsWhitelist.UnknownFlags = true, so that unknown flags are
// silently ignored rather than treated as fatal errors.
//
// After parsing, Apply examines whether the level flag was explicitly set and
// how its value should be interpreted:
//
//   - If no flag was found at all, or if the most recent occurrence produced
//     an empty value after trimming whitespace, the behavior depends on the
//     configuration:
//
//     If Required is true, Apply returns an error indicating that the
//     CLI flag is missing or empty.
//
//     If Required is false and HasDefault is true,
//     Apply sets the Threshold to DefaultLevel and returns nil.
//
//     If Required is false and HasDefault is false, Apply
//     leaves the Threshold unchanged and returns nil.
//
//   - If a non-empty value is found, Apply trims leading and trailing
//     whitespace and attempts to parse the result as a log level using
//     level.Parse.
//
//     If parsing fails, Apply returns an error and does not
//     modify the Threshold.
//
//     If parsing succeeds, the parsed level is written
//     into the underlying Threshold via SetLevel, and Apply returns nil.
//
// Apply is safe to call multiple times; each invocation re-scans the current
// argument list and may adjust the Threshold in response to updated flags.
func (h *Handler) Apply() error {
	if h.Level == nil {
		return fmt.Errorf("cli handler: Level Threshold is nil")
	}

	args := h.Args
	if args == nil {
		args = os.Args[1:]
	}

	flagName := h.FlagName
	if flagName == "" {
		flagName = DefaultFlagName
	}

	// Prepare a dedicated FlagSet to parse only the level-related flags.
	fs := pflag.NewFlagSet("rxlog-cli-level", pflag.ContinueOnError)

	// Suppress pflag’s default error/usage output; the handler reports all
	// errors via the returned error value instead.
	fs.SetOutput(io.Discard)

	// Allow unknown flags so that we do not interfere with the application’s
	// own flag parsing. We only care about the configured flag(s) here.
	fs.ParseErrorsAllowlist.UnknownFlags = true

	var value string

	if h.ShortFlag != "" {
		// Long + short form, e.g. --log-level / -L.
		fs.StringVarP(&value, flagName, h.ShortFlag, "", "logging level threshold")
	} else {
		// Only long form, e.g. --log-level.
		fs.StringVar(&value, flagName, "", "logging level threshold")
	}

	if err := fs.Parse(args); err != nil {
		// Any parse error (other than unknown flags, which are whitelisted)
		// is treated as a configuration error.
		return fmt.Errorf("cli handler: failed to parse flags: %w", err)
	}

	// If the flag was never provided, Changed will be false even if the
	// default value happens to be non-empty. We explicitly rely on this to
	// distinguish "no override" from "explicit value".
	if !fs.Changed(flagName) {
		if h.Required {
			return fmt.Errorf("cli handler: required flag %s not provided", h.flagSynopsis())
		}
		if h.HasDefault {
			h.Level.SetLevel(h.DefaultLevel)
		}
		return nil
	}

	value = strings.TrimSpace(value)
	if value == "" {
		if h.Required {
			return fmt.Errorf("cli handler: required flag %s is empty", h.flagSynopsis())
		}
		if h.HasDefault {
			h.Level.SetLevel(h.DefaultLevel)
		}
		return nil
	}

	parsed, err := level.Parse(value)
	if err != nil {
		return fmt.Errorf("cli handler: invalid level %q from %s: %w", value, h.flagSynopsis(), err)
	}

	h.Level.SetLevel(parsed)
	return nil
}

// MustApply is a convenience helper that calls Apply and panics if it
// returns a non-nil error.
//
// It is intended for use in process initialization paths where an
// invalid CLI-based configuration should prevent the application from
// starting successfully.
func (h *Handler) MustApply() {
	if err := h.Apply(); err != nil {
		panic(err)
	}
}

// flagSynopsis returns a human-readable representation of the configured
// flag(s) for use in error messages.
//
// For example:
//
//	"--log-level / -L"
//	"--verbosity"
//	"-L"
func (h *Handler) flagSynopsis() string {
	longName := ""
	if h.FlagName != "" {
		longName = "--" + h.FlagName
	}
	shortName := ""
	if h.ShortFlag != "" {
		shortName = "-" + h.ShortFlag
	}

	switch {
	case longName != "" && shortName != "":
		return longName + " / " + shortName
	case longName != "":
		return longName
	case shortName != "":
		return shortName
	default:
		// Fallback for misconfiguration; should not normally occur.
		return "--" + DefaultFlagName
	}
}
