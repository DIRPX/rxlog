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

package handler

// SkipOnErrorHandler skips the failing field and continues encoding.
//
// This handler reflects the most common production policy: a single bad
// field MUST NOT prevent the rest of the log entry from being written.
// Failures are not logged or surfaced directly by this handler; they are
// simply ignored at the field level.
type SkipOnErrorHandler struct{}

// HandleError signals that the failing field SHOULD be skipped and that
// encoding SHOULD continue for remaining fields.
//
// HandleError always returns nil and performs no logging or transformation
// of the error. It MUST be safe for concurrent use when used with a
// concurrent encoder.
func (h *SkipOnErrorHandler) HandleError(key string, value interface{}, err error) error {
	return nil // Skip and continue
}
