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

import (
	"time"

	"dirpx.dev/rxlog/rxapi/chrono/clock"
)

// DefaultClock is the default Clock implementation used by the logging
// subsystem.
//
// It is backed by the standard library's time package and provides real
// wall-clock time (time.Now) and standard tickers (time.NewTicker). This
// value is intended to be used as a process-wide singleton wherever a
// concrete Clock is required but no custom time source is configured.
//
// Callers MAY replace DefaultClock in their own wiring (for example, in
// tests or simulations) with an alternative implementation of clock.Clock
// that provides deterministic or virtual time.
var DefaultClock clock.Clock = systemClock{}

// systemClock is a zero-sized implementation of clock.Clock that delegates
// to the standard library time package.
//
// It is intentionally unexported: consumers should depend on the exported
// DefaultClock value (or define their own Clock implementation) rather than
// constructing systemClock directly. Multiple instances of systemClock are
// indistinguishable and stateless.
type systemClock struct{}

// Now returns the current time according to the underlying system clock.
//
// This method is a thin wrapper around time.Now and therefore reflects the
// process's view of wall-clock time, including any adjustments made by the
// operating system (for example, NTP corrections). The returned value MAY
// include a monotonic component as provided by Go's time.Time semantics.
func (systemClock) Now() time.Time {
	return time.Now()
}

// NewTicker creates a new *time.Ticker that delivers ticks at the specified
// interval.
//
// This method is a thin wrapper around time.NewTicker and therefore inherits
// its semantics: the returned ticker sends periodic time values on its C
// channel until Stop is called. Callers MUST ensure that Stop is eventually
// invoked to avoid leaking resources.
//
// The behavior with respect to very short durations, system clock changes,
// and scheduling delays is identical to time.NewTicker.
func (systemClock) NewTicker(duration time.Duration) *time.Ticker {
	return time.NewTicker(duration)
}
