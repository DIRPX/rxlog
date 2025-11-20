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

package ftype

// Type identifies which member of the Field union is active and defines how
// the underlying value MUST be interpreted and serialized.
//
// Implementations MUST set Type consistently with how the value is stored in
// the associated Field (e.g., which of Integer, String, or Interface is used).
// Callers and encoders MUST NOT assume any interpretation of the stored data
// that contradicts the Type value.
type Type uint8

const (
	// Unknown is the zero value for Type and represents an uninitialized
	// or invalid field kind.
	//
	// Using a Field with Unknown in an encoder MUST be treated as a programmer
	// error. Helper methods that attempt to serialize such a field (for example,
	// AddTo) SHOULD panic rather than silently emitting incorrect data.
	Unknown Type = iota

	// Array indicates that the field carries a value implementing the
	// ArrayMarshaler interface.
	//
	// When encountering Array, the encoder MUST treat Field.Interface as an
	// ArrayMarshaler and MUST serialize it as an array by delegating to the
	// ArrayMarshaler implementation, without relying on reflection.
	Array

	// Object indicates that the field carries a value implementing the
	// ObjectMarshaler interface.
	//
	// When encountering Object, the encoder MUST treat Field.Interface as an
	// ObjectMarshaler and MUST serialize it as a structured object by invoking
	// the ObjectMarshaler implementation to emit nested fields.
	Object

	// Binary indicates that the field contains an opaque sequence of bytes.
	//
	// For Binary, Field.Interface MUST hold a []byte. The encoder MUST NOT make
	// assumptions about character encoding (for example, UTF-8) and MAY emit
	// the bytes as raw binary, base-encoded text, or any other backend-specific
	// representation, as long as it is documented.
	Binary

	// Bool indicates that the field represents a boolean value.
	//
	// For Bool, the value MUST be derived from Field.Integer (for example,
	// 0 for false and 1 for true) and exposed to the encoder as a native
	// boolean. Callers MUST NOT store non-boolean data under this Type.
	Bool

	// ByteString indicates that the field carries a sequence of bytes that
	// SHOULD be interpreted as UTF-8 encoded text.
	//
	// For ByteString, Field.Interface MUST hold a []byte. It is semantically
	// similar to String, but MAY avoid an intermediate string allocation when
	// the data is already available as bytes. Encoders MAY validate UTF-8 or
	// treat the bytes as-is, but SHOULD document their behavior.
	ByteString

	// Complex128 indicates that the field carries a complex128 value.
	//
	// For Complex128, Field.Interface MUST hold a complex128. The encoder MAY
	// choose any reasonable textual or structured representation, but SHOULD
	// preserve enough information to reconstruct the original value if needed.
	Complex128

	// Complex64 indicates that the field carries a complex64 value.
	//
	// For Complex64, Field.Interface MUST hold a complex64. The concrete
	// encoded form MAY differ from Complex128, but the semantics are otherwise
	// analogous.
	Complex64

	// Duration indicates that the field carries a time.Duration value.
	//
	// For Duration, Field.Integer MUST store the duration as a number of
	// nanoseconds (int64), and the encoder MUST interpret it accordingly when
	// reconstructing a time.Duration. Encoders MAY choose textual or numeric
	// representations, but SHOULD document the chosen format.
	Duration

	// Float64 indicates that the field carries a float64 value.
	//
	// For Float64, Field.Integer MUST hold the IEEE-754 bit pattern of the
	// float64 value (for example, via math.Float64bits). The encoder MUST
	// reconstruct the float64 from these bits before emitting it.
	Float64

	// Float32 indicates that the field carries a float32 value.
	//
	// For Float32, Field.Integer MUST hold the IEEE-754 bit pattern of the
	// float32 value (for example, via math.Float32bits). The encoder MUST
	// reconstruct the float32 from these bits before emitting it.
	Float32

	// Int64 indicates that the field carries an int64 value.
	//
	// For Int64, Field.Integer MUST store the signed 64-bit integer directly
	// and the encoder MUST serialize it as a signed 64-bit numeric value.
	Int64

	// Int32 indicates that the field carries an int32 value.
	//
	// For Int32, Field.Integer MUST store the value widened to int64, and the
	// encoder MUST narrow it back to int32 before serialization.
	Int32

	// Int16 indicates that the field carries an int16 value.
	//
	// For Int16, Field.Integer MUST store the value widened to int64, and the
	// encoder MUST narrow it back to int16 before serialization.
	Int16

	// Int8 indicates that the field carries an int8 value.
	//
	// For Int8, Field.Integer MUST store the value widened to int64, and the
	// encoder MUST narrow it back to int8 before serialization.
	Int8

	// String indicates that the field carries a textual string value.
	//
	// For String, the textual data MUST be stored in Field.String and the
	// encoder MUST write it out as a string in the encoded output. Callers
	// SHOULD use String instead of ByteString when they already have a Go
	// string.
	String

	// Time indicates that the field carries a time.Time value represented
	// indirectly via an int64 UnixNano value.
	//
	// For Time, Field.Integer MUST store the UTC UnixNano timestamp. Optionally,
	// Field.Interface MAY hold a *time.Location that specifies the time zone
	// used when encoding. If Field.Interface is nil, the encoder SHOULD default
	// to UTC.
	Time

	// TimeFull indicates that the field carries a time.Time value stored
	// directly in Field.Interface.
	//
	// For TimeFull, Field.Interface MUST hold a time.Time. No normalization to
	// UnixNano is performed; the encoder receives the full time.Time object as-is
	// and MAY encode any of its components according to its own rules.
	TimeFull

	// Uint64 indicates that the field carries an uint64 value.
	//
	// For Uint64, Field.Integer MUST store the unsigned 64-bit integer (after
	// appropriate conversion) and the encoder MUST serialize it as an unsigned
	// 64-bit numeric value.
	Uint64

	// Uint32 indicates that the field carries an uint32 value.
	//
	// For Uint32, Field.Integer MUST store the value widened to int64, and the
	// encoder MUST narrow it back to uint32 before serialization.
	Uint32

	// Uint16 indicates that the field carries an uint16 value.
	//
	// For Uint16, Field.Integer MUST store the value widened to int64, and the
	// encoder MUST narrow it back to uint16 before serialization.
	Uint16

	// Uint8 indicates that the field carries an uint8 value.
	//
	// For Uint8, Field.Integer MUST store the value widened to int64, and the
	// encoder MUST narrow it back to uint8 before serialization.
	Uint8

	// Uintptr indicates that the field carries an uintptr value.
	//
	// For Uintptr, Field.Integer MUST store the value widened to int64, and the
	// encoder MUST narrow it back to uintptr before serialization. This Type is
	// primarily useful for debugging low-level pointer-heavy code and SHOULD
	// NOT be used for stable external APIs.
	Uintptr

	// Reflect indicates that the field carries an arbitrary interface{} value
	// in Field.Interface.
	//
	// When encountering Reflect, the encoder MUST use reflection to inspect
	// and serialize the underlying value. This Type is flexible but generally
	// slower than using more specific Types or explicit marshalers, and SHOULD
	// be reserved for cases where the exact shape of the value is not known
	// ahead of time.
	Reflect

	// Namespace marks the beginning of a new logical namespace (often
	// represented as a nested object or scope in the encoded output).
	//
	// When encountering Namespace, the encoder SHOULD open a new nested scope
	// under the field's key. All subsequent fields emitted while this namespace
	// is active SHOULD be grouped under that scope until the encoder closes it.
	Namespace

	// Stringer indicates that the field carries a value implementing
	// fmt.Stringer in Field.Interface.
	//
	// For Stringer, the encoder MUST obtain the textual representation by
	// calling the String() method. Helper code (such as encodeStringer) SHOULD
	// guard against panics (for example, due to nil receivers) and MAY encode
	// such failures as diagnostic strings instead of propagating the panic.
	Stringer

	// Error indicates that the field carries a value implementing the error
	// interface in Field.Interface.
	//
	// For Error, the encoder SHOULD serialize at least the error's message
	// (i.e., the result of Error()) and MAY attach additional structured
	// metadata if available. Callers SHOULD reserve this Type for actual error
	// values and MUST NOT use it for arbitrary strings.
	Error

	// Skip indicates that this field should be treated as a no-op and
	// completely ignored by the encoder.
	//
	// When encountering Skip, the encoder MUST NOT emit anything for this
	// field. This Type is useful as a sentinel when constructing fields
	// conditionally (for example, helpers MAY return Skip for "no-op" cases).
	Skip

	// Inline indicates that the field carries an ObjectMarshaler whose
	// fields should be serialized directly into the surrounding object.
	//
	// For Inline, Field.Interface MUST hold an ObjectMarshaler. The encoder
	// MUST invoke its MarshalLogObject method with the current encoder and
	// MUST NOT introduce an additional nesting level or key for this field.
	// This allows nested object fields to be inlined into the current logging
	// scope.
	Inline
)
