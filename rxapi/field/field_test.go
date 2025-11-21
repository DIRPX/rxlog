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

package field_test

import (
	"testing"

	rxfield "dirpx.dev/rxlog/rxapi/field"
	ftype "dirpx.dev/rxlog/rxapi/field/type"
)

// TestEqualsTypeMismatch verifies that Fields with different Types are
// considered unequal even if other data matches.
func TestEqualsTypeMismatch(t *testing.T) {
	f1 := rxfield.Field{
		Key:       "k",
		Type:      ftype.String,
		String:    "v",
		Integer:   0,
		Interface: nil,
	}
	f2 := rxfield.Field{
		Key:       "k",
		Type:      ftype.Int64,
		Integer:   42,
		String:    "",
		Interface: nil,
	}

	if f1.Equals(f2) {
		t.Fatalf("Fields with different Type should not be equal: %v vs %v", f1, f2)
	}
}

// TestEqualsKeyMismatch verifies that Fields with different Keys are
// considered unequal even if Type and payload match.
func TestEqualsKeyMismatch(t *testing.T) {
	f1 := rxfield.Field{
		Key:    "k1",
		Type:   ftype.String,
		String: "v",
	}
	f2 := rxfield.Field{
		Key:    "k2",
		Type:   ftype.String,
		String: "v",
	}

	if f1.Equals(f2) {
		t.Fatalf("Fields with different Key should not be equal: %v vs %v", f1, f2)
	}
}

// TestEqualsBinaryUsesBytesEqual ensures that Binary fields use content
// equality for []byte, not slice-header identity.
func TestEqualsBinaryUsesBytesEqual(t *testing.T) {
	b1 := []byte{1, 2, 3}
	b2 := []byte{1, 2, 3} // distinct slice with identical contents

	f1 := rxfield.Field{
		Key:       "bin",
		Type:      ftype.Binary,
		Interface: b1,
	}
	f2 := rxfield.Field{
		Key:       "bin",
		Type:      ftype.Binary,
		Interface: b2,
	}

	if !f1.Equals(f2) {
		t.Fatalf("Binary fields with same bytes should be equal: %v vs %v", f1, f2)
	}

	// Different contents must not be equal.
	f3 := rxfield.Field{
		Key:       "bin",
		Type:      ftype.Binary,
		Interface: []byte{1, 2, 4},
	}
	if f1.Equals(f3) {
		t.Fatalf("Binary fields with different bytes should not be equal: %v vs %v", f1, f3)
	}
}

// TestEqualsByteStringUsesBytesEqual ensures that ByteString fields use
// bytes.Equal for their []byte payloads.
func TestEqualsByteStringUsesBytesEqual(t *testing.T) {
	b1 := []byte("hello")
	b2 := []byte("hello")

	f1 := rxfield.Field{
		Key:       "bs",
		Type:      ftype.ByteString,
		Interface: b1,
	}
	f2 := rxfield.Field{
		Key:       "bs",
		Type:      ftype.ByteString,
		Interface: b2,
	}

	if !f1.Equals(f2) {
		t.Fatalf("ByteString fields with same bytes should be equal: %v vs %v", f1, f2)
	}

	f3 := rxfield.Field{
		Key:       "bs",
		Type:      ftype.ByteString,
		Interface: []byte("world"),
	}
	if f1.Equals(f3) {
		t.Fatalf("ByteString fields with different bytes should not be equal: %v vs %v", f1, f3)
	}
}

// complexPayload is a helper type used to verify DeepEqual-based comparison
// for complex Types (Array, Object, Error, Reflect).
type complexPayload struct {
	Name  string
	Value int
}

// TestEqualsComplexTypesUseDeepEqual verifies that Array, Object, Error,
// and Reflect types use reflect.DeepEqual on the Interface payload.
func TestEqualsComplexTypesUseDeepEqual(t *testing.T) {
	payload1 := complexPayload{Name: "x", Value: 1}
	payload2 := complexPayload{Name: "x", Value: 1}
	payload3 := complexPayload{Name: "y", Value: 2}

	types := []ftype.Type{
		ftype.Array,
		ftype.Object,
		ftype.Error,
		ftype.Reflect,
	}

	for _, ty := range types {
		ty := ty
		t.Run(string(ty), func(t *testing.T) {
			// Same structural payload => equal.
			f1 := rxfield.Field{
				Key:       "k",
				Type:      ty,
				Interface: payload1,
			}
			f2 := rxfield.Field{
				Key:       "k",
				Type:      ty,
				Interface: payload2, // distinct value with same contents
			}

			if !f1.Equals(f2) {
				t.Fatalf("Equals(%v, %v) = false, want true for Type %v", f1, f2, ty)
			}

			// Different structural payload => not equal.
			f3 := rxfield.Field{
				Key:       "k",
				Type:      ty,
				Interface: payload3,
			}
			if f1.Equals(f3) {
				t.Fatalf("Equals(%v, %v) = true, want false for Type %v", f1, f3, ty)
			}
		})
	}
}

// TestEqualsDefaultUsesStructEquality verifies that for Types not handled
// specially in Equals, the comparison falls back to full struct equality.
func TestEqualsDefaultUsesStructEquality(t *testing.T) {
	f1 := rxfield.Field{
		Key:       "num",
		Type:      ftype.Int64,
		Integer:   42,
		String:    "",
		Interface: nil,
	}
	f2 := rxfield.Field{
		Key:       "num",
		Type:      ftype.Int64,
		Integer:   42,
		String:    "",
		Interface: nil,
	}
	f3 := rxfield.Field{
		Key:       "num",
		Type:      ftype.Int64,
		Integer:   7,
		String:    "",
		Interface: nil,
	}

	if !f1.Equals(f2) {
		t.Fatalf("Fields with identical struct contents must be equal: %v vs %v", f1, f2)
	}
	if f1.Equals(f3) {
		t.Fatalf("Fields with different Integer must not be equal: %v vs %v", f1, f3)
	}
}
