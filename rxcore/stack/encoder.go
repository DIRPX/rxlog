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

package stack

import (
	"dirpx.dev/rxlog/rxapi/buffer"
)

// FullStackEncoder encodes a stack trace string as-is without any
// transformations or formatting.
//
// Format: <stack-trace>
//
// For example, if the stack trace is a multi-line string like:
//
//	goroutine 1 [running]:
//	main.foo(0x1, 0x2)
//	  /app/main.go:42 +0x123
//	main.main()
//	  /app/main.go:10 +0x456
//
// This encoder outputs it exactly as provided.
//
// When stack is empty, this encoder appends nothing to the buffer.
//
// This encoder is suitable for production logging where full stack traces
// are needed for debugging crashes, panics, or unexpected errors. The
// unmodified format ensures compatibility with stack trace parsers and
// error tracking tools.
func FullStackEncoder(dst *buffer.Buffer, stack string) *buffer.Buffer {
	if stack == "" {
		return dst
	}
	dst.AppendString(stack)
	return dst
}
