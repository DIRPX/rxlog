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
	"testing"
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
	timeenc "dirpx.dev/rxlog/rxcore/chrono/time/encoder"
)

// lEncodeWith is a helper that runs the given layout encoder on a fresh buffer
// and returns the resulting string.
func lEncodeWith(t *testing.T, layout string, loc *time.Location, ts time.Time) string {
	t.Helper()

	enc := timeenc.TimeLayoutEncoder(layout, loc)
	dst := &buffer.Buffer{}

	out := enc(dst, ts)
	if out == nil {
		t.Fatalf("TimeLayoutEncoder(%q, %v) returned nil buffer", layout, loc)
	}

	return string(out.Bytes())
}

// TestTimeLayoutEncoder_PreservesOriginalLocationWhenLocNil verifies that
// passing a nil location causes the encoder to format the timestamp using its
// existing Location without any conversion.
func TestTimeLayoutEncoder_PreservesOriginalLocationWhenLocNil(t *testing.T) {
	t.Helper()

	layout := "2006-01-02T15:04:05Z07:00"
	loc := time.FixedZone("TEST+02", 2*60*60)

	ts := time.Date(2025, 1, 2, 3, 4, 5, 0, loc)

	got := lEncodeWith(t, layout, nil, ts)
	want := ts.Format(layout)

	if got != want {
		t.Fatalf("TimeLayoutEncoder(layout, nil) = %q, want %q", got, want)
	}
}

// TestTimeLayoutEncoder_UsesProvidedLocation verifies that when a non-nil
// location is supplied, the encoder first converts the timestamp to that
// location via t.In(loc) before formatting.
func TestTimeLayoutEncoder_UsesProvidedLocation(t *testing.T) {
	t.Helper()

	layout := "2006-01-02T15:04:05Z07:00"
	locUTC := time.UTC
	locPlus3 := time.FixedZone("TEST+03", 3*60*60)

	// Start with a timestamp in UTC so that conversion to locPlus3 is visible
	// in the formatted output (offset changes from +00:00 to +03:00).
	ts := time.Date(2025, 1, 2, 3, 4, 5, 0, locUTC)

	got := lEncodeWith(t, layout, locPlus3, ts)
	want := ts.In(locPlus3).Format(layout)

	if got != want {
		t.Fatalf("TimeLayoutEncoder(layout, loc+3) = %q, want %q", got, want)
	}
}

// TestTimeLayoutEncoder_AppendsToExistingBuffer verifies that the encoder
// appends its formatted output to the existing contents of the buffer rather
// than replacing them.
func TestTimeLayoutEncoder_AppendsToExistingBuffer(t *testing.T) {
	t.Helper()

	layout := "2006-01-02 15:04:05"
	ts := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	enc := timeenc.TimeLayoutEncoder(layout, time.UTC)

	dst := &buffer.Buffer{}
	dst.AppendString("prefix:")

	out := enc(dst, ts)
	if out == nil {
		t.Fatal("TimeLayoutEncoder returned nil buffer")
	}

	wantSuffix := ts.In(time.UTC).Format(layout)
	got := string(out.Bytes())

	want := "prefix:" + wantSuffix
	if got != want {
		t.Fatalf("buffer after encoding = %q, want %q", got, want)
	}
}

// TestTimeLayoutEncoder_CanBeReusedAcrossCalls verifies that a single encoder
// instance is safe to reuse across multiple calls with different timestamps,
// and that it does not retain per-call state between invocations.
func TestTimeLayoutEncoder_CanBeReusedAcrossCalls(t *testing.T) {
	t.Helper()

	layout := time.RFC3339
	loc := time.UTC

	enc := timeenc.TimeLayoutEncoder(layout, loc)

	ts1 := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	ts2 := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)

	// First call.
	buf1 := &buffer.Buffer{}
	out1 := enc(buf1, ts1)
	if out1 == nil {
		t.Fatal("first call to encoder returned nil buffer")
	}
	got1 := string(out1.Bytes())
	want1 := ts1.In(loc).Format(layout)
	if got1 != want1 {
		t.Fatalf("first call: got %q, want %q", got1, want1)
	}

	// Second call with different timestamp and a fresh buffer.
	buf2 := &buffer.Buffer{}
	out2 := enc(buf2, ts2)
	if out2 == nil {
		t.Fatal("second call to encoder returned nil buffer")
	}
	got2 := string(out2.Bytes())
	want2 := ts2.In(loc).Format(layout)
	if got2 != want2 {
		t.Fatalf("second call: got %q, want %q", got2, want2)
	}
}
