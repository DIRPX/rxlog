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

package caller

import (
	"path/filepath"
	"strconv"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/caller"
)

// ShortCallerEncoder encodes caller information with only the base filename
// and line number, omitting the directory path.
//
// Format: <filename>:<line>
//
// For example, if the full path is "/app/internal/handlers/user.go:42",
// ShortCallerEncoder outputs "user.go:42".
//
// When caller.Defined is false, this encoder appends "<undefined>" to indicate
// that caller information is not available.
//
// This format is compact and suitable for console output or logs where the
// full path would add unnecessary noise, especially when the project structure
// is well-known to the developer.
func ShortCallerEncoder(dst *buffer.Buffer, c caller.Caller) *buffer.Buffer {
	if !c.Defined {
		dst.AppendString("<undefined>")
		return dst
	}

	// Extract base filename from full path.
	filename := filepath.Base(c.File)
	dst.AppendString(filename)
	dst.AppendByte(':')
	dst.AppendString(strconv.Itoa(c.Line))
	return dst
}

// FullCallerEncoder encodes caller information with the complete file path
// and line number.
//
// Format: <full-path>:<line>
//
// For example: "/app/internal/handlers/user.go:42"
//
// When caller.Defined is false, this encoder appends "<undefined>" to indicate
// that caller information is not available.
//
// This format provides maximum detail and is useful for production logs,
// error tracking systems, or situations where the full path is needed to
// disambiguate files with the same base name across different packages.
func FullCallerEncoder(dst *buffer.Buffer, c caller.Caller) *buffer.Buffer {
	if !c.Defined {
		dst.AppendString("<undefined>")
		return dst
	}

	dst.AppendString(c.File)
	dst.AppendByte(':')
	dst.AppendString(strconv.Itoa(c.Line))
	return dst
}
