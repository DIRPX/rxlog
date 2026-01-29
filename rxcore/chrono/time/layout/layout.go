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

package layout

import "time"

// These string layouts define commonly used time formatting patterns.
// They MAY be used directly with time.Format / time.Parse or passed to
// EncoderOfLayout to build reusable time encoders.
//
// Names are intentionally descriptive and explicit so that configuration
// remains self-documenting and unambiguous.

// ISO 8601–inspired layouts (not official standard identifiers, but
// widely used patterns).
const (
	// ISO8601Seconds formats a timestamp as an ISO 8601–style string with
	// second precision and a numeric time zone offset including a colon.
	//
	// Layout:
	//   2006-01-02T15:04:05Z07:00
	//
	// Example:
	//   2025-11-20T18:47:12+03:00
	ISO8601Seconds = "2006-01-02T15:04:05Z07:00"

	// ISO8601Millis formats a timestamp as an ISO 8601–style string with
	// millisecond precision and a numeric time zone offset without a colon
	// (similar to zap’s default style).
	//
	// Layout:
	//   2006-01-02T15:04:05.000Z0700
	//
	// Example:
	//   2025-11-20T18:47:12.123+0300
	ISO8601Millis = "2006-01-02T15:04:05.000Z0700"

	// ISO8601Micros formats a timestamp as an ISO 8601–style string with
	// microsecond precision and a numeric time zone offset including a colon.
	//
	// Layout:
	//   2006-01-02T15:04:05.000000Z07:00
	//
	// Example:
	//   2025-11-20T18:47:12.123456+03:00
	ISO8601Micros = "2006-01-02T15:04:05.000000Z07:00"

	// ISO8601Nanos formats a timestamp as an ISO 8601–style string with
	// nanosecond precision and a numeric time zone offset including a colon.
	//
	// Layout:
	//   2006-01-02T15:04:05.000000000Z07:00
	//
	// Example:
	//   2025-11-20T18:47:12.123456789+03:00
	ISO8601Nanos = "2006-01-02T15:04:05.000000000Z07:00"
)

// Common date/time layouts.
const (
	// Date formats only the calendar date (year, month, day) without any
	// time-of-day information.
	//
	// Layout:
	//   2006-01-02
	//
	// Example:
	//   2025-11-20
	Date = "2006-01-02"

	// Time formats only the clock time with second precision, without any
	// date or time zone information.
	//
	// Layout:
	//   15:04:05
	//
	// Example:
	//   18:47:12
	Time = "15:04:05"

	// TimeMillis formats only the clock time with millisecond precision.
	//
	// Layout:
	//   15:04:05.000
	//
	// Example:
	//   18:47:12.123
	TimeMillis = "15:04:05.000"
)

// Re-export Go's predefined layouts under this package.
//
// These constants are aliases of the layouts defined in the standard
// library's time package (time.RFC3339, time.UnixDate, etc.). They are
// provided so that configuration code using this package does not need
// to import time directly, while still relying on well-known formats.
const (
	RFC3339     = time.RFC3339
	RFC3339Nano = time.RFC3339Nano
	RFC1123     = time.RFC1123
	RFC1123Z    = time.RFC1123Z
	RFC822      = time.RFC822
	RFC822Z     = time.RFC822Z
	RFC850      = time.RFC850
	ANSIC       = time.ANSIC
	UnixDate    = time.UnixDate
	Kitchen     = time.Kitchen
	Stamp       = time.Stamp
	StampMilli  = time.StampMilli
	StampMicro  = time.StampMicro
	StampNano   = time.StampNano
)
