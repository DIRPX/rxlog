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

package level

import (
	"strings"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/level"
)

// LowercaseLevelEncoder serializes a Level to a lowercase string.
//
// For example, level.Info is encoded as "info", level.Error as "error", etc.
// This encoder uses the canonical lowercase names returned by Level.String().
//
// Example output: "trace", "debug", "info", "notice", "warn", "error",
// "critical", "fatal".
func LowercaseLevelEncoder(dst *buffer.Buffer, l level.Level) *buffer.Buffer {
	dst.AppendString(l.String())
	return dst
}

// CapitalLevelEncoder serializes a Level to an all-caps string.
//
// For example, level.Info is encoded as "INFO", level.Error as "ERROR", etc.
// The output is the uppercase version of the canonical level names.
//
// Example output: "TRACE", "DEBUG", "INFO", "NOTICE", "WARN", "ERROR",
// "CRITICAL", "FATAL".
func CapitalLevelEncoder(dst *buffer.Buffer, l level.Level) *buffer.Buffer {
	dst.AppendString(strings.ToUpper(l.String()))
	return dst
}
