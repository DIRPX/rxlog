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
	"dirpx.dev/rxlog/rxapi/encoder/base"
)

// ArrayMarshaler is the interface for custom array serialization.
//
// Implement this interface to provide zero-allocation encoding of custom
// array-like types. The encoder will call MarshalLogArray to serialize the
// array elements without using reflection.
//
// Example:
//
//	type UserIDs []int64
//
//	func (ids UserIDs) MarshalLogArray(enc base.ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
//	    for _, id := range ids {
//	        dst = enc.AppendInt64(dst, id)
//	    }
//	    return dst, nil
//	}
//
//	// Usage:
//	field := rxlog.Array("user_ids", UserIDs{1, 2, 3})
//	// JSON output: {"user_ids": [1, 2, 3]}
//
// The MarshalLogArray method receives an ArrayEncoder and a Buffer. It should
// append array elements to the buffer using the encoder's methods and return
// the updated buffer along with any error encountered.
//
// Important: You MUST NOT write array delimiters ('[', ']') yourself. The
// encoder handles framing automatically.
type ArrayMarshaler = base.ArrayMarshaler

// ArrayMarshalerFunc is a function adapter that allows using a plain function
// as an ArrayMarshaler.
//
// This is useful for creating ad-hoc array marshalers without defining a new type.
//
// Example:
//
//	// Define a function that encodes an array
//	encodeNumbers := rxlog.ArrayMarshalerFunc(func(enc base.ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
//	    for i := 1; i <= 5; i++ {
//	        dst = enc.AppendInt(dst, i)
//	    }
//	    return dst, nil
//	})
//
//	// Use it as a field
//	field := rxlog.Array("numbers", encodeNumbers)
//	// JSON output: {"numbers": [1, 2, 3, 4, 5]}
//
// This pattern is particularly useful for encoding dynamic arrays or when you
// need different encoding behavior based on context without creating multiple
// types.
type ArrayMarshalerFunc = base.ArrayMarshalerFunc

// ObjectMarshaler is the interface for custom object serialization.
//
// Implement this interface to provide zero-allocation encoding of custom
// struct types. The encoder will call MarshalLogObject to serialize the
// object's fields without using reflection.
//
// Example:
//
//	type User struct {
//	    ID   int64
//	    Name string
//	    Age  int
//	}
//
//	func (u User) MarshalLogObject(enc base.ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
//	    dst, _ = enc.AppendInt64(dst, "id", u.ID)
//	    dst, _ = enc.AppendString(dst, "name", u.Name)
//	    dst, _ = enc.AppendInt(dst, "age", u.Age)
//	    return dst, nil
//	}
//
//	// Usage:
//	user := User{ID: 123, Name: "Alice", Age: 30}
//	field := rxlog.Object("user", user)
//	// JSON output: {"user": {"id": 123, "name": "Alice", "age": 30}}
//
// The MarshalLogObject method receives an ObjectEncoder and a Buffer. It should
// append key-value pairs to the buffer using the encoder's methods and return
// the updated buffer along with any error encountered.
//
// Important: You MUST NOT write object delimiters ('{', '}') yourself. The
// encoder handles framing automatically.
//
// For fields that should be merged directly into the parent object without
// creating a nested structure, use the Inline() constructor instead of Object().
type ObjectMarshaler = base.ObjectMarshaler

// ObjectMarshalerFunc is a function adapter that allows using a plain function
// as an ObjectMarshaler.
//
// This is useful for creating ad-hoc object marshalers without defining a new type.
//
// Example:
//
//	// Define a function that encodes an object
//	encodeMetadata := rxlog.ObjectMarshalerFunc(func(enc base.ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
//	    dst, _ = enc.AppendString(dst, "version", "1.0")
//	    dst, _ = enc.AppendString(dst, "environment", "production")
//	    dst, _ = enc.AppendInt64(dst, "timestamp", time.Now().Unix())
//	    return dst, nil
//	})
//
//	// Use it as a field
//	field := rxlog.Object("metadata", encodeMetadata)
//	// JSON output: {"metadata": {"version": "1.0", "environment": "production", "timestamp": 1735689600}}
//
// This pattern is particularly useful for encoding context that's computed
// dynamically or when you need different encoding behavior based on runtime
// conditions without creating multiple types.
type ObjectMarshalerFunc = base.ObjectMarshalerFunc
