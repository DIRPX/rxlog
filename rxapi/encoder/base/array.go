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

// ArrayEncoder is a strongly-typed, encoding-agnostic interface for appending
// array-like values to the logging context using a caller-owned buffer.
//
// An ArrayEncoder represents a single, mutable array under construction. All
// encoded bytes are appended to the dst buffer passed into each method.
// Implementations MUST interpret each Append* call as appending exactly one
// logical element to the current array and MUST manage element separators
// (for example, commas in JSON) internally.
//
// ArrayEncoder values MUST NOT be used concurrently from multiple goroutines
// unless explicitly documented otherwise; callers SHOULD treat each instance
// as owned by a single goroutine for the duration of its use.
//
// ArrayEncoder extends PrimitiveArrayEncoder, so it MUST support appending
// all built-in scalar types in addition to the higher-level types defined
// below.
type ArrayEncoder interface {
	PrimitiveArrayEncoder

	// AppendDuration appends a single time.Duration value to the encoded
	// array.
	//
	// Implementations SHOULD represent the duration in a consistent format
	// (for example, a numeric count in a specific unit or a textual
	// representation) and SHOULD document the chosen format. This call MUST
	// append exactly one logical element to dst and return the updated buffer.
	AppendDuration(dst *buffer.Buffer, value time.Duration) *buffer.Buffer

	// AppendTime appends a single time.Time value to the encoded array.
	//
	// Implementations SHOULD document how time zones and monotonic clock
	// components are handled. This call MUST append exactly one logical
	// element to dst and MUST NOT open or close additional array/object
	// scopes beyond what is required to encode this element.
	AppendTime(dst *buffer.Buffer, value time.Time) *buffer.Buffer

	// AppendArray appends a nested array by delegating to the provided
	// ArrayMarshaler.
	//
	// Implementations MUST treat marshaler as a single nested array element.
	// They MUST write any necessary element separator for the parent array
	// into dst, encode the nested array (for example, by writing '[' ... ']'
	// and delegating element encoding to marshaler), and append the complete
	// nested array representation as one logical element.
	//
	// If encoding fails, AppendArray MUST return a non-nil error and MUST NOT
	// leave a partially written nested array element in dst (bytes written
	// before this element may remain).
	AppendArray(dst *buffer.Buffer, marshaler ArrayMarshaler) (*buffer.Buffer, error)

	// AppendObject appends a nested object by delegating to the provided
	// ObjectMarshaler.
	//
	// Implementations MUST treat marshaler as a single nested object element.
	// They MUST write any necessary element separator for the parent array
	// into dst, encode the nested object (for example, by writing '{' ... '}'
	// and delegating field encoding to marshaler), and append the complete
	// nested object representation as one logical element.
	//
	// If encoding fails, AppendObject MUST return a non-nil error and MUST NOT
	// leave a partially written nested object element in dst (bytes written
	// before this element may remain).
	AppendObject(dst *buffer.Buffer, marshaler ObjectMarshaler) (*buffer.Buffer, error)

	// AppendReflected appends an element by using reflection to serialize
	// an arbitrary Go value.
	//
	// This operation is intentionally slower and more allocation-heavy than
	// the strongly-typed Append* methods and SHOULD be used sparingly, only
	// when the value's shape is not known at compile time. Implementations
	// MUST treat value as a single logical element, MUST append its encoded
	// representation to dst, and MUST return a non-nil error if
	// reflection-based encoding fails. On error, implementations MUST NOT
	// leave a partially encoded element in dst.
	AppendReflected(dst *buffer.Buffer, value interface{}) (*buffer.Buffer, error)
}

