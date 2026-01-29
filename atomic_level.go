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

import "dirpx.dev/rxlog/rxcore/level"

// AtomicLevel is an alias for rxcore/level.AtomicLevel, representing
// a dynamically adjustable log level threshold.
//
// It provides thread-safe atomic operations for runtime level changes
// without recreating loggers or cores.
type AtomicLevel = level.AtomicLevel

// NewAtomicLevel constructs an AtomicLevel with the default threshold of Info.
//
// This is a convenience wrapper around rxcore/level.NewAtomicLevel.
// See rxcore/level.NewAtomicLevel for full documentation.
func NewAtomicLevel() AtomicLevel {
	return level.NewAtomicLevel()
}

// NewAtomicLevelAt constructs an AtomicLevel at the specified level.
//
// The supplied level must be valid (IsValid() == true), or this function
// will panic.
//
// This is a convenience wrapper around rxcore/level.NewAtomicLevelAt.
// See rxcore/level.NewAtomicLevelAt for full documentation.
func NewAtomicLevelAt(l Level) AtomicLevel {
	return level.NewAtomicLevelAt(l)
}

// ParseAtomicLevel parses a textual level into an AtomicLevel.
//
// On success, returns an initialized AtomicLevel set to the parsed level.
// On failure, returns an AtomicLevel at Info and a non-nil error.
//
// This is a convenience wrapper around rxcore/level.ParseAtomicLevel.
// See rxcore/level.ParseAtomicLevel for full documentation.
func ParseAtomicLevel(text string) (AtomicLevel, error) {
	return level.ParseAtomicLevel(text)
}

// MustParseAtomicLevel parses a textual level into an AtomicLevel and
// panics on error.
//
// This is intended for static initialization where level names are hard-coded.
//
// This is a convenience wrapper around rxcore/level.MustParseAtomicLevel.
// See rxcore/level.MustParseAtomicLevel for full documentation.
func MustParseAtomicLevel(text string) AtomicLevel {
	return level.MustParseAtomicLevel(text)
}
