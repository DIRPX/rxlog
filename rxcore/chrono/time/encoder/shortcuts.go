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
	"dirpx.dev/rxlog/rxcore/chrono/time/encoder/layout"
)

// These predefined encoders format time as human-readable strings or numeric
// timestamps. All of them are constructed via TimeLayoutEncoder (for textual
// formats) or inline closures (for Unix epoch encoders) and write directly
// into the provided buffer.Buffer.
//
// The encoders themselves are stateless and MAY be safely shared across
// goroutines. Callers MUST, however, respect the usual buffer ownership
// rules: each call MUST operate on a *buffer.Buffer that is not being
// concurrently modified elsewhere.

// ISO 8601–style encoders (UTC by default).
var (
	// ISO8601SecondsTimeEncoder formats time as an ISO 8601–style string with
	// second precision and a numeric time zone offset including a colon.
	//
	// Layout:
	//   2006-01-02T15:04:05Z07:00
	//
	// Example:
	//   2025-11-20T18:47:12+03:00
	ISO8601SecondsTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.ISO8601Seconds, time.UTC)

	// ISO8601MillisTimeEncoder formats time as an ISO 8601–style string with
	// millisecond precision and a numeric time zone offset without a colon
	// (compatible with zap’s default JSON format).
	//
	// Layout:
	//   2006-01-02T15:04:05.000Z0700
	//
	// Example:
	//   2025-11-20T18:47:12.123+0300
	ISO8601MillisTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.ISO8601Millis, time.UTC)

	// ISO8601MicrosTimeEncoder formats time as an ISO 8601–style string with
	// microsecond precision and a numeric time zone offset including a colon.
	//
	// Layout:
	//   2006-01-02T15:04:05.000000Z07:00
	ISO8601MicrosTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.ISO8601Micros, time.UTC)

	// ISO8601NanosTimeEncoder formats time as an ISO 8601–style string with
	// nanosecond precision and a numeric time zone offset including a colon.
	//
	// Layout:
	//   2006-01-02T15:04:05.000000000Z07:00
	ISO8601NanosTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.ISO8601Nanos, time.UTC)
)

// RFC 3339 encoders (canonical for JSON / API payloads).
var (
	// RFC3339TimeEncoder formats time using the RFC3339 layout with second
	// precision and a numeric time zone offset including a colon.
	//
	// Layout:
	//   2006-01-02T15:04:05Z07:00
	//
	// Example:
	//   2025-11-20T18:47:12+03:00
	RFC3339TimeEncoder time2.Encoder = TimeLayoutEncoder(layout.RFC3339, time.UTC)

	// RFC3339NanoTimeEncoder formats time using RFC3339Nano, providing
	// nanosecond precision while remaining human-readable.
	//
	// This is typically the most precise standardized textual format and is
	// well-suited for high-precision logging.
	RFC3339NanoTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.RFC3339Nano, time.UTC)
)

