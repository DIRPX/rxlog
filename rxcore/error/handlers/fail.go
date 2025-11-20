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

package handlers

// FailOnErrorHandler stops encoding immediately when an error occurs.
//
// This handler represents the strictest policy: any encoding failure is
// treated as fatal for the current log entry. It is useful during debugging
// or in environments where partial entries are considered unacceptable.
type FailOnErrorHandler struct{}

// HandleError propagates the error and requests that encoding be stopped.
//
// When invoked, HandleError returns the original error without modification.
// The encoder MUST treat a non-nil return value as a signal to abort encoding
// of the current entry and to propagate the error to its caller.
//
// HandleError MUST be safe for concurrent use when used with a concurrent
// encoder.
func (h *FailOnErrorHandler) HandleError(key string, value interface{}, err error) error {
	return err
}
