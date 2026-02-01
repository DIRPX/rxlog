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

package field

const (
	// CorrelationID is the field key for an application-level correlation
	// identifier that ties multiple logs together into a single logical
	// transaction or business flow.
	//
	// Unlike TraceID, this identifier MAY originate from a client or an
	// external system. Producers SHOULD use this field for end-to-end
	// correlation across services when distributed tracing is not available
	// or when a higher-level business identifier is needed.
	CorrelationID = "correlation_id"

	// TraceID is the field key for the distributed tracing trace identifier
	// (for example, W3C Trace Context or OpenTelemetry).
	//
	// When distributed tracing is enabled, producers SHOULD populate this
	// field with the current trace identifier so that logs can be joined with
	// traces in observability backends.
	TraceID = "trace_id"

	// SpanID is the field key for the distributed tracing span identifier
	// that links the log entry to a specific span within a trace.
	//
	// Producers that emit logs from within traced operations SHOULD set this
	// field so that individual log entries can be associated with the precise
	// span context in which they occurred.
	SpanID = "span_id"

	// ParentSpanID is the field key that records the identifier of the
	// parent span, if any.
	//
	// This is useful for reconstructing span hierarchies from log streams.
	ParentSpanID = "parent_span_id"

	// RequestID is the field key that identifies the current request within
	// a service or across services.
	//
	// This identifier is typically generated at the edge (for example, by an
	// API gateway or HTTP middleware) and propagated through the call chain.
	// Producers SHOULD set this field whenever they can associate a log entry
	// with a specific incoming request.
	RequestID = "request_id"
)
