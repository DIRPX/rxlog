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

// Package handler provides built-in implementations of the error.Handler
// interface used by rxlog encoders to decide how to react when a field
// fails to encode.
//
// The parent error package defines the Handler interface and the contract
// for HandleError: given a key, value, and a non-nil error raised by an
// encoder, a Handler implementation returns either:
//
//   - a non-nil error, signaling that encoding MUST stop immediately and
//     that the returned error MUST be propagated to the caller, or
//
//   - nil, signaling that the failing field MUST be treated as skipped and
//     that encoding of the remaining fields MUST continue.
//
// This package contains concrete policies that implement that contract for
// common use cases. Each policy is implemented as a small, stateless type
// whose HandleError method is safe for concurrent use when the associated
// encoder is used concurrently.
//
// # Overview of provided handler
//
// The handler package currently provides four primary strategies:
//
//   - FailOnErrorHandler: propagate the encoding error and abort encoding.
//
//   - SkipOnErrorHandler: ignore the failing field silently and continue
//     encoding the remaining fields.
//
//   - LogOnErrorHandler: log the failure via a user-supplied logging
//     function and then continue encoding, skipping only the failing field.
//
//   - ReplaceOnErrorHandler: request that the encoder substitute a
//     placeholder value for the failing field instead of the original
//     value, while allowing encoding of the entry to continue.
//
// All handler share the same high-level behavior with respect to the
// Handler interface: they never attempt to repair encoder-internal state,
// they do not mutate encoder configuration, and they do not perform I/O
// directly against the encoder's sinks. Their sole responsibility is to
// decide whether encoding should stop and, optionally, to trigger side
// effects such as logging or replacement.
//
// # FailOnErrorHandler
//
// FailOnErrorHandler is the simplest policy: it treats any encoding failure
// as fatal for the current log entry. Its HandleError implementation:
//
//   - returns the original error value it receives, without modification;
//
//   - thereby instructs the encoder to abort encoding of the current entry
//     and propagate that error to the caller of the encoding operation;
//
//   - performs no logging or transformation of the error.
//
// This handler is appropriate in contexts where partial log entries are not
// acceptable and where encoder failures should be surfaced immediately as
// hard errors.
//
// # SkipOnErrorHandler
//
// SkipOnErrorHandler is the opposite extreme: it treats encoding failures
// as field-local issues that should not prevent the rest of the log entry
// from being written. Its HandleError implementation:
//
//   - always returns nil, regardless of the specific error;
//
//   - instructs the encoder to treat the failing field as skipped; and
//
//   - does not log, wrap, or otherwise report the error directly.
//
// This policy is suitable when the logging system must be robust in the
// face of occasional serialization failures and where losing a single
// attribute is preferable to dropping an entire entry or failing an
// operation.
//
// # LogOnErrorHandler
//
// LogOnErrorHandler allows applications to observe encoding failures
// without aborting the current entry. It wraps a user-supplied logging
// callback and uses it to record details about the failure.
//
// Its HandleError implementation:
//
//   - invokes the configured Logger callback, if non-nil, with the field
//     key, the value that failed to encode, and the error produced by the
//     encoder;
//
//   - returns nil, signaling that the failing field should be skipped and
//     that remaining fields should continue to be encoded.
//
// The logging callback is responsible for deciding how and where to record
// the error (for example, to a separate logger, metrics system, or debug
// sink). HandleError itself does not perform any I/O beyond the callback
// invocation and MUST remain safe for concurrent use.
//
// # ReplaceOnErrorHandler
//
// ReplaceOnErrorHandler is designed for situations where a field must be
// present in the encoded output, even if its original value cannot be
// serialized successfully. The handler carries a Placeholder value that
// represents the safe substitute for failing fields.
//
// Its HandleError implementation:
//
//   - always returns nil, indicating that encoding of the current entry
//     should continue;
//
//   - does not perform the replacement itself; instead, it relies on the
//     encoder to detect that the active handler is a ReplaceOnErrorHandler
//     (for example, via type assertion) and to use the Placeholder value
//     instead of the original value when encoding the failing field.
//
// This design keeps the handler free of direct knowledge about encoder
// internals while still allowing a clear "replace instead of skip" policy.
// Encoders that support replacement MUST document how they detect the
// ReplaceOnErrorHandler and how they interpret its Placeholder field.
//
// # Concurrency and side effects
//
// All handler in this package are designed to be safe for concurrent use
// by multiple goroutines when used with concurrent encoders. In particular:
//
//   - FailOnErrorHandler and SkipOnErrorHandler are stateless and immutable.
//
//   - LogOnErrorHandler holds a Logger callback that MUST itself be safe
//     for concurrent use if the handler is shared by multiple goroutines.
//
//   - ReplaceOnErrorHandler holds a Placeholder value that SHOULD be treated
//     as immutable once the handler is in use.
//
// None of the handler modify encoder-internal state, and none of them
// retain references to encoder-specific data beyond the duration of a
// HandleError call. Any logging or replacement behavior they trigger is
// explicitly exposed through their public fields and documented semantics.
//
// # Usage guidance
//
// The handler in this package are intended to be used as building blocks
// when configuring encoders and cores:
//
//   - Applications that prefer strict behavior can use FailOnErrorHandler
//     to ensure that any serialization bug becomes an immediately visible
//     error.
//
//   - Applications that prioritize robustness can use SkipOnErrorHandler,
//     optionally combined with LogOnErrorHandler for observability.
//
//   - Applications that need to preserve schema shape can use
//     ReplaceOnErrorHandler with an appropriate Placeholder value.
//
// Custom handler can be implemented by satisfying the error.Handler
// interface directly, using these built-ins as reference implementations
// for correct interaction with the encoding pipeline.
package handler
