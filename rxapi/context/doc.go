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

// Package context groups abstractions for integrating Go's context.Context
// with the rxlog logging API.
//
// The rxlog project treats context.Context as the primary carrier for
// request-scoped, operation-scoped, and environment-scoped metadata such as
// trace identifiers, user or tenant information, and feature flags. This
// package provides a dedicated namespace for code that knows how to:
//
//   - interpret context.Context values in a logging-friendly way, and
//   - turn that interpretation into structured attributes that can be
//     attached to log entries.
//
// The root context package is intentionally minimal. It primarily serves as a
// container for sub-packages that implement specific context-related
// behaviors (for example, extracting structured fields from a context).
// These sub-packages define the concrete interfaces, adapters, and helpers;
// the root package describes the overall design goals and usage patterns.
//
// # Design goals
//
// The context namespace is designed around the following principles:
//
//   - Separation of concerns: obtaining metadata from context.Context and
//     attaching it to log entries are separate responsibilities. Code that
//     understands a particular context convention (such as framework- or
//     middleware-specific keys) can live in focused sub-packages, while the
//     rest of the logging pipeline remains agnostic to those details.
//
//   - Explicit contracts: context-related helpers expose narrow, well-
//     documented interfaces (for example, functions or interfaces that
//     convert context.Context into structured fields). This makes it clear
//     which parts of the system are responsible for interpreting context and
//     how their behavior can be tested and overridden.
//
//   - Testability: by centralizing context interpretation behind explicit
//     abstractions, applications can substitute test contexts, custom
//     extractors, or environment-specific adapters without changing the core
//     logging API.
//
//   - Non-intrusiveness: code in this namespace does not mandate any
//     particular scheme for storing values in context.Context. Instead, it
//     works with whatever conventions the application or its frameworks
//     already use, as long as those conventions can be expressed through the
//     provided abstractions.
//
// # Relationship to the rest of rxlog
//
// The context package does not define loggers, cores, encoders, or writers.
// It does not perform I/O and does not depend on specific logging backends.
// Instead, it provides a bridge between:
//
//   - the ephemeral, in-process metadata carried by context.Context, and
//   - the structured, serialized attributes carried by rxlog's field and
//     encoder APIs.
//
// Typical usage within the logging pipeline involves a component that:
//
//  1. receives a context.Context from application code,
//  2. delegates to one or more context-related helpers from this namespace
//     to derive structured attributes, and
//  3. attaches those attributes to the log entry alongside explicit fields
//     provided by the caller.
//
// # Importing and naming
//
// Because Go already defines a standard library package named context,
// applications that import this rxlog package typically use an alias to avoid
// confusion and name collisions, for example:
//
//	import (
//	    stdctx "context"
//	    rxctx "dirpx.dev/rxlog/rxapi/context"
//	)
//
// This documentation uses the unqualified name "context" to refer to this
// package and explicitly mentions context.Context when referring to the
// standard library type.
//
// In summary, the context package defines the conceptual "context bridge"
// for rxlog: it gathers the abstractions and helpers that know how to turn
// context.Context values into structured logging data, while remaining
// decoupled from specific logging backends and application frameworks.
package context
