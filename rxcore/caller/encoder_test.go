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

package caller_test

import (
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/caller"
	callerpkg "dirpx.dev/rxlog/rxcore/caller"
)

// encodeWith is a helper that runs the given caller encoder on a fresh buffer
// and returns the resulting string.
func encodeWith(t *testing.T, enc caller.Encoder, c caller.Caller) string {
	t.Helper()

	dst := &buffer.Buffer{}
	out := enc(dst, c)
	if out == nil {
		t.Fatalf("encoder returned nil buffer")
	}

	return string(out.Bytes())
}

// TestShortCallerEncoder_DefinedCaller verifies that ShortCallerEncoder
// produces the expected output with only the base filename and line number.
func TestShortCallerEncoder_DefinedCaller(t *testing.T) {
	t.Helper()

	c := caller.Caller{
		Defined:  true,
		PC:       0,
		File:     "/app/internal/handlers/user.go",
		Line:     42,
		Function: "main.handleUser",
	}

	got := encodeWith(t, callerpkg.ShortCallerEncoder, c)
	want := "user.go:42"
	if got != want {
		t.Errorf("ShortCallerEncoder = %q, want %q", got, want)
	}
}

// TestShortCallerEncoder_UndefinedCaller verifies that ShortCallerEncoder
// produces "<undefined>" when caller.Defined is false.
func TestShortCallerEncoder_UndefinedCaller(t *testing.T) {
	t.Helper()

	c := caller.Caller{
		Defined: false,
	}

	got := encodeWith(t, callerpkg.ShortCallerEncoder, c)
	want := "<undefined>"
	if got != want {
		t.Errorf("ShortCallerEncoder for undefined caller = %q, want %q", got, want)
	}
}

// TestShortCallerEncoder_AppendsToExistingBuffer verifies that the encoder
// appends its output to the existing contents of the buffer.
func TestShortCallerEncoder_AppendsToExistingBuffer(t *testing.T) {
	t.Helper()

	c := caller.Caller{
		Defined: true,
		File:    "/src/main.go",
		Line:    10,
	}

	dst := &buffer.Buffer{}
	dst.AppendString("prefix:")

	out := callerpkg.ShortCallerEncoder(dst, c)
	if out == nil {
		t.Fatal("ShortCallerEncoder returned nil buffer")
	}

	got := string(out.Bytes())
	want := "prefix:main.go:10"
	if got != want {
		t.Errorf("buffer after encoding = %q, want %q", got, want)
	}
}

// TestFullCallerEncoder_DefinedCaller verifies that FullCallerEncoder
// produces the expected output with the complete file path and line number.
func TestFullCallerEncoder_DefinedCaller(t *testing.T) {
	t.Helper()

	c := caller.Caller{
		Defined:  true,
		PC:       0,
		File:     "/app/internal/handlers/user.go",
		Line:     42,
		Function: "main.handleUser",
	}

	got := encodeWith(t, callerpkg.FullCallerEncoder, c)
	want := "/app/internal/handlers/user.go:42"
	if got != want {
		t.Errorf("FullCallerEncoder = %q, want %q", got, want)
	}
}

// TestFullCallerEncoder_UndefinedCaller verifies that FullCallerEncoder
// produces "<undefined>" when caller.Defined is false.
func TestFullCallerEncoder_UndefinedCaller(t *testing.T) {
	t.Helper()

	c := caller.Caller{
		Defined: false,
	}

	got := encodeWith(t, callerpkg.FullCallerEncoder, c)
	want := "<undefined>"
	if got != want {
		t.Errorf("FullCallerEncoder for undefined caller = %q, want %q", got, want)
	}
}

// TestFullCallerEncoder_AppendsToExistingBuffer verifies that the encoder
// appends its output to the existing contents of the buffer.
func TestFullCallerEncoder_AppendsToExistingBuffer(t *testing.T) {
	t.Helper()

	c := caller.Caller{
		Defined: true,
		File:    "/src/project/main.go",
		Line:    123,
	}

	dst := &buffer.Buffer{}
	dst.AppendString("caller=")

	out := callerpkg.FullCallerEncoder(dst, c)
	if out == nil {
		t.Fatal("FullCallerEncoder returned nil buffer")
	}

	got := string(out.Bytes())
	want := "caller=/src/project/main.go:123"
	if got != want {
		t.Errorf("buffer after encoding = %q, want %q", got, want)
	}
}

// TestShortVsFull_PathDifference verifies that ShortCallerEncoder strips
// the directory path while FullCallerEncoder retains it.
func TestShortVsFull_PathDifference(t *testing.T) {
	t.Helper()

	c := caller.Caller{
		Defined: true,
		File:    "/very/long/path/to/file.go",
		Line:    999,
	}

	short := encodeWith(t, callerpkg.ShortCallerEncoder, c)
	full := encodeWith(t, callerpkg.FullCallerEncoder, c)

	wantShort := "file.go:999"
	wantFull := "/very/long/path/to/file.go:999"

	if short != wantShort {
		t.Errorf("ShortCallerEncoder = %q, want %q", short, wantShort)
	}
	if full != wantFull {
		t.Errorf("FullCallerEncoder = %q, want %q", full, wantFull)
	}
}
