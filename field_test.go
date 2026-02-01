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

package rxlog_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"dirpx.dev/rxlog"
	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/encoder/base"
	ftype "dirpx.dev/rxlog/rxapi/field/type"
)

// Dummy implementations for array/object marshalers used in tests.

type dummyArrayMarshaler struct{}

func (dummyArrayMarshaler) MarshalLogArray(enc base.ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
	return nil, nil
}

type dummyObjectMarshaler struct{}

func (dummyObjectMarshaler) MarshalLogObject(enc base.ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
	return nil, nil
}

// Dummy stringer used for Stringer helpers and Any.

type dummyStringer struct{ s string }

func (d dummyStringer) String() string { return d.s }

func TestArrayShortcut(t *testing.T) {
	m := dummyArrayMarshaler{}
	f := rxlog.Array("k", m)

	if f.Key != "k" {
		t.Fatalf("Array.Key = %q, want %q", f.Key, "k")
	}
	if f.Type != ftype.Array {
		t.Fatalf("Array.Type = %v, want %v", f.Type, ftype.Array)
	}
	if _, ok := f.Interface.(dummyArrayMarshaler); !ok {
		t.Fatalf("Array.Interface has type %T, want dummyArrayMarshaler", f.Interface)
	}
}

func TestObjectShortcut(t *testing.T) {
	m := dummyObjectMarshaler{}
	f := rxlog.Object("k", m)

	if f.Key != "k" {
		t.Fatalf("Object.Key = %q, want %q", f.Key, "k")
	}
	if f.Type != ftype.Object {
		t.Fatalf("Object.Type = %v, want %v", f.Type, ftype.Object)
	}
	if _, ok := f.Interface.(dummyObjectMarshaler); !ok {
		t.Fatalf("Object.Interface has type %T, want dummyObjectMarshaler", f.Interface)
	}
}

func TestInlineShortcut(t *testing.T) {
	m := dummyObjectMarshaler{}
	f := rxlog.Inline(m)

	if f.Key != "" {
		t.Fatalf("Inline.Key = %q, want empty", f.Key)
	}
	if f.Type != ftype.Inline {
		t.Fatalf("Inline.Type = %v, want %v", f.Type, ftype.Inline)
	}
	if _, ok := f.Interface.(dummyObjectMarshaler); !ok {
		t.Fatalf("Inline.Interface has type %T, want dummyObjectMarshaler", f.Interface)
	}
}

func TestNamespaceShortcut(t *testing.T) {
	f := rxlog.Namespace("ns")

	if f.Key != "ns" {
		t.Fatalf("Namespace.Key = %q, want %q", f.Key, "ns")
	}
	if f.Type != ftype.Namespace {
		t.Fatalf("Namespace.Type = %v, want %v", f.Type, ftype.Namespace)
	}
}

func TestSkipShortcut(t *testing.T) {
	f := rxlog.Skip()

	if f.Type != ftype.Skip {
		t.Fatalf("Skip.Type = %v, want %v", f.Type, ftype.Skip)
	}
}

func TestBinaryAndBinaryPtrShortcuts(t *testing.T) {
	b := []byte{1, 2, 3}
	f := rxlog.Binary("k", b)

	if f.Key != "k" {
		t.Fatalf("Binary.Key = %q, want %q", f.Key, "k")
	}
	if f.Type != ftype.Binary {
		t.Fatalf("Binary.Type = %v, want %v", f.Type, ftype.Binary)
	}
	if got, ok := f.Interface.([]byte); !ok || !reflect.DeepEqual(got, b) {
		t.Fatalf("Binary.Interface = %#v, want %#v", f.Interface, b)
	}
	if &b[0] != &f.Interface.([]byte)[0] {
		t.Fatalf("Binary should not copy slice")
	}

	// Ptr nil -> Skip.
	if f := rxlog.BinaryPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("BinaryPtr(nil).Type = %v, want Skip", f.Type)
	}

	// Ptr non-nil -> Binary with same slice.
	b2 := []byte{4, 5}
	ptr := &b2
	f = rxlog.BinaryPtr("k", ptr)
	if f.Type != ftype.Binary {
		t.Fatalf("BinaryPtr.Type = %v, want %v", f.Type, ftype.Binary)
	}
	if got, ok := f.Interface.([]byte); !ok || &got[0] != &b2[0] {
		t.Fatalf("BinaryPtr should store underlying slice without copy")
	}
}

