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

package error_test

import (
	"errors"
	"fmt"
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
	aerr "dirpx.dev/rxlog/rxapi/error"
	errorpkg "dirpx.dev/rxlog/rxcore/error"
)

// encodeWith is a helper that runs the given error encoder on a fresh buffer
// and returns the resulting string.
func encodeWith(t *testing.T, enc aerr.Encoder, err error) string {
	t.Helper()

	dst := &buffer.Buffer{}
	out := enc(dst, err)
	if out == nil {
		t.Fatalf("encoder returned nil buffer")
	}

	return string(out.Bytes())
}

// TestSimpleErrorEncoder_NonNilError verifies that SimpleErrorEncoder outputs
// the error message from err.Error().
func TestSimpleErrorEncoder_NonNilError(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "simple error",
			err:  errors.New("connection refused"),
			want: "connection refused",
		},
		{
			name: "formatted error",
			err:  fmt.Errorf("failed to connect: %w", errors.New("timeout")),
			want: "failed to connect: timeout",
		},
		{
			name: "custom error type",
			err:  &customError{msg: "custom error message"},
			want: "custom error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeWith(t, errorpkg.SimpleErrorEncoder, tt.err)
			if got != tt.want {
				t.Errorf("SimpleErrorEncoder(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

// TestSimpleErrorEncoder_NilError verifies that SimpleErrorEncoder appends
// nothing when given a nil error.
func TestSimpleErrorEncoder_NilError(t *testing.T) {
	t.Helper()

	got := encodeWith(t, errorpkg.SimpleErrorEncoder, nil)
	if got != "" {
		t.Errorf("SimpleErrorEncoder(nil) = %q, want empty string", got)
	}
}

// TestSimpleErrorEncoder_AppendsToExistingBuffer verifies that the encoder
// appends its output to the existing contents of the buffer.
func TestSimpleErrorEncoder_AppendsToExistingBuffer(t *testing.T) {
	t.Helper()

	err := errors.New("test error")

	dst := &buffer.Buffer{}
	dst.AppendString("error=")

	out := errorpkg.SimpleErrorEncoder(dst, err)
	if out == nil {
		t.Fatal("SimpleErrorEncoder returned nil buffer")
	}

	got := string(out.Bytes())
	want := "error=test error"
	if got != want {
		t.Errorf("buffer after encoding = %q, want %q", got, want)
	}
}

// TestSimpleErrorEncoder_CanBeReusedAcrossCalls verifies that a single encoder
// instance is safe to reuse across multiple calls with different errors.
func TestSimpleErrorEncoder_CanBeReusedAcrossCalls(t *testing.T) {
	t.Helper()

	enc := errorpkg.SimpleErrorEncoder

	// First call.
	buf1 := &buffer.Buffer{}
	err1 := errors.New("first error")
	out1 := enc(buf1, err1)
	if out1 == nil {
		t.Fatal("first call to encoder returned nil buffer")
	}
	got1 := string(out1.Bytes())
	want1 := "first error"
	if got1 != want1 {
		t.Errorf("first call: got %q, want %q", got1, want1)
	}

	// Second call with different error and a fresh buffer.
	buf2 := &buffer.Buffer{}
	err2 := errors.New("second error")
	out2 := enc(buf2, err2)
	if out2 == nil {
		t.Fatal("second call to encoder returned nil buffer")
	}
	got2 := string(out2.Bytes())
	want2 := "second error"
	if got2 != want2 {
		t.Errorf("second call: got %q, want %q", got2, want2)
	}
}

// customError is a test error type that implements error.
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
