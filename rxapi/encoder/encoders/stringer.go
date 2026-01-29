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

package encoders

import (
	"fmt"
	"reflect"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/encoder/base"
)

// EncodeStringer encodes a value implementing fmt.Stringer into the provided
// ObjectEncoder as a single string field, appending bytes to dst and returning
// the updated Buffer.
//
// The function expects stringer to be either nil or to hold a value that
// implements fmt.Stringer. On the happy path, it calls String() on that value
// and encodes the result via enc.AddString.
//
// Panic handling:
//
//   - If calling String() or encoding triggers a panic and the underlying
//     value is a nil pointer, EncodeStringer recovers the panic and encodes
//     the literal "<nil>" as the field value, returning a nil error.
//   - If a panic occurs for any other reason, EncodeStringer recovers it and
//     returns a non-nil error describing the panic. In this case it does NOT
//     write a value for the field, leaving responsibility to the caller to
//     react (for example, by adding a "<Key>Error" companion field).
//
// The returned Buffer MUST be treated as the new authoritative destination;
// it MAY differ from dst if the underlying storage had to grow. This function
// MUST NOT call Free on any Buffer and MUST NOT retain enc, dst, or the
// returned Buffer beyond the duration of the call.
func EncodeStringer(
	dst *buffer.Buffer,
	key string,
	stringer interface{},
	enc base.ObjectEncoder,
) (out *buffer.Buffer, retErr error) {
	out = dst

	defer func() {
		if r := recover(); r != nil {
			v := reflect.ValueOf(stringer)
			if v.Kind() == reflect.Ptr && v.IsNil() {
				// Treat nil receiver as a valid "<nil>" representation.
				out = enc.AddString(out, key, "<nil>")
				retErr = nil
				return
			}

			// For non-nil receivers, surface the panic as an error so callers
			// can record it (e.g., via a "<Key>Error" field).
			retErr = fmt.Errorf("panic in Stringer.String: %v", r)
		}
	}()

	// Handle a nil interface without triggering a panic.
	if stringer == nil {
		out = enc.AddString(out, key, "<nil>")
		return out, nil
	}

	s := stringer.(fmt.Stringer).String()
	out = enc.AddString(out, key, s)
	return out, nil
}
