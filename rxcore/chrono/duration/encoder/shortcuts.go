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

package durenc

import (
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
	durationapi "dirpx.dev/rxlog/rxapi/chrono/duration"
)

// SecondsDurationEncoder serializes a time.Duration as a floating-point number
// of seconds elapsed.
//
// The value is written as a base-10 float64. This format is convenient for
// human-readable logs and many metrics systems, but note that floating-point
// representation may not preserve the original duration exactly.
var SecondsDurationEncoder durationapi.Encoder = func(dst *buffer.Buffer, d time.Duration) *buffer.Buffer {
	// Use float64 division to convert nanoseconds to seconds.
	seconds := float64(d) / float64(time.Second)
	dst.AppendFloat64(seconds)
	return dst
}

// MillisDurationEncoder serializes a time.Duration as an integer number of
// milliseconds elapsed since zero.
//
// The value is written as a base-10 int64. This format is compact and precise
// up to millisecond resolution.
var MillisDurationEncoder durationapi.Encoder = func(dst *buffer.Buffer, d time.Duration) *buffer.Buffer {
	// Go 1.13+ provides Milliseconds(); fall back to Nanoseconds()/1e6 if needed.
	dst.AppendInt64(d.Milliseconds())
	return dst
}

// MicrosDurationEncoder serializes a time.Duration as an integer number of
// microseconds elapsed since zero.
//
// The value is written as a base-10 int64.
var MicrosDurationEncoder durationapi.Encoder = func(dst *buffer.Buffer, d time.Duration) *buffer.Buffer {
	dst.AppendInt64(d.Microseconds())
	return dst
}

// NanosDurationEncoder serializes a time.Duration as an integer number of
// nanoseconds elapsed since zero.
//
// The value is written as a base-10 int64 and preserves the full precision of
// time.Duration.
var NanosDurationEncoder durationapi.Encoder = func(dst *buffer.Buffer, d time.Duration) *buffer.Buffer {
	dst.AppendInt64(d.Nanoseconds())
	return dst
}

// StringDurationEncoder serializes a time.Duration using its built-in String
// method (for example, "1s", "150ms", "200µs").
//
// This format is the most human-friendly and is reversible via time.ParseDuration,
// but may be less convenient for downstream systems that expect purely numeric
// values.
var StringDurationEncoder durationapi.Encoder = func(dst *buffer.Buffer, d time.Duration) *buffer.Buffer {
	dst.AppendString(d.String())
	return dst
}