func TestByteStringAndPtrShortcuts(t *testing.T) {
	b := []byte("hello")
	f := rxlog.ByteString("k", b)

	if f.Type != ftype.ByteString {
		t.Fatalf("ByteString.Type = %v, want %v", f.Type, ftype.ByteString)
	}
	if got, ok := f.Interface.([]byte); !ok || !reflect.DeepEqual(got, b) {
		t.Fatalf("ByteString.Interface = %#v, want %#v", f.Interface, b)
	}

	if f := rxlog.ByteStringPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("ByteStringPtr(nil).Type = %v, want Skip", f.Type)
	}
	b2 := []byte("world")
	ptr := &b2
	f = rxlog.ByteStringPtr("k", ptr)
	if f.Type != ftype.ByteString {
		t.Fatalf("ByteStringPtr.Type = %v, want %v", f.Type, ftype.ByteString)
	}
}

func TestStringAndPtrShortcuts(t *testing.T) {
	f := rxlog.String("k", "v")

	if f.Key != "k" || f.Type != ftype.String || f.String != "v" {
		t.Fatalf("String produced %+v, want Key=k, Type=String, String=v", f)
	}

	if f := rxlog.StringPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("StringPtr(nil).Type = %v, want Skip", f.Type)
	}

	s := "x"
	f = rxlog.StringPtr("k", &s)
	if f.Type != ftype.String || f.String != "x" {
		t.Fatalf("StringPtr produced %+v, want String=x", f)
	}
}

func TestBoolAndPtrShortcuts(t *testing.T) {
	fTrue := rxlog.Bool("k", true)
	if fTrue.Type != ftype.Bool || fTrue.Integer != 1 {
		t.Fatalf("Bool(true) = %+v, want Type=Bool, Integer=1", fTrue)
	}

	fFalse := rxlog.Bool("k", false)
	if fFalse.Type != ftype.Bool || fFalse.Integer != 0 {
		t.Fatalf("Bool(false) = %+v, want Type=Bool, Integer=0", fFalse)
	}

	if f := rxlog.BoolPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("BoolPtr(nil).Type = %v, want Skip", f.Type)
	}

	val := true
	fTrue = rxlog.BoolPtr("k", &val)
	if fTrue.Integer != 1 {
		t.Fatalf("BoolPtr(true).Integer = %d, want 1", fTrue.Integer)
	}
}

func TestIntShortcuts(t *testing.T) {
	fi64 := rxlog.Int64("k", -5)
	if fi64.Type != ftype.Int64 || fi64.Integer != -5 {
		t.Fatalf("Int64 = %+v, want Type=Int64, Integer=-5", fi64)
	}

	fi32 := rxlog.Int32("k", int32(-3))
	if fi32.Type != ftype.Int32 || fi32.Integer != int64(int32(-3)) {
		t.Fatalf("Int32 = %+v, want Type=Int32, Integer=%d", fi32, int32(-3))
	}

	fi16 := rxlog.Int16("k", int16(-2))
	if fi16.Type != ftype.Int16 || fi16.Integer != int64(int16(-2)) {
		t.Fatalf("Int16 = %+v, want Type=Int16, Integer=%d", fi16, int16(-2))
	}

	fi8 := rxlog.Int8("k", int8(-1))
	if fi8.Type != ftype.Int8 || fi8.Integer != int64(int8(-1)) {
		t.Fatalf("Int8 = %+v, want Type=Int8, Integer=%d", fi8, int8(-1))
	}

	fi := rxlog.Int("k", int(-42))
	if fi.Type != ftype.Int64 || fi.Integer != int64(-42) {
		t.Fatalf("Int = %+v, want Type=Int64, Integer=-42", fi)
	}

	// Ptr helpers -> Skip on nil.
	if f := rxlog.Int64Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Int64Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.Int32Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Int32Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.Int16Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Int16Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.Int8Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Int8Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.IntPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("IntPtr(nil).Type = %v, want Skip", f.Type)
	}
}

