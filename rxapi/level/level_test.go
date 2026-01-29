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

package level_test

import (
	"encoding"
	"fmt"
	"testing"

	"dirpx.dev/rxlog/rxapi/level"
)

// Compile-time interface checks to ensure Level implements the expected
// text marshaling interfaces.
var (
	_ encoding.TextMarshaler   = level.Level(0)
	_ encoding.TextUnmarshaler = (*level.Level)(nil)
)

// TestParse_ValidLevels verifies that Parse recognizes canonical names
// and common synonyms in a case-insensitive manner.
func TestParse_ValidLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect level.Level
	}{
		{"trace-lower", "trace", level.Trace},
		{"trace-upper", "TRACE", level.Trace},
		{"debug", "debug", level.Debug},
		{"info", "info", level.Info},
		{"info-synonym", "Information", level.Info},
		{"notice", "notice", level.Notice},
		{"warn", "warn", level.Warn},
		{"warn-synonym", "WARNING", level.Warn},
		{"error", "error", level.Error},
		{"error-synonym", "ERR", level.Error},
		{"critical", "critical", level.Critical},
		{"critical-synonym", "CRIT", level.Critical},
		{"fatal", "fatal", level.Fatal},
		{"with-spaces", "  info  ", level.Info},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := level.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if got != tt.expect {
				t.Fatalf("Parse(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

// TestParse_InvalidLevel verifies that unknown textual levels produce
// Invalid and a non-nil error.
func TestParse_InvalidLevel(t *testing.T) {
	t.Parallel()

	inputs := []string{"", "noop", "verbose", " infoo ", "warnn"}

	for _, in := range inputs {
		in := in
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			got, err := level.Parse(in)
			if err == nil {
				t.Fatalf("Parse(%q) returned nil error, got level %v, want non-nil error", in, got)
			}
			if got != level.Invalid {
				t.Fatalf("Parse(%q) = %v, want Invalid", in, got)
			}
		})
	}
}

// TestMustParse_PanicsOnError ensures MustParse panics when Parse would
// return an error.
func TestMustParse_PanicsOnError(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustParse did not panic for invalid level")
		}
	}()

	_ = level.MustParse("not-a-level")
}

// TestMustParse_SucceedsOnValid verifies that MustParse returns the
// expected level for valid textual representations.
func TestMustParse_SucceedsOnValid(t *testing.T) {
	t.Parallel()

	got := level.MustParse("INFO")
	if got != level.Info {
		t.Fatalf("MustParse(\"INFO\") = %v, want %v", got, level.Info)
	}
}

// TestLevel_String checks the canonical string representation for all
// defined levels and the fallback format for unknown numeric values.
func TestLevel_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		l      level.Level
		expect string
	}{
		{level.Trace, "trace"},
		{level.Debug, "debug"},
		{level.Info, "info"},
		{level.Notice, "notice"},
		{level.Warn, "warn"},
		{level.Error, "error"},
		{level.Critical, "critical"},
		{level.Fatal, "fatal"},
		{level.Invalid, "invalid"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.expect, func(t *testing.T) {
			t.Parallel()

			if got := tt.l.String(); got != tt.expect {
				t.Fatalf("Level(%v).String() = %q, want %q", tt.l, got, tt.expect)
			}
		})
	}

	// Unknown numeric value: expect fallback "Level(<n>)".
	unknown := level.Level(42)
	got := unknown.String()
	want := fmt.Sprintf("Level(%d)", int8(unknown))
	if got != want {
		t.Fatalf("String() for unknown level = %q, want %q", got, want)
	}
}

// TestLevel_MarshalText verifies that all valid levels marshal to their
// canonical textual representations and that invalid/out-of-range values
// produce an error.
func TestLevel_MarshalText(t *testing.T) {
	t.Parallel()

	// All known valid levels should marshal successfully.
	valid := []level.Level{
		level.Trace,
		level.Debug,
		level.Info,
		level.Notice,
		level.Warn,
		level.Error,
		level.Critical,
		level.Fatal,
	}

	for _, l := range valid {
		l := l
		t.Run(l.String(), func(t *testing.T) {
			t.Parallel()

			b, err := l.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText(%v) returned error: %v", l, err)
			}
			if string(b) != l.String() {
				t.Fatalf("MarshalText(%v) = %q, want %q", l, string(b), l.String())
			}
		})
	}

	// Invalid and out-of-range levels must fail.
	invalids := []level.Level{
		level.Invalid,
		level.Level(127),
	}

	for _, l := range invalids {
		l := l
		t.Run(fmt.Sprintf("invalid-%d", int8(l)), func(t *testing.T) {
			t.Parallel()

			if _, err := l.MarshalText(); err == nil {
				t.Fatalf("MarshalText(%v) returned nil error, want non-nil", l)
			}
		})
	}
}

