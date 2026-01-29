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

package time_test

import (
	"strconv"
	"strings"
	"testing"
	stdtime "time"

	"dirpx.dev/rxlog/rxapi/buffer"
	timeapi "dirpx.dev/rxlog/rxapi/chrono/time"
	timepkg "dirpx.dev/rxlog/rxcore/chrono/time"
	"dirpx.dev/rxlog/rxcore/chrono/time/layout"
)

// TestTextEncoders_LayoutsAndLocations verifies that all layout-based
// shortcut encoders format timestamps using the expected layout and
// location semantics.
func TestTextEncoders_LayoutsAndLocations(t *testing.T) {
	t.Helper()

	// Fixed timestamp so that all formats are deterministic.
	ts := stdtime.Date(2025, 1, 2, 3, 4, 5, 123_456_789, stdtime.FixedZone("TEST+03", 3*60*60))

	tests := []struct {
		name   string
		enc    timeapi.Encoder
		layout string
		loc    *stdtime.Location // nil means "preserve ts.Location"
	}{
		// ISO 8601 variants (UTC).
		{"iso8601-seconds", timepkg.ISO8601SecondsTimeEncoder, layout.ISO8601Seconds, stdtime.UTC},
		{"iso8601-millis", timepkg.ISO8601MillisTimeEncoder, layout.ISO8601Millis, stdtime.UTC},
		{"iso8601-micros", timepkg.ISO8601MicrosTimeEncoder, layout.ISO8601Micros, stdtime.UTC},
		{"iso8601-nanos", timepkg.ISO8601NanosTimeEncoder, layout.ISO8601Nanos, stdtime.UTC},

		// RFC 3339 family (UTC).
		{"rfc3339", timepkg.RFC3339TimeEncoder, layout.RFC3339, stdtime.UTC},
		{"rfc3339-nano", timepkg.RFC3339NanoTimeEncoder, layout.RFC3339Nano, stdtime.UTC},

		// Other textual layouts.
		{"rfc1123", timepkg.RFC1123TimeEncoder, layout.RFC1123, stdtime.UTC},
		{"rfc1123z", timepkg.RFC1123ZTimeEncoder, layout.RFC1123Z, stdtime.UTC},
		{"rfc822", timepkg.RFC822TimeEncoder, layout.RFC822, stdtime.UTC},
		{"rfc822z", timepkg.RFC822ZTimeEncoder, layout.RFC822Z, stdtime.UTC},
		{"rfc850", timepkg.RFC850TimeEncoder, layout.RFC850, stdtime.UTC},

		// ANSIC / UnixDate use stdtime.Local.
		{"ansic", timepkg.ANSICTimeEncoder, layout.ANSIC, stdtime.Local},
		{"unixdate", timepkg.UnixDateTimeEncoder, layout.UnixDate, stdtime.Local},

		// "Nil location" variants preserve ts.Location.
		{"kitchen", timepkg.KitchenTimeEncoder, layout.Kitchen, nil},
		{"stamp", timepkg.StampTimeEncoder, layout.Stamp, nil},
		{"stamp-milli", timepkg.StampMilliTimeEncoder, layout.StampMilli, nil},
		{"stamp-micro", timepkg.StampMicroTimeEncoder, layout.StampMicro, nil},
		{"stamp-nano", timepkg.StampNanoTimeEncoder, layout.StampNano, nil},

		// Date / time-only encoders (UTC).
		{"date-only", timepkg.DateOnlyTimeEncoder, layout.Date, stdtime.UTC},
		{"time-only", timepkg.TimeOnlyTimeEncoder, layout.Time, stdtime.UTC},
		{"time-millis-only", timepkg.TimeMillisOnlyTimeEncoder, layout.TimeMillis, stdtime.UTC},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			dst := &buffer.Buffer{}
			out := tt.enc(dst, ts)
			if out == nil {
				t.Fatalf("encoder %q returned nil buffer", tt.name)
			}

			got := string(out.Bytes())

			var want string
			if tt.loc == nil {
				// Preserve original timestamp Location.
				want = ts.Format(tt.layout)
			} else {
				want = ts.In(tt.loc).Format(tt.layout)
			}

			if got != want {
				t.Fatalf("encoder %q = %q, want %q", tt.name, got, want)
			}
		})
	}
}

// TestUnixEpochEncoders_Values verifies that numeric Unix* encoders emit
// integer timestamps in the expected units.
func TestUnixEpochEncoders_Values(t *testing.T) {
	t.Helper()

	// Non-trivial timestamp so all units differ.
	ts := stdtime.Unix(1672531200, 123_456_789) // seconds + nanos

	tests := []struct {
		name string
		enc  timeapi.Encoder
		want int64
	}{
		{"unix-seconds", timepkg.UnixSecondsTimeEncoder, ts.Unix()},
		{"unix-millis", timepkg.UnixMillisTimeEncoder, ts.UnixMilli()},
		{"unix-micros", timepkg.UnixMicrosTimeEncoder, ts.UnixMicro()},
		{"unix-nanos", timepkg.UnixNanosTimeEncoder, ts.UnixNano()},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			dst := &buffer.Buffer{}
			out := tt.enc(dst, ts)
			if out == nil {
				t.Fatalf("encoder %q returned nil buffer", tt.name)
			}

			s := string(out.Bytes())
			got, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				t.Fatalf("failed to parse %q as int64: %v", s, err)
			}
			if got != tt.want {
				t.Fatalf("encoder %q = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

// TestUnixSecondsTimeEncoder_AppendsToExistingBuffer verifies that the
// numeric encoders append to existing buffer contents instead of replacing
// them.
func TestUnixSecondsTimeEncoder_AppendsToExistingBuffer(t *testing.T) {
	t.Helper()

	ts := stdtime.Unix(1000, 0) // 1000 seconds after epoch

	dst := &buffer.Buffer{}
	dst.AppendString("prefix:")

	out := timepkg.UnixSecondsTimeEncoder(dst, ts)
	if out == nil {
		t.Fatal("UnixSecondsTimeEncoder returned nil buffer")
	}

	s := string(out.Bytes())
	if !strings.HasPrefix(s, "prefix:") {
		t.Fatalf("buffer after encoding = %q, want prefix %q", s, "prefix:")
	}

	numPart := strings.TrimPrefix(s, "prefix:")
	got, err := strconv.ParseInt(numPart, 10, 64)
	if err != nil {
		t.Fatalf("failed to parse numeric part %q: %v", numPart, err)
	}

	if want := ts.Unix(); got != want {
		t.Fatalf("UnixSecondsTimeEncoder appended %d, want %d", got, want)
	}
}
