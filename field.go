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

package rxlog

import (
	"fmt"
	"math"
	"time"

	"dirpx.dev/rxlog/rxapi/encoder/base"
	"dirpx.dev/rxlog/rxapi/field"
	ftype "dirpx.dev/rxlog/rxapi/field/type"
)

// Type aliases for rxapi/field types.
type (
	// Field is an alias for rxapi/field.Field, representing a structured
	// key-value pair for logging.
	Field = field.Field

	// FieldType is an alias for rxapi/field/type.Type, identifying the
	// type of value stored in a Field.
	FieldType = ftype.Type

	// ArrayMarshaler is an alias for rxapi/encoder/base.ArrayMarshaler,
	// used for custom array serialization in fields.
	ArrayMarshaler = base.ArrayMarshaler

	// ObjectMarshaler is an alias for rxapi/encoder/base.ObjectMarshaler,
	// used for custom object serialization in fields.
	ObjectMarshaler = base.ObjectMarshaler
)

// Field type constants providing convenient access to rxapi/field/type values.
const (
	// UnknownFieldType is the zero value for FieldType, representing an
	// uninitialized or invalid field.
	UnknownFieldType = ftype.Unknown

	// ArrayFieldType indicates the field carries an ArrayMarshaler.
	ArrayFieldType = ftype.Array

	// ObjectFieldType indicates the field carries an ObjectMarshaler.
	ObjectFieldType = ftype.Object

	// BinaryFieldType indicates the field contains opaque bytes.
	BinaryFieldType = ftype.Binary

	// BoolFieldType indicates the field represents a boolean value.
	BoolFieldType = ftype.Bool

	// ByteStringFieldType indicates the field carries UTF-8 encoded bytes.
	ByteStringFieldType = ftype.ByteString

	// Complex128FieldType indicates the field carries a complex128 value.
	Complex128FieldType = ftype.Complex128

	// Complex64FieldType indicates the field carries a complex64 value.
	Complex64FieldType = ftype.Complex64

	// DurationFieldType indicates the field carries a time.Duration.
	DurationFieldType = ftype.Duration

	// Float64FieldType indicates the field carries a float64 value.
	Float64FieldType = ftype.Float64

	// Float32FieldType indicates the field carries a float32 value.
	Float32FieldType = ftype.Float32

	// Int64FieldType indicates the field carries an int64 value.
	Int64FieldType = ftype.Int64

	// Int32FieldType indicates the field carries an int32 value.
	Int32FieldType = ftype.Int32

	// Int16FieldType indicates the field carries an int16 value.
	Int16FieldType = ftype.Int16

	// Int8FieldType indicates the field carries an int8 value.
	Int8FieldType = ftype.Int8

	// StringFieldType indicates the field carries a string value.
	StringFieldType = ftype.String

	// TimeFieldType indicates the field carries a time.Time via UnixNano.
	TimeFieldType = ftype.Time

	// TimeFullFieldType indicates the field carries a full time.Time.
	TimeFullFieldType = ftype.TimeFull

	// Uint64FieldType indicates the field carries a uint64 value.
	Uint64FieldType = ftype.Uint64

	// Uint32FieldType indicates the field carries a uint32 value.
	Uint32FieldType = ftype.Uint32

	// Uint16FieldType indicates the field carries a uint16 value.
	Uint16FieldType = ftype.Uint16

	// Uint8FieldType indicates the field carries a uint8 value.
	Uint8FieldType = ftype.Uint8

	// UintptrFieldType indicates the field carries a uintptr value.
	UintptrFieldType = ftype.Uintptr

	// ReflectFieldType indicates the field carries an arbitrary interface{}.
	ReflectFieldType = ftype.Reflect

	// NamespaceFieldType marks the beginning of a logical namespace.
	NamespaceFieldType = ftype.Namespace

	// StringerFieldType indicates the field carries a fmt.Stringer.
	StringerFieldType = ftype.Stringer

	// ErrorFieldType indicates the field carries an error value.
	ErrorFieldType = ftype.Error

	// SkipFieldType indicates the field should be ignored by encoders.
	SkipFieldType = ftype.Skip

	// InlineFieldType indicates the field's contents should be inlined.
	InlineFieldType = ftype.Inline
)

