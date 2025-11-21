/*
   Copyright 2025 The DIRPX Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you MAY not use this file except in compliance with the License.
   You MAY obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package level

// Threshold represents a mutable severity threshold used for level-based
// filtering.
//
// Conceptually, a Threshold is an Enabler whose decision is governed by a
// single minimum level. Events at or above this level SHOULD be considered
// enabled; events below it SHOULD be considered disabled. A typical
// implementation treats Level() as the current minimum and implements
// Enabled(l) as:
//
//	return l >= Level()
//
// Threshold is intended for components that need to both check and adjust
// the active log level at runtime (for example, via configuration reload,
// an admin HTTP endpoint, or command-line flags).
//
// Implementations MUST be safe for concurrent use by multiple goroutines.
// They MUST keep Enabled as lightweight and side-effect free as any other
// Enabler implementation, since it may be invoked on every log attempt.
type Threshold interface {
	Enabler

	// Level returns the current minimum enabled level.
	//
	// Events at this level or any more severe level SHOULD be considered
	// enabled by this Threshold, subject to whatever additional policies
	// the implementation documents (for example, clamping to a supported
	// range).
	Level() Level

	// SetLevel updates the minimum enabled level used by this Threshold.
	//
	// The supplied level MUST be a valid severity (IsValid() == true).
	// Implementations MAY panic on invalid values, or they MAY clamp or
	// reject them according to their documented policy. Callers SHOULD NOT
	// rely on any particular behavior unless it is explicitly documented by
	// the implementation.
	SetLevel(Level)
}
