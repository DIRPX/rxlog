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

package extractor

import (
	"context"

	"dirpx.dev/rxlog/api/field"
)

// Extractor extracts logging fields from a context.Context.
//
// Implementations MAY inspect values stored in the context (such as request
// identifiers, user information, or tracing metadata) and translate them into
// a slice of field.Field values. Callers MUST NOT assume that Extract returns
// a non-nil or non-empty slice; an Extractor MAY return nil or an empty slice
// if no relevant data is present.
type Extractor interface {
	Extract(context.Context) []field.Field
}

// Func is a function adapter that turns a plain function into an
// Extractor.
//
// Any function with the signature func(context.Context) []field.Field MAY be
// wrapped as a Func and used wherever an Extractor is required.
type Func func(context.Context) []field.Field

// Extract calls the underlying function.
//
// When used via this adapter, the function MUST obey the same contract as
// Extractor.Extract: it MAY return nil or an empty slice, and callers MUST NOT
// mutate any shared internal state without appropriate synchronization.
func (f Func) Extract(ctx context.Context) []field.Field {
	return f(ctx)
}

// Compose returns an Extractor that combines multiple Extractor instances.
//
// The returned Extractor invokes each provided Extractor in order and
// concatenates their results into a single slice. The composed Extractor:
//
//   - MUST call all provided extractors exactly once per Extract invocation,
//     in the order they were passed to Compose.
//   - MUST preserve the relative ordering of fields produced by each
//     individual Extractor within the combined result.
//   - MAY allocate a new slice on each call; callers MUST NOT rely on any
//     reuse of the backing array.
//
// If no extractors are provided, the returned Extractor will always return an
// empty slice. Callers SHOULD treat the result as read-only and MUST NOT
// modify the returned slice in ways that could affect other code paths.
func Compose(extractors ...Extractor) Extractor {
	return Func(func(ctx context.Context) []field.Field {
		var fields []field.Field
		for _, ext := range extractors {
			fields = append(fields, ext.Extract(ctx)...)
		}
		return fields
	})
}
