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

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"dirpx.dev/rxlog/rxapi/level"
)

// DefaultPath is the fallback file path used when Handler.Path is empty.
//
// The default is intentionally simple and relative to the current working
// directory. Applications that require a more specific or platform-dependent
// location (for example, /etc/myapp/log.level) SHOULD override the path via
// NewHandler's path argument or WithPath.
const DefaultPath = "rxlog.level"

// ReadFileFunc abstracts file reading for Handler.
//
// The concrete function is expected to behave like os.ReadFile: it should
// return the full contents of the file or an error. This indirection is
// primarily intended to make Handler easy to test without touching the
// real filesystem.
type ReadFileFunc func(path string) ([]byte, error)

// Handler configures a mutable log level threshold from a file on disk.
//
// It is intended primarily for startup-time configuration, but Apply MAY be
// called multiple times (for example, after an external configuration reload)
// to re-apply settings sourced from the same file.
//
// Semantics overview:
//
//   - The handler reads a text file whose (trimmed) contents are expected
//     to be a single log level name (e.g., "info", "warn", "error").
//   - On success, the parsed level is applied to the underlying Threshold.
//   - Missing or empty files can be treated as "no override", or as a reason
//     to fall back to a default level, or as a hard error, depending on
//     configuration (DefaultLevel, HasDefault, Required).
//
// Handler does not implement any file watching or polling logic by itself;
// it simply reads and interprets the file when Apply is invoked.
type Handler struct {
	// Level is the mutable log level threshold being controlled.
	//
	// This value MUST be non-nil. If it is nil at Apply time, Apply returns
	// an error.
	Level level.Threshold

	// Path is the filesystem path to the configuration file that carries
	// the desired log level.
	//
	// If empty, DefaultPath is used. Callers SHOULD prefer passing an
	// explicit path to NewHandler for clarity.
	Path string

	// DefaultLevel is the fallback level used when the configuration file
	// is missing or effectively empty.
	//
	// If HasDefault is false, DefaultLevel is ignored and Apply will leave
	// the Threshold unchanged when the file is missing or empty (unless
	// Required is true).
	DefaultLevel level.Level

	// HasDefault indicates whether DefaultLevel is meaningful.
	//
	// This extra flag avoids treating the zero value of level.Level as an
	// implicit default.
	HasDefault bool

	// Required controls behavior when the configuration file is missing
	// or empty.
	//
	// If Required is true and the file does not exist, cannot be read, or
	// contains only whitespace, Apply returns an error.
	//
	// If Required is false (the default), a missing/empty file is treated
	// as "no override" unless a DefaultLevel is configured.
	Required bool

	// ReadFile is the function used to read the configuration file.
	//
	// If nil, os.ReadFile is used. This indirection exists primarily to
	// make Handler easy to test without mutating process-wide filesystem
	// state.
	ReadFile ReadFileFunc
}

// NewHandler constructs a Handler bound to the provided mutable level
// threshold and file path.
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
//	h := filectl.NewHandler(&atomic, "/etc/myapp/log.level",
//	    filectl.WithDefaultLevel(level.Info), // optional fallback
//	)
//
//	if err := h.Apply(); err != nil {
//	    // Decide whether to fail fast or log-and-continue.
//	}
//
// If th is nil, NewHandler still returns a handler, but Apply will fail
// with an error until Level is set to a non-nil Threshold.
func NewHandler(th level.Threshold, path string, opts ...Option) *Handler {
	h := &Handler{
		Level:    th,
		Path:     path,
		ReadFile: os.ReadFile,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.Path == "" {
		h.Path = DefaultPath
	}
	if h.ReadFile == nil {
		h.ReadFile = os.ReadFile
	}

	return h
}

// Apply reads the configured file, interprets its contents as a log level,
// and updates the underlying Threshold accordingly.
//
// The method first verifies that the Handler has a non-nil Level. If Level
// is nil, Apply immediately returns an error, since there is no target to
// update.
//
// Next, Apply determines the effective configuration file path. If Path is
// non-empty, it is used as-is; otherwise, DefaultPath is used. The file is
// then read using the configured ReadFile function, or os.ReadFile if no
// custom reader is provided.
//
// If the file does not exist (for example, because it has not been created
// yet or the path is incorrect), the behavior depends on the Handler's
// configuration:
//   - If Required is true, Apply returns an error indicating that the file
//     is missing.
//   - If Required is false and HasDefault is true, Apply sets the Threshold
//     to DefaultLevel and returns nil.
//   - If Required is false and HasDefault is false, Apply leaves the
//     Threshold unchanged and returns nil.
//
// If the file exists but cannot be read for reasons other than "not found"
// (such as permission problems or I/O errors), Apply treats this as a hard
// failure and returns an error regardless of the Required flag, because the
// configuration source is present but unusable.
//
// When the file is read successfully, Apply trims leading and trailing
// whitespace from its contents. If the resulting string is empty, the
// behavior again depends on configuration:
//   - If Required is true, Apply returns an error indicating that the file
//     does not contain a usable value.
//   - If Required is false and HasDefault is true, Apply sets the Threshold
//     to DefaultLevel and returns nil.
//   - If Required is false and HasDefault is false, Apply leaves the
//     Threshold unchanged and returns nil.
//
// If the trimmed contents are non-empty, Apply attempts to parse them as a
// log level using level.Parse. If parsing fails, Apply returns an error
// describing the invalid value. If parsing succeeds, the parsed level is
// written into the underlying Threshold via SetLevel, and Apply returns nil.
//
// Apply is safe to call multiple times; each invocation re-reads the current
// file contents and may adjust the Threshold accordingly.
func (h *Handler) Apply() error {
	if h.Level == nil {
		return fmt.Errorf("file handler: Level Threshold is nil")
	}

	path := h.Path
	if strings.TrimSpace(path) == "" {
		path = DefaultPath
	}

	readFile := h.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}

	data, err := readFile(path)
	if err != nil {
		// Distinguish "file does not exist" from other I/O problems.
		if errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err) {
			if h.Required {
				return fmt.Errorf("file handler: required file %q does not exist: %w", path, err)
			}
			if h.HasDefault {
				h.Level.SetLevel(h.DefaultLevel)
			}
			// Missing file with no default -> no override.
			return nil
		}

		// Any other read error is treated as fatal for Apply.
		return fmt.Errorf("file handler: failed to read %q: %w", path, err)
	}

	value := strings.TrimSpace(string(data))
	if value == "" {
		if h.Required {
			return fmt.Errorf("file handler: required file %q is empty", path)
		}
		if h.HasDefault {
			h.Level.SetLevel(h.DefaultLevel)
		}
		// Empty file with no default -> no override.
		return nil
	}

	parsed, err := level.Parse(value)
	if err != nil {
		return fmt.Errorf("file handler: invalid level %q in %s: %w", value, path, err)
	}

	h.Level.SetLevel(parsed)
	return nil
}

// MustApply is a convenience helper that calls Apply and panics if it
// returns a non-nil error.
//
// It is intended for use in process initialization paths where an
// invalid file-based configuration should prevent the application from
// starting successfully.
//
// Example:
//
//	handler := filectl.NewHandler(&atomic, "/etc/myapp/log.level")
//	handler.MustApply()
func (h *Handler) MustApply() {
	if err := h.Apply(); err != nil {
		panic(err)
	}
}
