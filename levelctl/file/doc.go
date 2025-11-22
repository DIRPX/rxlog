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

// Package filectl provides file-based control over the logging level threshold.
//
// It is intended for scenarios where the minimum enabled log level of a
// running process should be configurable via a simple text file, without
// requiring code changes or recompilation.
//
// # Overview
//
// The core type in this package is Handler. It binds a mutable log level
// threshold (level.Threshold) to a configuration file on disk. When Apply
// is called, Handler reads the configured file, interprets its contents as
// a log level name (for example, "info", "warn", "error"), and updates the
// underlying threshold accordingly.
//
// Typical usage:
//
//	import (
//	    "log"
//
//	    "dirpx.dev/rxlog/rxapi/level"
//	    rxclvl "dirpx.dev/rxlog/rxcore/level"
//	    filectl "dirpx.dev/rxlog/levelctl/file"
//	)
//
//	func main() {
//	    // Create a mutable threshold (for example, an AtomicLevel).
//	    atomic := rxclvl.NewAtomicLevel()
//	    atomic.SetLevel(level.Info) // compile-time default
//
//	    // Bind it to a file-based configuration source.
//	    h := filectl.NewHandler(&atomic, "/etc/myapp/log.level",
//	        filectl.WithDefaultLevel(level.Info), // optional fallback
//	    )
//
//	    // Apply configuration at startup.
//	    if err := h.Apply(); err != nil {
//	        // Decide whether to fail fast or log-and-continue.
//	        log.Printf("failed to apply log level from file: %v", err)
//	    }
//
//	    // ... construct loggers from atomic, run application ...
//	}
//
// This pattern allows operators to influence the process-wide log level by
// editing a single file and either restarting the process or explicitly
// re-calling Handler.Apply at appropriate times.
//
// # File format and semantics
//
// Handler expects the configuration file to contain a single log level name,
// optionally surrounded by whitespace. For example:
//
//	info
//	warn
//	error
//
// Leading and trailing whitespace (including newlines) are ignored. The
// remaining string is passed to level.Parse for interpretation. If parsing
// succeeds, the resulting level is written into the underlying threshold via
// SetLevel. If parsing fails, Apply returns an error describing the invalid
// value and does not modify the threshold.
//
// The file is treated as a simple, line-oriented configuration mechanism on
// purpose: it is easy to understand, easy to edit manually, and does not
// depend on any specific structured configuration format.
//
// # Missing and empty files
//
// Handler distinguishes between several cases when reading the configuration:
//
//   - If the file does not exist at all (for example, because it has not been
//     created yet or the path is incorrect):
//
//   - If Required is true, Apply returns an error indicating that the file
//     is missing.
//
//   - If Required is false and HasDefault is true, Apply sets the
//     underlying threshold to DefaultLevel and returns nil.
//
//   - If Required is false and HasDefault is false, Apply leaves the
//     threshold unchanged and returns nil.
//
//   - If the file exists but cannot be read for reasons other than "not
//     found" (such as permission problems or I/O errors), Apply treats this
//     as a hard failure and returns an error regardless of Required. In
//     this case, the configuration source is present but unusable.
//
//   - If the file is read successfully but its contents are effectively
//     empty after trimming whitespace:
//
//   - If Required is true, Apply returns an error indicating that the
//     file does not contain a usable value.
//
//   - If Required is false and HasDefault is true, Apply sets the
//     threshold to DefaultLevel and returns nil.
//
//   - If Required is false and HasDefault is false, Apply leaves the
//     threshold unchanged and returns nil.
//
// These rules make it possible to use the same handler both for optional
// configuration (where missing files are benign) and for mandatory
// configuration (where absence or emptiness should be treated as an error).
//
// # Configuration and options
//
// Handlers are constructed via NewHandler, which binds a level.Threshold
// implementation and a configuration file path:
//
//	h := filectl.NewHandler(&atomic, "/etc/myapp/log.level")
//
// Callers MAY further customize behavior through functional options:
//
//   - WithPath(path string)
//     Overrides the file path used to read the log level. If Path is left
//     empty, DefaultPath is used as a fallback.
//
//   - WithDefaultLevel(lvl level.Level)
//     Configures a fallback level that is applied when the configuration
//     file is missing or effectively empty. When this option is used,
//     HasDefault is set to true.
//
//   - WithRequired(required bool)
//     Marks the configuration as required. When Required is true, a missing
//     or empty file causes Apply to return an error instead of silently
//     falling back to a default or leaving the threshold unchanged.
//
//   - WithReadFile(fn ReadFileFunc)
//     Overrides the function used to read the configuration file. By
//     default, Handler uses os.ReadFile. This hook exists primarily to
//     make Handler easy to test or to integrate with virtual filesystems.
//
// All configuration fields of Handler are initialized by NewHandler and MAY
// be further adjusted before the first call to Apply or MustApply.
//
// # Defaults and paths
//
// If NewHandler is called with an empty path and WithPath is not used, the
// handler falls back to DefaultPath, which is a simple relative path intended
// as a reasonable default (for example, "rxlog.level").
//
// Applications that require a more specific or platform-dependent location
// (such as a directory under /etc or a dedicated configuration directory)
// SHOULD pass an explicit path to NewHandler or configure it via WithPath,
// instead of relying on DefaultPath.
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
//   - Re-applying configuration periodically or on explicit trigger
//     (for example, in response to a SIGHUP or an admin command) by
//     calling Apply again, re-reading the file contents each time.
//
// Handler does not implement file watching, polling, or any scheduling
// mechanism; such concerns are deliberately left to the embedding
// application.
//
// # Error handling and MustApply
//
// For initialization paths where a misconfigured file-based level should
// prevent the application from starting successfully, Handler exposes
// MustApply, which simply calls Apply and panics on any non-nil error:
//
//	h := filectl.NewHandler(&atomic, "/etc/myapp/log.level")
//	h.MustApply() // panic if file is missing/invalid according to config
//
// This can be convenient for small services or tests where failing fast is
// preferable to continuing with an unexpected logging configuration.
//
// Larger systems may prefer to call Apply explicitly, log any errors, and
// continue with a previously configured or compile-time default level.
//
// # Security and operational considerations
//
// A file-based log level handler is inherently a privileged control surface:
// whoever can modify the configuration file effectively controls the minimum
// severity of events emitted by the process.
//
// Operators MUST ensure that:
//
//   - The file path chosen is protected with appropriate filesystem
//     permissions so that only trusted principals can modify it.
//   - Automation that writes this file (for example, configuration
//     management tools) is treated as part of the trusted operational
//     environment.
//   - Error handling paths for Apply are well-understood: in particular,
//     whether the application should continue running if the file cannot
//     be read or contains an invalid level.
//
// Misconfiguration (for example, unlinking the file or truncating it to an
// empty value while Required is true) will be surfaced as errors from Apply
// or panics from MustApply, depending on how the handler is invoked.
package filectl