// Array constructs a Field of Type Array carrying an ArrayMarshaler.
//
// The marshaler MUST NOT be nil. The encoder will delegate array serialization
// to this value without using reflection.
func Array(key string, m base.ArrayMarshaler) Field {
	return Field{
		Key:       key,
		Type:      ftype.Array,
		Interface: m,
	}
}

// Object constructs a Field of Type Object carrying an ObjectMarshaler.
//
// The marshaler MUST NOT be nil. The encoder will call MarshalLogObject to
// emit nested fields.
func Object(key string, m base.ObjectMarshaler) Field {
	return Field{
		Key:       key,
		Type:      ftype.Object,
		Interface: m,
	}
}

// Inline constructs a Field of Type Inline carrying an ObjectMarshaler whose
// fields will be inlined into the surrounding object.
//
// The encoder MUST NOT create an additional nesting level for this field.
func Inline(m base.ObjectMarshaler) Field {
	return Field{
		Type:      ftype.Inline,
		Interface: m,
	}
}

// Namespace constructs a Field of Type Namespace that opens a logical
// sub-namespace under the given key.
//
// Encoders typically implement this as a nested object or scope in the output.
func Namespace(key string) Field {
	return Field{
		Key:  key,
		Type: ftype.Namespace,
	}
}

// Skip constructs a Field of Type Skip which is a no-op placeholder.
//
// Encoders MUST ignore Skip fields and emit nothing for them.
func Skip() Field {
	return Field{
		Type: ftype.Skip,
	}
}

// Binary constructs a Field of Type Binary carrying an opaque byte slice.
//
// The slice is NOT copied; callers MUST NOT mutate it after constructing the
// field if encoders can access it concurrently.
func Binary(key string, b []byte) Field {
	return Field{
		Key:       key,
		Type:      ftype.Binary,
		Interface: b,
	}
}

// BinaryPtr constructs a Binary field from a *[]byte.
//
// If value is nil, Skip is returned.
func BinaryPtr(key string, value *[]byte) Field {
	if value == nil {
		return Skip()
	}
	return Binary(key, *value)
}

// ByteString constructs a Field of Type ByteString carrying a UTF-8 byte slice.
//
// The slice is NOT copied. Encoders MAY validate UTF-8 but are not required to.
func ByteString(key string, b []byte) Field {
	return Field{
		Key:       key,
		Type:      ftype.ByteString,
		Interface: b,
	}
}

// ByteStringPtr constructs a ByteString field from a *[]byte.
//
// If value is nil, Skip is returned.
func ByteStringPtr(key string, value *[]byte) Field {
	if value == nil {
		return Skip()
	}
	return ByteString(key, *value)
}

// String constructs a Field of Type String carrying a textual value.
func String(key, value string) Field {
	return Field{
		Key:    key,
		Type:   ftype.String,
		String: value,
	}
}

// StringPtr constructs a String field from a *string.
//
// If value is nil, Skip is returned.
func StringPtr(key string, value *string) Field {
	if value == nil {
		return Skip()
	}
	return String(key, *value)
}

// Bool constructs a Field of Type Bool carrying a boolean value.
//
// The value is stored as 0 (false) or 1 (true) in Field.Integer.
func Bool(key string, value bool) Field {
	var i int64
	if value {
		i = 1
	}
	return Field{
		Key:     key,
		Type:    ftype.Bool,
		Integer: i,
	}
}

// BoolPtr constructs a Bool field from a *bool.
//
// If value is nil, Skip is returned.
func BoolPtr(key string, value *bool) Field {
	if value == nil {
		return Skip()
	}
	return Bool(key, *value)
}

// Int64 constructs a Field of Type Int64 carrying an int64 value.
func Int64(key string, value int64) Field {
	return Field{
		Key:     key,
		Type:    ftype.Int64,
		Integer: value,
	}
}

// Int64Ptr constructs an Int64 field from a *int64.
//
// If value is nil, Skip is returned.
func Int64Ptr(key string, value *int64) Field {
	if value == nil {
		return Skip()
	}
	return Int64(key, *value)
}

