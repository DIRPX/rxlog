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

package core

import (
	"time"

	"dirpx.dev/rxlog/api/caller"
	"dirpx.dev/rxlog/api/level"
)

// Entry represents a complete, structured log event ready to be encoded.
//
// An Entry aggregates the rxcore attributes of a log record: severity Level,
// timestamp, logger name, human-readable message, call-site metadata, and an
// optional stack trace. The structured fields associated with the record MAY
// already have been serialized separately, depending on the logging pipeline.
//
// Entry values are typically pooled. Any function that accepts an *Entry
// MUST NOT retain references to it or any of its fields beyond the scope in
// which it is documented as valid. Callers that need to keep data from an
// Entry MUST copy the relevant values (for example, strings) before the
// Entry is returned to the pool.
type Entry struct {
	// Level is the severity level associated with this log event.
	//
	// Level MUST be a valid level.Level value as defined by the logging
	// subsystem. Consumers MAY use Level for filtering, routing, or mapping to
	// backend-specific severity notions. An implementation MUST NOT emit an
	// Entry with an undefined or invalid Level.
	Level level.Level

	// Time is the timestamp associated with this log event.
	//
	// Time SHOULD represent the moment when the Entry was created or when the
	// corresponding event was observed. Implementations SHOULD use a
	// monotonic-aware time source where applicable, but consumers MUST treat
	// Time as a wall-clock instant for display and storage purposes.
	//
	// If Time is the zero value, encoders MAY choose to omit it or replace it
	// with the current time, but such behavior SHOULD be documented.
	Time time.Time

	// LoggerName identifies the logical logger or subsystem that produced
	// this Entry.
	//
	// LoggerName MAY be empty, in which case consumers SHOULD treat the Entry
	// as originating from a default or unnamed logger. When non-empty, it
	// SHOULD be a stable, human-readable identifier and MAY use a hierarchical
	// naming scheme (for example, "service.api.auth").
	LoggerName string

	// Message is the primary human-readable message associated with this log
	// event.
	//
	// Message SHOULD provide a concise textual description of the event and
	// SHOULD be understandable without requiring access to additional fields,
	// although structured context MAY supply further detail. Implementations
	// SHOULD avoid storing large blobs or highly structured data in Message;
	// such data SHOULD be placed in structured fields instead.
	//
	// Message MAY be empty for purely structured logs, but consumers SHOULD be
	// prepared to display an Entry even when Message is blank.
	Message string

	// Caller captures call-site information (such as file, line, and function)
	// for the origin of this log event.
	//
	// When Caller.Defined is true, consumers MAY use it for source navigation,
	// correlation with stack traces, or filtering by location. When
	// Caller.Defined is false, all other Caller fields MUST be treated as
	// unspecified and MUST NOT be relied upon.
	Caller caller.Caller

	// Stack contains an OPTIONAL stack trace captured at or around the time
	// the log event was created.
	//
	// Stack MAY be empty if no stack trace was captured for this Entry. When
	// non-empty, its format is implementation-defined (for example, a
	// newline-separated list of frames) and SHOULD be documented by the
	// runtime package. Encoders MAY choose to omit Stack for lower-severity
	// levels and include it only for Error, Critical, or Fatal events.
	Stack string
}