func TestUintShortcuts(t *testing.T) {
	fu64 := rxlog.Uint64("k", uint64(42))
	if fu64.Type != ftype.Uint64 || fu64.Integer != int64(42) {
		t.Fatalf("Uint64 = %+v, want Type=Uint64, Integer=42", fu64)
	}

	fu32 := rxlog.Uint32("k", uint32(32))
	if fu32.Type != ftype.Uint32 || fu32.Integer != int64(uint32(32)) {
		t.Fatalf("Uint32 = %+v, want Type=Uint32, Integer=32", fu32)
	}

	fu16 := rxlog.Uint16("k", uint16(16))
	if fu16.Type != ftype.Uint16 || fu16.Integer != int64(uint16(16)) {
		t.Fatalf("Uint16 = %+v, want Type=Uint16, Integer=16", fu16)
	}

	fu8 := rxlog.Uint8("k", uint8(8))
	if fu8.Type != ftype.Uint8 || fu8.Integer != int64(uint8(8)) {
		t.Fatalf("Uint8 = %+v, want Type=Uint8, Integer=8", fu8)
	}

	fu := rxlog.Uint("k", uint(7))
	if fu.Type != ftype.Uint64 || fu.Integer != int64(7) {
		t.Fatalf("Uint = %+v, want Type=Uint64, Integer=7", fu)
	}

	fup := rxlog.Uintptr("k", uintptr(123))
	if fup.Type != ftype.Uintptr || fup.Integer != int64(uintptr(123)) {
		t.Fatalf("Uintptr = %+v, want Type=Uintptr, Integer=123", fup)
	}

	// Ptr helpers -> Skip on nil.
	if f := rxlog.Uint64Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Uint64Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.Uint32Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Uint32Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.Uint16Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Uint16Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.Uint8Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Uint8Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.UintPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("UintPtr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.UintptrPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("UintptrPtr(nil).Type = %v, want Skip", f.Type)
	}
}

func TestFloatShortcuts(t *testing.T) {
	f64 := rxlog.Float64("k", 1.5)
	if f64.Type != ftype.Float64 {
		t.Fatalf("Float64.Type = %v, want %v", f64.Type, ftype.Float64)
	}
	if got := math.Float64frombits(uint64(f64.Integer)); got != 1.5 {
		t.Fatalf("Float64 stored %v, want 1.5", got)
	}

	f32 := rxlog.Float32("k", float32(2.5))
	if f32.Type != ftype.Float32 {
		t.Fatalf("Float32.Type = %v, want %v", f32.Type, ftype.Float32)
	}
	if got := math.Float32frombits(uint32(f32.Integer)); got != float32(2.5) {
		t.Fatalf("Float32 stored %v, want 2.5", got)
	}

	if f := rxlog.Float64Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Float64Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.Float32Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Float32Ptr(nil).Type = %v, want Skip", f.Type)
	}
}

func TestComplexShortcuts(t *testing.T) {
	c128 := complex(1.0, -2.0)
	f128 := rxlog.Complex128("k", c128)
	if f128.Type != ftype.Complex128 {
		t.Fatalf("Complex128.Type = %v, want %v", f128.Type, ftype.Complex128)
	}
	if got, ok := f128.Interface.(complex128); !ok || got != c128 {
		t.Fatalf("Complex128.Interface = %#v, want %v", f128.Interface, c128)
	}

	c64 := complex(float32(3.0), float32(-4.0))
	f64 := rxlog.Complex64("k", c64)
	if f64.Type != ftype.Complex64 {
		t.Fatalf("Complex64.Type = %v, want %v", f64.Type, ftype.Complex64)
	}
	if got, ok := f64.Interface.(complex64); !ok || got != c64 {
		t.Fatalf("Complex64.Interface = %#v, want %v", f64.Interface, c64)
	}

	if f := rxlog.Complex128Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Complex128Ptr(nil).Type = %v, want Skip", f.Type)
	}
	if f := rxlog.Complex64Ptr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("Complex64Ptr(nil).Type = %v, want Skip", f.Type)
	}
}

