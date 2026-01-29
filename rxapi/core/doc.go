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

// Package core defines the core logging abstractions used by rxlog.
//
// At this layer, logging is expressed in terms of three closely related
// concepts:
//
//   - Entry: an immutable description of a single log event (level,
//     timestamp, logger name, message, caller, stack, structured fields,
//     and any other metadata that higher-level components attach).
//
//   - Core: the fundamental logging primitive that decides whether to log
//     an Entry, encodes it, and writes it to a configured sink.
//
//   - CheckedEntry: a pre-validated Entry paired with one or more Core
//     instances that have agreed to log it, used to amortize level and
//     predicate checks across multiple outputs.
//
// The core package does not define concrete loggers, encoders, or writers.
// Instead, it provides the contract that those components rely on when
// implementing logging behavior.
//
// # Entry semantics
//
// An Entry represents a single attempt to log an event. It is typically
// constructed from application-level inputs (such as a log level, message,
// and context-derived fields) and then passed into the logging pipeline.
// An Entry usually contains at least:
//
//   - the log Level (from the level package),
//
//   - a timestamp (time.Time),
//
//   - the logical logger name (from the name package),
//
//   - the message text,
//
//   - optional caller and stack information (from the caller and stack
//     packages), and
//
//   - a slice of field.Field values representing structured attributes.
//
// Entry values are treated as immutable once they are passed to Core.
// Implementations MUST NOT mutate an Entry they receive from callers;
// any Core that needs to adjust an Entry for its own purposes SHOULD work
// on a shallow copy instead.
//
// Entry is deliberately a plain struct rather than an interface so that
// it can be allocated and copied cheaply and used uniformly across all
// cores and hooks.
//
// # Core semantics
//
// Core is the central interface of the logging pipeline. Conceptually, a
// Core combines three responsibilities:
//
//   - level enabling: deciding whether an Entry at a given level is of
//     interest at all;
//
//   - entry construction: attaching additional fields or context that are
//     specific to this Core; and
//
//   - encoding and writing: serializing the Entry and fields and delivering
//     the resulting bytes to a sink.
//
// A typical Core interface has the following shape:
//
//	type Core interface {
//	    level.Enabler
//
//	    With(fields []field.Field) Core
//	    Check(entry Entry, fields []field.Field) *CheckedEntry
//	    Write(entry Entry, fields []field.Field) error
//	    Sync() error
//	}
//
// The methods have the following responsibilities:
//
//   - With returns a new Core that behaves like the receiver but has the
//     supplied fields attached to all subsequent log entries. The returned
//     Core MAY share internal state with the original as long as both are
//     safe for concurrent use by multiple goroutines.
//
//   - Check applies level and predicate checks to an Entry and its
//     associated fields. If this Core is not interested in the Entry,
//     Check returns nil. If it is interested, Check adds this Core to a
//     CheckedEntry and returns that CheckedEntry so that the caller can
//     invoke Write later. Check MUST NOT perform any encoding or I/O.
//
//   - Write encodes and writes the supplied Entry and fields to the Core's
//     configured sink. It is only called when Check has indicated that
//     the Entry is enabled for this Core. Write MAY be called multiple
//     times on the same Core concurrently; implementations MUST be safe
//     for concurrent use or explicitly documented otherwise.
//
//   - Sync flushes any buffered data and forwards corresponding Sync calls
//     to underlying sinks (for example, writer.WriteSyncer). It SHOULD be
//     idempotent and safe to call from multiple goroutines.
//
// Core implementations MUST embed or otherwise satisfy level.Enabler so
// that they can expose their level thresholds consistently. Higher-level
// components use this to perform fast pre-checks before constructing or
// encoding log entries.
//
// # CheckedEntry semantics
//
// CheckedEntry represents an Entry that has already passed initial level
// checks for one or more cores. It serves two purposes:
//
//   - It allows multiple cores to share the cost of level and predicate
//     evaluation: the logging pipeline can accumulate all cores that are
//     interested in an Entry into a single CheckedEntry and then invoke
//     Write on each of them exactly once.
//
//   - It provides a place to register hooks that should run if and when
//     the Entry is written, such as metric updates, counters, or side
//     effects that are associated with the logging action.
//
// A CheckedEntry typically contains:
//
//   - the original Entry;
//
//   - a slice of Core values that have indicated interest in the Entry via
//     their Check methods;
//
//   - optional hook metadata or callbacks that are invoked as part of the
//     write process.
//
// The usual lifecycle is:
//
//  1. A higher-level logging component constructs an Entry and a slice
//     of associated fields.
//
//  2. It calls Check on each Core. For each non-nil result, it merges the
//     returned CheckedEntry into a single aggregate CheckedEntry that
//     accumulates all interested cores.
//
//  3. If no Core is interested, the aggregate CheckedEntry is nil and the
//     log call becomes a no-op.
//
//  4. If one or more cores are interested, the caller invokes the Checked
//     method (or an equivalent operation) on the aggregate CheckedEntry,
//     which in turn calls Write on each underlying Core, running any
//     registered hooks before and/or after the writes.
//
// CheckedEntry implementations MUST ensure that each Core's Write method
// is invoked at most once per Entry and that hooks are executed exactly
// once in the intended order. They MUST be safe for concurrent use when
// the logging pipeline invokes CheckedEntry methods from multiple
// goroutines.
//
// # Concurrency and ownership
//
// Core values are intended to be shared across goroutines. Unless
// documented otherwise, implementations MUST be safe for concurrent use
// by multiple goroutines.
//
// In contrast, Entry and CheckedEntry values are typically short-lived and
// scoped to a single log operation. While a CheckedEntry may be shared for
// the duration of an individual log call, callers MUST NOT reuse a
// CheckedEntry for multiple distinct log operations.
//
// Core implementations that hold references to buffers, encoders, or
// writers MUST respect the ownership and lifetime rules defined by those
// packages. In particular, they MUST NOT retain or reuse buffers beyond
// their documented lifetimes and MUST coordinate access to shared writers
// using appropriate synchronization.
//
// # Usage context
//
// The core package defines the backbone of the rxlog logging pipeline.
// Higher-level loggers, hooks, and routing components build on these
// abstractions:
//
//   - Loggers construct Entry values and orchestrate calls to Core.Check,
//     Core.Write, and Core.Sync.
//
//   - Hook managers attach additional behavior before and after Core.Write
//     is invoked for a CheckedEntry.
//
//   - Composite cores fan out Entries to multiple underlying cores (for
//     example, different encoders or sinks) while sharing level checks via
//     CheckedEntry.
//
// By centralizing these responsibilities and contracts in one package,
// rxlog provides a stable, well-documented foundation for implementing
// custom logging behaviors, cores, and backends.
package core
