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

package field

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/encoder/base"
	"dirpx.dev/rxlog/rxapi/encoder/encoders"
	ftype "dirpx.dev/rxlog/rxapi/field/type"
)

// Field represents a single logical logging attribute (a key-value pair)
// that will eventually be serialized into the output.
//
// A Field is a compact, type-tagged container that allows the logger to defer
// expensive marshaling work until it is actually needed (for example, when the
// log level is enabled). The semantics of the stored data MUST be interpreted
// solely through the Type value and the conventions associated with that type.
//
// The concrete meaning of the stored data depends on the Type value:
//   - For simple scalar types (integers, unsigned integers, floats, booleans,
//     durations, times), the primary representation is usually stored in
//     Integer (and MAY use Interface for supplemental metadata such as
//     *time.Location).
//   - For string-like values, String or Interface MAY be used depending on the
//     specific Type.
//   - For more complex payloads (errors, marshalers, arbitrary interfaces,
//     reflected values, byte slices), Interface carries the underlying object.
//
// Callers and implementations MUST ensure that Type and the backing storage
// fields (Integer, String, Interface) are consistent with each other as
// described by the ftype.Type constants. Mismatches are considered programmer
// errors and MAY cause panics when encoding (for example, in AddTo).
type Field struct {
	// Key is the logical name of the field as it will appear in the encoded
	// log output (for example, "user_id", "request_id", "duration").
	//
	// Key SHOULD be stable and human-readable, since it typically forms part of
	// the public logging contract consumed by downstream systems (log search,
	// metrics extraction, dashboards, etc.). Callers SHOULD prefer lowercase
	// or snake_case keys for consistency, but specific naming conventions are
	// left to the application.
	//
	// Key MAY be empty, but such usage is discouraged; encoders and tools MAY
	// treat empty keys specially or ignore them.
	Key string

	// Type identifies which kind of value is stored in this Field and how the
	// remaining fields (Integer, String, Interface) MUST be interpreted when
	// the Field is serialized.
	//
	// Type MUST be a valid ftype.Type value and MUST be consistent with how the
	// Field was constructed. If Type and the backing storage (Integer, String,
	// Interface) are inconsistent, encoding helpers (such as AddTo) MAY panic
	// or produce undefined output. Callers MUST NOT rely on any behavior in
	// the presence of such inconsistencies.
	Type ftype.Type

	// Integer stores all integer-like data in a unified representation:
	//   - Signed and unsigned integers (int8/16/32/64, uint8/16/32/64, uintptr)
	//   - Booleans (commonly encoded as 0 for false and 1 for true)
	//   - Floating-point values (float32/float64) as raw IEEE-754 bit patterns
	//   - Durations and timestamps (e.g., time.Duration, time.Time as UnixNano)
	//
	// The exact interpretation of Integer MUST be derived from Type; consumers
	// MUST NOT read or interpret Integer without first checking Type and
	// applying the rules defined for that type. When Type indicates a variant
	// that does not use an integer-like representation, the value of Integer
	// is undefined and MUST be ignored.
	Integer int64

	// String carries string-based payloads for those Types that represent a
	// string value directly (for example, Type == ftype.String).
	//
	// String MAY also be used to hold preformatted error messages or other
	// textual data where a simple string is sufficient and no additional type
	// information is required. For Types that do not use a direct string
	// representation, the contents of String are undefined and encoders MUST
	// ignore it.
	String string

	// Interface stores all non-primitive or auxiliary values that do not fit
	// into Integer or String alone. Depending on Type, this MAY hold:
	//   - []byte for binary or byte-string fields
	//   - ArrayMarshaler or ObjectMarshaler implementations
	//   - error, fmt.Stringer, or arbitrary any for reflection-based encoding
	//   - time.Time or *time.Location for time-related fields
	//
	// The concrete Go type stored in Interface MUST be compatible with the
	// current Type, as documented by the ftype.Type constants. Callers MUST
	// only perform type assertions that are consistent with Type; attempting
	// to assert an incompatible type is a programmer error and WILL typically
	// result in a panic.
	//
	// For Types that do not require an auxiliary value, Interface MAY be nil
	// and MUST be ignored by encoders.
	Interface interface{}
}

