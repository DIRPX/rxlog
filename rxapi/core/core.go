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
	"dirpx.dev/rxlog/rxapi/field"
	"dirpx.dev/rxlog/rxapi/level"
)

// Core is the minimal, low-level logging backend interface.
//
// A Core combines two responsibilities:
//
//   - it decides whether a given Level is enabled (via the embedded
//     level.Enabler); and
//   - it encodes and writes log entries that pass that decision.
//
// Higher-level logger APIs SHOULD wrap Core to provide user-friendly logging
// functions. Core implementations MUST be safe for concurrent use by multiple
// goroutines unless explicitly documented otherwise.
type Core interface {
	level.Enabler

	// With returns a new Core that includes the provided structured context
	// fields on all subsequent log entries handled by that Core.
	//
	// The returned Core MUST preserve the behavior of the original Core and
	// MUST additionally attach the supplied fields to every Entry it writes.
	// Implementations SHOULD treat Cores as immutable: calling With SHOULD NOT
	// alter the receiver, and callers MAY continue to use both the original
	// and the derived Core independently.
	//
	// The fields slice MUST be treated as read-only for the duration of the
	// call; implementations MUST NOT retain or mutate the caller's slice
	// directly. If retention is required, implementations MUST copy the data.
	With([]field.Field) Core

	// Check determines whether the supplied Entry SHOULD be logged by this
	// Core, using the embedded level.Enabler and any additional internal logic.
	//
	// If the Entry SHOULD be logged, the Core MUST register itself with the
	// provided CheckedEntry (for example, by adding itself to an internal
	// list of Cores) and MUST return the resulting *CheckedEntry (which MAY be
	// the same pointer or a new one, depending on the implementation).
	// Otherwise, Check MUST return the original CheckedEntry unchanged, which
	// MAY be nil.
	//
	// Check MUST NOT perform the actual write; it MUST be a pure decision +
	// registration step. Callers MUST invoke Check before calling Write, and
	// MUST call Write only when Check has indicated that this Core should log
	// the Entry (typically via a non-nil, populated CheckedEntry).
	//
	// Implementations MUST NOT retain pointers to the Entry or fields beyond
	// the lifetime defined by the calling logger, since Entries and related
	// structures MAY be pooled.
	Check(Entry, *CheckedEntry) *CheckedEntry

	// Write serializes the Entry and any Fields supplied at the log site and
	// writes them to the Core's destination.
	//
	// When Write is called, the Core MUST attempt to log the Entry and Fields
	// unconditionally; it MUST NOT re-apply level checks or replicate the
	// logic of Check. Any filtering decisions MUST be made before Write is
	// invoked.
	//
	// The fields slice MUST be treated as read-only for the duration of the
	// call. Implementations MUST NOT retain or mutate the slice or its
	// elements after Write returns. If retention is required (for buffering or
	// asynchronous I/O), implementations MUST copy the relevant data.
	//
	// Write SHOULD return a non-nil error if it fails to encode or deliver the
	// log entry. Callers MAY choose to ignore such errors, but logging systems
	// that care about delivery guarantees SHOULD surface them.
	Write(Entry, []field.Field) error

	// Sync flushes any buffered log entries (if the Core performs buffering)
	// to their final destination.
	//
	// Implementations that do not perform buffering MAY implement Sync as a
	// no-op and SHOULD return nil. Implementations that buffer or batch writes
	// MUST ensure that Sync attempts to flush all pending records and SHOULD
	// return a non-nil error if the flush operation fails.
	//
	// Callers MAY invoke Sync during shutdown or at defined checkpoints to
	// improve durability of log records.
	Sync() error
}