// PrimitiveArrayEncoder is the subset of an array-like encoder interface that
// deals only with Go's built-in scalar types, writing directly into a
// caller-owned buffer.
//
// This interface is primarily intended as a target for helpers that encode
// primitive values (for example, duration and time encoders) without risking
// infinite recursion by accidentally calling higher-level ArrayEncoder
// methods that might delegate back into those helpers.
//
// Implementations MUST interpret each Append* call as appending a single
// element of the corresponding type to the encoded array. Methods in this
// interface MUST NOT perform any additional structural operations beyond
// those explicitly defined here (such as starting or ending nested objects).
type PrimitiveArrayEncoder interface {
	// BeginArray starts encoding an array into dst.
	//
	// Implementations MUST append the opening delimiter for an array
	// (for example, '[' in JSON) and update any required internal state
	// (such as element count or separator tracking). The updated buffer
	// MUST be returned.
	BeginArray(dst *buffer.Buffer) *buffer.Buffer

	// EndArray finishes encoding an array into dst.
	//
	// Implementations MUST append the closing delimiter for an array
	// (for example, ']' in JSON) and return the updated buffer. Any internal
	// state that tracks the current array scope MAY be reset once the array
	// is closed.
	EndArray(dst *buffer.Buffer) *buffer.Buffer

	// AppendBool appends a boolean value as a single array element.
	AppendBool(dst *buffer.Buffer, value bool) *buffer.Buffer

	// AppendByteString appends a UTF-8 byte slice as a single array element.
	//
	// Implementations MAY validate or normalize UTF-8 but are not required to.
	// Callers SHOULD prefer this method over arbitrary binary for textual
	// byte slices.
	AppendByteString(dst *buffer.Buffer, value []byte) *buffer.Buffer

	// AppendComplex128 appends a complex128 value as a single array element.
	AppendComplex128(dst *buffer.Buffer, value complex128) *buffer.Buffer

	// AppendComplex64 appends a complex64 value as a single array element.
	AppendComplex64(dst *buffer.Buffer, value complex64) *buffer.Buffer

	// AppendFloat64 appends a float64 value as a single array element.
	AppendFloat64(dst *buffer.Buffer, value float64) *buffer.Buffer

	// AppendFloat32 appends a float32 value as a single array element.
	AppendFloat32(dst *buffer.Buffer, value float32) *buffer.Buffer

	// AppendInt appends an int value as a single array element.
	//
	// Implementations MAY normalize int to int64 internally, as long as the
	// numeric value is preserved.
	AppendInt(dst *buffer.Buffer, value int) *buffer.Buffer

	// AppendInt64 appends an int64 value as a single array element.
	AppendInt64(dst *buffer.Buffer, value int64) *buffer.Buffer

	// AppendInt32 appends an int32 value as a single array element.
	AppendInt32(dst *buffer.Buffer, value int32) *buffer.Buffer

	// AppendInt16 appends an int16 value as a single array element.
	AppendInt16(dst *buffer.Buffer, value int16) *buffer.Buffer

	// AppendInt8 appends an int8 value as a single array element.
	AppendInt8(dst *buffer.Buffer, value int8) *buffer.Buffer

	// AppendString appends a string value as a single array element.
	AppendString(dst *buffer.Buffer, value string) *buffer.Buffer

	// AppendUint appends an uint value as a single array element.
	//
	// Implementations MAY normalize uint to uint64 internally, as long as the
	// numeric value is preserved.
	AppendUint(dst *buffer.Buffer, value uint) *buffer.Buffer

	// AppendUint64 appends a uint64 value as a single array element.
	AppendUint64(dst *buffer.Buffer, value uint64) *buffer.Buffer

	// AppendUint32 appends a uint32 value as a single array element.
	AppendUint32(dst *buffer.Buffer, value uint32) *buffer.Buffer

	// AppendUint16 appends a uint16 value as a single array element.
	AppendUint16(dst *buffer.Buffer, value uint16) *buffer.Buffer

	// AppendUint8 appends a uint8 value as a single array element.
	AppendUint8(dst *buffer.Buffer, value uint8) *buffer.Buffer

	// AppendUintptr appends an uintptr value as a single array element.
	//
	// This is mainly useful for low-level introspection or debugging.
	// Implementations SHOULD NOT rely on pointer representations being
	// stable across processes or runs.
	AppendUintptr(dst *buffer.Buffer, value uintptr) *buffer.Buffer
}

// ArrayMarshaler describes types that can encode themselves as the contents
// of a single logical array into an ArrayEncoder, writing bytes into a
// caller-provided buffer.
//
// A value that implements ArrayMarshaler is responsible only for emitting the
// *elements* of an array; it MUST NOT write any outer array framing (such as
// '[' or ']'). Framing is the responsibility of the surrounding encoder
// (for example, an AppendArray implementation on a parent ArrayEncoder).
//
// Implementations MAY use ArrayMarshaler to selectively omit sensitive or
// irrelevant information (for example, passwords or oversized blobs) instead
// of relying on reflection-based encoding. This provides tighter control over
// both performance characteristics and data exposure.
//
// ArrayMarshaler is used when a value is explicitly passed to APIs that treat
// it as an array payload (for example, helpers that append arrays or accept
// ArrayMarshaler directly). It is NOT used when generic reflection-based
// encoding is applied to arbitrary values.
type ArrayMarshaler interface {
	// MarshalLogArray encodes the receiver into the provided ArrayEncoder,
	// appending zero or more elements into dst and returning the updated
	// buffer.
	//
	// The enc parameter represents the array encoder for this call and MUST be
	// treated as a transient, single-use helper: it is valid only for the
	// duration of the method and MUST NOT be retained or used afterwards.
	// Implementations MUST emit all array elements through enc's Append*
	// methods so that internal state (such as separators and nesting) remains
	// consistent.
	//
	// The dst parameter is the current output buffer. MarshalLogArray MUST
	// append all encoded elements for this array to dst and return the
	// resulting Buffer. If the encoder needs to grow the underlying storage,
	// it MAY allocate a new Buffer and return that instead; callers MUST
	// always use the returned Buffer as the subsequent destination.
	//
	// On success, the returned error MUST be nil and the returned Buffer MUST
	// contain a complete sequence of elements for this array (without outer
	// '[' or ']' framing). On failure, the method MUST return a non-nil error.
	// In that case, dst MAY contain a partially written array (some elements
	// MAY already have been appended). Callers that require strict atomicity
	// MUST enforce it at a higher level, for example by encoding into a
	// temporary buffer and discarding it on error.
	//
	// MarshalLogArray MUST NOT retain references to enc or dst beyond the
	// duration of the call and MUST assume that enc is not safe for concurrent
	// use.
	MarshalLogArray(enc ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error)
}

// ArrayMarshalerFunc is a function adapter that turns a plain function into
// an ArrayMarshaler.
//
// Any function with the signature
//
//	func(enc ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error)
//
// MAY be wrapped as an ArrayMarshalerFunc and used wherever an
// ArrayMarshaler is required.
type ArrayMarshalerFunc func(enc ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error)

// MarshalLogArray calls the underlying function.
//
// When used via this adapter, the function MUST obey the same contract as
// ArrayMarshaler.MarshalLogArray: it MUST treat enc as a single-use encoder
// for the duration of the call, MUST append its output to dst and return the
// updated Buffer (which MAY be a different instance), and MUST return a
// non-nil error if encoding fails.
func (f ArrayMarshalerFunc) MarshalLogArray(enc ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
	return f(enc, dst)
}