// AddTo materializes this Field into the provided ObjectEncoder and appends
// its encoded representation to dst.
//
// The method translates the compact internal representation (Type, Integer,
// String, Interface) back into the high-level value expected by enc, then
// delegates to the corresponding ObjectEncoder method (or helper) for the
// concrete type.
//
// This method is performance-sensitive and assumes that the Field was
// constructed consistently with its Type. If Type and the underlying storage
// (Integer / String / Interface) do not match, the type assertions below will
// typically panic. In normal usage, Fields SHOULD be created via well-typed
// helper constructors; any mismatch SHOULD be considered a programmer bug.
//
// Some encoder operations (for example, AddArray, AddObject, AddReflected,
// and helpers like encodeStringer / encodeError) can return an error. Instead
// of propagating that error to the caller, AddTo encodes it as an additional
// string field with the key "<Key>Error" so that backends and operators CAN
// still see that serialization for this attribute failed, while the main log
// entry remains emitted.
//
// The returned Buffer MUST be treated as the new authoritative destination;
// it MAY differ from dst if the underlying storage had to grow.
func (f Field) AddTo(enc base.ObjectEncoder, dst *buffer.Buffer) *buffer.Buffer {
	var err error

	switch f.Type {
	case ftype.Array:
		// Array fields delegate their serialization to an ArrayMarshaler
		// implementation, enabling custom array encodings without reflection.
		dst, err = enc.AddArray(dst, f.Key, f.Interface.(base.ArrayMarshaler))

	case ftype.Object:
		// Object fields delegate their serialization to an ObjectMarshaler,
		// which is responsible for emitting nested fields.
		dst, err = enc.AddObject(dst, f.Key, f.Interface.(base.ObjectMarshaler))

	case ftype.Inline:
		// Inline object fields behave like Object, but their nested fields are
		// written directly into the surrounding scope rather than under a
		// dedicated key.
		dst, err = f.Interface.(base.ObjectMarshaler).MarshalLogObject(enc, dst)

	case ftype.Binary:
		// Binary fields carry opaque bytes; the encoder decides whether to emit
		// them as raw bytes or a base-encoded representation.
		dst = enc.AddBinary(dst, f.Key, f.Interface.([]byte))

	case ftype.Bool:
		// Boolean values are stored as an integer flag (0 for false, 1 for true).
		dst = enc.AddBool(dst, f.Key, f.Integer == 1)

	case ftype.ByteString:
		// ByteString fields carry UTF-8 encoded bytes, avoiding an extra string
		// allocation when data is already available as []byte.
		dst = enc.AddByteString(dst, f.Key, f.Interface.([]byte))

	case ftype.Complex128:
		dst = enc.AddComplex128(dst, f.Key, f.Interface.(complex128))

	case ftype.Complex64:
		dst = enc.AddComplex64(dst, f.Key, f.Interface.(complex64))

	case ftype.Duration:
		// Durations are stored as a nanosecond count in Integer.
		dst = enc.AddDuration(dst, f.Key, time.Duration(f.Integer))

	case ftype.Float64:
		// Float64 values are stored as their IEEE-754 bit pattern in Integer.
		dst = enc.AddFloat64(dst, f.Key, math.Float64frombits(uint64(f.Integer)))

	case ftype.Float32:
		// Float32 values are stored as their IEEE-754 bit pattern in Integer.
		dst = enc.AddFloat32(dst, f.Key, math.Float32frombits(uint32(f.Integer)))

	case ftype.Int64:
		dst = enc.AddInt64(dst, f.Key, f.Integer)

	case ftype.Int32:
		dst = enc.AddInt32(dst, f.Key, int32(f.Integer))

	case ftype.Int16:
		dst = enc.AddInt16(dst, f.Key, int16(f.Integer))

	case ftype.Int8:
		dst = enc.AddInt8(dst, f.Key, int8(f.Integer))

	case ftype.String:
		// String fields use the dedicated String storage for efficiency.
		dst = enc.AddString(dst, f.Key, f.String)

	case ftype.Time:
		// Time values are represented using an int64 UnixNano timestamp in
		// Integer. If a *time.Location is provided in Interface, it defines
		// the time zone used during encoding; otherwise, UTC is used.
		ts := time.Unix(0, f.Integer)
		if f.Interface != nil {
			ts = ts.In(f.Interface.(*time.Location))
		}
		dst = enc.AddTime(dst, f.Key, ts)

	case ftype.TimeFull:
		// TimeFull stores the full time.Time value directly in Interface.
		dst = enc.AddTime(dst, f.Key, f.Interface.(time.Time))

	case ftype.Uint64:
		dst = enc.AddUint64(dst, f.Key, uint64(f.Integer))

	case ftype.Uint32:
		dst = enc.AddUint32(dst, f.Key, uint32(f.Integer))

	case ftype.Uint16:
		dst = enc.AddUint16(dst, f.Key, uint16(f.Integer))

	case ftype.Uint8:
		dst = enc.AddUint8(dst, f.Key, uint8(f.Integer))

	case ftype.Uintptr:
		dst = enc.AddUintptr(dst, f.Key, uintptr(f.Integer))

	case ftype.Reflect:
		// Reflect fields rely on the encoder's reflection-based support to
		// inspect and serialize the underlying value in Interface.
		dst, err = enc.AddReflected(dst, f.Key, f.Interface)

	case ftype.Namespace:
		// Namespace opens a new logical scope; subsequent fields are grouped
		// under this name until the encoder closes the namespace.
		dst = enc.OpenNamespace(dst, f.Key)

	case ftype.Stringer:
		// Stringer fields obtain their representation by calling String() on
		// the stored fmt.Stringer. EncodeStringer adds panic safety and
		// encodes the resulting string as a field.
		dst, err = encoders.EncodeStringer(dst, f.Key, f.Interface, enc)

	case ftype.Error:
		// Error fields serialize values implementing error, typically by
		// capturing their Error() message and possibly additional metadata.
		dst, err = encoders.EncodeError(dst, f.Key, f.Interface.(error), enc)

	case ftype.Skip:
		// Skip explicitly represents a no-op field: nothing is written for it.
		// This is useful for conditional field construction.

	default:
		// Any unknown Type value indicates a bug in field construction or an
		// out-of-sync enum definition.
		panic(fmt.Sprintf("unknown field type: %v", f.Type))
	}

	// If any of the encoder calls reported an error, surface it as a companion
	// "<Key>Error" field instead of returning it, so logging can continue.
	if err != nil {
		// We intentionally ignore a potential error from AddString here:
		// if recording the "<Key>Error" field itself fails, the logger is
		// already in a degraded state. Escalating further by propagating
		// another error would provide little additional value.
		dst = enc.AddString(dst, fmt.Sprintf("%sError", f.Key), err.Error())
	}

	return dst
}

