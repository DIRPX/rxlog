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

package base

import (
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
)

// ObjectEncoder is a strongly-typed, encoding-agnostic interface for adding a
// map- or struct-like object to the logging context, writing directly into a
// caller-owned buffer.
//
// An ObjectEncoder represents a single, mutable logical object under
// construction. All encoded bytes are appended to the dst buffer passed into
// each method. This lets callers control buffer ownership and reuse (for
// example, via pooling) and minimizes allocations.
//
// Implementations MUST treat dst as the primary output buffer and MUST append
// data to it instead of allocating new buffers, except when growing the
// underlying byte slice is unavoidable. In such cases, the returned Buffer MAY
// differ from the input if the underlying slice has grown.
//
// ObjectEncoder values MUST NOT be used concurrently from multiple goroutines
// unless explicitly documented otherwise; callers SHOULD treat each instance as
// owned by a single goroutine for the duration of its use.
type ObjectEncoder interface {
	// BeginObject starts encoding an object into dst.
	//
	// Implementations MUST append the opening delimiter for an object (for
	// example, '{' for JSON) and return the updated buffer. They MAY update
	// internal state (such as depth or field count) to determine when
	// separators are needed before subsequent fields.
	BeginObject(dst *buffer.Buffer) *buffer.Buffer

	// EndObject finishes encoding an object into dst.
	//
	// Implementations MUST append the closing delimiter for an object (for
	// example, '}' for JSON) and return the updated buffer. Any internal
	// state that tracks the current object scope MAY be reset or rolled back
	// to a parent scope once the object is closed.
	EndObject(dst *buffer.Buffer) *buffer.Buffer

	// AddKey writes a field key and any necessary syntax into dst.
	//
	// Implementations MUST insert any required separators based on the current
	// object state (for example, commas between fields in JSON), encode the
	// key according to the target format (for example, as a quoted string),
	// and append the format-specific key/value separator (for example, ':').
	// This method only writes the key and associated syntax; callers MUST
	// write the value using one of the Add* methods or a marshaler immediately
	// afterwards.
	AddKey(dst *buffer.Buffer, key string) *buffer.Buffer

	// AddArray adds a field with the given key whose value is a nested array
	// encoded by the provided ArrayMarshaler.
	//
	// Implementations MUST treat marshaler as responsible for encoding the
	// entire array value for this key. On failure, AddArray MUST return a
	// non-nil error and MUST NOT append a partially written array value for
	// this field (though bytes written before this field remain intact).
	AddArray(dst *buffer.Buffer, key string, marshaler ArrayMarshaler) (*buffer.Buffer, error)

	// AddObject adds a field with the given key whose value is a nested object
	// encoded by the provided ObjectMarshaler.
	//
	// Implementations MUST treat marshaler as responsible for encoding the
	// entire object value for this key. On failure, AddObject MUST return a
	// non-nil error and MUST NOT append a partially written object value for
	// this field (bytes written before this field are preserved).
	AddObject(dst *buffer.Buffer, key string, marshaler ObjectMarshaler) (*buffer.Buffer, error)

	// AddBinary adds a field with the given key whose value is an arbitrary
	// sequence of bytes.
	//
	// Implementations MUST NOT assume any particular text encoding for value
	// and MAY represent it as raw binary or a base-encoded string, depending
	// on the target format and documented behavior.
	AddBinary(dst *buffer.Buffer, key string, value []byte) *buffer.Buffer

	// AddByteString adds a field with the given key whose value is a sequence
	// of bytes intended to represent UTF-8 encoded text.
	//
	// Implementations MAY validate UTF-8 but are not required to. Callers
	// SHOULD prefer AddByteString over AddBinary when the underlying data is
	// textual. If the encoder supports UTF-8 normalization, it MAY replace
	// invalid sequences according to configuration.
	AddByteString(dst *buffer.Buffer, key string, value []byte) *buffer.Buffer

	// AddBool adds a boolean field with the given key.
	AddBool(dst *buffer.Buffer, key string, value bool) *buffer.Buffer

	// AddComplex128 adds a complex128 field with the given key.
	AddComplex128(dst *buffer.Buffer, key string, value complex128) *buffer.Buffer

	// AddComplex64 adds a complex64 field with the given key.
	AddComplex64(dst *buffer.Buffer, key string, value complex64) *buffer.Buffer

	// AddDuration adds a time.Duration field with the given key.
	//
	// Implementations SHOULD document how durations are represented (for
	// example, numeric values in a specific unit or textual duration strings)
	// and MUST honor any configured DurationEncoder at a higher level.
	AddDuration(dst *buffer.Buffer, key string, value time.Duration) *buffer.Buffer

	// AddFloat64 adds a float64 field with the given key.
	AddFloat64(dst *buffer.Buffer, key string, value float64) *buffer.Buffer

	// AddFloat32 adds a float32 field with the given key.
	AddFloat32(dst *buffer.Buffer, key string, value float32) *buffer.Buffer

	// AddInt adds an int field with the given key.
	//
	// Implementations MAY normalize int to a fixed width internally, but MUST
	// preserve the numeric value.
	AddInt(dst *buffer.Buffer, key string, value int) *buffer.Buffer

	// AddInt64 adds an int64 field with the given key.
	AddInt64(dst *buffer.Buffer, key string, value int64) *buffer.Buffer

	// AddInt32 adds an int32 field with the given key.
	AddInt32(dst *buffer.Buffer, key string, value int32) *buffer.Buffer

	// AddInt16 adds an int16 field with the given key.
	AddInt16(dst *buffer.Buffer, key string, value int16) *buffer.Buffer

	// AddInt8 adds an int8 field with the given key.
	AddInt8(dst *buffer.Buffer, key string, value int8) *buffer.Buffer

	// AddString adds a string field with the given key.
	AddString(dst *buffer.Buffer, key, value string) *buffer.Buffer

	// AddTime adds a time.Time field with the given key.
	//
	// Implementations SHOULD document how time zones and monotonic components
	// are handled in the encoded representation and MUST honor any configured
	// TimeEncoder at a higher level.
	AddTime(dst *buffer.Buffer, key string, value time.Time) *buffer.Buffer

	// AddUint adds an uint field with the given key.
	//
	// As with AddInt, implementations MAY normalize to a fixed width
	// internally, but MUST preserve the numeric value.
	AddUint(dst *buffer.Buffer, key string, value uint) *buffer.Buffer

	// AddUint64 adds an uint64 field with the given key.
	AddUint64(dst *buffer.Buffer, key string, value uint64) *buffer.Buffer

	// AddUint32 adds an uint32 field with the given key.
	AddUint32(dst *buffer.Buffer, key string, value uint32) *buffer.Buffer

	// AddUint16 adds an uint16 field with the given key.
	AddUint16(dst *buffer.Buffer, key string, value uint16) *buffer.Buffer

	// AddUint8 adds an uint8 field with the given key.
	AddUint8(dst *buffer.Buffer, key string, value uint8) *buffer.Buffer

	// AddUintptr adds an uintptr field with the given key.
	//
	// This is primarily useful for low-level debugging. Implementations
	// SHOULD NOT rely on any particular pointer format being stable across
	// processes or executions.
	AddUintptr(dst *buffer.Buffer, key string, value uintptr) *buffer.Buffer

	// AddReflected adds a field with the given key whose value is encoded
	// using reflection.
	//
	// Reflection-based encoding is intentionally slower and more
	// allocation-heavy than the strongly-typed Add* methods and SHOULD be
	// used sparingly, only when the value’s shape is not known at compile
	// time. Implementations MUST treat value as a single logical field and
	// MUST return a non-nil error if reflection-based encoding fails. On
	// error, no partially written representation of this field MAY be
	// appended.
	AddReflected(dst *buffer.Buffer, key string, value interface{}) (*buffer.Buffer, error)

	// OpenNamespace opens a nested namespace under the given key so that
	// subsequent fields are logically grouped beneath it.
	//
	// Implementations MAY encode this as a nested object (for example,
	// writing `"key": { ... }` in JSON) or via another backend-specific
	// mechanism. OpenNamespace MUST append all required syntax for beginning
	// the namespace to dst and update internal state so that subsequent Add*
	// calls are routed into that namespace until it is closed according to
	// the encoder’s rules (typically when the enclosing object is ended).
	OpenNamespace(dst *buffer.Buffer, key string) *buffer.Buffer
}

