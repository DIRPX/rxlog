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

package level

// Enabler decides whether log events at a given severity Level SHOULD be
// emitted.
//
// Implementations of Enabler are intended to provide deterministic,
// severity-based filtering. Concerns such as sampling, rate limiting, or
// content-based filtering SHOULD be implemented in higher-level components
// (for example, as separate predicates or cores), and MUST NOT be encoded
// into Enabled in a way that makes its behavior non-deterministic.
//
// A common implementation pattern is to treat a concrete Level value as a
// threshold: an enabler representing a minimum level MIN SHOULD return true
// for MIN and all more severe levels, and false for less severe levels.
// For example, if MIN is Warn, then Enabled(Warn), Enabled(Error), and
// Enabled(Fatal) SHOULD return true, while Enabled(Info) and Enabled(Debug)
// SHOULD return false.
type Enabler interface {
	// Enabled reports whether logging at the provided Level SHOULD be allowed.
	//
	// Enabled MUST NOT have side effects and MUST be safe to call concurrently
	// from multiple goroutines. Callers MAY invoke Enabled frequently (for
	// example, on every log attempt) and therefore implementations SHOULD keep
	// this check as lightweight as possible.
	Enabled(Level) bool
}

// EnablerFunc is an adapter that allows ordinary functions to satisfy
// the Enabler interface.
//
// It is defined as a function type that takes a Level and returns a bool,
// making it easy to implement custom, severity-based filtering logic inline
// without declaring a separate struct type.
//
// Example:
//
//	var onlyErrors Enabler = EnablerFunc(func(l Level) bool {
//	    return l >= Error
//	})
//
// In all other respects, EnablerFunc instances MUST obey the same contract
// as any other Enabler implementation:
//
//   - Enabled MUST be free of side effects;
//   - Enabled MUST be safe for concurrent use by multiple goroutines;
//   - Enabled SHOULD be as lightweight as possible, because it may be
//     invoked on every logging attempt.
type EnablerFunc func(Level) bool

// Enabled implements the Enabler interface for EnablerFunc.
//
// It simply calls the underlying function f with the provided Level and
// returns its boolean result. The behavioral and concurrency guarantees are
// inherited from the function implementation: callers MUST ensure that the
// wrapped function is side-effect free, concurrency-safe, and sufficiently
// cheap to call on hot paths.
func (f EnablerFunc) Enabled(lvl Level) bool {
	return f(lvl)
}
