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
	"strings"
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/encoder/base"
	"dirpx.dev/rxlog/rxapi/encoder/encoders"
)

// stringerRecordingEncoder is a minimal ObjectEncoder test double that only
// records AddString calls for verification in tests.
type stringerRecordingEncoder struct {
	base.ObjectEncoder // embedded to satisfy interface

	keys   []string
	values []string
}

type SimpleStringer struct{ s string }

func (s SimpleStringer) String() string {
	return s.s
}

func newStringerRecordingEncoder() *stringerRecordingEncoder {
	return &stringerRecordingEncoder{
		keys:   make([]string, 0, 4),
		values: make([]string, 0, 4),
	}
}

// AddString records the key/value pair and appends the value to the buffer.
func (e *stringerRecordingEncoder) AddString(dst *buffer.Buffer, key, value string) *buffer.Buffer {
	e.keys = append(e.keys, key)
	e.values = append(e.values, value)
	dst.AppendString(value)
	return dst
}

func (e *stringerRecordingEncoder) last() (key, value string, ok bool) {
	if len(e.keys) == 0 {
		return "", "", false
	}
	i := len(e.keys) - 1
	return e.keys[i], e.values[i], true
}

// panicNilStringer is a fmt.Stringer whose String method always panics.
// We use a *panicNilStringer(nil) value to simulate a typed nil receiver
// that causes String() to panic.
type panicNilStringer struct{}

func (*panicNilStringer) String() string {
	panic("nil receiver")
}

// panicStringer is a non-pointer fmt.Stringer whose String method always
// panics. It is used to verify that EncodeStringer surfaces such panics as
// a non-nil error and does not encode a value.
type panicStringer struct{}

func (panicStringer) String() string {
	panic("unexpected panic")
}

func TestEncodeStringer_Nil(t *testing.T) {
	t.Helper()

	buf := &buffer.Buffer{}
	enc := newStringerRecordingEncoder()

	out, err := encoders.EncodeStringer(buf, "s", nil, enc)
	if err != nil {
		t.Fatalf("EncodeStringer(nil) returned error: %v", err)
	}
	if out == nil {
		t.Fatalf("EncodeStringer(nil) returned nil buffer")
	}

	key, value, ok := enc.last()
	if !ok {
		t.Fatalf("expected AddString to be called at least once")
	}
	if key != "s" {
		t.Errorf("unexpected key: got %q, want %q", key, "s")
	}
	if value != "<nil>" {
		t.Errorf("unexpected value: got %q, want %q", value, "<nil>")
	}
}

func TestEncodeStringer_NonNil(t *testing.T) {
	t.Helper()

	buf := &buffer.Buffer{}
	enc := newStringerRecordingEncoder()

	val := SimpleStringer{s: "value"}

	out, err := encoders.EncodeStringer(buf, "s", val, enc)
	if err != nil {
		t.Fatalf("EncodeStringer(non-nil) returned error: %v", err)
	}
	if out == nil {
		t.Fatalf("EncodeStringer(non-nil) returned nil buffer")
	}

	key, value, ok := enc.last()
	if !ok {
		t.Fatalf("expected AddString to be called at least once")
	}
	if key != "s" {
		t.Errorf("unexpected key: got %q, want %q", key, "s")
	}
	if value != "value" {
		t.Errorf("unexpected value: got %q, want %q", value, "value")
	}
}

func TestEncodeStringer_TypedNilPanic(t *testing.T) {
	t.Helper()

	buf := &buffer.Buffer{}
	enc := newStringerRecordingEncoder()

	// Construct a typed nil pointer value stored in an empty interface.
	var p *panicNilStringer
	var any interface{} = p
	if any == nil {
		t.Fatalf("expected typed nil pointer in interface{} to be non-nil")
	}

	out, err := encoders.EncodeStringer(buf, "s", any, enc)
	if err != nil {
		t.Fatalf("EncodeStringer(typed nil) returned error: %v", err)
	}
	if out == nil {
		t.Fatalf("EncodeStringer(typed nil) returned nil buffer")
	}

	key, value, ok := enc.last()
	if !ok {
		t.Fatalf("expected AddString to be called at least once")
	}
	if key != "s" {
		t.Errorf("unexpected key: got %q, want %q", key, "s")
	}
	if value != "<nil>" {
		t.Errorf("unexpected value for typed nil: got %q, want %q", value, "<nil>")
	}
}

func TestEncodeStringer_PanicNonNil(t *testing.T) {
	t.Helper()

	buf := &buffer.Buffer{}
	enc := newStringerRecordingEncoder()

	val := panicStringer{}

	out, err := encoders.EncodeStringer(buf, "s", val, enc)
	if err == nil {
		t.Fatalf("expected non-nil error when String() panics for non-nil value")
	}
	if out == nil {
		t.Fatalf("EncodeStringer(panic non-nil) returned nil buffer")
	}

	if !strings.Contains(err.Error(), "panic in Stringer.String") {
		t.Errorf("unexpected error text: got %q, want substring %q", err.Error(), "panic in Stringer.String")
	}

	if len(enc.keys) != 0 {
		t.Fatalf("expected no AddString calls on panic for non-nil value, got %d", len(enc.keys))
	}
}
