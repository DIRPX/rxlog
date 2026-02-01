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

// Package fields defines canonical field key names used by rxcore and
// higher layers of the logging stack.
//
// The constants in this package provide a curated vocabulary of attribute
// keys for common logging concerns. They are intended to be reused across
// services and components so that structured logs share a consistent,
// documented schema instead of relying on ad-hoc key naming.
//
// Most keys follow snake_case naming and are loosely aligned with common
// conventions from modern observability ecosystems (for example,
// OpenTelemetry, popular logging schemas), but the exact semantics are
// defined by the comments in each file.
//
// # Design goals
//
//   - Consistency:
//     By centralizing key names in this package, different components of a
//     system can agree on a shared schema for structured logs, reducing
//     drift such as "svc", "serviceName", and "service_name" being used
//     inconsistently for the same concept.
//
//   - Interoperability:
//     Downstream consumers (log routers, search backends, dashboards) can
//     be configured once for these canonical keys and then reused across
//     services and binaries that adopt the same vocabulary.
//
//   - Clarity:
//     The chosen names are intended to be explicit, readable ASCII
//     identifiers (snake_case), making logs easier to inspect directly as
//     well as to process programmatically.
//
//   - Separation of concerns:
//     Field keys are defined at the rxcore layer and may be referenced by
//     encoders, handlers, and higher-level APIs, but they are independent
//     of any particular logging backend or wire format.
//
// # Usage
//
// A typical usage pattern is:
//
//	// Construct a field using a canonical key.
//	enc.AddString(fields.Service, "auth")
//	enc.AddTime(fields.Timestamp, now)
//	enc.AddString(fields.HTTPMethod, r.Method)
//	enc.AddString(fields.AuthSubject, subject) // see auth.go
//
// Keys from this package are plain strings and can be used wherever a
// structured log key is required (for example, as Field.Key values in the
// rxapi/field package). Applications are free to introduce additional keys
// for domain-specific attributes, but reusing identifiers from this package
// where applicable is strongly recommended for consistency across the
// codebase.
//
// # Package organization
//
// The package is split into multiple source files, each covering a specific
// concern or domain. This improves discoverability and keeps individual
// files focused and maintainable.
//
// Core / generic schema:
//
//   - core.go
//     Core log schema and generic event metadata: schema version,
//     timestamp, level, message, logger name, component/subsystem,
//     operation/op, generic event name/ID, category and other broadly
//     applicable fields.
//
//   - base.go
//     Convenience and commonly used aliases or minimal shared subsets
//     that are expected to be referenced frequently across the codebase.
//
// Error, caller, and filesystem:
//
//   - error.go
//     Error, error_message, error_type, error_code, stacktrace, and
//     related exception metadata.
//
//   - caller.go
//     Caller/location metadata: combined caller field, file, line,
//     function name, or other source-level context for log sites.
//
//   - fs.go
//     File system–related fields such as file_path, file_name and
//     similar path-oriented attributes.
//
// Service, environment, and platform metadata:
//
//   - service_meta.go
//     Service identity and deployment metadata: service name and
//     namespace, version/build identifiers, environment (env), region,
//     zone, cluster, node/host, instance ID, process/runtime information.
//
//   - cloud.go
//     Cloud and orchestration metadata: cloud provider/region/zone,
//     Kubernetes cluster/namespace/pod/container/node and similar
//     platform-level attributes.
//
// Identity, auth, and security:
//
//   - identity.go
//     Correlation and identity fields: correlation_id, trace_id, span_id,
//     parent_span_id, request_id, session_id, user_id, tenant_id,
//     organization/customer identifiers, device_id, IP and related
//     contextual IDs.
//
//   - auth.go
//     Authentication and authorization metadata:
//     auth_subject, auth_scope, auth_role, auth_method, auth_token_id,
//     and closely related fields that describe who is authenticated and
//     under which privileges. :contentReference[oaicite:0]{index=0}
//
// Networking and protocols:
//
//   - http.go
//     HTTP request/response metadata: method, URL/path/route, host,
//     scheme, protocol, status code, user-agent, referer, remote/client
//     addresses, request/response sizes and related HTTP attributes.
//
//   - grpc.go
//     gRPC-specific metadata: gRPC service/method names and canonical
//     status codes for RPC results.
//
// Data access and messaging:
//
//   - db.go
//     Database and storage metadata: database system/name/user,
//     statements/operations, execution duration, rows affected, and
//     related low-level DB attributes.
//
//   - messaging.go
//     Messaging / queue / streaming metadata: system (for example,
//     Kafka, RabbitMQ), destination/topic/queue, destination kind,
//     partition, offset, message key and similar broker-centric fields.
//
// Metrics, telemetry, and domain-specific:
//
//   - metric.go
//     Metric-like fields and performance measurements: duration, latency,
//     attempt/retry counters, effective sampling rate, sampling decision
//     and other timing or sampling-related attributes.
//
//   - telemetry.go
//     Telemetry-oriented keys that tie logs into broader observability
//     signals (for example, fields that bridge logs with metrics/traces
//     beyond the raw trace/span identifiers).
//
//   - payment.go
//     Generic payment and commerce-related identifiers such as order_id,
//     payment_id and similar business keys that tend to recur across
//     services in payment-oriented domains.
//
// # Stability and evolution
//
// The semantics of existing keys in this package are intended to be stable,
// but the set of keys MAY grow over time as new common use cases emerge.
// Adding new constants is considered a backward-compatible change. Removing
// or renaming existing constants would be a breaking change and SHOULD be
// avoided once the package is in wide use.
//
// Applications SHOULD treat these keys as a shared contract for structured
// logging across the rxlog ecosystem and SHOULD prefer extending the schema
// in a consistent manner rather than redefining existing concepts under
// different names.
package field