// Equals reports whether this Field and the provided Field represent the same
// logical logging attribute.
//
// Equality is defined in a type- and representation-aware way:
//
//   - The Type and Key MUST match exactly, otherwise Equals returns false.
//   - For Binary and ByteString fields, the underlying byte slices are compared
//     using bytes.Equal so that two distinct []byte values with identical
//     contents are considered equal.
//   - For complex, non-primitive payloads (arrays, objects, errors, and
//     reflection-based values), Interface is compared using reflect.DeepEqual
//     to account for structural equality rather than pointer identity.
//   - For all other Types, the whole Field struct is compared using the ==
//     operator, which includes Integer, String, and Interface.
//
// This method is primarily intended for tests and diagnostic tooling rather
// than hot paths: reflect.DeepEqual MAY be relatively expensive. For Types
// that fall through to the default branch, Interface SHOULD only contain
// comparable values (or be nil), otherwise the struct equality (f == other)
// MAY panic due to Go's comparability rules.
func (f Field) Equals(other Field) bool {
	// Fast-path: Type and Key MUST match, otherwise the fields cannot
	// represent the same logical attribute.
	if f.Type != other.Type {
		return false
	}
	if f.Key != other.Key {
		return false
	}

	switch f.Type {
	case ftype.Binary, ftype.ByteString:
		// Compare byte slices by content, not by slice header identity.
		return bytes.Equal(f.Interface.([]byte), other.Interface.([]byte))

	case ftype.Array, ftype.Object, ftype.Error, ftype.Reflect:
		// For complex payloads, rely on structural equality rather than
		// pointer identity.
		return reflect.DeepEqual(f.Interface, other.Interface)

	default:
		// For all remaining Types, fall back to full struct equality.
		// This compares Integer, String, and Interface along with Type and Key.
		// Callers MUST ensure that Interface holds only comparable values
		// (or nil) for these Types.
		return f == other
	}
}
