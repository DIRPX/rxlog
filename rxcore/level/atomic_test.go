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

package level_test

import (
	"testing"

	apilevel "dirpx.dev/rxlog/rxapi/level"
	corelevel "dirpx.dev/rxlog/rxcore/level"
)

// TestAtomicLevel_ZeroValueDefaultsToInfo verifies that the zero value
// behaves as documented: it uses Info as the default threshold and does
// not panic when used for reads.
func TestAtomicLevel_ZeroValueDefaultsToInfo(t *testing.T) {
	var a corelevel.AtomicLevel

	if got, want := a.Level(), apilevel.Info; got != want {
		t.Fatalf("zero AtomicLevel.Level() = %v, want %v", got, want)
	}

	low, err := apilevel.Parse("trace")
	if err != nil {
		t.Fatalf("Parse(trace) failed: %v", err)
	}
	mid := apilevel.Info
	hi, err := apilevel.Parse("error")
	if err != nil {
		t.Fatalf("Parse(error) failed: %v", err)
	}

	if a.Enabled(low) {
		t.Fatalf("zero AtomicLevel.Enabled(trace) = true, want false")
	}
	if !a.Enabled(mid) {
		t.Fatalf("zero AtomicLevel.Enabled(info) = false, want true")
	}
	if !a.Enabled(hi) {
		t.Fatalf("zero AtomicLevel.Enabled(error) = false, want true")
	}
}

// TestNewAtomicLevel_DefaultInfo ensures that NewAtomicLevel constructs
// a fully initialized AtomicLevel with Info threshold.
func TestNewAtomicLevel_DefaultInfo(t *testing.T) {
	a := corelevel.NewAtomicLevel()

	if got, want := a.Level(), apilevel.Info; got != want {
		t.Fatalf("NewAtomicLevel().Level() = %v, want %v", got, want)
	}

	// Sanity check Enabled semantics for the default threshold.
	low, err := apilevel.Parse("trace")
	if err != nil {
		t.Fatalf("Parse(trace) failed: %v", err)
	}
	if a.Enabled(low) {
		t.Fatalf("NewAtomicLevel().Enabled(trace) = true, want false")
	}
	if !a.Enabled(apilevel.Info) {
		t.Fatalf("NewAtomicLevel().Enabled(info) = false, want true")
	}
}

// TestNewAtomicLevelAt_SetsRequestedLevel verifies that NewAtomicLevelAt
// initializes the AtomicLevel to the requested level.
func TestNewAtomicLevelAt_SetsRequestedLevel(t *testing.T) {
	a := corelevel.NewAtomicLevelAt(apilevel.Error)

	if got, want := a.Level(), apilevel.Error; got != want {
		t.Fatalf("NewAtomicLevelAt(error).Level() = %v, want %v", got, want)
	}
}

// TestAtomicLevel_SetLevel_ValidAndInvalid ensures SetLevel accepts valid
// levels and panics on invalid ones.
func TestAtomicLevel_SetLevel_ValidAndInvalid(t *testing.T) {
	var a corelevel.AtomicLevel

	// Valid level must be applied.
	a.SetLevel(apilevel.Warn)
	if got, want := a.Level(), apilevel.Warn; got != want {
		t.Fatalf("SetLevel(warn) -> Level() = %v, want %v", got, want)
	}

	// Invalid level MUST cause a panic.
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("SetLevel(apilevel.Invalid) did not panic")
		}
	}()
	a.SetLevel(apilevel.Invalid)
}

// TestAtomicLevel_CopiesShareState verifies that copies of AtomicLevel
// share the same underlying atomic state via the pointer field.
func TestAtomicLevel_CopiesShareState(t *testing.T) {
	a := corelevel.NewAtomicLevelAt(apilevel.Info)
	b := a // copy by value

	b.SetLevel(apilevel.Error)

	if got, want := a.Level(), apilevel.Error; got != want {
		t.Fatalf("after modifying copy, original.Level() = %v, want %v", got, want)
	}
	if got, want := b.Level(), apilevel.Error; got != want {
		t.Fatalf("copy.Level() = %v, want %v", got, want)
	}
}

// TestParseAtomicLevel_ValidAndInvalid verifies that ParseAtomicLevel
// parses valid text and returns an Info-initialized AtomicLevel on error.
func TestParseAtomicLevel_ValidAndInvalid(t *testing.T) {
	// Valid parse.
	a, err := corelevel.ParseAtomicLevel("error")
	if err != nil {
		t.Fatalf("ParseAtomicLevel(\"error\") returned error: %v", err)
	}
	if got, want := a.Level(), apilevel.Error; got != want {
		t.Fatalf("ParseAtomicLevel(\"error\").Level() = %v, want %v", got, want)
	}

	// Invalid parse: returned AtomicLevel should still be initialized to Info.
	b, err := corelevel.ParseAtomicLevel("no-such-level")
	if err == nil {
		t.Fatalf("ParseAtomicLevel(\"no-such-level\") error = nil, want non-nil")
	}
	if got, want := b.Level(), apilevel.Info; got != want {
		t.Fatalf("ParseAtomicLevel(\"no-such-level\").Level() = %v, want %v", got, want)
	}
}

