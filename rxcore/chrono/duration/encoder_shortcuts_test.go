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

package duration_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
	duration "dirpx.dev/rxlog/rxcore/chrono/duration"
)

func TestSecondsDurationEncoder_AppendsSecondsAsFloat(t *testing.T) {
	t.Helper()

	d := 1500 * time.Millisecond // 1.5 seconds

	dst := &buffer.Buffer{}
	dst.AppendString("prefix:")

	out := duration.SecondsDurationEncoder(dst, d)
	if out == nil {
		t.Fatalf("SecondsDurationEncoder returned nil buffer")
	}

	s := string(out.Bytes())
	if !strings.HasPrefix(s, "prefix:") {
		t.Fatalf("buffer after encoding does not preserve prefix: %q", s)
	}

	numPart := strings.TrimPrefix(s, "prefix:")
	if numPart == "" {
		t.Fatalf("no numeric part after prefix in %q", s)
	}

	got, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		t.Fatalf("failed to parse float %q: %v", numPart, err)
	}

	want := float64(d) / float64(time.Second)
	const eps = 1e-9
	if diff := got - want; diff < -eps || diff > eps {
		t.Fatalf("SecondsDurationEncoder = %v, want %v (diff=%v)", got, want, diff)
	}
}

func TestMillisDurationEncoder_AppendsMillisAsInt(t *testing.T) {
	t.Helper()

	d := 2*time.Second + 500*time.Millisecond // 2500 ms

	dst := &buffer.Buffer{}
	dst.AppendString("m:")

	out := duration.MillisDurationEncoder(dst, d)
	if out == nil {
		t.Fatalf("MillisDurationEncoder returned nil buffer")
	}

	s := string(out.Bytes())
	if !strings.HasPrefix(s, "m:") {
		t.Fatalf("buffer after encoding does not preserve prefix: %q", s)
	}

	numPart := strings.TrimPrefix(s, "m:")
	want := d.Milliseconds()
	if numPart != strconv.FormatInt(want, 10) {
		t.Fatalf("MillisDurationEncoder = %q, want %q", numPart, strconv.FormatInt(want, 10))
	}
}

func TestMicrosDurationEncoder_AppendsMicrosAsInt(t *testing.T) {
	t.Helper()

	d := 2*time.Second + 500*time.Millisecond // 2_500_000 µs

	dst := &buffer.Buffer{}
	dst.AppendString("u:")

	out := duration.MicrosDurationEncoder(dst, d)
	if out == nil {
		t.Fatalf("MicrosDurationEncoder returned nil buffer")
	}

	s := string(out.Bytes())
	if !strings.HasPrefix(s, "u:") {
		t.Fatalf("buffer after encoding does not preserve prefix: %q", s)
	}

	numPart := strings.TrimPrefix(s, "u:")
	want := d.Microseconds()
	if numPart != strconv.FormatInt(want, 10) {
		t.Fatalf("MicrosDurationEncoder = %q, want %q", numPart, strconv.FormatInt(want, 10))
	}
}

func TestNanosDurationEncoder_AppendsNanosAsInt(t *testing.T) {
	t.Helper()

	d := 2*time.Second + 500*time.Millisecond

	dst := &buffer.Buffer{}
	dst.AppendString("n:")

	out := duration.NanosDurationEncoder(dst, d)
	if out == nil {
		t.Fatalf("NanosDurationEncoder returned nil buffer")
	}

	s := string(out.Bytes())
	if !strings.HasPrefix(s, "n:") {
		t.Fatalf("buffer after encoding does not preserve prefix: %q", s)
	}

	numPart := strings.TrimPrefix(s, "n:")
	want := d.Nanoseconds()
	if numPart != strconv.FormatInt(want, 10) {
		t.Fatalf("NanosDurationEncoder = %q, want %q", numPart, strconv.FormatInt(want, 10))
	}
}

func TestStringDurationEncoder_AppendsDurationString(t *testing.T) {
	t.Helper()

	d := 2*time.Second + 150*time.Millisecond

	dst := &buffer.Buffer{}
	dst.AppendString("s:")

	out := duration.StringDurationEncoder(dst, d)
	if out == nil {
		t.Fatalf("StringDurationEncoder returned nil buffer")
	}

	s := string(out.Bytes())
	if !strings.HasPrefix(s, "s:") {
		t.Fatalf("buffer after encoding does not preserve prefix: %q", s)
	}

	numPart := strings.TrimPrefix(s, "s:")
	want := d.String()
	if numPart != want {
		t.Fatalf("StringDurationEncoder = %q, want %q", numPart, want)
	}
}