// Int32 constructs a Field of Type Int32 carrying an int32 value.
//
// The value is widened to int64 in storage and narrowed back by encoders.
func Int32(key string, value int32) Field {
	return Field{
		Key:     key,
		Type:    ftype.Int32,
		Integer: int64(value),
	}
}

// Int32Ptr constructs an Int32 field from a *int32.
//
// If value is nil, Skip is returned.
func Int32Ptr(key string, value *int32) Field {
	if value == nil {
		return Skip()
	}
	return Int32(key, *value)
}

// Int16 constructs a Field of Type Int16 carrying an int16 value.
func Int16(key string, value int16) Field {
	return Field{
		Key:     key,
		Type:    ftype.Int16,
		Integer: int64(value),
	}
}

// Int16Ptr constructs an Int16 field from a *int16.
//
// If value is nil, Skip is returned.
func Int16Ptr(key string, value *int16) Field {
	if value == nil {
		return Skip()
	}
	return Int16(key, *value)
}

// Int8 constructs a Field of Type Int8 carrying an int8 value.
func Int8(key string, value int8) Field {
	return Field{
		Key:     key,
		Type:    ftype.Int8,
		Integer: int64(value),
	}
}

// Int8Ptr constructs an Int8 field from a *int8.
//
// If value is nil, Skip is returned.
func Int8Ptr(key string, value *int8) Field {
	if value == nil {
		return Skip()
	}
	return Int8(key, *value)
}

// Int constructs a Field of Type Int64 from a Go int value.
//
// This normalizes platform-dependent int to the canonical Int64 representation.
func Int(key string, value int) Field {
	return Field{
		Key:     key,
		Type:    ftype.Int64,
		Integer: int64(value),
	}
}

// IntPtr constructs an Int64 field from a *int.
//
// If value is nil, Skip is returned.
func IntPtr(key string, value *int) Field {
	if value == nil {
		return Skip()
	}
	return Int(key, *value)
}

// Uint64 constructs a Field of Type Uint64 carrying an uint64 value.
//
// The value is converted to int64 using Go's standard conversion rules.
// Encoders MUST reconstruct the original unsigned value using uint64(f.Integer).
func Uint64(key string, value uint64) Field {
	return Field{
		Key:     key,
		Type:    ftype.Uint64,
		Integer: int64(value),
	}
}

// Uint64Ptr constructs an Uint64 field from a *uint64.
//
// If value is nil, Skip is returned.
func Uint64Ptr(key string, value *uint64) Field {
	if value == nil {
		return Skip()
	}
	return Uint64(key, *value)
}

// Uint32 constructs a Field of Type Uint32 carrying an uint32 value.
func Uint32(key string, value uint32) Field {
	return Field{
		Key:     key,
		Type:    ftype.Uint32,
		Integer: int64(value),
	}
}

// Uint32Ptr constructs an Uint32 field from a *uint32.
//
// If value is nil, Skip is returned.
func Uint32Ptr(key string, value *uint32) Field {
	if value == nil {
		return Skip()
	}
	return Uint32(key, *value)
}

// Uint16 constructs a Field of Type Uint16 carrying an uint16 value.
func Uint16(key string, value uint16) Field {
	return Field{
		Key:     key,
		Type:    ftype.Uint16,
		Integer: int64(value),
	}
}

// Uint16Ptr constructs an Uint16 field from a *uint16.
//
// If value is nil, Skip is returned.
func Uint16Ptr(key string, value *uint16) Field {
	if value == nil {
		return Skip()
	}
	return Uint16(key, *value)
}

// Uint8 constructs a Field of Type Uint8 carrying an uint8 value.
func Uint8(key string, value uint8) Field {
	return Field{
		Key:     key,
		Type:    ftype.Uint8,
		Integer: int64(value),
	}
}

// Uint8Ptr constructs an Uint8 field from a *uint8.
//
// If value is nil, Skip is returned.
func Uint8Ptr(key string, value *uint8) Field {
	if value == nil {
		return Skip()
	}
	return Uint8(key, *value)
}

