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

package writer

import "io"

// WriteSyncer abstracts a writeable log destination that also supports an
// explicit flush operation.
//
// This interface combines io.Writer with a Sync method, which allows callers
// to force any buffered data to be written to the underlying storage (for
// example, an OS file, network socket, or custom buffer).
//
// Implementations SHOULD be safe for concurrent use by multiple goroutines,
// since loggers typically share a small set of outputs across many call sites.
// Any internal buffering MUST honor the Sync contract described below.
type WriteSyncer interface {
	io.Writer

	// Sync flushes any buffered data to the underlying destination and
	// ensures that previously written bytes are durably handed off to the
	// operating system or transport.
	//
	// For implementations backed by files, Sync SHOULD correspond to
	// fsync-like semantics (or the closest available approximation). For
	// purely in-memory or streaming destinations, Sync MAY be a no-op, but
	// this behavior SHOULD be documented.
	//
	// On success, Sync MUST return nil. On failure, it MUST return a
	// non-nil error describing the underlying problem so that callers can
	// decide whether to retry, route logs elsewhere, or surface the issue.
	Sync() error
}