// TestMustParseAtomicLevel_PanicsOnError ensures that MustParseAtomicLevel
// panics for invalid input.
func TestMustParseAtomicLevel_PanicsOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustParseAtomicLevel(\"invalid\") did not panic")
		}
	}()
	_ = corelevel.MustParseAtomicLevel("definitely-not-a-level")
}

// TestSetLevelString_ValidAndInvalid ensures that SetLevelString updates
// the threshold on valid input and leaves it unchanged on error.
func TestSetLevelString_ValidAndInvalid(t *testing.T) {
	a := corelevel.NewAtomicLevelAt(apilevel.Info)

	if err := a.SetLevelString("error"); err != nil {
		t.Fatalf("SetLevelString(\"error\") returned error: %v", err)
	}
	if got, want := a.Level(), apilevel.Error; got != want {
		t.Fatalf("after SetLevelString(\"error\"), Level() = %v, want %v", got, want)
	}

	// Invalid input must not change the current level.
	if err := a.SetLevelString("not-a-level"); err == nil {
		t.Fatalf("SetLevelString(\"not-a-level\") error = nil, want non-nil")
	}
	if got, want := a.Level(), apilevel.Error; got != want {
		t.Fatalf("after invalid SetLevelString, Level() = %v, want %v", got, want)
	}
}

// TestMustSetLevelString_PanicsOnError verifies that MustSetLevelString
// panics for invalid text.
func TestMustSetLevelString_PanicsOnError(t *testing.T) {
	a := corelevel.NewAtomicLevel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustSetLevelString(\"bad\") did not panic")
		}
	}()
	a.MustSetLevelString("bad-level")
}

// TestAtomicLevel_EnabledFollowsThreshold checks that Enabled respects
// the current threshold.
func TestAtomicLevel_EnabledFollowsThreshold(t *testing.T) {
	a := corelevel.NewAtomicLevelAt(apilevel.Warn)

	low, err := apilevel.Parse("info")
	if err != nil {
		t.Fatalf("Parse(info) failed: %v", err)
	}
	hi, err := apilevel.Parse("error")
	if err != nil {
		t.Fatalf("Parse(error) failed: %v", err)
	}

	if a.Enabled(low) {
		t.Fatalf("Enabled(info) = true at warn threshold, want false")
	}
	if !a.Enabled(apilevel.Warn) {
		t.Fatalf("Enabled(warn) = false at warn threshold, want true")
	}
	if !a.Enabled(hi) {
		t.Fatalf("Enabled(error) = false at warn threshold, want true")
	}
}

// TestAtomicLevel_StringDelegatesToLevel verifies that String returns
// the textual form of the current Level.
func TestAtomicLevel_StringDelegatesToLevel(t *testing.T) {
	a := corelevel.NewAtomicLevelAt(apilevel.Error)

	if got, want := a.String(), "error"; got != want {
		t.Fatalf("AtomicLevel.String() = %q, want %q", got, want)
	}
}

// TestAtomicLevel_UnmarshalText_ValidAndInvalid ensures that UnmarshalText
// updates the level on valid input and leaves it unchanged on invalid input.
func TestAtomicLevel_UnmarshalText_ValidAndInvalid(t *testing.T) {
	var a corelevel.AtomicLevel

	// Valid text.
	if err := a.UnmarshalText([]byte("warn")); err != nil {
		t.Fatalf("UnmarshalText(\"warn\") returned error: %v", err)
	}
	if got, want := a.Level(), apilevel.Warn; got != want {
		t.Fatalf("after UnmarshalText(\"warn\"), Level() = %v, want %v", got, want)
	}

	// Invalid text should not change the level.
	if err := a.UnmarshalText([]byte("unknown-level")); err == nil {
		t.Fatalf("UnmarshalText(\"unknown-level\") error = nil, want non-nil")
	}
	if got, want := a.Level(), apilevel.Warn; got != want {
		t.Fatalf("after invalid UnmarshalText, Level() = %v, want %v", got, want)
	}
}

// TestAtomicLevel_MarshalTextMatchesLevel verifies that MarshalText
// matches the underlying Level's MarshalText representation.
func TestAtomicLevel_MarshalTextMatchesLevel(t *testing.T) {
	a := corelevel.NewAtomicLevelAt(apilevel.Error)

	text, err := a.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() returned error: %v", err)
	}

	levelText, err := a.Level().MarshalText()
	if err != nil {
		t.Fatalf("Level().MarshalText() returned error: %v", err)
	}

	if string(text) != string(levelText) {
		t.Fatalf("AtomicLevel.MarshalText() = %q, want %q", string(text), string(levelText))
	}
}
