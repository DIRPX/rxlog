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

package level_test

import (
	"strings"
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/level"
	levelpkg "dirpx.dev/rxlog/rxcore/level"
)

// encodeWith is a helper that runs the given level encoder on a fresh buffer
// and returns the resulting string.
func encodeWith(t *testing.T, enc level.Encoder, lvl level.Level) string {
	t.Helper()

	dst := &buffer.Buffer{}
	out := enc(dst, lvl)
	if out == nil {
		t.Fatalf("encoder returned nil buffer")
	}

	return string(out.Bytes())
}

// TestLowercaseLevelEncoder_AllLevels verifies that LowercaseLevelEncoder
// produces the expected lowercase string for each defined level.
func TestLowercaseLevelEncoder_AllLevels(t *testing.T) {
	t.Helper()

	tests := []struct {
		level level.Level
		want  string
	}{
		{level.Trace, "trace"},
		{level.Debug, "debug"},
		{level.Info, "info"},
		{level.Notice, "notice"},
		{level.Warn, "warn"},
		{level.Error, "error"},
		{level.Critical, "critical"},
		{level.Fatal, "fatal"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := encodeWith(t, levelpkg.LowercaseLevelEncoder, tt.level)
			if got != tt.want {
				t.Errorf("LowercaseLevelEncoder(%v) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

// TestLowercaseLevelEncoder_AppendsToExistingBuffer verifies that the encoder
// appends its output to the existing contents of the buffer rather than
// replacing them.
func TestLowercaseLevelEncoder_AppendsToExistingBuffer(t *testing.T) {
	t.Helper()

	dst := &buffer.Buffer{}
	dst.AppendString("prefix:")

	out := levelpkg.LowercaseLevelEncoder(dst, level.Info)
	if out == nil {
		t.Fatal("LowercaseLevelEncoder returned nil buffer")
	}

	got := string(out.Bytes())
	want := "prefix:info"
	if got != want {
		t.Errorf("buffer after encoding = %q, want %q", got, want)
	}
}

// TestLowercaseLevelEncoder_CanBeReusedAcrossCalls verifies that a single
// encoder instance is safe to reuse across multiple calls with different
// levels, and that it does not retain per-call state between invocations.
func TestLowercaseLevelEncoder_CanBeReusedAcrossCalls(t *testing.T) {
	t.Helper()

	enc := levelpkg.LowercaseLevelEncoder

	// First call.
	buf1 := &buffer.Buffer{}
	out1 := enc(buf1, level.Info)
	if out1 == nil {
		t.Fatal("first call to encoder returned nil buffer")
	}
	got1 := string(out1.Bytes())
	want1 := "info"
	if got1 != want1 {
		t.Errorf("first call: got %q, want %q", got1, want1)
	}

	// Second call with different level and a fresh buffer.
	buf2 := &buffer.Buffer{}
	out2 := enc(buf2, level.Error)
	if out2 == nil {
		t.Fatal("second call to encoder returned nil buffer")
	}
	got2 := string(out2.Bytes())
	want2 := "error"
	if got2 != want2 {
		t.Errorf("second call: got %q, want %q", got2, want2)
	}
}

// TestCapitalLevelEncoder_AllLevels verifies that CapitalLevelEncoder
// produces the expected uppercase string for each defined level.
func TestCapitalLevelEncoder_AllLevels(t *testing.T) {
	t.Helper()

	tests := []struct {
		level level.Level
		want  string
	}{
		{level.Trace, "TRACE"},
		{level.Debug, "DEBUG"},
		{level.Info, "INFO"},
		{level.Notice, "NOTICE"},
		{level.Warn, "WARN"},
		{level.Error, "ERROR"},
		{level.Critical, "CRITICAL"},
		{level.Fatal, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := encodeWith(t, levelpkg.CapitalLevelEncoder, tt.level)
			if got != tt.want {
				t.Errorf("CapitalLevelEncoder(%v) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

// TestCapitalLevelEncoder_AppendsToExistingBuffer verifies that the encoder
// appends its output to the existing contents of the buffer rather than
// replacing them.
func TestCapitalLevelEncoder_AppendsToExistingBuffer(t *testing.T) {
	t.Helper()

	dst := &buffer.Buffer{}
	dst.AppendString("prefix:")

	out := levelpkg.CapitalLevelEncoder(dst, level.Warn)
	if out == nil {
		t.Fatal("CapitalLevelEncoder returned nil buffer")
	}

	got := string(out.Bytes())
	want := "prefix:WARN"
	if got != want {
		t.Errorf("buffer after encoding = %q, want %q", got, want)
	}
}

// TestCapitalLevelEncoder_CanBeReusedAcrossCalls verifies that a single
// encoder instance is safe to reuse across multiple calls with different
// levels, and that it does not retain per-call state between invocations.
func TestCapitalLevelEncoder_CanBeReusedAcrossCalls(t *testing.T) {
	t.Helper()

	enc := levelpkg.CapitalLevelEncoder

	// First call.
	buf1 := &buffer.Buffer{}
	out1 := enc(buf1, level.Debug)
	if out1 == nil {
		t.Fatal("first call to encoder returned nil buffer")
	}
	got1 := string(out1.Bytes())
	want1 := "DEBUG"
	if got1 != want1 {
		t.Errorf("first call: got %q, want %q", got1, want1)
	}

	// Second call with different level and a fresh buffer.
	buf2 := &buffer.Buffer{}
	out2 := enc(buf2, level.Critical)
	if out2 == nil {
		t.Fatal("second call to encoder returned nil buffer")
	}
	got2 := string(out2.Bytes())
	want2 := "CRITICAL"
	if got2 != want2 {
		t.Errorf("second call: got %q, want %q", got2, want2)
	}
}

// TestLowercaseVsCapital_ProduceConsistentCase verifies that both encoders
// produce the same text modulo case transformation.
func TestLowercaseVsCapital_ProduceConsistentCase(t *testing.T) {
	t.Helper()

	levels := []level.Level{
		level.Trace, level.Debug, level.Info, level.Notice,
		level.Warn, level.Error, level.Critical, level.Fatal,
	}

	for _, lvl := range levels {
		t.Run(lvl.String(), func(t *testing.T) {
			lower := encodeWith(t, levelpkg.LowercaseLevelEncoder, lvl)
			upper := encodeWith(t, levelpkg.CapitalLevelEncoder, lvl)

			if strings.ToUpper(lower) != upper {
				t.Errorf("LowercaseLevelEncoder(%v)=%q uppercased to %q != CapitalLevelEncoder(%v)=%q",
					lvl, lower, strings.ToUpper(lower), lvl, upper)
			}
		})
	}
}
