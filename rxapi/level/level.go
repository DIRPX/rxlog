/*
   Copyright 2025 The DIRPX Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you MAY not use this file except in compliance with the License.
   You MAY obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package level

import (
	"fmt"
	"strings"
)

// Level represents the logging severity used across rxlog.
//
// The numeric ordering of Level values is normative: lower values are more
// verbose and less severe; higher values are less verbose and more severe.
// Filtering implementations MUST treat a chosen Level value as an inclusive
// threshold and emit all records at that Level or any more severe Level.
type Level int8

const (
	// Trace is the most verbose level.
	//
	// Callers SHOULD use Trace only for fine-grained, low-level diagnostic
	// information that is not required for normal operation (for example,
	// detailed state dumps, internal algorithm steps, or tight loops).
	// Implementations SHOULD disable Trace by default in most environments,
	// and MAY enable it only for ad-hoc debugging or targeted investigations.
	Trace Level = iota - 1

	// Debug is a verbose diagnostic level intended for regular debugging.
	//
	// Callers SHOULD use Debug for information that explains control flow,
	// decisions, and state changes, but that would be too noisy at Info.
	// Implementations MAY enable Debug in development and during active issue
	// investigation, and SHOULD disable it by default in high-traffic or
	// performance-sensitive production environments.
	Debug

	// Info is the baseline informational level for normal operation.
	//
	// Callers SHOULD use Info for high-level lifecycle events and steady-state
	// observability, such as service startup/shutdown, configuration summaries,
	// and key state transitions. Implementations SHOULD enable Info in all
	// normal deployments so that operators MAY continuously observe system
	// health and major events.
	Info

	// Notice marks noteworthy events that are unusual but not yet warnings.
	//
	// Callers MAY use Notice when something out of the ordinary happens that
	// remains within the expected operational envelope but might be relevant
	// for trend analysis or early detection of issues (for example, rare but
	// valid code paths, soft limits being approached, or automatic recovery
	// that succeeded without user-visible impact).
	Notice

	// Warn indicates unexpected or undesirable situations that are not fatal
	// and do not immediately violate correctness, but MAY require attention.
	//
	// Callers SHOULD use Warn for conditions that, if left unaddressed, could
	// evolve into errors, such as transient failures with automatic retry,
	// degraded performance, or partial feature unavailability. Operators
	// SHOULD treat recurring or sustained Warn entries as signals that the
	// system MAY be trending towards failure.
	Warn

	// Error indicates failures after which the process can continue running,
	// but the attempted operation did not succeed as intended.
	//
	// Callers MUST assume that an Error entry corresponds to a failed logical
	// operation or a user-visible failure mode. Implementations SHOULD make
	// Error entries clearly visible to operators and SHOULD support alerting
	// or remediation workflows around them. Error logs MUST NOT be used for
	// purely informational messages.
	Error

	// Critical represents severe errors that significantly impact correctness,
	// availability, or data integrity, and typically require urgent action.
	//
	// Callers SHOULD use Critical when the system is still running but its
	// ability to operate correctly is compromised (for example, persistent
	// inability to reach a rxcore dependency, unrecoverable data inconsistencies,
	// or repeated errors in critical execution paths). Operators MAY treat
	// Critical as a paging or incident trigger, and monitoring systems SHOULD
	// prioritize Critical events above Error.
	Critical

	// Fatal indicates unrecoverable errors after which the process MUST exit
	// or be terminated as soon as practical.
	//
	// Callers MUST NOT continue normal execution after logging at Fatal. Code
	// that logs at Fatal MUST terminate the process (or the relevant component)
	// as soon as it is safe to do so, because continuing execution is assumed
	// to be unsafe or misleading (for example, corrupted in-memory state,
	// failed mandatory initialization, or violation of fundamental invariants).
	Fatal

	// _min defines the lowest Level value that is treated as valid for normal
	// threshold configuration and user-visible filtering.
	//
	// Any Level value strictly less than _min MUST be treated
	// as out of range for configuration and MUST NOT be accepted as a valid
	// logging threshold in public APIs. Callers MAY still use such values for
	// internal or hard-coded log records, but MUST NOT expose them as valid
	// configuration options to end users.
	_min Level = Trace

	// _max defines the highest Level value that is considered valid for normal
	// configuration and parsing.
	//
	// Any Level value strictly greater than _max MUST be treated as out of
	// range. Parsing and validation logic SHOULD report an error or return
	// Invalid rather than silently clamping the value to _max.
	_max Level = Fatal

	// Invalid is a sentinel Level value used to represent an out-of-range or
	// unrecognized level (for example, the result of a failed parse).
	//
	// Invalid is always strictly greater than _max and MUST NOT be used as the
	// Level of an actual log entry. Callers and parsers MAY return Invalid to
	// signal an error condition instead of panicking, and callers that receive
	// Invalid MUST treat it as a configuration or input error.
	Invalid = _max + 1
)

// Parse converts a textual log level into its Level representation.
//
// The comparison is case-insensitive and ignores leading/trailing whitespace.
// Common spellings and synonyms are supported, for example:
//
//	"trace"                   -> Trace
//	"debug"                   -> Debug
//	"info", "information"     -> Info
//	"notice"                  -> Notice
//	"warn", "warning"         -> Warn
//	"error", "err"            -> Error
//	"critical", "crit"        -> Critical
//	"fatal"                   -> Fatal
//
// If the string does not match any known level, Parse returns Invalid and a
// non-nil error. Callers MUST treat such errors as configuration or input
// problems and MUST NOT silently coerce unknown values to a default.
func Parse(s string) (Level, error) {
	name := strings.TrimSpace(strings.ToLower(s))
	switch name {
	case "trace":
		return Trace, nil
	case "debug":
		return Debug, nil
	case "info", "information", "informational":
		return Info, nil
	case "notice":
		return Notice, nil
	case "warn", "warning":
		return Warn, nil
	case "error", "err":
		return Error, nil
	case "critical", "crit":
		return Critical, nil
	case "fatal":
		return Fatal, nil
	default:
		return Invalid, fmt.Errorf("level.Parse: unknown level %q", s)
	}
}

// MustParse is a convenience wrapper around Parse that panics on error.
//
// MustParse is intended for use in static initialization code where level
// names are hard-coded and any failure indicates a programmer error or a
// misconfigured build. It MUST NOT be used for untrusted or user-provided
// input, where returning an error is preferable.
//
// Example:
//
//	var defaultLevel = level.MustParse("info")
func MustParse(s string) Level {
	lvl, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return lvl
}

// String returns the canonical lowercase textual representation of the level.
//
// For defined levels, String returns one of:
//
//	"trace", "debug", "info", "notice", "warn",
//	"error", "critical", "fatal", "invalid"
//
// For values that do not correspond to any known constant, String returns a
// fallback representation of the form "Level(<numeric>)". This is primarily
// intended for debugging and SHOULD NOT be relied upon in stable logs or
// external formats.
func (l Level) String() string {
	switch l {
	case Trace:
		return "trace"
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Notice:
		return "notice"
	case Warn:
		return "warn"
	case Error:
		return "error"
	case Critical:
		return "critical"
	case Fatal:
		return "fatal"
	case Invalid:
		return "invalid"
	default:
		return fmt.Sprintf("Level(%d)", int8(l))
	}
}

// MarshalText implements encoding.TextMarshaler for Level.
//
// It serializes the level to its canonical lowercase name (for example,
// "debug", "info", "error"). Only defined, non-Invalid levels (Trace through
// Fatal, inclusive) are considered marshalable. Attempting to marshal Invalid
// or a value outside this range results in an error.
//
// This behavior is appropriate for configuration and wire formats such as
// JSON, YAML, or TOML that rely on textual level names.
func (l Level) MarshalText() ([]byte, error) {
	if !l.IsValid() {
		return nil, fmt.Errorf("level.MarshalText: cannot marshal invalid level %d", int8(l))
	}
	return []byte(l.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for Level.
//
// It accepts the same set of textual names as Parse, performing a
// case-insensitive comparison and ignoring leading/trailing whitespace.
// On success, the receiver is updated to the parsed level and nil is returned.
//
// On failure, the receiver is left unchanged and a non-nil error is returned.
// Callers MUST treat such errors as configuration or input problems.
func (l *Level) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*l = parsed
	return nil
}

// IsValid reports whether l is one of the defined severity levels that may be
// used for actual log entries (_min through _max, inclusive).
//
// It returns false for Invalid and for any value outside the [ _min, _max ]
// numeric range.
func (l Level) IsValid() bool {
	return l >= _min && l <= _max
}