func TestDurationShortcuts(t *testing.T) {
	d := 2*time.Second + 500*time.Millisecond
	f := rxlog.Duration("k", d)

	if f.Type != ftype.Duration || f.Integer != int64(d) {
		t.Fatalf("Duration = %+v, want Type=Duration, Integer=%d", f, d)
	}

	if f := rxlog.DurationPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("DurationPtr(nil).Type = %v, want Skip", f.Type)
	}
}

func TestTimeShortcuts(t *testing.T) {
	loc := time.FixedZone("TEST", 3*60*60)
	tm := time.Date(2025, 1, 2, 3, 4, 5, 123_000_000, loc)

	// Time: normalized to UTC, location stored if not UTC.
	f := rxlog.Time("k", tm)
	if f.Type != ftype.Time {
		t.Fatalf("Time.Type = %v, want %v", f.Type, ftype.Time)
	}
	if f.Key != "k" {
		t.Fatalf("Time.Key = %q, want %q", f.Key, "k")
	}
	if f.Integer != tm.UTC().UnixNano() {
		t.Fatalf("Time.Integer = %d, want %d", f.Integer, tm.UTC().UnixNano())
	}
	if gotLoc, ok := f.Interface.(*time.Location); !ok || gotLoc != loc {
		t.Fatalf("Time.Interface = %#v, want *time.Location %v", f.Interface, loc)
	}

	// TimePtr nil -> Skip.
	if f := rxlog.TimePtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("TimePtr(nil).Type = %v, want Skip", f.Type)
	}

	// TimeInLocation: t converted to loc before UnixNano.
	tm2 := time.Date(2025, 5, 6, 7, 8, 9, 0, time.UTC)
	f = rxlog.TimeInLocation("k", tm2, loc)
	if f.Type != ftype.Time {
		t.Fatalf("TimeInLocation.Type = %v, want %v", f.Type, ftype.Time)
	}
	if f.Integer != tm2.In(loc).UnixNano() {
		t.Fatalf("TimeInLocation.Integer = %d, want %d", f.Integer, tm2.In(loc).UnixNano())
	}
	if f.Interface != loc {
		t.Fatalf("TimeInLocation.Interface = %#v, want %v", f.Interface, loc)
	}

	// TimeFull: full time in Interface.
	now := time.Now()
	f = rxlog.TimeFull("k", now)
	if f.Type != ftype.TimeFull {
		t.Fatalf("TimeFull.Type = %v, want %v", f.Type, ftype.TimeFull)
	}
	if got, ok := f.Interface.(time.Time); !ok || !got.Equal(now) {
		t.Fatalf("TimeFull.Interface = %#v, want %v", f.Interface, now)
	}
	if f := rxlog.TimeFullPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("TimeFullPtr(nil).Type = %v, want Skip", f.Type)
	}
}

func TestErrorAndErrorPtrShortcuts(t *testing.T) {
	err := fmt.Errorf("boom")
	f := rxlog.Error("k", err)

	if f.Type != ftype.Error {
		t.Fatalf("Error.Type = %v, want %v", f.Type, ftype.Error)
	}
	if got, ok := f.Interface.(error); !ok || got != err {
		t.Fatalf("Error.Interface = %#v, want %v", f.Interface, err)
	}

	// Error allows nil error in Interface.
	f = rxlog.Error("k", nil)
	if f.Type != ftype.Error {
		t.Fatalf("Error(nil).Type = %v, want %v", f.Type, ftype.Error)
	}
	if f.Interface != nil {
		t.Fatalf("Error(nil).Interface = %#v, want nil", f.Interface)
	}

	// ErrorPtr nil pointer -> Skip.
	if f := rxlog.ErrorPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("ErrorPtr(nil).Type = %v, want Skip", f.Type)
	}

	// ErrorPtr non-nil pointer with nil value.
	var e error
	f = rxlog.ErrorPtr("k", &e)
	if f.Type != ftype.Error || f.Interface != nil {
		t.Fatalf("ErrorPtr(&nil) = %+v, want Type=Error, Interface=nil", f)
	}
}

