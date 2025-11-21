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

package caller

// Resolver abstracts the mechanism for capturing call-site information
// (file, line, function, program counter) from the current goroutine stack.
//
// The primary purpose of a Resolver is to decouple call-site inspection
// (for example, via runtime.Caller) from the rest of the logging pipeline,
// enabling alternative implementations (caching, symbolization, testing
// fakes, or integrations with third-party stack-walking libraries).
//
// Unless explicitly documented otherwise, implementations SHOULD be safe for
// concurrent use by multiple goroutines, since loggers typically resolve
// call sites from many call points in parallel.
type Resolver interface {
	// Caller returns information about the call site located skip stack frames
	// above the point where this method is invoked.
	//
	// The skip parameter follows the conventional runtime.Caller contract:
	//
	//   - skip == 0 refers to the frame of Resolver.Caller itself;
	//   - skip == 1 refers to the immediate caller of Resolver.Caller;
	//   - higher values walk further up the call stack.
	//
	// Implementations SHOULD clearly document any deviations from this
	// convention (for example, if they internally adjust skip to hide wrapper
	// frames). If the requested frame cannot be resolved, the returned Caller
	// value MUST have Defined set to false, and all other fields SHOULD be
	// treated as unspecified.
	Caller(skip int) Caller
}
