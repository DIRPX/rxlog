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

// Package clock defines the time-source abstraction used by the rxlog logging
// pipeline.
//
// Instead of calling time.Now or time.NewTicker directly, higher-level
// components depend on the Clock interface defined in this package. This makes
// the source of time configurable, testable, and replaceable without changing
// the rest of the logging stack.
//
// # Overview
//
// The clock package revolves around a single central concept:
//
//   - Clock: an interface that provides the current time and constructs
//     tickers configured to deliver periodic time values.
//
// A Clock implementation is responsible for two operations:
//
//   - reporting the "current" time via Now, and
//   - creating tickers via NewTicker that emit time values at a specified
//     interval.
//
// The concrete notion of "current" time is left to the implementation. A Clock
// MAY represent real wall-clock time, process time, simulated time, or any
// other monotonic or non-monotonic time source, as long as it satisfies the
// documented contract.
//
// # Relationship to the standard library
//
// The Clock interface is intentionally aligned with the standard library's
// time package:
//
//   - Now returns a time.Time value, just like time.Now.
//   - NewTicker returns a *time.Ticker configured to tick at the requested
//     interval, analogous to time.NewTicker.
//
// This alignment allows existing code and mental models around time.Time,
// time.Duration, and time.Ticker to carry over directly. At the same time, the
// indirection through Clock makes it possible to:
//
//   - replace the real system clock with a deterministic or virtual clock in
//     tests and simulations;
//   - centralize clock selection in application wiring, rather than scattering
//     time.Now calls throughout the codebase;
//   - run the same logging code against different time sources in different
//     environments (for example, real time in production and simulated time in
//     integration tests).
//
// # Clock semantics
//
// A Clock implementation MUST satisfy the following high-level guarantees:
//
//   - Now MUST be safe for concurrent use by multiple goroutines. It SHOULD be
//     fast and non-blocking, since it may be called for every log entry in
//     high-throughput applications.
//
//   - NewTicker MUST return a *time.Ticker configured to deliver ticks at the
//     requested interval. The returned ticker MUST behave analogously to
//     time.NewTicker:
//
//   - it MUST send periodic time.Time values on its C channel;
//
//   - it MUST stop delivering ticks after Stop is called;
//
//   - it MUST be safe for use from multiple goroutines as per the standard
//     library contract.
//
//   - If a given Clock cannot support tickers (for example, a minimal test
//     implementation), it MUST document this limitation and SHOULD panic or
//     fail fast in NewTicker rather than silently returning a broken ticker.
//
// Callers that use NewTicker are responsible for calling Stop on the returned
// ticker to avoid resource leaks. The Clock implementation itself MUST NOT
// attempt to infer or enforce when a ticker is no longer needed.
//
// # Usage in the logging pipeline
//
// In a typical rxlog configuration:
//
//   - When a log entry is created, the logging infrastructure calls Clock.Now
//     to obtain the timestamp that will be stored on the entry.
//
//   - Periodic tasks within the logging subsystem (for example, background
//     maintenance or metrics) MAY use Clock.NewTicker instead of
//     time.NewTicker directly, so that they follow the same time source as
//     the rest of the system.
//
// By depending on clock.Clock rather than the concrete system clock, the
// logging pipeline remains agnostic to how time is produced. Applications can
// choose an appropriate Clock implementation for each environment while
// keeping the rxlog API and behavior consistent.
//
// # Implementations
//
// This package only defines the Clock interface and its contract. Concrete
// implementations (for example, a Clock that delegates to the real system
// time) are provided in separate packages outside the api layer. Users are
// free to supply their own implementations that satisfy the interface and the
// guarantees described above.
package clock
