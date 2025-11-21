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

package clock

import "time"

// Clock abstracts a time source used by the logging subsystem.
//
// This interface exists to decouple time acquisition from concrete
// implementations (for example, time.Now and time.NewTicker), making it
// possible to provide custom or deterministic clocks in tests and in
// specialized runtime environments (such as simulators or replay tools).
//
// Implementations SHOULD be safe for concurrent use by multiple goroutines,
// since loggers typically access time from many call sites in parallel.
type Clock interface {
	// Now returns the current time according to this Clock.
	//
	// Implementations MAY return wall-clock time, monotonic time, or a
	// combination, but they SHOULD document the semantics they provide
	// (for example, whether the value corresponds to time.Now). Callers MUST
	// NOT assume a particular time zone; they SHOULD treat the returned value
	// as an opaque instant and format/convert it as needed.
	Now() time.Time

	// NewTicker creates and returns a *time.Ticker associated with this Clock,
	// configured to deliver ticks at the specified interval.
	//
	// The returned ticker MUST behave analogously to time.NewTicker: it MUST
	// send periodic time values on its C channel until Stop is called, and it
	// MUST be safe for use from multiple goroutines as per the standard
	// library contract.
	//
	// Callers MUST ensure that Stop is eventually called on the returned
	// ticker to avoid resource leaks. Implementations MAY return a ticker that
	// uses a custom notion of time (for example, a mocked or virtual clock),
	// but they SHOULD document any behavior that differs from the standard
	// time.Ticker.
	NewTicker(time.Duration) *time.Ticker
}
