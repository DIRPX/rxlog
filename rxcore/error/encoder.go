/*
   Copyright 2026 The DIRPX Authors.

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

package error

import (
	"dirpx.dev/rxlog/rxapi/buffer"
)

// SimpleErrorEncoder encodes an error using its Error() method, producing
// a plain string representation.
//
// Format: <error-message>
//
// For example, if err.Error() returns "connection refused", this encoder
// outputs "connection refused".
//
// When err is nil, this encoder appends nothing to the buffer, treating
// the absence of an error as a no-op rather than outputting a placeholder.
//
// This encoder is suitable for most logging scenarios where the error message
// itself is sufficient and you don't need structured error details or
// wrapped error chains.
func SimpleErrorEncoder(dst *buffer.Buffer, err error) *buffer.Buffer {
	if err == nil {
		return dst
	}
	dst.AppendString(err.Error())
	return dst
}
