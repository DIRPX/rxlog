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

// Package writer defines low-level abstractions for writing serialized log
// entries to output destinations.
//
// The central concept in this package is the WriteSyncer interface, which
// combines the byte-oriented streaming semantics of io.Writer with an
// explicit Sync method for flushing buffered data. Higher-level components
// in the logging pipeline (such as cores or encoders) depend on WriteSyncer
// to hand off completed log records to files, network connections, in-memory
// buffers, or other sinks without needing to know the details of each
// destination.
//
// # Overview
//
// A WriteSyncer represents a sink that accepts bytes and can be asked to
// flush any buffered state:
//
//   - Write(p []byte) (n int, err error) appends data to the destination,
//     following the standard io.Writer contract.
//
//   - Sync() error flushes buffered data to the underlying destination and
//     ensures, as far as the implementation is able, that previously written
//     bytes have been durably handed off to the operating system or transport.
//
// This design separates three concerns:
//
//   - encoding: how log entries are transformed into bytes,
//
//   - writing: where those bytes are sent (for example, files, sockets,
//     streams, or in-memory buffers), and
//
//   - durability: when callers can reasonably assume that previously written
//     bytes have left the process boundary or reached stable storage.
//
// # Write semantics
//
// The Write method of a WriteSyncer MUST satisfy the same basic contract as
// io.Writer:
//
//   - On success, it returns n == len(p) and a nil error, indicating that all
//     bytes in p were accepted by the sink.
//
//   - If it returns a non-nil error, it MAY also report a non-zero n to
//     indicate a partial write; callers MUST be prepared to handle this
//     possibility according to their needs.
//
// For logging use cases, WriteSyncer implementations SHOULD treat each call
// to Write as an atomic append of a complete serialized log entry (or
// fragment), and SHOULD avoid partial writes where possible. In particular:
//
//   - For file-backed sinks, Write SHOULD delegate to the underlying file
//     descriptor in a way that either writes the full buffer or fails.
//
//   - For buffered or batched sinks, implementations SHOULD treat the
//     provided slice as a single record to be enqueued or appended.
//
// Regardless of internal strategy, implementations MUST NOT modify the
// contents of the p slice after Write returns, and MUST NOT retain references
// to p beyond the duration of the call, unless explicitly and carefully
// documented.
//
// # Sync semantics
//
// The Sync method provides a best-effort mechanism for flushing buffered data
// and improving durability guarantees:
//
//   - For sinks backed by files or other storage with fsync-like semantics,
//     Sync SHOULD trigger an appropriate flush (for example, an fsync or
//     fdatasync call, or the closest available approximation).
//
//   - For sinks that do not have meaningful durability semantics (for example,
//     purely in-memory buffers or unbuffered streaming connections), Sync MAY
//     be a no-op. In such cases, the implementation SHOULD document that Sync
//     does not provide additional guarantees.
//
// On success, Sync MUST return nil. On failure, it MUST return a non-nil
// error describing the underlying problem so that callers can decide whether
// to retry, route logs elsewhere, degrade functionality, or surface the
// issue.
//
// Callers are responsible for deciding when to invoke Sync. Common patterns
// include:
//
//   - calling Sync during process shutdown,
//
//   - syncing periodically from a background goroutine, or
//
//   - relying on external rotation tools or buffering semantics and calling
//     Sync infrequently or not at all.
//
// # Concurrency model
//
// Unless explicitly documented otherwise, WriteSyncer implementations MUST be
// safe for concurrent use by multiple goroutines. This requirement exists
// because loggers are typically called from many goroutines concurrently and
// fan out encoded log entries to shared sinks.
//
// In particular:
//
//   - concurrent calls to Write and Sync on the same WriteSyncer MUST NOT
//     cause data races or violate the guarantees of the underlying transport;
//
//   - implementations SHOULD ensure that writes from different goroutines are
//     serialized in some well-defined order (for example, using internal
//     locking) or delegated to an underlying component that guarantees
//     correct serialization.
//
// # Usage context
//
// The writer package is intentionally small and self-contained. It does not
// perform encoding itself, and it does not know about higher-level logging
// concepts such as entries, fields, or cores. Instead, it provides:
//
//   - the WriteSyncer interface as a common contract for log sinks, and
//
//   - a foundation on which more complex writer implementations (including
//     adapters, fan-out writers, and wrappers around existing io.Writer
//     instances) can be built.
//
// In the broader rxlog pipeline:
//
//   - encoders construct serialized log entries into buffers,
//
//   - cores or other components hand those buffers to a configured
//     WriteSyncer, and
//
//   - application code decides how to configure and manage the underlying
//     destinations that each WriteSyncer represents.
//
// By centralizing sink semantics in this package, rxlog keeps the boundary
// between "encoding logs" and "delivering logs" explicit and composable.
package writer
