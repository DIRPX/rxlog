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

	"dirpx.dev/rxlog/rxapi/core"
	"dirpx.dev/rxlog/rxapi/field"
)

// Manager coordinates all registered logging hooks and controls their
// execution order and lifecycle.
//
// A Manager keeps track of pre-write, post-write, and error hooks and is
// responsible for invoking them at the appropriate points in the logging
// pipeline. Implementations MAY support both synchronous and asynchronous
// execution for hooks that implement AsyncHook.
//
// Managers used in loggers that are shared across goroutines MUST be safe for
// concurrent use.
type Manager interface {
	// AddPreWrite registers a PreWriteHook with the given priority.
	//
	// Hooks with lower numeric priority values MUST be executed before hooks
	// with higher values, allowing callers to control ordering explicitly
	// (for example, normalization hooks before audit hooks). If multiple
	// hooks share the same priority, their relative ordering is
	// implementation-defined but SHOULD be stable within a single process.
	AddPreWrite(hook PreWriteHook, priority int)

	// AddPostWrite registers a PostWriteHook with the given priority.
	//
	// Post-write hooks are invoked after the rxcore has attempted to write an
	// entry. As with pre-write hooks, lower numeric priority values MUST be
	// executed earlier than higher values. Implementations MAY execute hooks
	// that report Async() == true in the background, provided that such hooks
	// are safe for concurrent execution.
	AddPostWrite(hook PostWriteHook, priority int)

	// AddError registers an ErrorHook.
	//
	// Error hooks are invoked when the rxcore reports a write failure. The
	// Manager MAY maintain multiple error hooks; all registered hooks MUST
	// be invoked for each error, subject to any asynchronous execution
	// policy implemented by the Manager.
	AddError(hook ErrorHook)

	// Remove unregisters any hook with the given name.
	//
	// The name MUST correspond to the value returned by Hook.Name(). If
	// multiple hooks share the same name, the Manager MAY remove all of them
	// or only the first match; this behavior SHOULD be documented by the
	// implementation. If no hook with the given name is found, Remove MUST
	// be a no-op.
	Remove(name string)

	// FirePreWrite invokes all registered PreWriteHook instances for the
	// given entry and fields.
	//
	// Hooks MUST be invoked in ascending priority order. Each hook MAY return
	// modified versions of the entry and fields, and those modifications MUST
	// be passed as input to subsequent hooks. If any hook returns a non-nil
	// error, the Manager MUST stop invoking further pre-write hooks and MUST
	// return that error to the caller. In this case, the caller MUST treat
	// the write as cancelled and MUST NOT pass the entry to the rxcore.
	//
	// The returned entry and field slice represent the final, possibly
	// transformed values that SHOULD be used for encoding and writing when
	// no error is returned.
	FirePreWrite(context.Context, core.Entry, []field.Field) (core.Entry, []field.Field, error)

	// FirePostWrite invokes all registered PostWriteHook instances for the
	// given entry, fields, and write result.
	//
	// Hooks MUST observe, but MUST NOT affect, the outcome of the write
	// operation. Any errors returned by individual hooks MUST NOT be
	// propagated to the caller of FirePostWrite; they MAY be logged or
	// otherwise handled internally by the Manager. Implementations MAY choose
	// to execute hooks that advertise Async() == true in the background, but
	// synchronous hooks MUST complete before FirePostWrite returns.
	FirePostWrite(ctx context.Context, entry core.Entry, fields []field.Field, written int, err error)

	// FireError invokes all registered ErrorHook instances for the given
	// entry and write error.
	//
	// This method is called when the rxcore reports a non-nil error from
	// Write. All registered error hooks MUST be invoked, regardless of each
	// other's behavior. Any errors produced by the hooks themselves MUST NOT
	// affect the original write result and SHOULD be handled internally
	// (for example, logged to a fallback logger).
	FireError(ctx context.Context, entry core.Entry, fields []field.Field, err error)

	// Stop waits for all asynchronous hook activity to complete and prevents
	// new asynchronous work from being scheduled.
	//
	// Implementations that support asynchronous hooks MUST ensure that Stop
	// blocks until all in-flight asynchronous hook invocations have
	// finished. After Stop returns, no further asynchronous execution MUST
	// occur. Callers SHOULD invoke Stop during shutdown to ensure that
	// background hook work (such as metrics, forwarding, or external calls)
	// has completed before process exit.
	Stop()
}
