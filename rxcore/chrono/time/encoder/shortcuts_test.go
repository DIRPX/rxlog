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
	"strconv"
	"strings"
	"testing"
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
	timeapi "dirpx.dev/rxlog/rxapi/chrono/time"
	timeenc "dirpx.dev/rxlog/rxcore/chrono/time/encoder"
	"dirpx.dev/rxlog/rxcore/chrono/time/encoder/layout"
)

// TestTextEncoders_LayoutsAndLocations verifies that all layout-based
// shortcut encoders format timestamps using the expected layout and
// location semantics.
func TestTextEncoders_LayoutsAndLocations(t *testing.T) {
	t.Helper()

	// Fixed timestamp so that all formats are deterministic.
	ts := time.Date(2025, 1, 2, 3, 4, 5, 123_456_789, time.FixedZone("TEST+03", 3*60*60))

	tests := []struct {
		name   string
		enc    timeapi.Encoder
		layout string
		loc    *time.Location // nil means "preserve ts.Location"
	}{
		// ISO 8601 variants (UTC).
		{"iso8601-seconds", timeenc.ISO8601SecondsTimeEncoder, layout.ISO8601Seconds, time.UTC},
		{"iso8601-millis", timeenc.ISO8601MillisTimeEncoder, layout.ISO8601Millis, time.UTC},
		{"iso8601-micros", timeenc.ISO8601MicrosTimeEncoder, layout.ISO8601Micros, time.UTC},
		{"iso8601-nanos", timeenc.ISO8601NanosTimeEncoder, layout.ISO8601Nanos, time.UTC},

		// RFC 3339 family (UTC).
		{"rfc3339", timeenc.RFC3339TimeEncoder, layout.RFC3339, time.UTC},
		{"rfc3339-nano", timeenc.RFC3339NanoTimeEncoder, layout.RFC3339Nano, time.UTC},

		// Other textual layouts.
		{"rfc1123", timeenc.RFC1123TimeEncoder, layout.RFC1123, time.UTC},
		{"rfc1123z", timeenc.RFC1123ZTimeEncoder, layout.RFC1123Z, time.UTC},
		{"rfc822", timeenc.RFC822TimeEncoder, layout.RFC822, time.UTC},
		{"rfc822z", timeenc.RFC822ZTimeEncoder, layout.RFC822Z, time.UTC},
		{"rfc850", timeenc.RFC850TimeEncoder, layout.RFC850, time.UTC},

		// ANSIC / UnixDate use time.Local.
		{"ansic", timeenc.ANSICTimeEncoder, layout.ANSIC, time.Local},
		{"unixdate", timeenc.UnixDateTimeEncoder, layout.UnixDate, time.Local},

		// "Nil location" variants preserve ts.Location.
		{"kitchen", timeenc.KitchenTimeEncoder, layout.Kitchen, nil},
		{"stamp", timeenc.StampTimeEncoder, layout.Stamp, nil},
		{"stamp-milli", timeenc.StampMilliTimeEncoder, layout.StampMilli, nil},
		{"stamp-micro", timeenc.StampMicroTimeEncoder, layout.StampMicro, nil},
		{"stamp-nano", timeenc.StampNanoTimeEncoder, layout.StampNano, nil},

		// Date / time-only encoders (UTC).
		{"date-only", timeenc.DateOnlyTimeEncoder, layout.Date, time.UTC},
		{"time-only", timeenc.TimeOnlyTimeEncoder, layout.Time, time.UTC},
		{"time-millis-only", timeenc.TimeMillisOnlyTimeEncoder, layout.TimeMillis, time.UTC},
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
	ts := time.Unix(1672531200, 123_456_789) // seconds + nanos

	tests := []struct {
		name string
		enc  timeapi.Encoder
		want int64
	}{
		{"unix-seconds", timeenc.UnixSecondsTimeEncoder, ts.Unix()},
		{"unix-millis", timeenc.UnixMillisTimeEncoder, ts.UnixMilli()},
		{"unix-micros", timeenc.UnixMicrosTimeEncoder, ts.UnixMicro()},
		{"unix-nanos", timeenc.UnixNanosTimeEncoder, ts.UnixNano()},
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

	ts := time.Unix(1000, 0) // 1000 seconds after epoch

	dst := &buffer.Buffer{}
	dst.AppendString("prefix:")

	out := timeenc.UnixSecondsTimeEncoder(dst, ts)
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
