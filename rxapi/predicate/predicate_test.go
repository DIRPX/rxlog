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

package predicate_test

import (
	"testing"

	"dirpx.dev/rxlog/rxapi/core"
	"dirpx.dev/rxlog/rxapi/field"
	"dirpx.dev/rxlog/rxapi/predicate"
)

// dummyEntry returns a zero-value Entry suitable for tests.
//
// The predicate combinators in this package do not inspect Entry contents,
// so zero value is sufficient for black-box tests.
func dummyEntry() core.Entry {
	return core.Entry{}
}

func TestFuncAdapterCallsUnderlyingFunction(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	fields := []field.Field{field.String("k", "v")}

	var (
		called    bool
		gotEntry  core.Entry
		gotFields []field.Field
	)

	fn := predicate.Func(func(entry core.Entry, fs []field.Field) bool {
		called = true
		gotEntry = entry
		gotFields = fs
		return true
	})

	if !fn.ShouldLog(e, fields) {
		t.Fatalf("Func.ShouldLog returned false, want true")
	}
	if !called {
		t.Fatalf("Func.ShouldLog did not call underlying function")
	}
	if gotEntry != e {
		t.Fatalf("Func.ShouldLog received unexpected entry: %#v", gotEntry)
	}
	if len(gotFields) != len(fields) {
		t.Fatalf("Func.ShouldLog received %d fields, want %d", len(gotFields), len(fields))
	}
}

func TestAndEmptyReturnsTrue(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	and := predicate.And() // no predicates
	if !and.ShouldLog(e, fields) {
		t.Fatalf("And() with no predicates returned false, want true (identity for AND)")
	}
}

func TestAndAllTrueReturnsTrue(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	p1 := predicate.Func(func(core.Entry, []field.Field) bool { return true })
	p2 := predicate.Func(func(core.Entry, []field.Field) bool { return true })

	and := predicate.And(p1, p2)
	if !and.ShouldLog(e, fields) {
		t.Fatalf("And(true, true) = false, want true")
	}
}

func TestAndAnyFalseReturnsFalseAndShortCircuits(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	pFalse := predicate.Func(func(core.Entry, []field.Field) bool { return false })

	var called int
	pCount := predicate.Func(func(core.Entry, []field.Field) bool {
		called++
		return true
	})

	and := predicate.And(pFalse, pCount)
	if and.ShouldLog(e, fields) {
		t.Fatalf("And(false, true) = true, want false")
	}
	if called != 0 {
		t.Fatalf("And should short-circuit on false; second predicate called %d times, want 0", called)
	}
}

func TestOrEmptyReturnsFalse(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	or := predicate.Or() // no predicates
	if or.ShouldLog(e, fields) {
		t.Fatalf("Or() with no predicates returned true, want false (identity for OR)")
	}
}

func TestOrAnyTrueReturnsTrueAndShortCircuits(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	pTrue := predicate.Func(func(core.Entry, []field.Field) bool { return true })

	var called int
	pCount := predicate.Func(func(core.Entry, []field.Field) bool {
		called++
		return true
	})

	or := predicate.Or(pTrue, pCount)
	if !or.ShouldLog(e, fields) {
		t.Fatalf("Or(true, _) = false, want true")
	}
	if called != 0 {
		t.Fatalf("Or should short-circuit on true; second predicate called %d times, want 0", called)
	}
}

func TestOrAllFalseReturnsFalse(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	p1 := predicate.Func(func(core.Entry, []field.Field) bool { return false })
	p2 := predicate.Func(func(core.Entry, []field.Field) bool { return false })

	or := predicate.Or(p1, p2)
	if or.ShouldLog(e, fields) {
		t.Fatalf("Or(false, false) = true, want false")
	}
}

func TestNotInvertsResult(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	pTrue := predicate.Func(func(core.Entry, []field.Field) bool { return true })
	pFalse := predicate.Func(func(core.Entry, []field.Field) bool { return false })

	if !predicate.Not(pFalse).ShouldLog(e, fields) {
		t.Fatalf("Not(false) = false, want true")
	}
	if predicate.Not(pTrue).ShouldLog(e, fields) {
		t.Fatalf("Not(true) = true, want false")
	}
}

func TestXorEmptyReturnsFalse(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	xor := predicate.Xor() // no predicates
	if xor.ShouldLog(e, fields) {
		t.Fatalf("Xor() with no predicates returned true, want false")
	}
}

func TestXorSingleBehavesLikePredicate(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	var called int
	p := predicate.Func(func(core.Entry, []field.Field) bool {
		called++
		return true
	})

	xor := predicate.Xor(p)
	if !xor.ShouldLog(e, fields) {
		t.Fatalf("Xor(p) with single true predicate = false, want true")
	}
	if called != 1 {
		t.Fatalf("Xor(p) should call predicate exactly once, called %d times", called)
	}
}

func TestXorExactlyOneTrue(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	pFalse := predicate.Func(func(core.Entry, []field.Field) bool { return false })
	pTrue := predicate.Func(func(core.Entry, []field.Field) bool { return true })

	// Only second is true.
	xor := predicate.Xor(pFalse, pTrue, pFalse)
	if !xor.ShouldLog(e, fields) {
		t.Fatalf("Xor(false, true, false) = false, want true")
	}

	// Only first is true.
	xor = predicate.Xor(pTrue, pFalse, pFalse)
	if !xor.ShouldLog(e, fields) {
		t.Fatalf("Xor(true, false, false) = false, want true")
	}
}

func TestXorZeroOrMultipleTrueReturnFalse(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	pFalse := predicate.Func(func(core.Entry, []field.Field) bool { return false })
	pTrue := predicate.Func(func(core.Entry, []field.Field) bool { return true })

	// Zero true.
	if xor := predicate.Xor(pFalse, pFalse); xor.ShouldLog(e, fields) {
		t.Fatalf("Xor(false, false) = true, want false")
	}

	// Two true.
	if xor := predicate.Xor(pTrue, pTrue); xor.ShouldLog(e, fields) {
		t.Fatalf("Xor(true, true) = true, want false")
	}

	// More than one true among three.
	if xor := predicate.Xor(pTrue, pFalse, pTrue); xor.ShouldLog(e, fields) {
		t.Fatalf("Xor(true, false, true) = true, want false")
	}
}

func TestXorDoesNotShortCircuit(t *testing.T) {
	t.Helper()

	e := dummyEntry()
	var fields []field.Field

	var called int
	p1 := predicate.Func(func(core.Entry, []field.Field) bool {
		called++
		return true
	})
	p2 := predicate.Func(func(core.Entry, []field.Field) bool {
		called++
		return false
	})

	xor := predicate.Xor(p1, p2)
	_ = xor.ShouldLog(e, fields)

	if called != 2 {
		t.Fatalf("Xor must evaluate all predicates; called = %d, want 2", called)
	}
}
