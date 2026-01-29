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
	"strconv"
	"time"
)

// Size defines the default initial capacity (in bytes) for newly allocated
// Buffer instances.
//
// Implementations that construct Buffer values (for example, a Pool factory)
// MAY use this constant to preallocate the underlying byte slice so that
// small writes do not immediately trigger reallocations. Callers MUST NOT
// assume that a Buffer will always have exactly this capacity: it is a
// starting point and MAY grow as data is appended.
const Size = 1024

// Buffer is a concrete, growable byte buffer used as the backing storage for
// encoded log records or other serialized payloads.
//
// Internally, Buffer maintains a single byte slice (data) that grows as
// callers write to it. The buffer is intended to be reused via an associated
// Pool to reduce allocations and GC pressure.
//
// Concurrency:
//   - A Buffer instance is NOT safe for concurrent use by multiple goroutines.
//     All access (reads and writes) MUST be externally synchronized if a
//     Buffer is shared.
//   - After a Buffer has been returned to its Pool via Free, it MUST NOT be
//     used again by the caller; doing so constitutes a use-after-free bug and
//     MAY lead to data races or corrupted output.
type Buffer struct {
	// data holds the current contents of the buffer. The slice length reflects
	// the amount of valid data, while the capacity reflects the allocated
	// storage that can be reused for future writes.
	data []byte

	// pool is the originating Pool responsible for recycling this Buffer.
	// Free uses this reference to return the Buffer for reuse. It MUST be
	// treated as opaque by callers.
	pool Pool
}

// Buffer implements Interface at compile time, ensuring that *Buffer satisfies
// the expected buffer contract used by the rest of the logging pipeline.
var _ Interface = (*Buffer)(nil)

// NewWithPool creates a new Buffer with the specified initial capacity and
// pool reference.
//
// This constructor is intended for use by internal pooling implementations
// that need to create buffers with custom pool bindings. Most callers SHOULD
// use a pool's Get() method rather than calling NewWithPool directly.
//
// The capacity parameter specifies the initial capacity of the underlying
// byte slice. The actual capacity MAY be larger due to allocator rounding.
//
// The pool parameter MUST NOT be nil and MUST remain valid for the lifetime
// of the buffer. When the buffer is freed via Free(), it will be returned
// to this pool.
//
// Example (internal pool implementation):
//
//	func (p *MyPool) newBuffer() *buffer.Buffer {
//	    return buffer.NewWithPool(p, 512)
//	}
func NewWithPool(pool Pool, capacity int) *Buffer {
	return &Buffer{
		data: make([]byte, 0, capacity),
		pool: pool,
	}
}