// ObjectMarshaler describes types that can encode themselves as the contents
// of a single logical object into an ObjectEncoder, writing bytes into a
// caller-provided buffer.
//
// A value that implements ObjectMarshaler is responsible only for emitting
// the *inside* of an object (that is, its fields); it MUST NOT write any
// outer framing for the object itself (such as '{' or '}'). Framing is the
// responsibility of the surrounding encoder (for example, an AddObject
// implementation on a parent ObjectEncoder).
//
// Implementations MUST treat enc as a single-use encoder for the duration of
// the call and MUST NOT retain it or use it after the method returns. All
// encoded bytes for this object MUST be appended to dst, and the updated
// Buffer MUST be returned. If the underlying buffer needs to grow, the
// returned Buffer MAY differ from the input dst; callers MUST always use the
// returned value as the new destination.
//
// On failure, implementations MUST return a non-nil error. In that case, they
// MUST NOT leave a partially encoded representation of this object in dst:
// bytes written before this object was started MAY remain, but the object
// itself MUST NOT be left in a half-written state.
type ObjectMarshaler interface {
	// MarshalLogObject encodes the receiver into the provided ObjectEncoder,
	// appending the object’s fields into dst and returning the updated buffer.
	//
	// The enc parameter represents the logical object encoder for this call.
	// The implementation MUST treat enc as a transient, single-use helper:
	// it is valid only for the duration of the call and MUST NOT be retained
	// or accessed afterwards. All field-level operations (for example,
	// AddString, AddInt, AddObject) MUST go through enc so that any internal
	// state (such as key separators or depth tracking) remains consistent.
	//
	// The dst parameter is the current output buffer. MarshalLogObject MUST
	// append all encoded bytes for this object to dst and return the resulting
	// Buffer. If the encoder needs to grow the underlying storage, it MAY
	// allocate a new Buffer and return that instead; callers MUST always use
	// the returned Buffer as the subsequent destination.
	//
	// On success, the returned error MUST be nil and the returned Buffer MUST
	// contain a complete, well-formed object payload for this value (without
	// outer braces, which are handled by the caller). On failure, the method
	// MUST return a non-nil error and MUST ensure that no partial
	// representation of this object is left in dst; any bytes written prior
	// to starting this object MAY remain, but the object itself MUST NOT be
	// half-written.
	MarshalLogObject(enc ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error)
}

// ObjectMarshalerFunc is a function adapter that turns a plain function into
// an ObjectMarshaler.
//
// Any function with the signature
//
//	func(enc ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error)
//
// MAY be wrapped as an ObjectMarshalerFunc and used wherever an
// ObjectMarshaler is required.
type ObjectMarshalerFunc func(enc ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error)

// MarshalLogObject calls the underlying function.
//
// When used via this adapter, the function MUST obey the same contract as
// ObjectMarshaler.MarshalLogObject: it MUST treat enc as a single-use
// encoder for the duration of the call, MUST append its output to dst and
// return the updated Buffer (which MAY be a different instance), and MUST
// return a non-nil error if encoding fails without leaving the object
// half-written.
func (f ObjectMarshalerFunc) MarshalLogObject(enc ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
	return f(enc, dst)
}
