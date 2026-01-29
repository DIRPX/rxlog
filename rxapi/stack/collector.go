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

package stack

// Collector builds stack trace strings from the current goroutine stack.
//
// A Collector abstracts the mechanism for capturing and formatting stack
// information (frames, file/line pairs, function names, program counters)
// into a single textual representation, typically for inclusion in log
// entries or error reports.
//
// Implementations SHOULD be safe for concurrent use by multiple goroutines,
// since stack traces are commonly captured from many call sites in parallel.
// Any internal caching, symbolization, or buffer reuse MUST preserve this
// concurrency guarantee.
type Collector interface {
	// Stack captures and formats a stack trace for the current goroutine and
	// returns it as a string.
	//
	// The skip argument controls how many stack frames above the call to Stack
	// are omitted from the result. Its semantics SHOULD mirror those of
	// runtime.Caller: a skip value of 0 refers to the frame of Collector.Stack
	// itself, a value of 1 refers to the direct caller of Collector.Stack, and
	// larger values walk further up the call stack.
	//
	// The depth argument limits how many frames are included in the returned
	// trace. Implementations SHOULD interpret positive values as an upper
	// bound on the number of frames to encode. If depth is zero or negative,
	// implementations MAY treat this as “no explicit limit” and capture as
	// many frames as are reasonably available, subject to their own internal
	// safety limits.
	//
	// The returned string SHOULD use a multi-line, human-oriented format that
	// is suitable for inclusion in logs (for example, one frame per line with
	// function, file, and line number). The exact format is
	// implementation-defined but SHOULD be documented and kept stable over
	// time so that downstream tools can parse or display it consistently.
	Stack(skip, depth int) string
}