// Write appends the contents of p to the buffer, growing the underlying slice
// as needed, and reports the number of bytes written.
//
// This method implements io.Writer. On success, it always returns len(p) and a
// nil error. It MUST NOT return a short write without an accompanying error.
//
// Write MAY cause the underlying slice to be reallocated if the current
// capacity is insufficient. Any slices obtained from Bytes prior to such a
// reallocation MUST be considered invalid once Write returns.
func (b *Buffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

// WriteByte appends a single byte v to the buffer and reports success.
//
// This method is intentionally compatible with the signatures of
// bytes.Buffer and bufio.Writer:
//
//   - It ALWAYS returns a nil error on success.
//   - It NEVER performs a short write: the byte is either appended
//     (barring a panic) or the call does not return normally.
//
// Callers MAY use WriteByte wherever an io.ByteWriter is expected. Internally,
// this is a thin wrapper around AppendByte and exists primarily for API
// compatibility and ergonomics.
func (b *Buffer) WriteByte(v byte) error {
	b.AppendByte(v)
	return nil
}

// WriteString appends the contents of s to the buffer and reports the number
// of bytes written.
//
// This method is intentionally compatible with the signatures of
// bytes.Buffer and bufio.Writer:
//
//   - The returned error is ALWAYS nil; short writes do not occur.
//   - The returned count is ALWAYS equal to len(s), assuming no panic.
//
// Callers MAY use WriteString wherever an io.StringWriter is expected.
// Internally, this is a thin wrapper around AppendString.
func (b *Buffer) WriteString(s string) (int, error) {
	b.AppendString(s)
	return len(s), nil
}

// AppendTime appends a textual representation of t to the buffer using
// the provided layout.
//
// The layout MUST follow Go's standard reference time pattern conventions
// (for example, "2006-01-02 15:04:05.000Z07:00"). Formatting is performed by
// time.Time.AppendFormat, which appends directly into the underlying byte
// slice and MAY grow the slice if necessary.
//
// This method does not return a value; callers SHOULD assume that b.data has
// been extended in-place (subject to possible reallocation of its backing
// array).
func (b *Buffer) AppendTime(t time.Time, layout string) {
	b.data = t.AppendFormat(b.data, layout)
}

// AppendDuration appends d as a base-10 integer representing the number of
// nanoseconds contained in the duration.
//
// Higher-level duration encoders (for example, those that format values as
// "1s", "10ms", etc.) SHOULD convert d into the desired unit and then call
// the appropriate numeric append helper (such as AppendInt64 or AppendFloat64)
// instead of relying on this low-level representation directly.
func (b *Buffer) AppendDuration(d time.Duration) {
	b.data = strconv.AppendInt(b.data, int64(d), 10)
}

// AppendByte appends a single byte v to the buffer.
//
// This is a low-level helper that simply extends b.data by one byte. It
// SHOULD be preferred over constructing a one-byte slice when writing
// single characters or delimiters.
func (b *Buffer) AppendByte(v byte) {
	b.data = append(b.data, v)
}

// AppendBytes appends the contents of v to the buffer as-is.
//
// No encoding, escaping, or interpretation is performed. Callers SHOULD only
// use this method when v already contains the exact byte representation that
// SHOULD appear in the output.
func (b *Buffer) AppendBytes(v []byte) {
	b.data = append(b.data, v...)
}

// AppendString appends the bytes of s to the buffer.
//
// The string’s bytes are copied into the underlying slice. Callers that need
// to write large strings repeatedly SHOULD consider buffering or pooling at
// a higher level to avoid unnecessary allocations of the string itself.
func (b *Buffer) AppendString(s string) {
	b.data = append(b.data, s...)
}

// AppendBool appends the textual representation of v ("true" or "false")
// in lower case.
//
// The format matches strconv.AppendBool and is stable for downstream
// consumers that expect canonical JSON-like boolean literals.
func (b *Buffer) AppendBool(v bool) {
	b.data = strconv.AppendBool(b.data, v)
}

// AppendComplex128 appends v in the textual form "<real><+|-><imag>i".
//
// Both real and imaginary parts are formatted using 'g' with full 64-bit
// precision via strconv.AppendFloat. Example outputs include:
//   - 1.23+4.56i
//   - -1-2i
//
// The exact formatting is stable with respect to strconv’s documented
// behavior for format 'g' with prec == -1 and bitSize == 64.
func (b *Buffer) AppendComplex128(v complex128) {
	r := real(v)
	im := imag(v)

	// real part
	b.data = strconv.AppendFloat(b.data, r, 'g', -1, 64)

	// sign for imaginary part
	if im >= 0 {
		b.data = append(b.data, '+')
	}

	// imaginary magnitude
	b.data = strconv.AppendFloat(b.data, im, 'g', -1, 64)
	b.data = append(b.data, 'i')
}

// AppendComplex64 appends v in the textual form "<real><+|-><imag>i".
//
// Both real and imaginary parts are formatted using 'g' with 32-bit
// precision. This mirrors AppendComplex128 but uses bitSize == 32.
func (b *Buffer) AppendComplex64(v complex64) {
	r := float64(real(v))
	im := float64(imag(v))

	b.data = strconv.AppendFloat(b.data, r, 'g', -1, 32)

	if im >= 0 {
		b.data = append(b.data, '+')
	}

	b.data = strconv.AppendFloat(b.data, im, 'g', -1, 32)
	b.data = append(b.data, 'i')
}

// AppendFloat64 appends v using 'g' format with full 64-bit precision.
//
// This matches common logging conventions (for example, zerolog-style):
// values are rendered compactly when possible while preserving precision.
// The exact formatting is determined by strconv.AppendFloat with
// format 'g', prec -1, and bitSize 64.
func (b *Buffer) AppendFloat64(v float64) {
	b.data = strconv.AppendFloat(b.data, v, 'g', -1, 64)
}

// AppendFloat32 appends v using 'g' format with 32-bit precision.
//
// The value is first converted to float64 for use with strconv.AppendFloat,
// but the bitSize argument is set to 32, so the textual representation
// reflects float32 precision.
func (b *Buffer) AppendFloat32(v float32) {
	b.data = strconv.AppendFloat(b.data, float64(v), 'g', -1, 32)
}

// AppendInt appends v as a base-10 signed integer.
//
// The representation matches strconv.AppendInt with base 10.
func (b *Buffer) AppendInt(v int) {
	b.data = strconv.AppendInt(b.data, int64(v), 10)
}

// AppendInt64 appends v as a base-10 signed integer.
//
// This is the canonical helper for 64-bit integer logging when no additional
// formatting is required.
func (b *Buffer) AppendInt64(v int64) {
	b.data = strconv.AppendInt(b.data, v, 10)
}

// AppendInt32 appends v as a base-10 signed integer.
//
// The value is widened to int64 for formatting; the numeric value is
// preserved exactly.
func (b *Buffer) AppendInt32(v int32) {
	b.data = strconv.AppendInt(b.data, int64(v), 10)
}

// AppendInt16 appends v as a base-10 signed integer.
//
// As with other integer helpers, the value is converted to int64 for
// formatting while preserving its numeric value.
func (b *Buffer) AppendInt16(v int16) {
	b.data = strconv.AppendInt(b.data, int64(v), 10)
}

// AppendInt8 appends v as a base-10 signed integer.
func (b *Buffer) AppendInt8(v int8) {
	b.data = strconv.AppendInt(b.data, int64(v), 10)
}

// AppendUint appends v as a base-10 unsigned integer.
//
// The representation matches strconv.AppendUint with base 10.
func (b *Buffer) AppendUint(v uint) {
	b.data = strconv.AppendUint(b.data, uint64(v), 10)
}

// AppendUint64 appends v as a base-10 unsigned integer.
func (b *Buffer) AppendUint64(v uint64) {
	b.data = strconv.AppendUint(b.data, v, 10)
}

// AppendUint32 appends v as a base-10 unsigned integer.
func (b *Buffer) AppendUint32(v uint32) {
	b.data = strconv.AppendUint(b.data, uint64(v), 10)
}

// AppendUint16 appends v as a base-10 unsigned integer.
func (b *Buffer) AppendUint16(v uint16) {
	b.data = strconv.AppendUint(b.data, uint64(v), 10)
}

// AppendUint8 appends v as a base-10 unsigned integer.
func (b *Buffer) AppendUint8(v uint8) {
	b.data = strconv.AppendUint(b.data, uint64(v), 10)
}

// AppendUintptr appends v as a hexadecimal pointer-like representation.
//
// The value is encoded as "0x" followed by lowercase hexadecimal digits
// (base 16). This is primarily intended for low-level debugging or
// introspection of pointer-sized values and SHOULD NOT be relied on as a
// stable identifier across processes or executions.
func (b *Buffer) AppendUintptr(v uintptr) {
	b.data = append(b.data, '0', 'x')
	b.data = strconv.AppendUint(b.data, uint64(v), 16)
}

// Bytes returns the underlying byte slice containing the current contents of
// the buffer.
//
// The returned slice is a view into the internal storage and MUST be treated
// as read-only by callers. Mutating it in place MAY corrupt the buffer state
// or violate invariants expected by pooling logic.
//
// The slice remains valid until the buffer grows (via further writes) or the
// Buffer is returned to its Pool via Free.
func (b *Buffer) Bytes() []byte {
	return b.data
}

// Len reports the current length of the buffer, in bytes.
//
// This is equivalent to len(b.Bytes()) but does not allocate and is the
// preferred way to query the logical size of the buffered data.
func (b *Buffer) Len() int {
	return len(b.data)
}

// Cap reports the current capacity of the underlying byte slice, in bytes.
//
// This is primarily useful for diagnostics and optimization (for example,
// understanding growth behavior). Callers MUST NOT rely on Cap remaining
// stable across writes or pool round-trips: it MAY change as the buffer
// grows or is reused.
func (b *Buffer) Cap() int {
	return cap(b.data)
}

// Reset clears the logical contents of the buffer while preserving its
// allocated capacity.
//
// After Reset, Len will be zero, but Cap will remain unchanged. This makes
// Reset suitable for reusing buffers without incurring additional allocations.
//
// Reset does NOT return the Buffer to its Pool; callers that are done with
// the Buffer MUST still call Free (or otherwise release it) to allow the
// underlying storage to be recycled.
func (b *Buffer) Reset() {
	b.data = b.data[:0]
}

// Free returns the Buffer to its originating Pool so that it can be reused.
//
// After Free is called, the caller MUST treat the Buffer as invalid and MUST
// NOT read from, write to, or otherwise use it. Doing so constitutes
// undefined behavior from the caller’s perspective and MAY lead to data races
// or memory corruption in pool users.
//
// Each Buffer instance MUST be freed at most once. Calling Free more than
// once or mixing direct Pool.Put calls with Free on the same instance is not
// supported and MAY break Pool invariants.
func (b *Buffer) Free() {
	b.pool.Put(b)
}
