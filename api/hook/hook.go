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

package hook

import (
	"context"

	"dirpx.dev/rxlog/api/core"
	"dirpx.dev/rxlog/api/field"
	"dirpx.dev/rxlog/api/name"
)

// Hook is the base interface for all hook implementations used in the logging
// pipeline.
//
// A Hook represents a named, pluggable component that can be attached to a
// logger or encoder to observe, modify, or augment log entries at various
// stages (for example, before encoding, before writing, or after errors).
//
// By embedding name.Namer, every Hook MUST provide a stable, human-readable
// name via Name(). Callers MAY use this name for configuration, diagnostics,
// or metrics (for example, enabling or disabling specific hooks by name).
// Implementations SHOULD ensure that Name() is deterministic for a given
// instance and SHOULD document any naming conventions they rely on.
type Hook interface {
	name.Namer
}

// PreWriteHook is invoked immediately before a log entry is encoded and
// written by the logging rxcore.
//
// A PreWriteHook MAY inspect and modify both the Entry and its associated
// fields, or it MAY veto the write entirely by returning a non-nil error.
// Implementations SHOULD be side-effect free with respect to global state,
// except for intentional logging, metrics, or tracing side effects.
//
// Hooks that are attached to loggers used from multiple goroutines MUST be
// safe for concurrent use.
type PreWriteHook interface {
	Hook

	// PreWrite is called before encoding and writing a log entry.
	//
	// The method receives the original Entry and its current set of fields.
	// It MAY return a modified Entry (for example, changing the level or
	// message) and/or a modified slice of fields (for example, adding,
	// removing, or rewriting fields).
	//
	// If PreWrite returns a non-nil error, the write MUST be cancelled and
	// the log entry MUST NOT be encoded or emitted by the caller. The caller
	// MAY choose to log or surface that error separately, but the original
	// entry is considered vetoed.
	//
	// If PreWrite returns nil error, the caller MUST use the returned Entry
	// and field slice (which MAY be the same values that were passed in) for
	// subsequent encoding and writing.
	PreWrite(context.Context, core.Entry, []field.Field) (core.Entry, []field.Field, error)
}

// PostWriteHook is invoked after the logging rxcore has attempted to write
// a log entry to its underlying output(s).
//
// A PostWriteHook observes the final entry, its fields, the write result,
// and MAY perform side effects such as metrics emission, additional logging,
// tracing, or forwarding the record to secondary destinations. It MUST NOT
// attempt to modify or re-write the original log entry: by the time this hook
// is called, the write has already been attempted.
//
// Hooks attached to loggers that are used from multiple goroutines MUST be
// safe for concurrent use.
type PostWriteHook interface {
	Hook

	// PostWrite is called after the write attempt for a given log entry has
	// completed, regardless of whether it succeeded or failed.
	//
	// The ctx argument carries the ambient context associated with the log
	// call. Implementations MUST treat ctx as read-only and MUST NOT store it
	// beyond the lifetime of the call.
	//
	// The entry argument represents the final Entry that was passed to the
	// rxcore for writing, including all applied transformations (for example,
	// level or message adjustments made by hooks).
	//
	// The fields slice contains the final set of structured fields associated
	// with the entry at the time of the write attempt. Implementations MUST
	// treat this slice and its elements as read-only and MUST NOT retain or
	// mutate it after the call returns.
	//
	// The written argument is the number of bytes that the rxcore reports as
	// successfully written for this entry. The exact semantics of this value
	// (for example, whether it is per-sink, aggregated across sinks, or
	// best-effort in the presence of buffering) are defined by the rxcore
	// implementation and SHOULD be documented there. When err is non-nil,
	// written MAY be zero or MAY represent a partial write.
	//
	// The err argument is the error returned by the rxcore's write operation,
	// or nil if the write was reported as successful. PostWriteHook is
	// intended to observe this result; it MUST NOT attempt to retroactively
	// cancel, replay, or roll back the write.
	//
	// The error returned from PostWrite MUST NOT affect the outcome of the
	// already-attempted write: the rxcore MUST treat the write result as final
	// regardless of the hook's return value. Callers MAY choose to log or
	// otherwise surface a non-nil error from PostWrite for diagnostic
	// purposes, but they MUST NOT attempt to suppress or re-emit the original
	// log entry solely based on this return value.
	PostWrite(ctx context.Context, entry core.Entry, fields []field.Field, written int, err error) error
}

// ErrorHook is invoked when the logging rxcore reports a failure while writing
// a log entry to its underlying output(s).
//
// An ErrorHook observes write failures from Core.Write and MAY perform
// side effects such as emitting metrics, forwarding the error to a separate
// logger, or triggering alerts. It MUST NOT assume that the original write
// can be retried successfully by the caller, and it MUST NOT attempt to roll
// back any partial effects that the rxcore may already have produced.
//
// Hooks attached to loggers that are used from multiple goroutines MUST be
// safe for concurrent use.
type ErrorHook interface {
	Hook

	// OnError is called when Core.Write returns a non-nil error for a given
	// log entry.
	//
	// The ctx argument carries the ambient context associated with the log
	// call. Implementations MUST treat ctx as read-only and MUST NOT retain
	// it beyond the duration of the call.
	//
	// The entry argument is the Entry that was passed to the rxcore for writing,
	// including any transformations applied before the write attempt (for
	// example, changes made by PreWrite hooks).
	//
	// The fields slice contains the final set of structured fields associated
	// with the entry at the time of the failed write. Implementations MUST
	// treat this slice and its elements as read-only and MUST NOT retain or
	// mutate it after the call returns.
	//
	// The err argument is the error returned by Core.Write and MUST be
	// non-nil. OnError MUST treat this error as informative: it MAY log it,
	// emit metrics, or trigger alerts, but it MUST NOT assume that the write
	// can or will be retried.
	//
	// OnError MUST be safe for concurrent use when the associated logger or
	// rxcore is used concurrently.
	OnError(ctx context.Context, entry core.Entry, fields []field.Field, err error)
}

// AsyncHook marks a hook as eligible for asynchronous execution.
//
// This interface is typically implemented by hook types (such as PreWriteHook,
// PostWriteHook, or ErrorHook) that may be run off the main logging fast path.
// When a hook reports that it is asynchronous, the logging implementation MAY
// execute it in the background (for example, in a separate goroutine or worker)
// so that it does not block the caller that emits the log entry.
//
// Implementations that return true from Async MUST be fully safe for
// concurrent use and MUST tolerate being invoked after the originating log
// call has already completed. They MUST NOT assume ordering guarantees relative
// to other asynchronous hooks or to the log write itself.
type AsyncHook interface {
	// Async reports whether this hook prefers to run asynchronously.
	//
	// When Async returns true, the logging implementation MAY execute the hook
	// in the background and MUST NOT let failures in the hook block or delay
	// the main logging path. Asynchronous hooks MUST be written defensively:
	// they SHOULD assume that the log entry has already been committed and
	// that timing, ordering, and delivery are best-effort.
	//
	// When Async returns false, the logging implementation SHOULD execute the
	// hook synchronously in the logging call path, before returning control to
	// the caller, subject to the semantics of the specific hook type.
	Async() bool
}
