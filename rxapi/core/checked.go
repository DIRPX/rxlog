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

import "dirpx.dev/rxlog/rxapi/writer"

// CheckedEntry represents a pre-filtered log entry that has been validated
// against one or more Cores and is ready to be written.
//
// CheckedEntry is intended as a fast-path optimization: callers first ask
// a logger to check whether an Entry SHOULD be logged, and only if a non-nil
// *CheckedEntry is returned do they proceed with the actual write. When the
// log level is disabled (or no Core is interested in the Entry), the logger
// MAY return a nil *CheckedEntry, making the decision a single pointer check.
//
// CheckedEntry instances are typically short-lived and MAY be pooled. Code
// that receives a *CheckedEntry MUST NOT retain it beyond the lifetime
// documented by the logger implementation.
type CheckedEntry struct {
	Entry

	// cores holds the set of Core instances that have indicated interest in
	// receiving this Entry.
	//
	// Each Core in this slice SHOULD have had its Check method invoked and
	// MUST have decided that the Entry SHOULD be logged. When the logger
	// eventually writes the CheckedEntry, it MUST dispatch the Entry (and any
	// associated fields) to each Core in this slice, typically by calling
	// Core.Write.
	//
	// This field is unexported; external callers MUST NOT rely on its presence
	// or layout. Logger internals that manipulate cores MUST ensure that the
	// slice is not modified concurrently.
	cores []Core

	// ErrorOutput is an optional sink used to report internal logging errors.
	//
	// The logging pipeline MAY use ErrorOutput to report:
	//   - unsafe reuse of a dirty CheckedEntry (double-Write scenarios),
	//   - errors returned by Core.Write implementations.
	//
	// When non-nil, ErrorOutput is expected to implement writer.WriteSyncer
	// interface.
	//
	// If ErrorOutput is nil, internal errors SHOULD be silently ignored to
	// avoid infinite recursion or log-spam loops.
	ErrorOutput writer.WriteSyncer

	// dirty tracks whether this CheckedEntry has already been consumed
	// (i.e., written to its associated cores) at least once.
	//
	// Logger internals MUST set dirty to true immediately before or after
	// performing the first write operation with this CheckedEntry. Any
	// subsequent attempt to write a dirty CheckedEntry MUST be treated as
	// misuse (a double-Write scenario).
	//
	// Implementations MAY enforce this contract by panicking in debug builds,
	// returning an error, or silently ignoring additional writes, but the
	// behavior SHOULD be documented. Callers external to the logger internals
	// MUST NOT rely on or modify dirty directly.
	dirty bool
}
