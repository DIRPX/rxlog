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

package timeenc

import (
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
	time2 "dirpx.dev/rxlog/rxapi/chrono/time"
)

// TimeLayoutEncoder constructs a time encoder that formats time.Time values
// using the specified layout and optional location.
//
// The returned Encoder applies the following rules:
//
//   - The layout parameter MUST follow Go's standard reference time semantics
//     (for example, "2006-01-02T15:04:05Z07:00") as used by time.Format.
//   - If loc is non-nil, each timestamp t is first converted via t.In(loc)
//     before formatting; if loc is nil, t is formatted as-is, preserving its
//     existing location.
//   - The formatted result is appended to dst as plain text via AppendString.
//     No quoting or additional framing is added by this helper.
//
// The returned Encoder is safe for concurrent use across goroutines as long
// as callers respect the usual buffer ownership rules (that is, each call
// MUST provide a *buffer.Buffer that is not shared concurrently). The closure
// itself captures only layout and loc, which are treated as immutable.
func TimeLayoutEncoder(layout string, loc *time.Location) time2.Encoder {
	return func(dst *buffer.Buffer, t time.Time) *buffer.Buffer {
		if loc != nil {
			t = t.In(loc)
		}

		s := t.Format(layout)
		dst.AppendString(s)

		return dst
	}
}
