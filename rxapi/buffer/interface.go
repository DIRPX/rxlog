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

package buffer

import (
	"io"
	"time"
)

// Interface represents a reusable, pooled buffer that stores an encoded log
// entry or any other serialized payload. Implementations SHOULD be backed by
// a contiguous byte slice and SHOULD return the buffer to an internal pool
// when it is no longer needed.
//
// An Interface is NOT safe for concurrent use by multiple goroutines unless the
// implementation explicitly documents otherwise. Callers MUST treat each
// Interface instance as owned by a single goroutine at a time, and MUST respect
// the lifetime rules of Bytes, Reset, and Free described below.
type Interface interface {
	io.Writer

	// WriteByte appends a single byte v to the buffer and reports success.
	//
	// This method is compatible with io.ByteWriter and ALWAYS returns a nil
	// error on success. The byte is either appended (barring a panic) or the
	// call does not return normally.
	//
	// Implementations SHOULD use this method when writing single characters
	// or delimiters, as it is more efficient than constructing a one-byte
	// slice and calling Write.
	WriteByte(v byte) error

	// WriteString appends the contents of s to the buffer and reports the
	// number of bytes written.
	//
	// This method is compatible with io.StringWriter. The returned error is
	// ALWAYS nil; short writes do not occur. The returned count is ALWAYS
	// equal to len(s), assuming no panic.
	//
	// Implementations SHOULD prefer this method over Write when the input is
	// already a string, as it MAY avoid copying the string to a temporary
	// byte slice.
	WriteString(s string) (int, error)

	// AppendTime appends a textual representation of t to the buffer using
	// the provided layout.
	//
	// The layout MUST follow Go's standard reference time pattern conventions
	// (for example, "2006-01-02 15:04:05.000Z07:00"). Formatting is performed
	// by time.Time.AppendFormat, which appends directly into the underlying
	// byte slice and MAY grow the slice if necessary.
	//
	// This method does not return a value; callers SHOULD assume that the
	// buffer has been extended in-place (subject to possible reallocation of
	// its backing array).
	AppendTime(t time.Time, layout string)

	// AppendDuration appends d as a base-10 integer representing the number
	// of nanoseconds contained in the duration.
	//
	// Higher-level duration encoders (for example, those that format values
	// as "1s", "10ms", etc.) SHOULD convert d into the desired unit and then
	// call the appropriate numeric append helper (such as AppendInt64 or
	// AppendFloat64) instead of relying on this low-level representation
	// directly.
	AppendDuration(d time.Duration)

	// AppendByte appends a single byte v to the buffer.
	//
	// This is a low-level helper that simply extends the buffer by one byte.
	// It SHOULD be preferred over constructing a one-byte slice when writing
	// single characters or delimiters.
	AppendByte(v byte)

	// AppendBytes appends the contents of v to the buffer as-is.
	//
	// No encoding, escaping, or interpretation is performed. Callers SHOULD
	// only use this method when v already contains the exact byte
	// representation that SHOULD appear in the output.
	AppendBytes(v []byte)

	// AppendString appends the bytes of s to the buffer.
	//
	// The string's bytes are copied into the underlying slice. Callers that
	// need to write large strings repeatedly SHOULD consider buffering or
	// pooling at a higher level to avoid unnecessary allocations of the
	// string itself.
	AppendString(s string)

	// AppendBool appends the textual representation of v ("true" or "false")
	// in lower case.
	//
	// The format matches strconv.AppendBool and is stable for downstream
	// consumers that expect canonical JSON-like boolean literals.
	AppendBool(v bool)

	// AppendComplex128 appends v in the textual form "<real><+|-><imag>i".
	//
	// Both real and imaginary parts are formatted using 'g' with full 64-bit
	// precision via strconv.AppendFloat. Example outputs include:
	//   - 1.23+4.56i
	//   - -1-2i
	//
	// The exact formatting is stable with respect to strconv's documented
	// behavior for format 'g' with prec == -1 and bitSize == 64.
	AppendComplex128(v complex128)

	// AppendComplex64 appends v in the textual form "<real><+|-><imag>i".
	//
	// Both real and imaginary parts are formatted using 'g' with 32-bit
	// precision. This mirrors AppendComplex128 but uses bitSize == 32.
	AppendComplex64(v complex64)

	// AppendFloat64 appends v using 'g' format with full 64-bit precision.
	//
	// This matches common logging conventions (for example, zerolog-style):
	// values are rendered compactly when possible while preserving precision.
	// The exact formatting is determined by strconv.AppendFloat with format
	// 'g', prec -1, and bitSize 64.
	AppendFloat64(v float64)

	// AppendFloat32 appends v using 'g' format with 32-bit precision.
	//
	// The value is first converted to float64 for use with strconv.AppendFloat,
	// but the bitSize argument is set to 32, so the textual representation
	// reflects float32 precision.
	AppendFloat32(v float32)

	// AppendInt appends v as a base-10 signed integer.
	//
	// The representation matches strconv.AppendInt with base 10.
	AppendInt(v int)

	// AppendInt64 appends v as a base-10 signed integer.
	//
	// This is the canonical helper for 64-bit integer logging when no
	// additional formatting is required.
	AppendInt64(v int64)

	// AppendInt32 appends v as a base-10 signed integer.
	//
	// The value is widened to int64 for formatting; the numeric value is
	// preserved exactly.
	AppendInt32(v int32)

	// AppendInt16 appends v as a base-10 signed integer.
	//
	// As with other integer helpers, the value is converted to int64 for
	// formatting while preserving its numeric value.
	AppendInt16(v int16)

	// AppendInt8 appends v as a base-10 signed integer.
	AppendInt8(v int8)

	// AppendUint appends v as a base-10 unsigned integer.
	//
	// The representation matches strconv.AppendUint with base 10.
	AppendUint(v uint)

	// AppendUint64 appends v as a base-10 unsigned integer.
	AppendUint64(v uint64)

	// AppendUint32 appends v as a base-10 unsigned integer.
	AppendUint32(v uint32)

	// AppendUint16 appends v as a base-10 unsigned integer.
	AppendUint16(v uint16)

	// AppendUint8 appends v as a base-10 unsigned integer.
	AppendUint8(v uint8)

	// AppendUintptr appends v as a hexadecimal pointer-like representation.
	//
	// The value is encoded as "0x" followed by lowercase hexadecimal digits
	// (base 16). This is primarily intended for low-level debugging or
	// introspection of pointer-sized values and SHOULD NOT be relied on as a
	// stable identifier across processes or executions.
	AppendUintptr(v uintptr)

	// Bytes returns a view of the underlying encoded data as a byte slice.
	//
	// The returned slice is valid only until Free is called; after Free, its
	// contents, length, and capacity MUST be considered invalid and MUST NOT
	// be accessed. Callers MUST NOT retain the returned slice beyond the
	// lifetime of the Interface.
	//
	// The caller MUST treat the returned slice as immutable. Modifying it in
	// place MAY corrupt the buffer pool or violate assumptions made by other
	// components that reuse the same underlying storage.
	Bytes() []byte

	// Len reports the current length of the encoded data, in bytes.
	//
	// Len MUST be equivalent to len(Bytes()) at the time of the call, but
	// MUST NOT allocate or otherwise incur additional cost beyond reading the
	// current logical length. Implementations MAY compute Len more cheaply
	// than constructing or returning a fresh slice.
	Len() int

	// Cap reports the current capacity of the underlying buffer, in bytes.
	//
	// Cap SHOULD reflect the total number of bytes that can be written without
	// requiring a reallocation. Callers MAY use Cap to reason about potential
	// growth or to make decisions about preallocation, but MUST NOT rely on
	// any particular growth strategy or upper bound.
	Cap() int

	// Reset clears the logical contents of the buffer while preserving its
	// underlying storage so that it can be efficiently reused for building a
	// new payload.
	//
	// After Reset returns, Len MUST be 0. Implementations SHOULD preserve the
	// existing capacity (i.e., Cap SHOULD remain unchanged), but MAY shrink or
	// grow the underlying storage in exceptional cases (for example, to enforce
	// maximum size limits).
	//
	// Reset MUST NOT return the buffer to any pool and MUST NOT invalidate
	// existing ownership; for releasing a buffer, callers MUST use Free.
	Reset()

	// Free releases the buffer back to its originating pool so that its
	// storage can be reused.
	//
	// After Free returns, the Interface MUST NOT be used again: all methods,
	// including Bytes, Len, Cap, Reset, and Write, are invalid to call, and
	// any previously obtained slices from Bytes become unsafe to access.
	//
	// Each Interface instance MUST be freed at most once. Double-free or use
	// after Free MAY corrupt the pool, leak memory, or cause undefined
	// behavior in callers. Implementations SHOULD document whether Free may
	// perform additional checks (such as debug assertions) in development
	// builds.
	Free()
}
