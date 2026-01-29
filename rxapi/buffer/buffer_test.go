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

package buffer_test

import (
	"strconv"
	"testing"
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
)

func newBuffer() *buffer.Buffer {
	// Zero value of Buffer must be usable.
	return &buffer.Buffer{}
}

func TestBufferWriteAndWriteString(t *testing.T) {
	b := newBuffer()

	n, err := b.Write([]byte("foo"))
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != 3 {
		t.Fatalf("Write returned n = %d, want 3", n)
	}
	if got := string(b.Bytes()); got != "foo" {
		t.Fatalf("after first Write, buffer = %q, want %q", got, "foo")
	}

	n, err = b.WriteString("bar")
	if err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	if n != len("bar") {
		t.Fatalf("WriteString returned n = %d, want %d", n, len("bar"))
	}
	if got := string(b.Bytes()); got != "foobar" {
		t.Fatalf("after WriteString, buffer = %q, want %q", got, "foobar")
	}
}

func TestBufferWriteByte(t *testing.T) {
	b := newBuffer()

	if err := b.WriteByte('a'); err != nil {
		t.Fatalf("WriteByte returned error: %v", err)
	}
	if err := b.WriteByte('b'); err != nil {
		t.Fatalf("WriteByte returned error: %v", err)
	}

	if got := string(b.Bytes()); got != "ab" {
		t.Fatalf("buffer after WriteByte = %q, want %q", got, "ab")
	}
}

func TestBufferAppendTime(t *testing.T) {
	b := newBuffer()

	// Fixed point in time to ensure deterministic formatting.
	tm := time.Date(2025, 1, 2, 3, 4, 5, 123_000_000, time.FixedZone("TEST", 3*60*60))
	layout := "2006-01-02T15:04:05.000Z07:00"

	b.AppendTime(tm, layout)

	want := tm.Format(layout)
	if got := string(b.Bytes()); got != want {
		t.Fatalf("AppendTime = %q, want %q", got, want)
	}
}

func TestBufferAppendDuration(t *testing.T) {
	b := newBuffer()

	d := 2*time.Second + 500*time.Millisecond
	b.AppendDuration(d)

	want := strconv.FormatInt(int64(d), 10)
	if got := string(b.Bytes()); got != want {
		t.Fatalf("AppendDuration = %q, want %q", got, want)
	}
}

func TestBufferAppendByteAndBytesAndString(t *testing.T) {
	b := newBuffer()

	b.AppendByte('[')
	b.AppendBytes([]byte("foo"))
	b.AppendByte(',')
	b.AppendString("bar")
	b.AppendByte(']')

	if got := string(b.Bytes()); got != "[foo,bar]" {
		t.Fatalf("combined appends = %q, want %q", got, "[foo,bar]")
	}
}

func TestBufferAppendBool(t *testing.T) {
	b := newBuffer()

	b.AppendBool(true)
	b.AppendByte(',')
	b.AppendBool(false)

	if got := string(b.Bytes()); got != "true,false" {
		t.Fatalf("AppendBool sequence = %q, want %q", got, "true,false")
	}
}

func TestBufferAppendComplex128(t *testing.T) {
	tests := []struct {
		name string
		val  complex128
		want string
	}{
		{"positiveImag", 1 + 2i, "1+2i"},
		{"negativeImag", 1 - 2i, "1-2i"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBuffer()
			b.AppendComplex128(tt.val)

			got := string(b.Bytes())
			if got != tt.want {
				t.Fatalf("AppendComplex128(%v) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

func TestBufferAppendComplex64(t *testing.T) {
	tests := []struct {
		name string
		val  complex64
		want string
	}{
		{"positiveImag", complex64(1 + 2i), "1+2i"},
		{"negativeImag", complex64(1 - 2i), "1-2i"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newBuffer()
			b.AppendComplex64(tt.val)

			got := string(b.Bytes())
			if got != tt.want {
				t.Fatalf("AppendComplex64(%v) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

func TestBufferAppendFloats(t *testing.T) {
	b := newBuffer()

	b.AppendFloat64(1.5)
	b.AppendByte(',')
	b.AppendFloat32(2.5)

	got := string(b.Bytes())
	// With 'g' formatting and these simple values, representation is stable.
	if got != "1.5,2.5" {
		t.Fatalf("AppendFloat64/AppendFloat32 = %q, want %q", got, "1.5,2.5")
	}
}

func TestBufferAppendInts(t *testing.T) {
	b := newBuffer()

	b.AppendInt(-1)
	b.AppendByte(',')
	b.AppendInt64(-2)
	b.AppendByte(',')
	b.AppendInt32(-3)
	b.AppendByte(',')
	b.AppendInt16(-4)
	b.AppendByte(',')
	b.AppendInt8(-5)

	want := "-1,-2,-3,-4,-5"
	if got := string(b.Bytes()); got != want {
		t.Fatalf("AppendInt* = %q, want %q", got, want)
	}
}

func TestBufferAppendUints(t *testing.T) {
	b := newBuffer()

	b.AppendUint(1)
	b.AppendByte(',')
	b.AppendUint64(2)
	b.AppendByte(',')
	b.AppendUint32(3)
	b.AppendByte(',')
	b.AppendUint16(4)
	b.AppendByte(',')
	b.AppendUint8(5)

	want := "1,2,3,4,5"
	if got := string(b.Bytes()); got != want {
		t.Fatalf("AppendUint* = %q, want %q", got, want)
	}
}

func TestBufferAppendUintptr(t *testing.T) {
	b := newBuffer()

	b.AppendUintptr(0x1234)

	got := string(b.Bytes())
	want := "0x1234"
	if got != want {
		t.Fatalf("AppendUintptr = %q, want %q", got, want)
	}
}

func TestBufferBytesLenCapAndReset(t *testing.T) {
	b := newBuffer()

	if b.Len() != 0 {
		t.Fatalf("initial Len = %d, want 0", b.Len())
	}
	if b.Cap() != 0 {
		// Zero value is allowed to have 0-cap slice.
		t.Fatalf("initial Cap = %d, want 0", b.Cap())
	}

	b.AppendString("hello")
	if gotLen, wantLen := b.Len(), len("hello"); gotLen != wantLen {
		t.Fatalf("Len after AppendString = %d, want %d", gotLen, wantLen)
	}
	if got := string(b.Bytes()); got != "hello" {
		t.Fatalf("Bytes after AppendString = %q, want %q", got, "hello")
	}
	prevCap := b.Cap()
	if prevCap < b.Len() {
		t.Fatalf("Cap (%d) should be >= Len (%d)", prevCap, b.Len())
	}

	b.Reset()
	if gotLen := b.Len(); gotLen != 0 {
		t.Fatalf("Len after Reset = %d, want 0", gotLen)
	}
	if gotCap := b.Cap(); gotCap != prevCap {
		t.Fatalf("Cap after Reset = %d, want %d (capacity must be preserved)", gotCap, prevCap)
	}
	if len(b.Bytes()) != 0 {
		t.Fatalf("Bytes after Reset should be empty, got %q", string(b.Bytes()))
	}
}
