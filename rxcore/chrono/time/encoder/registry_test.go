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

package timeenc_test

import (
	"strings"
	"testing"
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
	timeapi "dirpx.dev/rxlog/rxapi/chrono/time"
	timeenc "dirpx.dev/rxlog/rxcore/chrono/time/encoder"
)

// regEncodeWith is a helper that runs the given time encoder on a fresh buffer
// and returns its textual representation.
func regEncodeWith(t *testing.T, enc timeapi.Encoder, ts time.Time) string {
	t.Helper()

	dst := &buffer.Buffer{}
	out := enc(dst, ts)
	if out == nil {
		t.Fatalf("encoder returned nil buffer")
	}
	return string(out.Bytes())
}

// TestFromString_KnownNames verifies that FromString resolves all predefined
// symbolic names and that resulting encoders behave identically to the
// exported shortcut variables.
func TestFromString_KnownNames(t *testing.T) {
	t.Helper()

	// Fixed timestamp to make all encodings deterministic.
	ts := time.Date(2025, 1, 2, 3, 4, 5, 123_456_789, time.FixedZone("TEST+03", 3*60*60))

	tests := []struct {
		name string
		want timeapi.Encoder
	}{
		// ISO 8601–style layouts.
		{"iso8601", timeenc.ISO8601MillisTimeEncoder},
		{"iso8601-millis", timeenc.ISO8601MillisTimeEncoder},
		{"iso8601-sec", timeenc.ISO8601SecondsTimeEncoder},
		{"iso8601-seconds", timeenc.ISO8601SecondsTimeEncoder},
		{"iso8601-micros", timeenc.ISO8601MicrosTimeEncoder},
		{"iso8601-nanos", timeenc.ISO8601NanosTimeEncoder},

		// RFC 3339 and related.
		{"rfc3339", timeenc.RFC3339TimeEncoder},
		{"rfc3339-nano", timeenc.RFC3339NanoTimeEncoder},
		{"rfc1123", timeenc.RFC1123TimeEncoder},
		{"rfc1123z", timeenc.RFC1123ZTimeEncoder},
		{"rfc822", timeenc.RFC822TimeEncoder},
		{"rfc822z", timeenc.RFC822ZTimeEncoder},
		{"rfc850", timeenc.RFC850TimeEncoder},

		// Other textual layouts.
		{"ansic", timeenc.ANSICTimeEncoder},
		{"unixdate", timeenc.UnixDateTimeEncoder},
		{"kitchen", timeenc.KitchenTimeEncoder},
		{"stamp", timeenc.StampTimeEncoder},
		{"stamp-millis", timeenc.StampMilliTimeEncoder},
		{"stamp-micros", timeenc.StampMicroTimeEncoder},
		{"stamp-nanos", timeenc.StampNanoTimeEncoder},

		// Date / time-only.
		{"date", timeenc.DateOnlyTimeEncoder},
		{"time", timeenc.TimeOnlyTimeEncoder},
		{"time-millis", timeenc.TimeMillisOnlyTimeEncoder},

		// Unix epoch representations.
		{"unix", timeenc.UnixSecondsTimeEncoder},
		{"unix-seconds", timeenc.UnixSecondsTimeEncoder},
		{"unix-secs", timeenc.UnixSecondsTimeEncoder},
		{"unix-millis", timeenc.UnixMillisTimeEncoder},
		{"unix-micros", timeenc.UnixMicrosTimeEncoder},
		{"unix-nanos", timeenc.UnixNanosTimeEncoder},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			gotEnc, err := timeenc.FromString(tt.name)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.name, err)
			}
			if gotEnc == nil {
				t.Fatalf("FromString(%q) returned nil encoder", tt.name)
			}

			got := regEncodeWith(t, gotEnc, ts)
			want := regEncodeWith(t, tt.want, ts)
			if got != want {
				t.Fatalf("encoder(%q) output = %q, want %q", tt.name, got, want)
			}
		})
	}
}

// TestFromString_UnknownName verifies that FromString returns (nil, error)
// for unknown encoder names.
func TestFromString_UnknownName(t *testing.T) {
	t.Helper()

	const name = "no-such-time-encoder"

	enc, err := timeenc.FromString(name)
	if err == nil {
		t.Fatalf("FromString(%q) returned nil error, want non-nil", name)
	}
	if enc != nil {
		t.Fatalf("FromString(%q) returned non-nil encoder, want nil", name)
	}
	if msg := err.Error(); !strings.Contains(msg, "unknown time encoder") {
		t.Fatalf("error = %q, want substring %q", msg, "unknown time encoder")
	}
}

// TestMustFromString_Succeeds verifies that MustFromString succeeds and does
// not panic for a known encoder name.
func TestMustFromString_Succeeds(t *testing.T) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustFromString(\"rfc3339\") panicked: %v", r)
		}
	}()

	enc := timeenc.MustFromString("rfc3339")
	if enc == nil {
		t.Fatalf("MustFromString(\"rfc3339\") returned nil encoder")
	}

	// Sanity-check behavior.
	ts := time.Now()
	got := regEncodeWith(t, enc, ts)
	want := regEncodeWith(t, timeenc.RFC3339TimeEncoder, ts)
	if got != want {
		t.Fatalf("MustFromString(\"rfc3339\") encoder output = %q, want %q", got, want)
	}
}

// TestMustFromString_PanicsOnUnknown verifies that MustFromString panics
// if the encoder name is not registered.
func TestMustFromString_PanicsOnUnknown(t *testing.T) {
	t.Helper()

	const name = "definitely-not-registered"

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustFromString(%q) did not panic for unknown encoder", name)
		}
	}()

	_ = timeenc.MustFromString(name)
}

// TestRegister_AddAndOverride verifies that Register can add encoders under
// a new name and override an existing registration.
func TestRegister_AddAndOverride(t *testing.T) {
	t.Helper()

	const name = "custom-test-time-encoder"

	// e1: prefix "e1:" + Unix seconds.
	e1 := func(dst *buffer.Buffer, t time.Time) *buffer.Buffer {
		dst.AppendString("e1:")
		dst.AppendInt64(t.Unix())
		return dst
	}

	// e2: prefix "e2:" + Unix milliseconds.
	e2 := func(dst *buffer.Buffer, t time.Time) *buffer.Buffer {
		dst.AppendString("e2:")
		dst.AppendInt64(t.UnixMilli())
		return dst
	}

	ts := time.Date(2025, 5, 6, 7, 8, 9, 0, time.UTC)

	// First registration.
	timeenc.Register(name, e1)
	enc1, err := timeenc.FromString(name)
	if err != nil {
		t.Fatalf("FromString(%q) after first Register returned error: %v", name, err)
	}
	got1 := regEncodeWith(t, enc1, ts)
	want1 := regEncodeWith(t, e1, ts)
	if got1 != want1 {
		t.Fatalf("encoder(%q) after first Register = %q, want %q", name, got1, want1)
	}

	// Override registration.
	timeenc.Register(name, e2)
	enc2, err := timeenc.FromString(name)
	if err != nil {
		t.Fatalf("FromString(%q) after override returned error: %v", name, err)
	}
	got2 := regEncodeWith(t, enc2, ts)
	want2 := regEncodeWith(t, e2, ts)
	if got2 != want2 {
		t.Fatalf("encoder(%q) after override = %q, want %q", name, got2, want2)
	}

	// Ensure behavior actually changed.
	if got1 == got2 {
		t.Fatalf("override appears to have no effect: before=%q, after=%q", got1, got2)
	}
}
