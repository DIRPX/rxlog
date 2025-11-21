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

package level

import (
	"fmt"
	"sync/atomic"

	"dirpx.dev/rxlog/rxapi/level"
)

// AtomicLevel is a dynamically adjustable log level threshold.
//
// It wraps a single atomic.Int32 value that stores the numeric representation
// of a level.Level. All operations on AtomicLevel are safe for concurrent use
// by multiple goroutines.
//
// AtomicLevel is intended for scenarios where the minimum enabled log level
// must be changed at runtime (for example, via an admin HTTP endpoint or a
// configuration reload) without recreating loggers or cores.
//
// Zero value:
//
//   - The zero value of AtomicLevel (with l == nil) is treated as having an
//     Info threshold when read via Level() or Enabled().
//   - The first call to SetLevel or UnmarshalText lazily initializes the
//     underlying atomic.Int32.
//   - Callers that want to be explicit SHOULD prefer NewAtomicLevel or
//     NewAtomicLevelAt over relying on the zero value semantics.
type AtomicLevel struct {
	l *atomic.Int32
}

// NewAtomicLevel constructs an AtomicLevel with the default threshold of Info.
//
// The returned value is fully initialized, safe for immediate use, and may be
// copied by value. Copies share the same underlying atomic state via the
// pointer field.
func NewAtomicLevel() AtomicLevel {
	var a AtomicLevel
	a.SetLevel(level.Info)
	return a
}

// NewAtomicLevelAt is a convenience constructor that builds an AtomicLevel
// and immediately sets it to the provided level.
//
// The supplied level MUST be valid (level.IsValid() == true). If it is not,
// NewAtomicLevelAt will panic via SetLevel.
func NewAtomicLevelAt(l level.Level) AtomicLevel {
	a := NewAtomicLevel()
	a.SetLevel(l)
	return a
}

// ParseAtomicLevel parses a textual representation of a level into an
// AtomicLevel whose current threshold is set to the parsed value.
//
// It accepts the same set of names and synonyms as level.Parse (case-
// insensitive, with leading/trailing whitespace ignored). On success,
// the returned AtomicLevel is initialized and safe for immediate use.
//
// On failure, ParseAtomicLevel returns an AtomicLevel initialized to
// Info and a non-nil error. Callers MUST treat such errors as
// configuration/input problems and MUST NOT silently ignore them.
func ParseAtomicLevel(text string) (AtomicLevel, error) {
	a := NewAtomicLevel()

	l, err := level.Parse(text)
	if err != nil {
		return a, err
	}

	a.SetLevel(l)
	return a, nil
}

// MustParseAtomicLevel is a convenience helper that parses the given textual
// level into an AtomicLevel and panics on error.
//
// This is appropriate for static initialization code where level names are
// hard-coded, and any failure indicates a programmer error or a misconfigured
// build. It MUST NOT be used for untrusted or user-provided input.
func MustParseAtomicLevel(text string) AtomicLevel {
	a, err := ParseAtomicLevel(text)
	if err != nil {
		panic(err)
	}
	return a
}

// Level returns the current minimum enabled level.
//
// It is safe to call Level concurrently with SetLevel and Enabled.
//
// If the underlying atomic storage has not yet been initialized (l == nil),
// Level returns level.Info as the default threshold. This ensures that the
// zero value of AtomicLevel behaves reasonably even before explicit
// initialization.
func (lvl *AtomicLevel) Level() level.Level {
	if lvl.l == nil {
		// Lazy default for zero value.
		return level.Info
	}
	return level.Level(lvl.l.Load())
}

// SetLevel updates the minimum enabled level.
//
// The supplied level MUST be a valid severity (l.IsValid() == true). Passing
// level.Invalid or an out-of-range value is considered a programmer error and
// will cause SetLevel to panic. This avoids silently accepting misconfigured
// thresholds.
//
// SetLevel is safe for concurrent use with Level and Enabled. All copies of
// an AtomicLevel share the same underlying atomic state due to the pointer
// field, so updating any copy updates the effective threshold for all of them.
func (lvl *AtomicLevel) SetLevel(l level.Level) {
	if !l.IsValid() {
		panic(fmt.Sprintf("AtomicLevel.SetLevel: invalid level %d", int8(l)))
	}

	if lvl.l == nil {
		lvl.l = &atomic.Int32{}
	}

	lvl.l.Store(int32(l))
}

// Enabled reports whether events at the candidate level l should be emitted
// under the current threshold.
//
// It returns true if and only if l >= lvl.Level(). Enabled assumes that l is
// a valid level.Level and does not perform additional validation, because it
// is typically called on hot paths in the logging pipeline.
//
// Enabled is safe for concurrent use with Level and SetLevel.
func (lvl *AtomicLevel) Enabled(l level.Level) bool {
	return l >= lvl.Level()
}

// String returns the string representation of the current threshold level.
//
// This is equivalent to lvl.Level().String().
func (lvl *AtomicLevel) String() string {
	return lvl.Level().String()
}

// UnmarshalText implements encoding.TextUnmarshaler for AtomicLevel.
//
// It accepts the same textual representations as level.Parse (for example,
// "debug", "info", "warn", "error", with case-insensitive matching).
//
// On success, the receiver is updated to the parsed level, and the underlying
// atomic storage is initialized if necessary. On failure, the receiver is left
// unchanged and a non-nil error is returned.
func (lvl *AtomicLevel) UnmarshalText(text []byte) error {
	parsed, err := level.Parse(string(text))
	if err != nil {
		return err
	}

	lvl.SetLevel(parsed)
	return nil
}

// MarshalText implements encoding.TextMarshaler for AtomicLevel.
//
// It marshals the current threshold level using level.Level.MarshalText(), so
// the textual representation is consistent with other uses of Level in
// configuration and wire formats.
func (lvl *AtomicLevel) MarshalText() ([]byte, error) {
	return lvl.Level().MarshalText()
}
