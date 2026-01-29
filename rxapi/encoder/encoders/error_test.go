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

package encoders_test

import (
	"errors"
	"strings"
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/encoder/base"
	"dirpx.dev/rxlog/rxapi/encoder/encoders"
)

// errorRecordingEncoder is a minimal ObjectEncoder test double that only
// implements AddString and records all key/value pairs passed through it.
//
// All other ObjectEncoder methods are satisfied via the embedded
// base.ObjectEncoder interface, but they remain unused in these tests.
type errorRecordingEncoder struct {
	base.ObjectEncoder // embedded to satisfy the full interface

	keys   []string
	values []string
}

func newRecordingEncoder() *errorRecordingEncoder {
	return &errorRecordingEncoder{
		keys:   make([]string, 0, 4),
		values: make([]string, 0, 4),
	}
}

// AddString records the key/value pair and appends the value to the buffer.
// This is the only ObjectEncoder method exercised by EncodeError.
func (e *errorRecordingEncoder) AddString(dst *buffer.Buffer, key, value string) *buffer.Buffer {
	e.keys = append(e.keys, key)
	e.values = append(e.values, value)

	// Append to the buffer so that callers can inspect that output if desired.
	dst.AppendString(value)
	return dst
}

func (e *errorRecordingEncoder) last() (key, value string, ok bool) {
	if len(e.keys) == 0 {
		return "", "", false
	}
	idx := len(e.keys) - 1
	return e.keys[idx], e.values[idx], true
}

// panicNilError is an error implementation whose Error method always panics.
// In tests we wrap a *panicNilError(nil) into an error interface to simulate
// the "typed nil" case where calling Error() leads to a panic with a nil
// pointer underlying value.
type panicNilError struct{}

func (p *panicNilError) Error() string {
	panic("nil receiver")
}

// panicError is a non-pointer error type whose Error method always panics.
// It is used to verify that EncodeError surfaces such panics as a non-nil
// returned error instead of encoding a field value.
type panicError struct{}

func (p panicError) Error() string {
	panic("unexpected panic")
}

// TestEncodeError_NilError verifies that a nil error is encoded as the literal
// "<nil>" and does not result in a failure.
func TestEncodeError_NilError(t *testing.T) {
	t.Helper()

	buf := &buffer.Buffer{}
	enc := newRecordingEncoder()

	out, err := encoders.EncodeError(buf, "err", nil, enc)
	if err != nil {
		t.Fatalf("EncodeError(nil) returned unexpected error: %v", err)
	}
	if out == nil {
		t.Fatalf("EncodeError(nil) returned nil buffer")
	}

	key, value, ok := enc.last()
	if !ok {
		t.Fatalf("expected AddString to be called at least once")
	}
	if key != "err" {
		t.Errorf("unexpected key recorded: got %q, want %q", key, "err")
	}
	if value != "<nil>" {
		t.Errorf("unexpected value recorded: got %q, want %q", value, "<nil>")
	}
}

// TestEncodeError_NonNil verifies that a regular non-nil error is encoded
// using its Error() string.
func TestEncodeError_NonNil(t *testing.T) {
	t.Helper()

	buf := &buffer.Buffer{}
	enc := newRecordingEncoder()

	errValue := errors.New("something went wrong")

	out, err := encoders.EncodeError(buf, "err", errValue, enc)
	if err != nil {
		t.Fatalf("EncodeError(non-nil) returned unexpected error: %v", err)
	}
	if out == nil {
		t.Fatalf("EncodeError(non-nil) returned nil buffer")
	}

	key, value, ok := enc.last()
	if !ok {
		t.Fatalf("expected AddString to be called at least once")
	}
	if key != "err" {
		t.Errorf("unexpected key recorded: got %q, want %q", key, "err")
	}
	want := errValue.Error()
	if value != want {
		t.Errorf("unexpected value recorded: got %q, want %q", value, want)
	}
}

// TestEncodeError_TypedNilPanic verifies that when Error() panics and the
// underlying error value is a nil pointer, EncodeError recovers and encodes
// "<nil>" while returning a nil error.
func TestEncodeError_TypedNilPanic(t *testing.T) {
	t.Helper()

	buf := &buffer.Buffer{}
	enc := newRecordingEncoder()

	// Construct an error interface that holds a typed nil pointer.
	var e *panicNilError
	var errValue error = e
	if errValue == nil {
		t.Fatalf("expected typed nil error interface to be non-nil")
	}

	out, err := encoders.EncodeError(buf, "err", errValue, enc)
	if err != nil {
		t.Fatalf("EncodeError(typed nil) returned unexpected error: %v", err)
	}
	if out == nil {
		t.Fatalf("EncodeError(typed nil) returned nil buffer")
	}

	key, value, ok := enc.last()
	if !ok {
		t.Fatalf("expected AddString to be called at least once")
	}
	if key != "err" {
		t.Errorf("unexpected key recorded: got %q, want %q", key, "err")
	}
	if value != "<nil>" {
		t.Errorf("unexpected value recorded for typed nil: got %q, want %q", value, "<nil>")
	}
}

// TestEncodeError_PanicNonNil verifies that when Error() panics on a non-nil
// non-pointer error value, EncodeError surfaces a non-nil error and does not
// attempt to encode a string field.
func TestEncodeError_PanicNonNil(t *testing.T) {
	t.Helper()

	buf := &buffer.Buffer{}
	enc := newRecordingEncoder()

	errValue := panicError{}

	out, err := encoders.EncodeError(buf, "err", errValue, enc)
	if err == nil {
		t.Fatalf("expected non-nil error when Error() panics for non-pointer value")
	}
	if out == nil {
		t.Fatalf("EncodeError(panic non-nil) returned nil buffer")
	}
	if !strings.Contains(err.Error(), "panic in error.Error") {
		t.Errorf("unexpected error text: got %q, want substring %q", err.Error(), "panic in error.Error")
	}

	if len(enc.keys) != 0 {
		t.Fatalf("expected no calls to AddString on panic for non-nil value, got %d", len(enc.keys))
	}
}
