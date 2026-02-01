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

// Field constructor shortcuts providing convenient access to rxapi/field
// constructor functions.
var (
	// Array constructs a Field carrying an ArrayMarshaler.
	Array = field.Array

	// Object constructs a Field carrying an ObjectMarshaler.
	Object = field.Object

	// Inline constructs a Field whose fields will be inlined.
	Inline = field.Inline

	// Namespace constructs a Field that opens a logical sub-namespace.
	Namespace = field.Namespace

	// Skip constructs a no-op placeholder Field.
	Skip = field.Skip

	// Binary constructs a Field carrying an opaque byte slice.
	Binary = field.Binary

	// BinaryPtr constructs a Binary field from a *[]byte.
	BinaryPtr = field.BinaryPtr

	// ByteString constructs a Field carrying a UTF-8 byte slice.
	ByteString = field.ByteString

	// ByteStringPtr constructs a ByteString field from a *[]byte.
	ByteStringPtr = field.ByteStringPtr

	// String constructs a Field carrying a textual value.
	String = field.String

	// StringPtr constructs a String field from a *string.
	StringPtr = field.StringPtr

	// Bool constructs a Field carrying a boolean value.
	Bool = field.Bool

	// BoolPtr constructs a Bool field from a *bool.
	BoolPtr = field.BoolPtr

	// Int64 constructs a Field carrying an int64 value.
	Int64 = field.Int64

	// Int64Ptr constructs an Int64 field from a *int64.
	Int64Ptr = field.Int64Ptr

	// Int32 constructs a Field carrying an int32 value.
	Int32 = field.Int32

	// Int32Ptr constructs an Int32 field from a *int32.
	Int32Ptr = field.Int32Ptr

	// Int16 constructs a Field carrying an int16 value.
	Int16 = field.Int16

	// Int16Ptr constructs an Int16 field from a *int16.
	Int16Ptr = field.Int16Ptr

	// Int8 constructs a Field carrying an int8 value.
	Int8 = field.Int8

	// Int8Ptr constructs an Int8 field from a *int8.
	Int8Ptr = field.Int8Ptr

	// Int constructs a Field carrying an int value.
	Int = field.Int

	// IntPtr constructs an Int field from a *int.
	IntPtr = field.IntPtr

	// Uint64 constructs a Field carrying a uint64 value.
	Uint64 = field.Uint64

	// Uint64Ptr constructs a Uint64 field from a *uint64.
	Uint64Ptr = field.Uint64Ptr

	// Uint32 constructs a Field carrying a uint32 value.
	Uint32 = field.Uint32

	// Uint32Ptr constructs a Uint32 field from a *uint32.
	Uint32Ptr = field.Uint32Ptr

	// Uint16 constructs a Field carrying a uint16 value.
	Uint16 = field.Uint16

	// Uint16Ptr constructs a Uint16 field from a *uint16.
	Uint16Ptr = field.Uint16Ptr

	// Uint8 constructs a Field carrying a uint8 value.
	Uint8 = field.Uint8

	// Uint8Ptr constructs a Uint8 field from a *uint8.
	Uint8Ptr = field.Uint8Ptr

	// Uint constructs a Field carrying a uint value.
	Uint = field.Uint

	// UintPtr constructs a Uint field from a *uint.
	UintPtr = field.UintPtr

	// Uintptr constructs a Field carrying a uintptr value.
	Uintptr = field.Uintptr

	// UintptrPtr constructs a Uintptr field from a *uintptr.
	UintptrPtr = field.UintptrPtr

	// Float64 constructs a Field carrying a float64 value.
	Float64 = field.Float64

	// Float64Ptr constructs a Float64 field from a *float64.
	Float64Ptr = field.Float64Ptr

	// Float32 constructs a Field carrying a float32 value.
	Float32 = field.Float32

	// Float32Ptr constructs a Float32 field from a *float32.
	Float32Ptr = field.Float32Ptr

	// Complex128 constructs a Field carrying a complex128 value.
	Complex128 = field.Complex128

	// Complex128Ptr constructs a Complex128 field from a *complex128.
	Complex128Ptr = field.Complex128Ptr

	// Complex64 constructs a Field carrying a complex64 value.
	Complex64 = field.Complex64

	// Complex64Ptr constructs a Complex64 field from a *complex64.
	Complex64Ptr = field.Complex64Ptr

	// Duration constructs a Field carrying a time.Duration value.
	Duration = field.Duration

	// DurationPtr constructs a Duration field from a *time.Duration.
	DurationPtr = field.DurationPtr

	// Time constructs a Field from a time.Time (normalized to UTC).
	Time = field.Time

	// TimePtr constructs a Time field from a *time.Time.
	TimePtr = field.TimePtr

	// TimeFull constructs a Field carrying a full time.Time value.
	TimeFull = field.TimeFull

	// TimeFullPtr constructs a TimeFull field from a *time.Time.
	TimeFullPtr = field.TimeFullPtr

	// Error constructs a Field carrying an error value.
	Error = field.Error

	// ErrorPtr constructs an Error field from a *error.
	ErrorPtr = field.ErrorPtr

	// Stringer constructs a Field carrying a fmt.Stringer value.
	Stringer = field.Stringer

	// StringerPtr constructs a Stringer field from a *fmt.Stringer.
	StringerPtr = field.StringerPtr

	// Reflect constructs a Field carrying an arbitrary value.
	Reflect = field.Reflect

	// Any chooses a Field constructor based on the dynamic type of v.
	Any = field.Any
)

// TimeInLocation constructs a Time field by first converting t to loc.
//
// This is a convenience wrapper around field.TimeInLocation. See
// rxapi/field.TimeInLocation for full documentation.
func TimeInLocation(key string, t time.Time, loc *time.Location) Field {
	return field.TimeInLocation(key, t, loc)
}
