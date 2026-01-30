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

package name

import (
	"dirpx.dev/rxlog/rxapi/buffer"
)

// FullNameEncoder encodes the complete logger name as-is without any
// transformations.
//
// Format: <name>
//
// For example, if the logger name is "service.handler.user", this encoder
// outputs "service.handler.user" exactly.
//
// This encoder is suitable for structured logging formats (JSON, logfmt)
// where the full hierarchical name helps with log aggregation, filtering,
// and grouping by component or subsystem.
//
// When name is empty, this encoder appends nothing to the buffer.
func FullNameEncoder(dst *buffer.Buffer, n string) *buffer.Buffer {
	if n == "" {
		return dst
	}
	dst.AppendString(n)
	return dst
}
