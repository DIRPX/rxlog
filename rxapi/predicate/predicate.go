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

package predicate

import (
	"dirpx.dev/rxlog/rxapi/core"
	"dirpx.dev/rxlog/rxapi/field"
)

// Predicate decides whether a given log event SHOULD be emitted.
//
// Implementations of Predicate SHOULD be side-effect free and SHOULD NOT
// mutate the provided Entry or fields. A Predicate MUST return true if the
// event is allowed to be logged and false if it MUST be suppressed.
type Predicate interface {
	ShouldLog(core.Entry, []field.Field) bool
}

// Func is a function adapter that turns a plain function into
// a Predicate.
//
// Any function with the signature func(entry.Entry, []field.Field) bool MAY be
// wrapped as a Func and used wherever a Predicate is required.
type Func func(core.Entry, []field.Field) bool

// ShouldLog calls the underlying function.
//
// When used via this adapter, the function MUST obey the same contract as
// Predicate.ShouldLog: it SHOULD be free of side effects and MUST be safe to
// call multiple times with the same arguments.
func (f Func) ShouldLog(e core.Entry, fields []field.Field) bool {
	return f(e, fields)
}

// And returns a Predicate that represents a logical conjunction (AND)
// of the provided predicates.
//
// The resulting Predicate MUST return true if and only if ALL supplied
// predicates return true for the given Entry and fields. Predicates are
// evaluated in the order provided and evaluation MUST short-circuit: as soon
// as one predicate returns false, subsequent predicates MUST NOT be evaluated.
//
// If no predicates are provided, the returned Predicate MUST always return
// true (logical identity for AND).
func And(predicates ...Predicate) Predicate {
	return Func(func(e core.Entry, fields []field.Field) bool {
		for _, p := range predicates {
			if !p.ShouldLog(e, fields) {
				return false
			}
		}
		return true
	})
}

// Or returns a Predicate that represents a logical disjunction (OR)
// of the provided predicates.
//
// The resulting Predicate MUST return true if at least one of the supplied
// predicates returns true for the given Entry and fields. Predicates are
// evaluated in the order provided and evaluation MUST short-circuit: as soon
// as one predicate returns true, subsequent predicates MUST NOT be evaluated.
//
// If no predicates are provided, the returned Predicate MUST always return
// false (logical identity for OR).
func Or(predicates ...Predicate) Predicate {
	return Func(func(e core.Entry, fields []field.Field) bool {
		for _, p := range predicates {
			if p.ShouldLog(e, fields) {
				return true
			}
		}
		return false
	})
}

// Not returns a Predicate that represents a logical negation (NOT)
// of the provided predicate.
//
// The resulting Predicate MUST return true if and only if the supplied
// predicate returns false for the given Entry and fields, and MUST return
// false if and only if the supplied predicate returns true.
//
// This operator is useful for inverting existing predicates. For example,
// Not(Level(minLevel)) will log everything that the original predicate would
// suppress, and suppress everything that the original predicate would log.
func Not(p Predicate) Predicate {
	return Func(func(e core.Entry, fields []field.Field) bool {
		return !p.ShouldLog(e, fields)
	})
}

// Xor returns a Predicate that represents a logical exclusive OR (XOR)
// of the provided predicates.
//
// The resulting Predicate MUST return true if and only if EXACTLY ONE of the
// supplied predicates returns true for the given Entry and fields. If zero
// predicates return true, or if more than one predicate returns true, the
// result MUST be false.
//
// Unlike And and Or, Xor does NOT short-circuit: all predicates MUST be
// evaluated to determine the final result.
//
// If no predicates are provided, the returned Predicate MUST always return
// false (no predicate is true).
//
// If only one predicate is provided, the returned Predicate MUST behave
// identically to that predicate (exactly one candidate is always that one).
func Xor(predicates ...Predicate) Predicate {
	return Func(func(e core.Entry, fields []field.Field) bool {
		trueCount := 0
		for _, p := range predicates {
			if p.ShouldLog(e, fields) {
				trueCount++
				if trueCount > 1 {
					// More than one predicate is true; XOR MUST be false.
					return false
				}
			}
		}
		return trueCount == 1
	})
}
