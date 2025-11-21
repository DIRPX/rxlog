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

// LogOnErrorHandler logs encoding errors to a configured callback and then
// allows encoding to continue.
//
// This handler is useful in production when you want to preserve as many
// log entries as possible while still gaining visibility into fields that
// fail to encode. Each error is reported to Logger, but the failing field
// is skipped and subsequent fields continue to be processed.
type LogOnErrorHandler struct {
	// Logger is invoked for each encoding error.
	//
	// When Logger is non-nil, it SHOULD record or report the failure (for
	// example, by writing to stderr, a dedicated error log, or a metrics /
	// monitoring system). When Logger is nil, encoding errors handled by this
	// handler are effectively discarded.
	//
	// Implementations of Logger MUST be safe for concurrent use, since
	// encoding typically occurs from multiple goroutines.
	Logger func(key string, value interface{}, err error)
}

// HandleError logs the encoding failure and requests that encoding continue.
//
// When an error occurs, HandleError forwards key, value, and err to Logger
// if Logger is non-nil. It then returns nil to signal to the encoder that
// the failing field SHOULD be skipped and that remaining fields SHOULD
// continue to be encoded.
//
// HandleError MUST NOT modify encoder-internal state and MUST be safe for
// concurrent use when used with a concurrent encoder.
func (h *LogOnErrorHandler) HandleError(key string, value interface{}, err error) error {
	if h.Logger != nil {
		h.Logger(key, value, err)
	}
	return nil // Continue after logging
}
