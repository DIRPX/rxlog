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

// Package clictl provides CLI-based control over the logging level threshold.
//
// It is intended for scenarios where the minimum enabled log level of a
// running process should be configurable via command-line flags, without
// requiring configuration files or environment variables.
//
// # Overview
//
// The core type in this package is Handler. It binds a mutable log level
// threshold (level.Threshold) to one or more command-line flags. When
// Apply is called, Handler scans a chosen argument list (typically os.Args)
// for the configured flag(s), interprets the associated value as a log level
// name (for example, "debug", "info", "warn", "error"), and updates the
// underlying threshold accordingly.
//
// Typical usage:
//
//	import (
//	    "log"
//
//	    "dirpx.dev/rxlog/rxapi/level"
//	    rxclvl "dirpx.dev/rxlog/rxcore/level"
//	    clictl "dirpx.dev/rxlog/levelctl/cli"
//	)
//
//	func main() {
//	    // Create a mutable threshold (for example, an AtomicLevel).
//	    atomic := rxclvl.NewAtomicLevel()
//	    atomic.SetLevel(level.Info) // compile-time default
//
//	    // Bind it to CLI flags such as --log-level / -L.
//	    h := clictl.NewHandler(&atomic,
//	        clictl.WithFlagName("log-level"), // --log-level
//	        clictl.WithShortFlag("L"),         // -L
//	        clictl.WithDefaultLevel(level.Info),
//	    )
//
//	    // Apply configuration once arguments are available.
//	    if err := h.Apply(); err != nil {
//	        log.Printf("failed to apply log level from CLI: %v", err)
//	    }
//
//	    // ... construct loggers from atomic, run application ...
//	}
//
// This pattern allows operators and tooling to influence the process-wide
// log level by passing flags at process startup.
//
// # Flag forms and semantics
//
// Handler uses a dedicated flag set (based on spf13/pflag) to parse only the
// configured level-related flag(s). It does not modify the global flag set
// and is designed to coexist with the application's own CLI parser.
//
// By default, Handler looks for a long flag named:
//
//	--log-level
//
// Callers MAY override this name via WithFlagName. In addition, a one-letter
// short form may be configured via WithShortFlag; for example, WithShortFlag("L")
// will recognize:
//
//	-L=debug
//	-L debug
//
// For the long flag, Handler recognizes:
//
//	--log-level=info
//	--log-level info
//
// For the short flag (if present), it recognizes:
//
//	-L=info
//	-L info
//
// If the same flag appears multiple times, the last occurrence wins, consistent
// with common CLI parsing conventions.
//
// Internally, the flag is declared with an empty string as its default value.
// After parsing, Handler can reliably distinguish between "no flag at all"
// and "flag explicitly provided with a value".
//
// # Missing and empty values
//
// After parsing the argument list, Handler inspects the resulting value:
//
//   - If the flag was never provided at all, or if the final value is empty
//     or consists only of whitespace:
//
//     If Required is true, Apply returns an error indicating that the
//     CLI flag is missing or empty.
//
//     If Required is false and HasDefault is true, Apply sets the
//     underlying threshold to DefaultLevel and returns nil.
//
//     If Required is false and HasDefault is false, Apply leaves the
//     threshold unchanged and returns nil.
//
//   - If a non-empty value is present:
//
//     The value is trimmed of leading and trailing whitespace.
//
//     Handler attempts to parse it as a log level using level.Parse.
//
//     If parsing fails, Apply returns an error and does not modify the
//     threshold.
//
//     If parsing succeeds, the parsed level is written into the underlying
//     threshold via SetLevel, and Apply returns nil.
//
// These rules allow the same handler to be used both for optional CLI overrides
// (where the flag is best-effort) and for mandatory CLI configuration (where
// absence or emptiness should be treated as an error).
//
// # Argument source
//
// Handler determines which arguments to inspect as follows:
//
//   - If Args is non-nil, it uses that slice as-is. Callers can supply a
//     custom argument list via WithArgs, which is especially useful for
//     tests or when embedding Handler into larger CLI frameworks.
//
//   - If Args is nil, Handler falls back to os.Args[1:], that is, the process'
//     arguments excluding the program name.
//
// The argument list is not modified in place; WithArgs stores a defensive
// copy of the provided slice to avoid accidental aliasing.
//
// # Unknown flags and coexistence
//
// Handler constructs its own pflag.FlagSet with the following properties:
//
//   - ParseErrorsWhitelist.UnknownFlags is set so that unknown flags are
//     ignored rather than treated as fatal errors.
//   - Error output from the internal FlagSet is suppressed; any parse errors
//     are surface through the error returned by Apply.
//
// This design ensures that Handler can be used in applications that already
// use their own CLI parsing logic: the handler only cares about its own flag,
// and unknown flags pass through without conflict.
//
// # Configuration and options
//
// Handlers are constructed via NewHandler and customized via functional
// options:
//
//	h := clictl.NewHandler(&atomic,
//	    clictl.WithFlagName("verbosity"),
//	    clictl.WithShortFlag("v"),
//	    clictl.WithDefaultLevel(level.Info),
//	    clictl.WithRequired(true),
//	    clictl.WithArgs([]string{"--verbosity=debug"}), // optional
//	)
//
// The available options are:
//
//   - WithFlagName(name string)
//     Overrides the long flag name. If name is left empty, DefaultFlagName
//     is used as a fallback.
//
//   - WithShortFlag(short string)
//     Configures a one-character short flag (for example, "L" for "-L").
//     Passing an empty string disables short-flag support.
//
//   - WithDefaultLevel(lvl level.Level)
//     Configures a fallback level that is applied when the flag is missing
//     or empty and Required is false. When this option is used, HasDefault
//     is set to true.
//
//   - WithRequired(required bool)
//     Marks the CLI flag as required. When Required is true, a missing or
//     empty flag causes Apply to return an error instead of silently
//     falling back to a default or leaving the threshold unchanged.
//
//   - WithArgs(args []string)
//     Overrides the argument list to be parsed. Handler stores a copy of
//     the provided slice. Passing nil resets Args to nil, causing Apply to
//     use os.Args[1:].
//
// All configuration fields of Handler are initialized by NewHandler and MAY
// be further adjusted before the first call to Apply or MustApply.
//
// # Concurrency and lifecycle
//
// Handler itself does not maintain mutable internal state beyond its
// configuration and the reference to Level. It delegates all updates to the
// provided level.Threshold, which MUST be safe for concurrent use if Apply
// is called from multiple goroutines.
//
// Typical usage patterns include:
//
//   - Applying configuration once at process startup (most common).
//   - Re-applying configuration in a specialized setup phase that inspects
//     CLI arguments before constructing loggers.
//
// Handler does not implement any scheduling or re-parsing logic by itself;
// such concerns are left to the embedding application.
//
// # Error handling and MustApply
//
// For initialization paths where an invalid CLI-based configuration should
// prevent the application from starting successfully, Handler exposes
// MustApply, which simply calls Apply and panics on any non-nil error:
//
//	h := clictl.NewHandler(&atomic)
//	h.MustApply() // panic if the flag is missing/invalid according to config
//
// This can be convenient for small services or tests where failing fast is
// preferable to continuing with an unexpected logging configuration.
//
// Larger systems may prefer to call Apply explicitly, log any errors, and
// continue with a previously configured or compile-time default level.
//
// # Relationship to other control surfaces
//
// The clictl package is complementary to other control mechanisms such as
// envctl (environment-based) and filectl (file-based). All three packages
// operate on the same level.Threshold abstraction and can be combined if
// desired, for example by applying a compile-time default, then an optional
// file override, then an environment override, and finally a CLI override.
package clictl
