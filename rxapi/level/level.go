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

// Enabler decides whether log events at a given severity Level SHOULD be
// emitted.
//
// Implementations of Enabler are intended to provide deterministic,
// severity-based filtering. Concerns such as sampling, rate limiting, or
// content-based filtering SHOULD be implemented in higher-level components
// (for example, as separate predicates or cores), and MUST NOT be encoded
// into Enabled in a way that makes its behavior non-deterministic.
//
// A common implementation pattern is to treat a concrete Level value as a
// threshold: an enabler representing a minimum level MIN SHOULD return true
// for MIN and all more severe levels, and false for less severe levels.
// For example, if MIN is Warn, then Enabled(Warn), Enabled(Error), and
// Enabled(Fatal) SHOULD return true, while Enabled(Info) and Enabled(Debug)
// SHOULD return false.
type Enabler interface {
	// Enabled reports whether logging at the provided Level SHOULD be allowed.
	//
	// Enabled MUST NOT have side effects and MUST be safe to call concurrently
	// from multiple goroutines. Callers MAY invoke Enabled frequently (for
	// example, on every log attempt) and therefore implementations SHOULD keep
	// this check as lightweight as possible.
	Enabled(Level) bool
}

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
	// Any Level value strictly less than _min (such as Trace) MUST be treated
	// as out of range for configuration and MUST NOT be accepted as a valid
	// logging threshold in public APIs. Callers MAY still use such values for
	// internal or hard-coded log records, but MUST NOT expose them as valid
	// configuration options to end users.
	_min Level = Debug

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
