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

// Package httpctl provides HTTP control endpoints for the logging subsystem.
//
// It is intended for operational use: operators and automation can inspect
// and modify logging behavior of a running process over HTTP, without
// requiring process restarts or direct access to in-process configuration.
//
// # Overview
//
// The core type in this package is Handler. It binds a mutable log level
// threshold (level.Threshold) to a standard http.Handler, exposing a small,
// well-defined protocol for reading and updating the process-wide minimum
// log level.
//
// Typical usage:
//
//	import (
//	    "net/http"
//
//	    httpctl "dirpx.dev/rxlog/httpctl"
//	    "dirpx.dev/rxlog/rxapi/level"
//	    rxclvl "dirpx.dev/rxlog/rxcore/level"
//	)
//
//	func main() {
//	    // Create a mutable threshold (for example, an AtomicLevel).
//	    atomic := rxclvl.NewAtomicLevel()
//	    atomic.SetLevel(level.Info)
//
//	    // Bind it to an HTTP endpoint.
//	    handler := httpctl.NewHandler(&atomic)
//
//	    // Mount on an internal admin mux.
//	    http.Handle("/debug/log-level", handler)
//	    _ = http.ListenAndServe("127.0.0.1:6060", nil)
//	}
//
// With this setup, operators can query and adjust logging at runtime via
// HTTP requests to /debug/log-level.
//
// # Protocol
//
// Handler implements the following semantics:
//
//   - GET  /…  -> returns the current level as JSON:
//     { "<fields.Level>": "<level-name>" }.
//   - HEAD /…  -> same headers as GET, but with an empty body.
//   - POST /…  -> parses a new level and updates the threshold.
//   - PUT  /…  -> same as POST; provided for idempotent-style usage.
//   - PATCH /… -> same as POST; provided for convenience.
//
// For mutation methods (POST/PUT/PATCH), the desired log level is resolved
// in the following order of precedence:
//
//  1. Query parameter: ?<ParamName>=<level>, if non-empty.
//  2. JSON body (Content-Type: application/json):
//     { "<ParamName>": "<level>" }.
//  3. Plain-text body: entire body interpreted as a single level string.
//
// On success, mutation requests respond with HTTP 200 and a JSON payload
// carrying the effective level under the key defined by fields.Level:
//
//	{ "<fields.Level>": "<normalized-level>" }
//
// On client errors (for example, malformed JSON or an unknown level string),
// the handler responds with HTTP 400 and a JSON error payload using the key
// defined by fields.Error:
//
//	{ "<fields.Error>": "<human-readable-message>" }
//
// If the handler is misconfigured (for example, Handler.Level is nil) or
// an unexpected internal failure occurs, serveHTTP returns a non-nil error
// and the outer ServeHTTP wrapper responds with HTTP 500 and a simple
// text/plain diagnostic.
//
// # Configuration and options
//
// Handler instances are constructed via NewHandler. The constructor
// accepts a level.Threshold implementation (for example, *rxcore/level.AtomicLevel)
// and an optional list of functional options defined in this package:
//
//   - WithParamName(name string)
//     Overrides the query/body field name used when reading the desired
//     level. If omitted, DefaultLevelParamName (derived from fields.Level)
//     is used.
//
//   - WithMaxBodyBytes(n int64)
//     Sets an upper bound on the number of bytes read from the HTTP request
//     body when parsing a new level. If n <= 0, DefaultMaxBodyBytes is used.
//     This is a defensive limit to avoid unbounded reads.
//
//   - WithReadOnly(readOnly bool)
//     If set to true, disables mutation methods (POST/PUT/PATCH). In this
//     mode, the handler effectively becomes a read-only introspection
//     endpoint; mutation attempts receive HTTP 405 Method Not Allowed.
//
// All configuration fields of Handler are initialized by NewHandler
// and MAY be further adjusted by callers before the handler is exposed.
//
// # JSON schema and field names
//
// HTTP responses produced by this package use JSON keys derived from the
// global field-name definitions in dirpx.dev/rxlog/rxapi/field/fields, in
// particular:
//
//   - fields.Level for the current log level.
//   - fields.Error for error messages.
//
// This ensures that operational endpoints and log records share the same
// naming conventions and can be consumed by downstream tooling in a uniform
// way.
//
// # Concurrency and safety
//
// Handler is safe for concurrent use by multiple goroutines as long as
// the underlying level.Threshold implementation is itself safe for concurrent
// use.
//
// The handler:
//   - Does not maintain any mutable internal state beyond the provided
//     Threshold.
//   - Delegates all reads and updates to level.Threshold methods.
//   - Performs only local request/response processing without global side
//     effects beyond log-level changes.
//
// Callers SHOULD ensure that the bound Threshold is the same instance that
// their logger configuration uses; otherwise, changing the level via HTTP
// will not affect the active loggers.
//
// # Security considerations
//
// This package deliberately does not implement any authentication,
// authorization, or rate limiting. Exposing log-level control over HTTP is
// inherently a privileged operation.
//
// Callers MUST protect endpoints created with Handler appropriately,
// for example by:
//
//   - Binding only to an internal interface (e.g., 127.0.0.1).
//   - Mounting under an internal admin mux that is not reachable from the
//     public Internet.
//   - Placing the endpoint behind a reverse proxy or API gateway that
//     enforces authentication and access control.
//
// Failure to do so may allow untrusted parties to alter logging behavior,
// potentially suppressing important diagnostics or flooding downstream
// systems with excessive log volume.
package httpctl
