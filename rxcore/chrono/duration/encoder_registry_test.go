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

package durenc_test

import (
	"strings"
	"testing"
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
	durationapi "dirpx.dev/rxlog/rxapi/chrono/duration"
	durenc "dirpx.dev/rxlog/rxcore/chrono/duration/encoder"
)

// encodeWith is a helper that runs the given encoder on a fresh buffer and
// returns its textual representation. It is used to compare encoders by
// behavior rather than by function identity (which is not comparable).
func encodeWith(t *testing.T, enc durationapi.Encoder, d time.Duration) string {
	t.Helper()

	dst := &buffer.Buffer{}
	out := enc(dst, d)
	if out == nil {
		t.Fatalf("encoder returned nil buffer")
	}
	return string(out.Bytes())
}

// TestFromString_KnownNames verifies that FromString resolves known symbolic
// names and that the resulting encoders behave identically to the exported
// shortcut variables (StringDurationEncoder, SecondsDurationEncoder, etc.).
func TestFromString_KnownNames(t *testing.T) {
	// Non-trivial duration to exercise all encoders.
	d := 1500 * time.Millisecond // 1.5s

	tests := []struct {
		name string
		want durationapi.Encoder
	}{
		// String representation.
		{"string", durenc.StringDurationEncoder},

		// Seconds variants.
		{"seconds", durenc.SecondsDurationEncoder},
		{"secs", durenc.SecondsDurationEncoder},
		{"seconds-f64", durenc.SecondsDurationEncoder},
		{"secs-float64", durenc.SecondsDurationEncoder},

		// Millis variants.
		{"millis", durenc.MillisDurationEncoder},
		{"milliseconds", durenc.MillisDurationEncoder},
		{"ms", durenc.MillisDurationEncoder},
		{"millis-int64", durenc.MillisDurationEncoder},
		{"ms-int64", durenc.MillisDurationEncoder},

		// Micros variants (incl. Unicode µs).
		{"micros", durenc.MicrosDurationEncoder},
		{"microseconds", durenc.MicrosDurationEncoder},
		{"µs", durenc.MicrosDurationEncoder},
		{"micros-int64", durenc.MicrosDurationEncoder},
		{"microsecs-int64", durenc.MicrosDurationEncoder},

		// Nanos variants.
		{"nanos", durenc.NanosDurationEncoder},
		{"nanoseconds", durenc.NanosDurationEncoder},
		{"ns", durenc.NanosDurationEncoder},
		{"nanos-int64", durenc.NanosDurationEncoder},
		{"nanosecs-int64", durenc.NanosDurationEncoder},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotEnc, err := durenc.FromString(tt.name)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.name, err)
			}
			if gotEnc == nil {
				t.Fatalf("FromString(%q) returned nil encoder", tt.name)
			}

			got := encodeWith(t, gotEnc, d)
			want := encodeWith(t, tt.want, d)
			if got != want {
				t.Fatalf("encoder(%q) output = %q, want %q", tt.name, got, want)
			}
		})
	}
}

// TestFromString_UnknownName verifies that FromString returns a nil encoder
// and a non-nil error for unknown names.
func TestFromString_UnknownName(t *testing.T) {
	const name = "no-such-encoder"

	enc, err := durenc.FromString(name)
	if err == nil {
		t.Fatalf("FromString(%q) returned nil error, want non-nil", name)
	}
	if enc != nil {
		t.Fatalf("FromString(%q) returned non-nil encoder, want nil", name)
	}
	if msg := err.Error(); !strings.Contains(msg, "unknown duration encoder") {
		t.Fatalf("FromString(%q) error = %q, want substring %q", name, msg, "unknown duration encoder")
	}
}

// TestMustFromString_Succeeds ensures MustFromString does not panic and
// returns a non-nil encoder for a known name.
func TestMustFromString_Succeeds(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustFromString(\"string\") panicked: %v", r)
		}
	}()

	enc := durenc.MustFromString("string")
	if enc == nil {
		t.Fatalf("MustFromString(\"string\") returned nil encoder")
	}

	// Sanity-check: encoder should behave like StringDurationEncoder.
	d := 250 * time.Millisecond
	got := encodeWith(t, enc, d)
	want := encodeWith(t, durenc.StringDurationEncoder, d)
	if got != want {
		t.Fatalf("MustFromString(\"string\") encoder output = %q, want %q", got, want)
	}
}

// TestMustFromString_PanicsOnUnknown verifies that MustFromString panics
// when asked to resolve an unknown encoder name.
func TestMustFromString_PanicsOnUnknown(t *testing.T) {
	const name = "definitely-not-registered"

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustFromString(%q) did not panic for unknown encoder", name)
		}
	}()

	_ = durenc.MustFromString(name)
}

// TestRegister_AddAndOverride verifies that Register can both install a new
// encoder under a previously unused name and override an existing entry.
func TestRegister_AddAndOverride(t *testing.T) {
	const name = "custom-test-encoder"

	// e1 encodes a fixed prefix and the duration in nanoseconds.
	e1 := func(dst *buffer.Buffer, d time.Duration) *buffer.Buffer {
		dst.AppendString("e1:")
		dst.AppendInt64(d.Nanoseconds())
		return dst
	}

	// e2 encodes a different prefix and the duration in milliseconds.
	e2 := func(dst *buffer.Buffer, d time.Duration) *buffer.Buffer {
		dst.AppendString("e2:")
		dst.AppendInt64(d.Milliseconds())
		return dst
	}

	d := 1234 * time.Millisecond

	// First registration: name -> e1.
	durenc.Register(name, e1)
	enc1, err := durenc.FromString(name)
	if err != nil {
		t.Fatalf("FromString(%q) after first Register returned error: %v", name, err)
	}
	if got, want := encodeWith(t, enc1, d), encodeWith(t, e1, d); got != want {
		t.Fatalf("encoder(%q) after first Register = %q, want %q", name, got, want)
	}

	// Override registration: name -> e2.
	durenc.Register(name, e2)
	enc2, err := durenc.FromString(name)
	if err != nil {
		t.Fatalf("FromString(%q) after override returned error: %v", name, err)
	}
	got := encodeWith(t, enc2, d)
	want := encodeWith(t, e2, d)
	if got != want {
		t.Fatalf("encoder(%q) after override = %q, want %q", name, got, want)
	}

	// Ensure we actually changed behavior relative to e1 for this duration.
	if prev := encodeWith(t, e1, d); got == prev {
		t.Fatalf("override appears to have no effect: prev=%q, got=%q", prev, got)
	}
}