// Uint constructs a Field of Type Uint64 from a Go uint value.
func Uint(key string, value uint) Field {
	return Field{
		Key:     key,
		Type:    ftype.Uint64,
		Integer: int64(value),
	}
}

// UintPtr constructs an Uint64 field from a *uint.
//
// If value is nil, Skip is returned.
func UintPtr(key string, value *uint) Field {
	if value == nil {
		return Skip()
	}
	return Uint(key, *value)
}

// Uintptr constructs a Field of Type Uintptr carrying an uintptr value.
//
// The value is widened to int64 and MUST be interpreted as uintptr by encoders.
func Uintptr(key string, value uintptr) Field {
	return Field{
		Key:     key,
		Type:    ftype.Uintptr,
		Integer: int64(value),
	}
}

// UintptrPtr constructs an Uintptr field from a *uintptr.
//
// If value is nil, Skip is returned.
func UintptrPtr(key string, value *uintptr) Field {
	if value == nil {
		return Skip()
	}
	return Uintptr(key, *value)
}

// Float64 constructs a Field of Type Float64 carrying a float64 value.
//
// The IEEE-754 bits are stored in Field.Integer via math.Float64bits.
func Float64(key string, value float64) Field {
	return Field{
		Key:     key,
		Type:    ftype.Float64,
		Integer: int64(math.Float64bits(value)),
	}
}

// Float64Ptr constructs a Float64 field from a *float64.
//
// If value is nil, Skip is returned.
func Float64Ptr(key string, value *float64) Field {
	if value == nil {
		return Skip()
	}
	return Float64(key, *value)
}

// Float32 constructs a Field of Type Float32 carrying a float32 value.
//
// The IEEE-754 bits are stored in Field.Integer via math.Float32bits.
func Float32(key string, value float32) Field {
	return Field{
		Key:     key,
		Type:    ftype.Float32,
		Integer: int64(math.Float32bits(value)),
	}
}

// Float32Ptr constructs a Float32 field from a *float32.
//
// If value is nil, Skip is returned.
func Float32Ptr(key string, value *float32) Field {
	if value == nil {
		return Skip()
	}
	return Float32(key, *value)
}

// Complex128 constructs a Field of Type Complex128 carrying a complex128 value.
func Complex128(key string, value complex128) Field {
	return Field{
		Key:       key,
		Type:      ftype.Complex128,
		Interface: value,
	}
}

// Complex128Ptr constructs a Complex128 field from a *complex128.
//
// If value is nil, Skip is returned.
func Complex128Ptr(key string, value *complex128) Field {
	if value == nil {
		return Skip()
	}
	return Complex128(key, *value)
}

// Complex64 constructs a Field of Type Complex64 carrying a complex64 value.
func Complex64(key string, value complex64) Field {
	return Field{
		Key:       key,
		Type:      ftype.Complex64,
		Interface: value,
	}
}

// Complex64Ptr constructs a Complex64 field from a *complex64.
//
// If value is nil, Skip is returned.
func Complex64Ptr(key string, value *complex64) Field {
	if value == nil {
		return Skip()
	}
	return Complex64(key, *value)
}

// Duration constructs a Field of Type Duration carrying a time.Duration value.
//
// The duration is stored in nanoseconds in Field.Integer.
func Duration(key string, value time.Duration) Field {
	return Field{
		Key:     key,
		Type:    ftype.Duration,
		Integer: int64(value),
	}
}

// DurationPtr constructs a Duration field from a *time.Duration.
//
// If value is nil, Skip is returned.
func DurationPtr(key string, value *time.Duration) Field {
	if value == nil {
		return Skip()
	}
	return Duration(key, *value)
}

// Time constructs a Field of Type Time from the provided time.Time.
//
// The timestamp is normalized to UTC and stored as UnixNano in Field.Integer.
// If the original location is not UTC, it is stored in Field.Interface.
func Time(key string, t time.Time) Field {
	utc := t.UTC()
	var loc *time.Location
	if t.Location() != time.UTC {
		loc = t.Location()
	}
	return Field{
		Key:       key,
		Type:      ftype.Time,
		Integer:   utc.UnixNano(),
		Interface: loc,
	}
}

