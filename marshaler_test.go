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

package rxlog_test

import (
	"testing"

	"dirpx.dev/rxlog"
	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/encoder/base"
	ftype "dirpx.dev/rxlog/rxapi/field/type"
)

// testArrayMarshaler demonstrates ArrayMarshaler implementation.
// This is a minimal test implementation - real implementations would
// use the encoder to properly encode array elements.
type testArrayMarshaler []int64

func (t testArrayMarshaler) MarshalLogArray(_ base.ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
	// Minimal implementation for testing - just returns the buffer unchanged.
	// Real implementations would use enc.AppendInt64, enc.AppendString, etc.
	return dst, nil
}

// testObjectMarshaler demonstrates ObjectMarshaler implementation.
// This is a minimal test implementation - real implementations would
// use the encoder to properly encode object fields.
type testObjectMarshaler struct {
	ID   int64
	Name string
}

func (t testObjectMarshaler) MarshalLogObject(_ base.ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
	// Minimal implementation for testing - just returns the buffer unchanged.
	// Real implementations would use enc.AppendInt64, enc.AppendString, etc.
	return dst, nil
}

func TestArrayMarshaler(t *testing.T) {
	// Verify that rxlog.ArrayMarshaler is usable and can be implemented.
	var _ rxlog.ArrayMarshaler = testArrayMarshaler{}

	arr := testArrayMarshaler{1, 2, 3}
	dst := &buffer.Buffer{}

	// Call MarshalLogArray to verify it compiles and runs.
	result, err := arr.MarshalLogArray(nil, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != dst {
		t.Fatalf("expected result to be dst")
	}
}

func TestArrayMarshalerFunc(t *testing.T) {
	// Verify that ArrayMarshalerFunc implements ArrayMarshaler.
	fn := rxlog.ArrayMarshalerFunc(func(_ base.ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
		return dst, nil
	})

	var _ rxlog.ArrayMarshaler = fn

	dst := &buffer.Buffer{}
	result, err := fn.MarshalLogArray(nil, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != dst {
		t.Fatalf("expected result to be dst")
	}
}

func TestObjectMarshaler(t *testing.T) {
	// Verify that rxlog.ObjectMarshaler is usable and can be implemented.
	var _ rxlog.ObjectMarshaler = testObjectMarshaler{}

	obj := testObjectMarshaler{ID: 123, Name: "Alice"}
	dst := &buffer.Buffer{}

	// Call MarshalLogObject to verify it compiles and runs.
	result, err := obj.MarshalLogObject(nil, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != dst {
		t.Fatalf("expected result to be dst")
	}
}

func TestObjectMarshalerFunc(t *testing.T) {
	// Verify that ObjectMarshalerFunc implements ObjectMarshaler.
	fn := rxlog.ObjectMarshalerFunc(func(_ base.ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
		return dst, nil
	})

	var _ rxlog.ObjectMarshaler = fn

	dst := &buffer.Buffer{}
	result, err := fn.MarshalLogObject(nil, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != dst {
		t.Fatalf("expected result to be dst")
	}
}

func TestMarshalerUsageWithFields(t *testing.T) {
	// Test that marshalers work with the Array() field constructor.
	t.Run("ArrayWithCustomType", func(t *testing.T) {
		arr := testArrayMarshaler{100, 200, 300}
		field := rxlog.Array("user_ids", arr)

		if field.Type != ftype.Array {
			t.Errorf("expected type %v, got %v", ftype.Array, field.Type)
		}
		if field.Key != "user_ids" {
			t.Errorf("expected key %q, got %q", "user_ids", field.Key)
		}
		if _, ok := field.Interface.(testArrayMarshaler); !ok {
			t.Fatalf("field.Interface has type %T, want testArrayMarshaler", field.Interface)
		}
	})

	// Test that marshalers work with the Object() field constructor.
	t.Run("ObjectWithCustomType", func(t *testing.T) {
		obj := testObjectMarshaler{ID: 42, Name: "Bob"}
		field := rxlog.Object("user", obj)

		if field.Type != ftype.Object {
			t.Errorf("expected type %v, got %v", ftype.Object, field.Type)
		}
		if field.Key != "user" {
			t.Errorf("expected key %q, got %q", "user", field.Key)
		}
		if _, ok := field.Interface.(testObjectMarshaler); !ok {
			t.Fatalf("field.Interface has type %T, want testObjectMarshaler", field.Interface)
		}
	})

	// Test that ArrayMarshalerFunc works with the Array() field constructor.
	t.Run("ArrayWithFunc", func(t *testing.T) {
		fn := rxlog.ArrayMarshalerFunc(func(_ base.ArrayEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
			return dst, nil
		})
		field := rxlog.Array("items", fn)

		if field.Type != ftype.Array {
			t.Errorf("expected type %v, got %v", ftype.Array, field.Type)
		}
		if field.Key != "items" {
			t.Errorf("expected key %q, got %q", "items", field.Key)
		}
	})

	// Test that ObjectMarshalerFunc works with the Object() field constructor.
	t.Run("ObjectWithFunc", func(t *testing.T) {
		fn := rxlog.ObjectMarshalerFunc(func(_ base.ObjectEncoder, dst *buffer.Buffer) (*buffer.Buffer, error) {
			return dst, nil
		})
		field := rxlog.Object("meta", fn)

		if field.Type != ftype.Object {
			t.Errorf("expected type %v, got %v", ftype.Object, field.Type)
		}
		if field.Key != "meta" {
			t.Errorf("expected key %q, got %q", "meta", field.Key)
		}
	})

	// Test that Inline() works with ObjectMarshaler.
	t.Run("InlineWithObjectMarshaler", func(t *testing.T) {
		obj := testObjectMarshaler{ID: 1, Name: "test"}
		field := rxlog.Inline(obj)

		if field.Type != ftype.Inline {
			t.Errorf("expected type %v, got %v", ftype.Inline, field.Type)
		}
		if field.Key != "" {
			t.Errorf("expected empty key, got %q", field.Key)
		}
		if _, ok := field.Interface.(testObjectMarshaler); !ok {
			t.Fatalf("field.Interface has type %T, want testObjectMarshaler", field.Interface)
		}
	})
}
