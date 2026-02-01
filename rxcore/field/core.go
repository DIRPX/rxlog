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
	// LogSchemaVersion is the field key used to record the version identifier
	// of the log schema in use.
	//
	// Producers SHOULD populate this field when the overall structure or
	// semantics of emitted logs may evolve over time, so that consumers can
	// distinguish between different schema revisions. The value stored under
	// this key is typically a version string (for example, "v1", "v2") or a
	// more detailed identifier, but the exact format is application-defined.
	LogSchemaVersion = "log_schema"

	// Timestamp is the field key that records the time at which the log entry
	// was created.
	//
	// Encoders SHOULD emit the timestamp in a well-defined, documented
	// format (for example, RFC3339/RFC3339Nano in UTC, or Unix epoch in
	// seconds/milliseconds). Consumers MUST NOT assume a particular format
	// beyond what the encoder documents.
	Timestamp = "ts"

	// Level is the field key that encodes the severity or verbosity of the
	// log entry.
	//
	// Typical values include textual representations such as "trace",
	// "debug", "info", "warn", and "error", but numeric encodings are also
	// possible. Encoders SHOULD document the mapping they use so that
	// downstream consumers can interpret levels consistently.
	Level = "level"

	// Message is the field key that carries the primary human-readable text
	// of the log entry.
	//
	// The value stored under this key SHOULD be short and descriptive, giving
	// a concise summary of the event. Additional structured context SHOULD be
	// placed in other fields rather than embedded into the message string.
	Message = "msg"

	// Logger is the field key that carries the logical logger name associated
	// with the entry.
	//
	// Typical values identify a subsystem or logical logger (for example,
	// "http.server", "sql", "cache"). Producers SHOULD populate this field
	// when they use multiple named loggers so that downstream consumers can
	// group and filter by logger identity.
	Logger = "logger"

	// LoggerName is an alias of Logger preserved for compatibility with older
	// code and schemas.
	//
	// New code SHOULD prefer Logger, but both constants resolve to the same
	// underlying key name ("logger").
	LoggerName = Logger

	// Component is the field key that identifies a higher-level part of the
	// service that emits the log entry.
	//
	// Examples include "ingress", "egress", "scheduler", or "worker_pool".
	// Producers MAY use this key to group logs by functional area or
	// subsystem within a single service.
	Component = "component"

	// Subsystem is the field key that identifies a more fine-grained part
	// inside a component.
	//
	// Examples include "jwt", "tls", "routing", or "storage". Producers MAY
	// use this field to provide additional structure within a component, but
	// they SHOULD keep the value space reasonably bounded to avoid excessive
	// cardinality.
	Subsystem = "subsystem"

	// Operation is the field key that describes the current operation or
	// action associated with the log entry.
	//
	// Typical values are aligned with handler names or business operations
	// such as "route", "issue_token", or "list_users". Producers SHOULD use
	// this field to capture the logical operation being performed rather than
	// low-level implementation details.
	Operation = "op"
	
	// EventName is the field key that records a more generic event or
	// activity name, especially when logs represent discrete domain events.
	//
	// Producers MAY use this when they treat logs as event streams with
	// named event types (for example, "user.created", "invoice.paid").
	EventName = "event"

	// EventID is the field key that carries a unique identifier for the
	// specific log event or domain event.
	//
	// Values are typically UUIDs or other globally unique identifiers.
	// Producers MAY use this to support idempotency, deduplication, or
	// cross-system joins.
	EventID = "event_id"

	// Category is the field key that groups related log entries under a
	// higher-level classification.
	//
	// Examples include "security", "audit", "performance", or "business".
	// This is a flexible categorization mechanism orthogonal to Level.
	Category = "category"
)
