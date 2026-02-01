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
	"errors"
	"testing"
	"time"

	"dirpx.dev/rxlog"
	ftype "dirpx.dev/rxlog/rxapi/field/type"
)

// TestFieldShortcuts verifies that all field constructor shortcuts
// exported from the root package work correctly.
func TestFieldShortcuts(t *testing.T) {
	tests := []struct {
		name      string
		field     rxlog.Field
		wantType  rxlog.FieldType
		wantKey   string
		wantValue interface{} // expected value for verification
	}{
		{
			name:     "String",
			field:    rxlog.String("key", "value"),
			wantType: ftype.String,
			wantKey:  "key",
		},
		{
			name:     "Int",
			field:    rxlog.Int("count", 42),
			wantType: ftype.Int64,
			wantKey:  "count",
		},
		{
			name:     "Int64",
			field:    rxlog.Int64("num", 100),
			wantType: ftype.Int64,
			wantKey:  "num",
		},
		{
			name:     "Bool",
			field:    rxlog.Bool("flag", true),
			wantType: ftype.Bool,
			wantKey:  "flag",
		},
		{
			name:     "Float64",
			field:    rxlog.Float64("pi", 3.14159),
			wantType: ftype.Float64,
			wantKey:  "pi",
		},
		{
			name:     "Duration",
			field:    rxlog.Duration("elapsed", time.Second),
			wantType: ftype.Duration,
			wantKey:  "elapsed",
		},
		{
			name:     "Error",
			field:    rxlog.Error("err", errors.New("test error")),
			wantType: ftype.Error,
			wantKey:  "err",
		},
		{
			name:     "Skip",
			field:    rxlog.Skip(),
			wantType: ftype.Skip,
		},
		{
			name:     "Binary",
			field:    rxlog.Binary("data", []byte{1, 2, 3}),
			wantType: ftype.Binary,
			wantKey:  "data",
		},
		{
			name:     "ByteString",
			field:    rxlog.ByteString("bytes", []byte("hello")),
			wantType: ftype.ByteString,
			wantKey:  "bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.Type != tt.wantType {
				t.Errorf("field type = %v, want %v", tt.field.Type, tt.wantType)
			}
			if tt.wantKey != "" && tt.field.Key != tt.wantKey {
				t.Errorf("field key = %q, want %q", tt.field.Key, tt.wantKey)
			}
		})
	}
}

// TestFieldPointerShortcuts verifies that pointer field constructors
// handle nil values correctly by returning Skip fields.
func TestFieldPointerShortcuts(t *testing.T) {
	var (
		nilString   *string
		nilInt      *int
		nilBool     *bool
		nilFloat64  *float64
		nilDuration *time.Duration
		nilError    *error
	)

	tests := []struct {
		name  string
		field rxlog.Field
	}{
		{"StringPtr(nil)", rxlog.StringPtr("key", nilString)},
		{"IntPtr(nil)", rxlog.IntPtr("key", nilInt)},
		{"BoolPtr(nil)", rxlog.BoolPtr("key", nilBool)},
		{"Float64Ptr(nil)", rxlog.Float64Ptr("key", nilFloat64)},
		{"DurationPtr(nil)", rxlog.DurationPtr("key", nilDuration)},
		{"ErrorPtr(nil)", rxlog.ErrorPtr("key", nilError)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.Type != ftype.Skip {
				t.Errorf("expected Skip type for nil pointer, got %v", tt.field.Type)
			}
		})
	}

	// Test non-nil pointers work correctly
	s := "test"
	i := 42
	if f := rxlog.StringPtr("key", &s); f.Type != ftype.String {
		t.Errorf("StringPtr with non-nil should produce String type, got %v", f.Type)
	}
	if f := rxlog.IntPtr("key", &i); f.Type != ftype.Int64 {
		t.Errorf("IntPtr with non-nil should produce Int64 type, got %v", f.Type)
	}
}

// TestFieldTypeConstants verifies that field type constants are correctly
// exported from the root package.
func TestFieldTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant rxlog.FieldType
		want     ftype.Type
	}{
		{"UnknownFieldType", rxlog.UnknownFieldType, ftype.Unknown},
		{"StringFieldType", rxlog.StringFieldType, ftype.String},
		{"Int64FieldType", rxlog.Int64FieldType, ftype.Int64},
		{"BoolFieldType", rxlog.BoolFieldType, ftype.Bool},
		{"Float64FieldType", rxlog.Float64FieldType, ftype.Float64},
		{"DurationFieldType", rxlog.DurationFieldType, ftype.Duration},
		{"TimeFieldType", rxlog.TimeFieldType, ftype.Time},
		{"ErrorFieldType", rxlog.ErrorFieldType, ftype.Error},
		{"SkipFieldType", rxlog.SkipFieldType, ftype.Skip},
		{"ArrayFieldType", rxlog.ArrayFieldType, ftype.Array},
		{"ObjectFieldType", rxlog.ObjectFieldType, ftype.Object},
		{"InlineFieldType", rxlog.InlineFieldType, ftype.Inline},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("constant %s = %v, want %v", tt.name, tt.constant, tt.want)
			}
		})
	}
}

// TestTimeInLocation verifies that the TimeInLocation wrapper function
// works correctly.
func TestTimeInLocation(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("could not load America/New_York timezone")
	}

	now := time.Now()
	f := rxlog.TimeInLocation("timestamp", now, loc)

	if f.Type != ftype.Time {
		t.Errorf("TimeInLocation type = %v, want %v", f.Type, ftype.Time)
	}
	if f.Key != "timestamp" {
		t.Errorf("TimeInLocation key = %q, want %q", f.Key, "timestamp")
	}
}

// TestAnyFieldConstructor verifies that the Any constructor correctly
// dispatches to appropriate type-specific constructors.
func TestAnyFieldConstructor(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		wantType rxlog.FieldType
	}{
		{"nil", nil, ftype.Skip},
		{"string", "hello", ftype.String},
		{"int", 42, ftype.Int64},
		{"bool", true, ftype.Bool},
		{"float64", 3.14, ftype.Float64},
		{"duration", time.Second, ftype.Duration},
		{"error", errors.New("err"), ftype.Error},
		{"[]byte", []byte("data"), ftype.ByteString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := rxlog.Any("key", tt.value)
			if f.Type != tt.wantType {
				t.Errorf("Any(%v) type = %v, want %v", tt.value, f.Type, tt.wantType)
			}
		})
	}
}
