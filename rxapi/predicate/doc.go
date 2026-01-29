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

// Package predicate defines composable boolean conditions used to decide
// whether a log entry should be recorded.
//
// While level-based filtering answers the question "is this severity enabled
// at all?", predicates provide a more general mechanism for answering
// "should this particular entry, with these fields, be logged?". They allow
// applications and infrastructure code to express routing, sampling, and
// filtering policies as reusable boolean conditions.
//
// # Overview
//
// The package is built around three related concepts:
//
//   - Predicate: an interface representing a yes/no decision based on the
//     properties of a log entry (and possibly its fields or other context).
//
//   - Func: an adapter that allows ordinary functions with the appropriate
//     signature to be used wherever a Predicate is expected.
//
//   - Composition helpers (such as And, Or, Not) for combining multiple
//     predicates into more complex conditions with short-circuit semantics.
//
// Together, these abstractions make it possible to build sophisticated
// filtering logic while keeping the core logging pipeline agnostic to the
// specifics of those policies.
//
// # Predicate semantics
//
// A Predicate conceptually evaluates a log event and returns a boolean
// answer indicating whether that event is "enabled" under the current
// policy. The exact method signature is defined by the Predicate interface
// in this package, but all implementations MUST obey the following rules:
//
//   - Predicates MUST be side-effect free with respect to the logging
//     pipeline. Evaluating a predicate MUST NOT mutate the log entry, its
//     fields, or shared global state in a way that affects subsequent
//     logging behavior.
//
//   - Predicates SHOULD be pure functions of their inputs: given the same
//     entry and fields, they SHOULD always return the same result, unless
//     they are explicitly designed to incorporate external state (for
//     example, sampling counters or configuration that can change over
//     time).
//
//   - Predicates MUST be safe for concurrent use by multiple goroutines,
//     since they are typically shared across many loggers and cores in a
//     multi-goroutine application.
//
//   - Predicates SHOULD be inexpensive to evaluate, as they may be called on
//     every log attempt before any expensive encoding or I/O is performed.
//
// Typical implementations include:
//
//   - checking the log level against a threshold in combination with other
//     criteria;
//
//   - matching logger names or prefixes to enable or disable specific
//     subsystems;
//
//   - inspecting structured fields for particular keys or values;
//
//   - consulting dynamic configuration or feature flags.
//
// # Func adapter
//
// The Func adapter type allows plain functions to satisfy the Predicate
// interface. This is convenient for callers that want to define small,
// focused predicates inline or as standalone functions:
//
//   - It avoids the boilerplate of declaring a struct type for every
//     predicate.
//
//   - It makes it easy to capture configuration via closures (for example,
//     allowed logger names, patterns, or thresholds) while still presenting
//     the unified Predicate interface to the rest of the logging pipeline.
//
// As with all Predicate implementations, functions wrapped by Func MUST be
// safe for concurrent use and SHOULD avoid observable side effects.
//
// # Composition helpers
//
// To support building more complex conditions from simpler ones, the
// package provides helpers that combine predicates using boolean logic,
// typically with short-circuit semantics:
//
//   - And(p1, p2, ...) returns a predicate that evaluates each operand in
//     sequence and returns false as soon as any operand returns false. Only
//     if all operands return true does the composite predicate return true.
//
//   - Or(p1, p2, ...) returns a predicate that evaluates each operand in
//     sequence and returns true as soon as any operand returns true. Only
//     if all operands return false does the composite predicate return
//     false.
//
//   - Not(p) returns a predicate that negates the result of p.
//
// These helpers are designed to be transparent:
//
//   - They do not introduce additional side effects beyond calling their
//     operands.
//
//   - They preserve short-circuit behavior: once the result is determined,
//     remaining predicates are not evaluated.
//
//   - They are themselves safe for concurrent use as long as the underlying
//     predicates are.
//
// Some implementations may also provide convenience predicates such as
// "always true" or "always false" for use as defaults or sentinels. Their
// behavior is trivial but can simplify configuration code.
//
// # Usage context
//
// The predicate package is intentionally small and self-contained. It does
// not know about encoders, writers, or concrete logging backends. Instead,
// it provides a shared vocabulary for expressing boolean policies over log
// events.
//
// In the broader rxlog pipeline, predicates are typically used:
//
//   - by cores to decide whether to accept an entry beyond simple level
//     checks;
//
//   - by hook managers to decide whether particular hooks should run for a
//     given entry;
//
//   - by application code to implement routing and sampling (for example,
//     only log certain categories of requests or only sample a subset of
//     high-volume events).
//
// By centralizing predicate semantics and composition in this package,
// rxlog allows applications to build rich filtering logic from small,
// reusable pieces while keeping the core logging API clean and declarative.
package predicate
