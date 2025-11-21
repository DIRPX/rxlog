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

// Package clock defines an abstraction over time used throughout the logging
// stack.
//
// Instead of calling time.Now or time.NewTicker directly, components that need
// the current time or periodic ticks SHOULD depend on the Clock interface
// exposed by this package. This indirection makes time-dependent code easier
// to test, easier to reason about, and more flexible in environments where
// “real” wall-clock time is not appropriate (for example, simulations or
// deterministic replay).
//
// # Core abstraction
//
// The central type in this package is the Clock interface, which provides two
// operations:
//
//   - Now() time.Time
//     Returns the current time as a time.Time value. Implementations MAY
//     return wall-clock time, monotonic time, or a virtual time source,
//     but SHOULD document which semantics they provide. Callers MUST NOT
//     make assumptions about the exact origin of the time beyond what the
//     implementation guarantees.
//
//   - NewTicker(d time.Duration) *time.Ticker
//     Creates a ticker that delivers periodic “ticks” at the given
//     interval. The returned value behaves like the standard library’s
//     time.Ticker: it exposes a C channel that receives time values at (or
//     near) the requested period, and it MUST be stopped by the caller
//     when no longer needed to avoid leaks.
//
// Any concrete implementation of Clock MUST be safe for concurrent use by
// multiple goroutines if it is intended to be used with process-wide loggers
// or shared infrastructure.
//
// # Default implementation: DefaultClock
//
// The package provides a default implementation, DefaultClock, which wraps the
// standard library’s time package:
//
//   - Now delegates to time.Now.
//   - NewTicker delegates to time.NewTicker.
//
// DefaultClock is intended to be used as a process-wide singleton wherever a
// concrete Clock is required but no custom time source is configured. It
// provides real wall-clock time with the usual properties of time.Now, and
// its tickers behave exactly like time.Ticker instances.
//
// Code that is not explicitly concerned with testing or virtual time MAY use
// DefaultClock directly, but SHOULD still depend on the Clock interface at
// API boundaries so that the implementation can be swapped when necessary.
//
// # Testing and virtual time
//
// A primary motivation for this package is to make time-based behavior
// testable and controllable. Callers MAY provide alternative Clock
// implementations that:
//
//   - return deterministic or manually controlled times from Now,
//   - expose tickers driven by explicit “advance” operations instead of real
//     elapsed time,
//   - simulate slow, fast, or skewed clocks for resilience testing.
//
// For example, unit tests can inject a fake Clock into logging components,
// allowing assertions on timestamps or timeouts without waiting for real
// time to pass. Integration or simulation environments can use virtual time
// to run long scenarios quickly and repeatably.
//
// # Usage guidelines
//
//   - Library code that logs or schedules periodic tasks SHOULD accept a
//     Clock interface where practical, defaulting to DefaultClock if no
//     specific implementation is supplied.
//
//   - Applications that do not need custom time behavior MAY ignore this
//     package and rely on the defaults wired into higher-level logging
//     configuration.
//
//   - Implementations of Clock MUST clearly document any deviations from the
//     behavior of the standard time package (for example, whether Now
//     advances automatically, whether tickers can drift, or whether time can
//     move backwards).
package clock