// TestLevel_UnmarshalText verifies that UnmarshalText parses known levels
// and leaves the receiver unchanged on error.
func TestLevel_UnmarshalText(t *testing.T) {
	t.Parallel()

	// Successful cases.
	okCases := []struct {
		input  string
		expect level.Level
	}{
		{"trace", level.Trace},
		{"DEBUG", level.Debug},
		{" info ", level.Info},
		{"WARNING", level.Warn},
		{"err", level.Error},
		{"crit", level.Critical},
		{"fatal", level.Fatal},
	}

	for _, tt := range okCases {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			var l level.Level
			// Initialize to a known different value to ensure it is overwritten.
			l = level.Trace

			if err := l.UnmarshalText([]byte(tt.input)); err != nil {
				t.Fatalf("UnmarshalText(%q) returned error: %v", tt.input, err)
			}
			if l != tt.expect {
				t.Fatalf("after UnmarshalText(%q), level = %v, want %v", tt.input, l, tt.expect)
			}
		})
	}

	// Failure case: receiver must remain unchanged.
	t.Run("invalid-input", func(t *testing.T) {
		t.Parallel()

		var l level.Level = level.Info
		err := l.UnmarshalText([]byte("not-a-level"))
		if err == nil {
			t.Fatalf("UnmarshalText(\"not-a-level\") returned nil error, want non-nil")
		}
		if l != level.Info {
			t.Fatalf("UnmarshalText on invalid input modified receiver: got %v, want %v", l, level.Info)
		}
	})
}

// TestLevel_IsValid verifies that IsValid returns true for all defined
// levels and false for Invalid and out-of-range values.
func TestLevel_IsValid(t *testing.T) {
	t.Parallel()

	valid := []level.Level{
		level.Trace,
		level.Debug,
		level.Info,
		level.Notice,
		level.Warn,
		level.Error,
		level.Critical,
		level.Fatal,
	}

	for _, l := range valid {
		l := l
		t.Run(l.String(), func(t *testing.T) {
			t.Parallel()

			if !l.IsValid() {
				t.Fatalf("IsValid() for %v = false, want true", l)
			}
		})
	}

	invalid := []level.Level{
		level.Invalid,
		level.Level(127),
		level.Level(-100),
	}

	for _, l := range invalid {
		l := l
		t.Run(fmt.Sprintf("invalid-%d", int8(l)), func(t *testing.T) {
			t.Parallel()

			if l.IsValid() {
				t.Fatalf("IsValid() for %v = true, want false", l)
			}
		})
	}
}

func TestLevel_Enabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level     level.Level
		threshold level.Level
		want      bool
	}{
		// Same level is enabled
		{level.Trace, level.Trace, true},
		{level.Debug, level.Debug, true},
		{level.Info, level.Info, true},
		{level.Warn, level.Warn, true},
		{level.Error, level.Error, true},

		// Higher levels are enabled
		{level.Error, level.Info, true},
		{level.Error, level.Debug, true},
		{level.Error, level.Trace, true},
		{level.Warn, level.Info, true},
		{level.Critical, level.Error, true},
		{level.Fatal, level.Critical, true},

		// Lower levels are not enabled
		{level.Debug, level.Info, false},
		{level.Trace, level.Debug, false},
		{level.Info, level.Warn, false},
		{level.Warn, level.Error, false},
		{level.Error, level.Fatal, false},
	}

	for _, tt := range tests {
		tt := tt
		name := fmt.Sprintf("%s.Enabled(%s)", tt.level, tt.threshold)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tt.level.Enabled(tt.threshold)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLevel_Higher(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level level.Level
		other level.Level
		want  bool
	}{
		// Same level is not higher
		{level.Info, level.Info, false},
		{level.Debug, level.Debug, false},
		{level.Error, level.Error, false},

		// Higher levels
		{level.Error, level.Warn, true},
		{level.Warn, level.Info, true},
		{level.Info, level.Debug, true},
		{level.Debug, level.Trace, true},
		{level.Fatal, level.Error, true},
		{level.Critical, level.Warn, true},

		// Lower levels
		{level.Trace, level.Debug, false},
		{level.Debug, level.Info, false},
		{level.Info, level.Warn, false},
		{level.Warn, level.Error, false},
		{level.Error, level.Fatal, false},
	}

	for _, tt := range tests {
		tt := tt
		name := fmt.Sprintf("%s.Higher(%s)", tt.level, tt.other)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tt.level.Higher(tt.other)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLevel_Lower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level level.Level
		other level.Level
		want  bool
	}{
		// Same level is not lower
		{level.Info, level.Info, false},
		{level.Debug, level.Debug, false},
		{level.Error, level.Error, false},

		// Lower levels
		{level.Trace, level.Debug, true},
		{level.Debug, level.Info, true},
		{level.Info, level.Warn, true},
		{level.Warn, level.Error, true},
		{level.Error, level.Fatal, true},

		// Higher levels
		{level.Debug, level.Trace, false},
		{level.Info, level.Debug, false},
		{level.Warn, level.Info, false},
		{level.Error, level.Warn, false},
		{level.Fatal, level.Error, false},
	}

	for _, tt := range tests {
		tt := tt
		name := fmt.Sprintf("%s.Lower(%s)", tt.level, tt.other)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tt.level.Lower(tt.other)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
