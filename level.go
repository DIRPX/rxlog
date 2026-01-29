/*
   Copyright 2026 The DIRPX Authors.

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

package rxlog

import "dirpx.dev/rxlog/rxapi/level"

// Type aliases for rxapi/level types.
type (
	// Level is an alias for rxapi/level.Level, representing log severity.
	Level = level.Level

	// LevelEnabler is an alias for rxapi/level.Enabler, which decides whether
	// a given log level is enabled.
	LevelEnabler = level.Enabler

	// LevelEnablerFunc is an alias for rxapi/level.EnablerFunc, allowing
	// ordinary functions to implement the Enabler interface.
	LevelEnablerFunc = level.EnablerFunc

	// LevelThreshold is an alias for rxapi/level.Threshold, representing
	// a mutable level threshold.
	LevelThreshold = level.Threshold

	// LevelEncoder is an alias for rxapi/level.Encoder, which serializes
	// log levels into buffers.
	LevelEncoder = level.Encoder
)

// Level constants providing convenient access to rxapi/level severity values.
const (
	// TraceLevel is the most verbose level, intended for fine-grained debugging.
	TraceLevel = level.Trace

	// DebugLevel is a verbose diagnostic level for regular debugging.
	DebugLevel = level.Debug

	// InfoLevel is the baseline informational level for normal operation.
	InfoLevel = level.Info

	// NoticeLevel marks noteworthy events that are unusual but not warnings.
	NoticeLevel = level.Notice

	// WarnLevel indicates unexpected situations that may require attention.
	WarnLevel = level.Warn

	// ErrorLevel indicates failures where the operation did not succeed.
	ErrorLevel = level.Error

	// CriticalLevel represents severe errors requiring urgent action.
	CriticalLevel = level.Critical

	// FatalLevel indicates unrecoverable errors requiring process termination.
	FatalLevel = level.Fatal

	// InvalidLevel is a sentinel value for out-of-range or unrecognized levels.
	InvalidLevel = level.Invalid
)

// ParseLevel converts a textual log level into its Level representation.
//
// This is a convenience wrapper around level.Parse. See rxapi/level.Parse
// for full documentation on supported formats and error handling.
func ParseLevel(s string) (Level, error) {
	return level.Parse(s)
}

// MustParseLevel is a convenience wrapper around level.MustParse that panics
// on error.
//
// This is intended for use in static initialization where level names are
// hard-coded. See rxapi/level.MustParse for full documentation.
func MustParseLevel(s string) Level {
	return level.MustParse(s)
}
