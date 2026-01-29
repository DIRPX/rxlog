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

// Package hook defines extensibility points that allow applications and
// infrastructure code to observe and influence the logging pipeline.
//
// Hooks are small pieces of logic that are invoked at well-defined moments
// in the lifecycle of a log entry. Typical use cases include:
//
//   - enriching entries with additional fields,
//   - redacting or transforming sensitive data,
//   - updating metrics or counters,
//   - routing entries to auxiliary systems, and
//   - implementing sampling or auditing policies.
//
// The package provides three main abstractions:
//
//   - Hook: a common marker for hook implementations and shared metadata,
//
//   - PreWriteHook: a hook invoked before an entry is encoded and written,
//     which may modify or veto the entry, and
//
//   - PostWriteHook: a hook invoked after an entry has been written (or a
//     write has failed), which may observe the outcome and perform
//     side-effects such as metrics or alerts.
//
// In addition, a hook manager API coordinates registration and execution
// of hooks, including priority ordering and stage dispatch.
//
// # Pre-write hooks
//
// Pre-write hooks run after a log Entry and its associated fields have been
// constructed, but before any encoding or I/O is performed. They are the
// first opportunity for application code to:
//
//   - attach additional fields based on the entry contents,
//
//   - normalize or redact structured data,
//
//   - enforce policy (for example, deny logging of certain events), or
//
//   - decide that a particular event should be dropped entirely.
//
// Conceptually, a pre-write hook is a function that receives the Entry and
// its fields and returns possibly modified values and an error. The exact
// method signature is defined by the PreWriteHook interface, but all
// implementations MUST obey the following rules:
//
//   - Pre-write hooks MAY mutate the Entry and field slice they receive, or
//     they MAY construct and return new values. The exact mutation semantics
//     SHOULD be documented by each implementation.
//
//   - A non-nil error return indicates that logging SHOULD be aborted for
//     this entry. The hook manager MUST treat such an error as a veto for
//     the current entry and MUST NOT proceed to encode or write it.
//
//   - A nil error return indicates that logging MAY proceed, subject to
//     subsequent hooks and core-level checks.
//
//   - Pre-write hooks MUST be safe for concurrent use by multiple goroutines
//     when shared, since they are typically attached to cores or loggers
//     that are used concurrently.
//
//   - Pre-write hooks SHOULD avoid expensive or blocking operations where
//     possible, as they run on the critical path before any encoding or I/O.
//
// # Post-write hooks
//
// Post-write hooks run after the logging pipeline has attempted to write an
// entry to its sink. They receive the original Entry and fields (as seen at
// write time) along with the result of the write operation (for example, a
// nil error on success or a non-nil error on failure).
//
// Post-write hooks are intended for observation and side effects:
//
//   - updating metrics (counts of written entries, failures, retries),
//
//   - forwarding a copy of selected entries to additional systems,
//
//   - triggering alerts in response to write failures, and
//
//   - recording diagnostics for later analysis.
//
// Implementations of PostWriteHook MUST treat the Entry and fields they
// receive as read-only; they MUST NOT attempt to modify values that are
// still in use by encoders or cores. If a hook needs to retain data beyond
// the duration of the call, it MUST make its own copy.
//
// A post-write hook MUST NOT attempt to re-enter the same logging pipeline
// recursively with the same hook configuration; doing so can easily lead to
// infinite recursion or cycles. If a post-write hook needs to log its own
// diagnostics, it SHOULD use a separate logger instance or a configuration
// that disables the hook for those internal logs.
//
// # Hook metadata and priorities
//
// The hook package allows hooks to carry metadata such as names and
// priorities. A typical arrangement is:
//
//   - each hook has a stable, human-readable name used for diagnostics and
//     configuration; and
//
//   - each hook is registered with an integer priority that determines its
//     relative order of execution within a given stage.
//
// Priority ordering follows these rules:
//
//   - Hooks with lower numeric priority values run before hooks with higher
//     numeric priority values.
//
//   - Hooks that share the same numeric priority value are executed in a
//     stable order determined by registration; implementations SHOULD
//     preserve registration order for determinism.
//
//   - Priority comparisons are local to a particular stage; pre-write and
//     post-write hooks are ordered independently.
//
// # Managers and execution
//
// The hook package provides a manager abstraction responsible for
// registering hooks and executing them at the appropriate stage. While the
// exact API surface may vary, a manager typically exposes operations such
// as:
//
//   - adding or removing pre-write and post-write hooks with associated
//     priorities, and
//
//   - executing all hooks for a particular stage in priority order when
//     driven by the logging pipeline.
//
// Managers MUST enforce the following invariants:
//
//   - They MUST be safe for concurrent use by multiple goroutines.
//
//   - They MUST respect priority ordering and preserve relative order for
//     hooks with equal priority values.
//
//   - For pre-write hooks, they MUST stop execution as soon as a hook
//     returns a non-nil error and MUST propagate that error to the caller,
//     causing the entry to be dropped.
//
//   - For post-write hooks, they SHOULD attempt to run all registered hooks
//     even if some hooks observe or report errors, so that a failure in one
//     observer does not prevent others from running. Any errors returned by
//     post-write hooks SHOULD be aggregated or reported through a mechanism
//     defined by higher-level code.
//
// # Usage context
//
// The hook package is designed to be used by cores, logger implementations,
// and infrastructure components that need to extend logging behavior:
//
//   - Cores invoke pre-write hooks after constructing an Entry and before
//     encoding or writing it.
//
//   - After Write completes (successfully or with an error), cores invoke
//     post-write hooks so that observers can react to the outcome.
//
//   - Applications configure which hooks are installed (and with what
//     priorities) when wiring their logging stack.
//
// The package does not perform encoding or I/O itself and does not depend on
// concrete backends. It defines the contracts that hook implementations and
// managers MUST follow so that the rest of the logging pipeline can remain
// agnostic to the specific behaviors attached via hooks.
package hook
