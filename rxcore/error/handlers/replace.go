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

// ReplaceOnErrorHandler replaces the failing field's value with a placeholder.
//
// This handler is useful when you want to preserve the presence of a field
// but explicitly indicate that encoding of the original value failed.
// The placeholder can be any value that the encoder knows how to serialize.
type ReplaceOnErrorHandler struct {
	// Placeholder is the value to use instead of the failing value.
	//
	// Typical placeholders include static strings (such as "<error>" or
	// "[encoding failed]"), nil, or small structured payloads that describe
	// the failure (such as a map or struct with an "error" field). The encoder
	// implementation MUST document how it interprets Placeholder and how it
	// obtains it from this handler.
	Placeholder interface{}
}

// HandleError indicates that the failing field SHOULD be replaced rather
// than dropped or cause encoding to fail.
//
// This method itself only signals the replacement policy by returning nil;
// the actual substitution with Placeholder MUST be implemented by the
// encoder. The encoder MUST detect that the active handler is a
// ReplaceOnErrorHandler (for example, via type assertion) and MUST use the
// Placeholder value instead of the original value when encoding the field.
//
// HandleError MUST be safe for concurrent use when used with a concurrent
// encoder.
func (h *ReplaceOnErrorHandler) HandleError(key string, value interface{}, err error) error {
	// Note: The actual replacement happens in the encoder implementation.
	// This handler only signals that replacement is desired.
	return nil
}
