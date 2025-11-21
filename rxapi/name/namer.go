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

package name

// Namer provides a stable, human-readable name for an implementing type.
//
// Implementations SHOULD return a name that is suitable for use in logs,
// metrics, or diagnostic output (for example, a component identifier, logger
// name, or subsystem label). The returned value SHOULD be deterministic for a
// given instance and SHOULD NOT change over the lifetime of that instance.
//
// Callers MUST NOT assume any particular format (such as casing or delimiter
// style) unless explicitly documented by the implementation, but MAY rely on
// Name returning a non-empty string when a meaningful name is available.
type Namer interface {
	// Name returns a logical name associated with the receiver.
	//
	// Name SHOULD be a concise, human-readable identifier. Implementations
	// MAY return an empty string if no meaningful name is available, and
	// callers MUST be prepared to handle that case gracefully.
	Name() string
}
