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

package extractor_test

import (
	"context"
	"testing"

	"dirpx.dev/rxlog/rxapi/context/extractor"
	"dirpx.dev/rxlog/rxapi/field"
)

// ctxKey is a private key type used to verify that Func receives the
// context passed to Extract.
type ctxKey struct{ name string }

func TestFuncAdapterCallsUnderlyingFunction(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey{"k"}, "v")

	var (
		called   bool
		gotValue any
	)

	fn := extractor.Func(func(c context.Context) []field.Field {
		called = true
		gotValue = c.Value(ctxKey{"k"})
		return []field.Field{field.String("k", "v")}
	})

	fields := fn.Extract(ctx)

	if !called {
		t.Fatalf("Func.Extract did not call underlying function")
	}
	if gotValue != "v" {
		t.Fatalf("Func.Extract received unexpected context value: got %v, want %v", gotValue, "v")
	}
	if len(fields) != 1 {
		t.Fatalf("Func.Extract returned %d fields, want 1", len(fields))
	}
	want := field.String("k", "v")
	if fields[0] != want {
		t.Fatalf("Func.Extract returned field %v, want %v", fields[0], want)
	}
}

func TestComposeNoExtractorsReturnsEmpty(t *testing.T) {
	combined := extractor.Compose() // no extractors
	ctx := context.Background()

	fields := combined.Extract(ctx)
	if len(fields) != 0 {
		t.Fatalf("Compose() with no extractors returned %d fields, want 0", len(fields))
	}
}

func TestComposePreservesOrderAcrossExtractors(t *testing.T) {
	ctx := context.Background()

	ext1 := extractor.Func(func(context.Context) []field.Field {
		return []field.Field{
			field.String("e1", "a"),
			field.String("e1", "b"),
		}
	})
	ext2 := extractor.Func(func(context.Context) []field.Field {
		return []field.Field{
			field.String("e2", "c"),
		}
	})
	ext3 := extractor.Func(func(context.Context) []field.Field {
		return []field.Field{
			field.String("e3", "d"),
			field.String("e3", "e"),
		}
	})

	combined := extractor.Compose(ext1, ext2, ext3)
	got := combined.Extract(ctx)

	want := []field.Field{
		field.String("e1", "a"),
		field.String("e1", "b"),
		field.String("e2", "c"),
		field.String("e3", "d"),
		field.String("e3", "e"),
	}

	if len(got) != len(want) {
		t.Fatalf("Compose returned %d fields, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Compose result[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestComposeHandlesNilAndEmptyResults(t *testing.T) {
	ctx := context.Background()

	// First extractor returns nil slice.
	extNil := extractor.Func(func(context.Context) []field.Field {
		return nil
	})

	// Second extractor returns an explicitly empty slice.
	extEmpty := extractor.Func(func(context.Context) []field.Field {
		return []field.Field{}
	})

	// Third extractor returns actual fields.
	extFields := extractor.Func(func(context.Context) []field.Field {
		return []field.Field{
			field.String("k", "v"),
		}
	})

	combined := extractor.Compose(extNil, extEmpty, extFields)
	got := combined.Extract(ctx)

	if len(got) != 1 {
		t.Fatalf("Compose(nil, empty, fields) returned %d fields, want 1", len(got))
	}
	want := field.String("k", "v")
	if got[0] != want {
		t.Fatalf("Compose(nil, empty, fields)[0] = %v, want %v", got[0], want)
	}
}
