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

	"dirpx.dev/rxlog/api/buffer"
	"dirpx.dev/rxlog/api/encoder/base"
)

// EncodeError encodes an error value into the provided ObjectEncoder as a
// single string field, appending bytes to dst and returning the updated
// Buffer.
//
// On the happy path, it calls err.Error() and encodes the resulting string via
// enc.AddString. If err is nil (including typed nils), it encodes the literal
// "<nil>" instead.
//
// Panic handling:
//
//   - If calling Error() or encoding triggers a panic and the underlying
//     error value is a nil pointer, EncodeError recovers the panic and
//     encodes "<nil>" as the field value, returning a nil error.
//   - If a panic occurs for any other reason, EncodeError recovers it and
//     returns a non-nil error describing the panic. In this case it does NOT
//     write a value for the field, leaving responsibility to the caller to
//     react (for example, by adding a "<Key>Error" companion field).
//
// The returned Buffer MUST be treated as the new authoritative destination;
// it MAY differ from dst if the underlying storage had to grow. This function
// MUST NOT call Free on any Buffer and MUST NOT retain enc, dst, or the
// returned Buffer beyond the duration of the call.
func EncodeError(
	dst *buffer.Buffer,
	key string,
	err error,
	enc base.ObjectEncoder,
) (out *buffer.Buffer, retErr error) {
	out = dst

	defer func() {
		if r := recover(); r != nil {
			v := reflect.ValueOf(err)
			if v.Kind() == reflect.Ptr && v.IsNil() {
				// Treat nil receiver as a valid "<nil>" representation.
				out = enc.AddString(out, key, "<nil>")
				retErr = nil
				return
			}

			// Surface unexpected panics as errors so callers can record them.
			retErr = fmt.Errorf("panic in error.Error: %v", r)
		}
	}()

	// Handle nil error (including typed nils) explicitly.
	if err == nil {
		out = enc.AddString(out, key, "<nil>")
		return out, nil
	}

	out = enc.AddString(out, key, err.Error())
	return out, nil
}
