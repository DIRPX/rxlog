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

// Type aliases for core field types.
//
// These aliases provide convenient access to field-related types from the root
// rxlog package without needing to import multiple sub-packages.
type (
	// Field represents a structured key-value pair for logging.
	//
	// Fields are the fundamental building blocks of structured logs. Each field
	// consists of a key (string), a type indicator, and the actual value stored
	// in one of several internal representations optimized for zero-allocation
	// encoding.
	//
	// Example:
	//
	//	f := rxlog.String("user", "alice")
	//	// f.Key = "user", f.Type = StringFieldType, f.String = "alice"
	Field = field.Field

	// FieldType identifies which member of the Field union is active.
	//
	// The type determines how encoders interpret and serialize the field's value.
	// For example, Int64FieldType indicates the value is in Field.Integer and
	// should be encoded as a signed 64-bit integer.
	FieldType = ftype.Type

	// ArrayMarshaler is the interface for custom array serialization.
	//
	// Implement this interface to provide zero-allocation encoding of custom
	// array-like types. The encoder will call MarshalLogArray to serialize the
	// array elements without using reflection.
	//
	// Example:
	//
	//	type UserIDs []int64
	//	func (ids UserIDs) MarshalLogArray(enc ArrayEncoder, dst *Buffer) (*Buffer, error) {
	//	    for _, id := range ids {
	//	        dst, _ = enc.AppendInt64(dst, id)
	//	    }
	//	    return dst, nil
	//	}
	ArrayMarshaler = base.ArrayMarshaler

	// ObjectMarshaler is the interface for custom object serialization.
	//
	// Implement this interface to provide zero-allocation encoding of custom
	// struct types. The encoder will call MarshalLogObject to serialize the
	// object's fields without using reflection.
	//
	// Example:
	//
	//	type User struct { ID int64; Name string }
	//	func (u User) MarshalLogObject(enc ObjectEncoder, dst *Buffer) (*Buffer, error) {
	//	    dst, _ = enc.AppendInt64(dst, "id", u.ID)
	//	    dst, _ = enc.AppendString(dst, "name", u.Name)
	//	    return dst, nil
	//	}
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

// Array constructs a Field that encodes an array-like value using a custom marshaler.
//
// The provided ArrayMarshaler will be invoked by the encoder to serialize the array
// elements without using reflection, enabling zero-allocation encoding of custom
// array types.
//
// Example:
//
//	type UserIDs []int64
//	func (ids UserIDs) MarshalLogArray(enc ArrayEncoder, dst *Buffer) (*Buffer, error) {
//	    for _, id := range ids {
//	        dst, _ = enc.AppendInt64(dst, id)
//	    }
//	    return dst, nil
//	}
//	field := rxlog.Array("user_ids", UserIDs{1, 2, 3})
//
// The marshaler MUST NOT be nil. Passing nil will result in undefined behavior.
func Array(key string, m base.ArrayMarshaler) Field {
	return Field{
		Key:       key,
		Type:      ftype.Array,
		Interface: m,
	}
}

// Object constructs a Field that encodes a structured object using a custom marshaler.
//
// The provided ObjectMarshaler will be invoked by the encoder to serialize nested
// fields without using reflection, enabling zero-allocation encoding of custom
// struct types.
//
// Example:
//
//	type User struct {
//	    ID   int64
//	    Name string
//	}
//	func (u User) MarshalLogObject(enc ObjectEncoder, dst *Buffer) (*Buffer, error) {
//	    dst, _ = enc.AppendInt64(dst, "id", u.ID)
//	    dst, _ = enc.AppendString(dst, "name", u.Name)
//	    return dst, nil
//	}
//	field := rxlog.Object("user", User{ID: 123, Name: "Alice"})
//
// The marshaler MUST NOT be nil. Passing nil will result in undefined behavior.
func Object(key string, m base.ObjectMarshaler) Field {
	return Field{
		Key:       key,
		Type:      ftype.Object,
		Interface: m,
	}
}

// Inline constructs a Field whose nested fields are merged into the parent object.
//
// Unlike Object, which creates a nested structure, Inline flattens the fields from
// the ObjectMarshaler directly into the enclosing log entry. This is useful for
// composing log contexts from multiple sources without adding nesting.
//
// Example:
//
//	type RequestContext struct {
//	    TraceID string
//	    UserID  int64
//	}
//	func (rc RequestContext) MarshalLogObject(enc ObjectEncoder, dst *Buffer) (*Buffer, error) {
//	    dst, _ = enc.AppendString(dst, "trace_id", rc.TraceID)
//	    dst, _ = enc.AppendInt64(dst, "user_id", rc.UserID)
//	    return dst, nil
//	}
//
//	// With Object: {"request": {"trace_id": "...", "user_id": 123}, "msg": "..."}
//	// With Inline: {"trace_id": "...", "user_id": 123, "msg": "..."}
//	field := rxlog.Inline(RequestContext{TraceID: "abc", UserID: 123})
//
// Note that Inline fields have no key, as their contents are merged directly.
func Inline(m base.ObjectMarshaler) Field {
	return Field{
		Type:      ftype.Inline,
		Interface: m,
	}
}

// Namespace constructs a Field that creates a nested scope for subsequent fields.
//
// This is primarily used by loggers to organize hierarchical field structures.
// Encoders typically implement this as opening a nested object in JSON or a new
// indentation level in text formats.
//
// Example:
//
//	logger.Info("request completed",
//	    rxlog.Namespace("http"),
//	    rxlog.Int("status", 200),
//	    rxlog.String("method", "GET"),
//	)
//	// JSON output: {"http": {"status": 200, "method": "GET"}, "msg": "request completed"}
//
// The behavior of Namespace depends on the encoder implementation. Most structured
// encoders will create a nested object under the given key.
func Namespace(key string) Field {
	return Field{
		Key:  key,
		Type: ftype.Namespace,
	}
}

// Skip constructs a no-op Field that encoders ignore completely.
//
// This is useful as a sentinel value when constructing fields conditionally,
// particularly in the *Ptr family of constructors which return Skip() for nil
// pointers. Encoders MUST skip these fields entirely without emitting any output.
//
// Example:
//
//	var name *string
//	field := rxlog.StringPtr("name", name)  // Returns Skip() when name is nil
//
//	// Conditional field:
//	func optionalField(include bool, key, val string) Field {
//	    if !include {
//	        return rxlog.Skip()
//	    }
//	    return rxlog.String(key, val)
//	}
func Skip() Field {
	return Field{
		Type: ftype.Skip,
	}
}

// Binary constructs a Field containing an opaque byte sequence.
//
// Use this for binary data like hashes, tokens, or protocol buffers that should
// not be interpreted as text. Encoders typically base64-encode binary fields or
// emit them as byte arrays depending on the output format.
//
// Example:
//
//	hash := sha256.Sum256(data)
//	field := rxlog.Binary("hash", hash[:])
//	// JSON output: {"hash": "base64encodedstring..."}
//
// IMPORTANT: The slice is stored by reference without copying. Callers MUST NOT
// modify the slice after constructing the field if it may be accessed concurrently
// by encoders.
func Binary(key string, b []byte) Field {
	return Field{
		Key:       key,
		Type:      ftype.Binary,
		Interface: b,
	}
}

// BinaryPtr constructs a Binary field from a byte slice pointer.
//
// Returns Skip() if the pointer is nil, allowing safe handling of optional
// binary data without explicit nil checks.
//
// Example:
//
//	var token *[]byte
//	if hasAuth {
//	    t := getAuthToken()
//	    token = &t
//	}
//	field := rxlog.BinaryPtr("auth_token", token)  // Skip if token is nil
func BinaryPtr(key string, value *[]byte) Field {
	if value == nil {
		return Skip()
	}
	return Binary(key, *value)
}

// ByteString constructs a Field containing UTF-8 encoded text as a byte slice.
//
// This is semantically equivalent to String but avoids allocating a string when
// the data is already available as bytes. Useful for zero-copy handling of text
// from I/O buffers or network packets.
//
// Example:
//
//	buf := make([]byte, n)
//	io.ReadFull(conn, buf)
//	field := rxlog.ByteString("payload", buf)
//	// Equivalent to: rxlog.String("payload", string(buf)) but without allocation
//
// The slice is stored by reference without copying. Encoders may or may not
// validate UTF-8 encoding depending on their implementation.
func ByteString(key string, b []byte) Field {
	return Field{
		Key:       key,
		Type:      ftype.ByteString,
		Interface: b,
	}
}

// ByteStringPtr constructs a ByteString field from a byte slice pointer.
//
// Returns Skip() if the pointer is nil.
func ByteStringPtr(key string, value *[]byte) Field {
	if value == nil {
		return Skip()
	}
	return ByteString(key, *value)
}

// String constructs a Field containing a text value.
//
// This is the most commonly used field constructor for logging arbitrary text data
// such as messages, identifiers, status codes, and descriptions.
//
// Example:
//
//	logger.Info("user logged in",
//	    rxlog.String("username", "alice"),
//	    rxlog.String("ip", "192.168.1.1"),
//	)
//	// JSON output: {"msg": "user logged in", "username": "alice", "ip": "192.168.1.1"}
func String(key, value string) Field {
	return Field{
		Key:    key,
		Type:   ftype.String,
		String: value,
	}
}

// StringPtr constructs a String field from a string pointer.
//
// Returns Skip() if the pointer is nil, enabling convenient optional string fields
// without explicit nil checks.
//
// Example:
//
//	var email *string
//	if user.HasEmail {
//	    e := user.Email
//	    email = &e
//	}
//	field := rxlog.StringPtr("email", email)  // Skip if email is nil
func StringPtr(key string, value *string) Field {
	if value == nil {
		return Skip()
	}
	return String(key, *value)
}

// Bool constructs a Field containing a boolean value.
//
// Commonly used for flags, status indicators, and conditional states.
//
// Example:
//
//	logger.Info("operation completed",
//	    rxlog.Bool("success", true),
//	    rxlog.Bool("cached", false),
//	)
//	// JSON output: {"msg": "operation completed", "success": true, "cached": false}
//
// The boolean value is stored as an integer (0 for false, 1 for true) in the
// field's internal representation for efficiency.
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

// BoolPtr constructs a Bool field from a boolean pointer.
//
// Returns Skip() if the pointer is nil.
func BoolPtr(key string, value *bool) Field {
	if value == nil {
		return Skip()
	}
	return Bool(key, *value)
}

// Int64 constructs a Field containing a signed 64-bit integer.
//
// Use this for IDs, counts, timestamps (as Unix time), and other numeric values
// that fit in a signed 64-bit range.
//
// Example:
//
//	logger.Info("user created",
//	    rxlog.Int64("user_id", 1234567890),
//	    rxlog.Int64("created_at", time.Now().Unix()),
//	)
//	// JSON output: {"msg": "user created", "user_id": 1234567890, "created_at": 1735689600}
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

// Int constructs a Field containing a signed integer.
//
// This is a convenience wrapper around Int64 for Go's platform-dependent int type.
// The value is normalized to int64 for consistent encoding regardless of platform
// (32-bit vs 64-bit).
//
// Example:
//
//	count := len(items)
//	field := rxlog.Int("count", count)
//	// Equivalent to: rxlog.Int64("count", int64(count))
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

// Duration constructs a Field containing a time duration.
//
// Durations are commonly used to log elapsed time, timeouts, and intervals.
// The duration is stored as nanoseconds internally, but encoders typically format
// it in a human-readable way (e.g., "2.5s", "100ms") depending on the encoder
// configuration.
//
// Example:
//
//	start := time.Now()
//	doWork()
//	elapsed := time.Since(start)
//	logger.Info("work completed", rxlog.Duration("elapsed", elapsed))
//	// JSON output: {"msg": "work completed", "elapsed": 2500000000}
//	// Or with duration encoder: {"msg": "work completed", "elapsed": "2.5s"}
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

// Time constructs a Field containing a timestamp.
//
// The timestamp is normalized to UTC and stored as nanoseconds since the Unix epoch.
// If the original time was not in UTC, the location is preserved for potential use
// by encoders that support timezone-aware formatting.
//
// Example:
//
//	now := time.Now()
//	logger.Info("event occurred", rxlog.Time("occurred_at", now))
//	// JSON output: {"msg": "event occurred", "occurred_at": "2025-01-02T15:04:05.123Z"}
//
// Most encoders will format timestamps as RFC3339/ISO8601 strings. Use TimeFull()
// if you need to preserve sub-second precision beyond nanoseconds.
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

// TimeInLocation constructs a Time field with the timestamp converted to a specific location.
//
// This is useful when you need to log times in a specific timezone (e.g., local time
// of an event or a user's timezone) rather than UTC. The location information is
// preserved for encoders that support timezone-aware formatting.
//
// Example:
//
//	loc, _ := time.LoadLocation("America/New_York")
//	eventTime := time.Now()
//	field := rxlog.TimeInLocation("event_time", eventTime, loc)
//	// Encoder may format as: "2025-01-02T10:04:05-05:00" (EST)
//
// If loc is nil, the time is used as-is without conversion.
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

// Error constructs a Field containing an error value.
//
// This is one of the most commonly used field constructors for logging failures
// and exceptional conditions. Encoders typically extract the error message via
// Error() and may include additional structured information like stack traces
// or wrapped error chains depending on the error type and encoder configuration.
//
// Example:
//
//	if err := doSomething(); err != nil {
//	    logger.Error("operation failed", rxlog.Error("error", err))
//	}
//	// JSON output: {"level": "error", "msg": "operation failed", "error": "connection timeout"}
//
// The error value may be nil; encoders will handle nil errors gracefully, typically
// by emitting "null" or omitting the field entirely depending on configuration.
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

// Any constructs a Field by automatically choosing the appropriate constructor
// based on the dynamic type of the value.
//
// This provides a convenient way to log values when the type is not known at compile
// time or when writing generic logging code. However, it introduces runtime type
// checking overhead and should be avoided in performance-critical code paths where
// the type is known.
//
// Supported types (in order of precedence):
//   - nil → Skip()
//   - string → String()
//   - []byte → ByteString()
//   - bool → Bool()
//   - int, int8, int16, int32, int64 → Int*()
//   - uint, uint8, uint16, uint32, uint64, uintptr → Uint*()
//   - float32, float64 → Float*()
//   - complex64, complex128 → Complex*()
//   - time.Time → TimeFull()
//   - time.Duration → Duration()
//   - error → Error()
//   - fmt.Stringer → Stringer()
//   - ArrayMarshaler → Array()
//   - ObjectMarshaler → Object()
//   - all other types → Reflect()
//
// Example:
//
//	var value interface{} = getUserInput()
//	logger.Info("received input", rxlog.Any("value", value))
//
// For best performance, prefer using type-specific constructors when the type
// is known at compile time.
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