func TestStringerAndPtrShortcuts(t *testing.T) {
	s := dummyStringer{s: "v"}
	f := rxlog.Stringer("k", s)

	if f.Type != ftype.Stringer {
		t.Fatalf("Stringer.Type = %v, want %v", f.Type, ftype.Stringer)
	}
	if got, ok := f.Interface.(dummyStringer); !ok || got.s != "v" {
		t.Fatalf("Stringer.Interface = %#v, want dummyStringer{v}", f.Interface)
	}

	if f := rxlog.StringerPtr("k", nil); f.Type != ftype.Skip {
		t.Fatalf("StringerPtr(nil).Type = %v, want Skip", f.Type)
	}

	var sp fmt.Stringer = dummyStringer{s: "x"}
	f = rxlog.StringerPtr("k", &sp)
	if f.Type != ftype.Stringer {
		t.Fatalf("StringerPtr.Type = %v, want %v", f.Type, ftype.Stringer)
	}
	if got, ok := f.Interface.(fmt.Stringer); !ok || got.String() != "x" {
		t.Fatalf("StringerPtr.Interface.String() = %q, want %q", got.String(), "x")
	}
}

func TestReflectShortcut(t *testing.T) {
	type payload struct {
		A int
		B string
	}
	p := payload{A: 1, B: "x"}

	f := rxlog.Reflect("k", p)
	if f.Key != "k" || f.Type != ftype.Reflect {
		t.Fatalf("Reflect = %+v, want Key=k, Type=Reflect", f)
	}
	if got, ok := f.Interface.(payload); !ok || got != p {
		t.Fatalf("Reflect.Interface = %#v, want %#v", f.Interface, p)
	}
}

func TestAnyShortcut_SelectsExpectedConstructors(t *testing.T) {
	loc := time.FixedZone("TEST", 3*60*60)

	cases := []struct {
		name string
		val  any
		typ  ftype.Type
	}{
		{"nil", nil, ftype.Skip},
		{"string", "str", ftype.String},
		{"bytes", []byte("bs"), ftype.ByteString},
		{"bool", true, ftype.Bool},
		{"int", int(1), ftype.Int64},
		{"int64", int64(1), ftype.Int64},
		{"int32", int32(1), ftype.Int32},
		{"int16", int16(1), ftype.Int16},
		{"int8", int8(1), ftype.Int8},
		{"uint", uint(1), ftype.Uint64},
		{"uint64", uint64(1), ftype.Uint64},
		{"uint32", uint32(1), ftype.Uint32},
		{"uint16", uint16(1), ftype.Uint16},
		{"uint8", uint8(1), ftype.Uint8},
		{"uintptr", uintptr(1), ftype.Uintptr},
		{"float64", float64(1.0), ftype.Float64},
		{"float32", float32(1.0), ftype.Float32},
		{"complex128", complex(1.0, 2.0), ftype.Complex128},
		{"complex64", complex(float32(1.0), float32(2.0)), ftype.Complex64},
		{"time.Time", time.Date(2025, 1, 2, 3, 4, 5, 0, loc), ftype.TimeFull},
		{"duration", time.Second, ftype.Duration},
		{"error", fmt.Errorf("boom"), ftype.Error},
		{"stringer", dummyStringer{"x"}, ftype.Stringer},
		{"array-marshaler", dummyArrayMarshaler{}, ftype.Array},
		{"object-marshaler", dummyObjectMarshaler{}, ftype.Object},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			f := rxlog.Any("k", tt.val)
			if f.Type != tt.typ {
				t.Fatalf("Any(%s).Type = %v, want %v", tt.name, f.Type, tt.typ)
			}
		})
	}
}