// TimePtr constructs a Time field from a *time.Time.
//
// If value is nil, Skip is returned.
func TimePtr(key string, value *time.Time) Field {
	if value == nil {
		return Skip()
	}
	return Time(key, *value)
}

// TimeInLocation constructs a Time field by first converting t to loc.
//
// The resulting UnixNano and location are stored in Integer and Interface.
func TimeInLocation(key string, t time.Time, loc *time.Location) Field {
	if loc != nil {
		t = t.In(loc)
	}
	return Field{
		Key:       key,
		Type:      ftype.Time,
		Integer:   t.UnixNano(),
		Interface: loc,
	}
}

// TimeFull constructs a Field of Type TimeFull carrying the full time.Time
// value in Field.Interface without normalization to UnixNano.
func TimeFull(key string, t time.Time) Field {
	return Field{
		Key:       key,
		Type:      ftype.TimeFull,
		Interface: t,
	}
}

// TimeFullPtr constructs a TimeFull field from a *time.Time.
//
// If value is nil, Skip is returned.
func TimeFullPtr(key string, value *time.Time) Field {
	if value == nil {
		return Skip()
	}
	return TimeFull(key, *value)
}

// Error constructs a Field of Type Error carrying an error value.
//
// The error may be nil; encoding helpers SHOULD handle nil and typed-nil
// receivers gracefully.
func Error(key string, err error) Field {
	return Field{
		Key:       key,
		Type:      ftype.Error,
		Interface: err,
	}
}

// ErrorPtr constructs an Error field from a *error.
//
// If value is nil, Skip is returned. If *value is nil, a nil error is stored.
func ErrorPtr(key string, value *error) Field {
	if value == nil {
		return Skip()
	}
	return Error(key, *value)
}

// Stringer constructs a Field of Type Stringer carrying a fmt.Stringer value.
//
// The value may be nil; encoding helpers are expected to guard against panics
// from String() and report them as diagnostic strings if needed.
func Stringer(key string, s fmt.Stringer) Field {
	return Field{
		Key:       key,
		Type:      ftype.Stringer,
		Interface: s,
	}
}

// StringerPtr constructs a Stringer field from a *fmt.Stringer.
//
// If value is nil, Skip is returned.
func StringerPtr(key string, value *fmt.Stringer) Field {
	if value == nil {
		return Skip()
	}
	return Stringer(key, *value)
}

// Reflect constructs a Field of Type Reflect carrying an arbitrary value.
//
// Encoders will typically use reflection to inspect and serialize the value.
// This is flexible but slower than using more specific Types.
func Reflect(key string, v interface{}) Field {
	return Field{
		Key:       key,
		Type:      ftype.Reflect,
		Interface: v,
	}
}

// Any chooses a Field constructor based on the dynamic type of v.
//
// This is a convenience helper and MAY introduce additional overhead compared
// to using specific constructors directly.
func Any(key string, v interface{}) Field {
	switch val := v.(type) {
	case nil:
		return Skip()
	case string:
		return String(key, val)
	case []byte:
		return ByteString(key, val)
	case bool:
		return Bool(key, val)
	case int:
		return Int(key, val)
	case int64:
		return Int64(key, val)
	case int32:
		return Int32(key, val)
	case int16:
		return Int16(key, val)
	case int8:
		return Int8(key, val)
	case uint:
		return Uint(key, val)
	case uint64:
		return Uint64(key, val)
	case uint32:
		return Uint32(key, val)
	case uint16:
		return Uint16(key, val)
	case uint8:
		return Uint8(key, val)
	case uintptr:
		return Uintptr(key, val)
	case float64:
		return Float64(key, val)
	case float32:
		return Float32(key, val)
	case complex128:
		return Complex128(key, val)
	case complex64:
		return Complex64(key, val)
	case time.Time:
		return TimeFull(key, val)
	case time.Duration:
		return Duration(key, val)
	case error:
		return Error(key, val)
	case fmt.Stringer:
		return Stringer(key, val)
	case base.ArrayMarshaler:
		return Array(key, val)
	case base.ObjectMarshaler:
		return Object(key, val)
	default:
		return Reflect(key, v)
	}
}