// Other common textual layouts.
var (
	// RFC1123TimeEncoder formats time using the RFC1123 layout.
	//
	// Example:
	//   Mon, 02 Jan 2006 15:04:05 MST
	RFC1123TimeEncoder time2.Encoder = TimeLayoutEncoder(layout.RFC1123, time.UTC)

	// RFC1123ZTimeEncoder formats time using the RFC1123Z layout, which uses
	// a numeric time zone offset instead of an abbreviation.
	//
	// Example:
	//   Mon, 02 Jan 2006 15:04:05 -0700
	RFC1123ZTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.RFC1123Z, time.UTC)

	// RFC822TimeEncoder formats time using the RFC822 layout.
	RFC822TimeEncoder time2.Encoder = TimeLayoutEncoder(layout.RFC822, time.UTC)

	// RFC822ZTimeEncoder formats time using the RFC822Z layout, which includes
	// an explicit numeric zone.
	RFC822ZTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.RFC822Z, time.UTC)

	// RFC850TimeEncoder formats time using the RFC850 layout (HTTP-date style).
	//
	// Example:
	//   Monday, 02-Jan-06 15:04:05 MST
	RFC850TimeEncoder time2.Encoder = TimeLayoutEncoder(layout.RFC850, time.UTC)

	// ANSICTimeEncoder formats time using the ANSIC layout.
	//
	// Example:
	//   Mon Jan _2 15:04:05 2006
	//
	// This encoder uses time.Local, which means the output reflects the local
	// time zone of the running process.
	ANSICTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.ANSIC, time.Local)

	// UnixDateTimeEncoder formats time using the UnixDate layout.
	//
	// Example:
	//   Mon Jan _2 15:04:05 MST 2006
	//
	// As with ANSICTimeEncoder, it uses time.Local for time zone resolution.
	UnixDateTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.UnixDate, time.Local)

	// KitchenTimeEncoder formats only the clock portion in a human-friendly
	// 12-hour form with AM/PM.
	//
	// Layout:
	//   3:04PM
	//
	// Example:
	//   6:47PM
	//
	// Passing nil as the location means the timestamp’s existing Location is
	// used as-is.
	KitchenTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.Kitchen, nil)

	// StampTimeEncoder formats time using time.Stamp, a short locale-like
	// representation intended primarily for human consumption.
	StampTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.Stamp, nil)

	// StampMilliTimeEncoder formats time using time.StampMilli, extending
	// Stamp with millisecond precision.
	StampMilliTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.StampMilli, nil)

	// StampMicroTimeEncoder formats time using time.StampMicro, extending
	// Stamp with microsecond precision.
	StampMicroTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.StampMicro, nil)

	// StampNanoTimeEncoder formats time using time.StampNano, extending
	// Stamp with nanosecond precision.
	StampNanoTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.StampNano, nil)
)

// Date / time-only textual encoders.
var (
	// DateOnlyTimeEncoder formats only the calendar date (YYYY-MM-DD)
	// in UTC, with no time-of-day component.
	//
	// Layout:
	//   2006-01-02
	DateOnlyTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.Date, time.UTC)

	// TimeOnlyTimeEncoder formats only the clock time with second precision
	// in UTC, with no date component.
	//
	// Layout:
	//   15:04:05
	TimeOnlyTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.Time, time.UTC)

	// TimeMillisOnlyTimeEncoder formats only the clock time with millisecond
	// precision in UTC.
	//
	// Layout:
	//   15:04:05.000
	TimeMillisOnlyTimeEncoder time2.Encoder = TimeLayoutEncoder(layout.TimeMillis, time.UTC)
)

// Numeric Unix timestamp encoders.
//
// These encoders write integer Unix timestamps in different units. They are
// useful for systems that prefer purely numeric time representations (for
// example, metrics, binary logs, or downstream consumers that treat time as
// an integer value rather than a formatted string).
var (
	// UnixSecondsTimeEncoder writes seconds since the Unix epoch
	// (1970-01-01T00:00:00Z) as a base-10 integer.
	//
	// Example:
	//   1672531200
	UnixSecondsTimeEncoder time2.Encoder = func(dst *buffer.Buffer, t time.Time) *buffer.Buffer {
		dst.AppendInt64(t.Unix())
		return dst
	}

	// UnixMillisTimeEncoder writes milliseconds since the Unix epoch as a
	// base-10 integer.
	//
	// Example:
	//   1672531200000
	UnixMillisTimeEncoder time2.Encoder = func(dst *buffer.Buffer, t time.Time) *buffer.Buffer {
		dst.AppendInt64(t.UnixMilli())
		return dst
	}

	// UnixMicrosTimeEncoder writes microseconds since the Unix epoch as a
	// base-10 integer.
	UnixMicrosTimeEncoder time2.Encoder = func(dst *buffer.Buffer, t time.Time) *buffer.Buffer {
		dst.AppendInt64(t.UnixMicro())
		return dst
	}

	// UnixNanosTimeEncoder writes nanoseconds since the Unix epoch as a
	// base-10 integer.
	UnixNanosTimeEncoder time2.Encoder = func(dst *buffer.Buffer, t time.Time) *buffer.Buffer {
		dst.AppendInt64(t.UnixNano())
		return dst
	}
)
