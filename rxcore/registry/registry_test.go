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

package registry_test

import (
	"strings"
	"testing"

	"dirpx.dev/rxlog/rxcore/registry"
)

// TestNew_CreatesEmptyRegistry verifies that New creates a registry with no items.
func TestNew_CreatesEmptyRegistry(t *testing.T) {
	t.Helper()

	reg := registry.New[string]()

	_, err := reg.FromString("anything")
	if err == nil {
		t.Fatal("FromString on empty registry returned nil error, want error")
	}
}

// TestRegister_AddsValue verifies that Register adds values that can be retrieved.
func TestRegister_AddsValue(t *testing.T) {
	t.Helper()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"simple", "test", "value"},
		{"with-dash", "test-key", "test-value"},
		{"with_underscore", "test_key", "test_value"},
		{"numeric", "test123", "value123"},
		{"empty_value", "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := registry.New[string]()
			reg.Register(tt.key, tt.value)

			got, err := reg.FromString(tt.key)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.key, err)
			}
			if got != tt.value {
				t.Errorf("FromString(%q) = %q, want %q", tt.key, got, tt.value)
			}
		})
	}
}

// TestRegister_OverwritesExistingValue verifies that Register replaces existing values.
func TestRegister_OverwritesExistingValue(t *testing.T) {
	t.Helper()

	reg := registry.New[string]()
	reg.Register("test", "initial")
	reg.Register("test", "replaced")

	got, err := reg.FromString("test")
	if err != nil {
		t.Fatalf("FromString(%q) returned error: %v", "test", err)
	}
	if got != "replaced" {
		t.Errorf("FromString(%q) = %q, want %q", "test", got, "replaced")
	}
}

// TestFromString_UnknownName verifies that FromString returns an error for unregistered names.
func TestFromString_UnknownName(t *testing.T) {
	t.Helper()

	tests := []struct {
		name         string
		registered   map[string]string
		lookup       string
		wantInError  string
	}{
		{
			name:        "no_registrations",
			registered:  map[string]string{},
			lookup:      "unknown",
			wantInError: "unknown",
		},
		{
			name:        "different_name",
			registered:  map[string]string{"known": "value"},
			lookup:      "unknown",
			wantInError: "unknown",
		},
		{
			name:        "case_sensitive",
			registered:  map[string]string{"test": "value"},
			lookup:      "Test",
			wantInError: "Test",
		},
		{
			name:        "whitespace_sensitive",
			registered:  map[string]string{"test": "value"},
			lookup:      " test",
			wantInError: " test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := registry.New[string]()
			for k, v := range tt.registered {
				reg.Register(k, v)
			}

			_, err := reg.FromString(tt.lookup)
			if err == nil {
				t.Fatalf("FromString(%q) returned nil error, want error", tt.lookup)
			}

			if !strings.Contains(err.Error(), tt.wantInError) {
				t.Errorf("error message %q does not contain %q", err.Error(), tt.wantInError)
			}
		})
	}
}

// TestFromString_ErrorMessageIncludesTypeName verifies that error messages
// include the type name extracted via reflection.
func TestFromString_ErrorMessageIncludesTypeName(t *testing.T) {
	t.Helper()

	tests := []struct {
		name         string
		wantTypeName string
	}{
		{"int", "int"},
		{"string", "string"},
		{"bool", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			switch tt.name {
			case "int":
				reg := registry.New[int]()
				_, err = reg.FromString("nonexistent")
			case "string":
				reg := registry.New[string]()
				_, err = reg.FromString("nonexistent")
			case "bool":
				reg := registry.New[bool]()
				_, err = reg.FromString("nonexistent")
			}

			if err == nil {
				t.Fatal("FromString returned nil error, want error")
			}

			if !strings.Contains(err.Error(), tt.wantTypeName) {
				t.Errorf("error message %q does not contain type name %q", err.Error(), tt.wantTypeName)
			}
		})
	}
}

// TestMustFromString_ReturnsValueOnSuccess verifies that MustFromString
// returns the value for registered names without panicking.
func TestMustFromString_ReturnsValueOnSuccess(t *testing.T) {
	t.Helper()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"simple", "test", "success"},
		{"lowercase", "lowercase", "lower"},
		{"uppercase", "uppercase", "upper"},
		{"with-dash", "encoder-json", "json-encoder"},
		{"multiple_words", "my_custom_encoder", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := registry.New[string]()
			reg.Register(tt.key, tt.value)

			got := reg.MustFromString(tt.key)
			if got != tt.value {
				t.Errorf("MustFromString(%q) = %q, want %q", tt.key, got, tt.value)
			}
		})
	}
}

// TestMustFromString_PanicsOnUnknown verifies that MustFromString panics
// when given an unregistered name.
func TestMustFromString_PanicsOnUnknown(t *testing.T) {
	t.Helper()

	reg := registry.New[string]()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustFromString(%q) did not panic", "unknown")
		}
	}()

	_ = reg.MustFromString("unknown")
}

// TestRegistry_WithFunctionType verifies that Registry works with function types.
func TestRegistry_WithFunctionType(t *testing.T) {
	t.Helper()

	type EncoderFunc func(s string) string

	reg := registry.New[EncoderFunc]()

	uppercase := func(s string) string { return strings.ToUpper(s) }
	lowercase := func(s string) string { return strings.ToLower(s) }
	reverse := func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}

	reg.Register("uppercase", uppercase)
	reg.Register("lowercase", lowercase)
	reg.Register("reverse", reverse)

	tests := []struct {
		name      string
		encoder   string
		input     string
		wantOut   string
	}{
		{"uppercase", "uppercase", "hello", "HELLO"},
		{"lowercase", "lowercase", "WORLD", "world"},
		{"reverse", "reverse", "abc", "cba"},
		{"uppercase_mixed", "uppercase", "HeLLo", "HELLO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := reg.FromString(tt.encoder)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.encoder, err)
			}
			if got := enc(tt.input); got != tt.wantOut {
				t.Errorf("%s encoder(%q) = %q, want %q", tt.encoder, tt.input, got, tt.wantOut)
			}
		})
	}
}

// testFormatter is a test interface for TestRegistry_WithInterfaceType.
type testFormatter interface {
	Format(s string) string
}

// testUpperFormatter implements testFormatter with uppercase formatting.
type testUpperFormatter struct{}

func (testUpperFormatter) Format(s string) string { return strings.ToUpper(s) }

// testLowerFormatter implements testFormatter with lowercase formatting.
type testLowerFormatter struct{}

func (testLowerFormatter) Format(s string) string { return strings.ToLower(s) }

// TestRegistry_WithInterfaceType verifies that Registry works with interface types.
func TestRegistry_WithInterfaceType(t *testing.T) {
	t.Helper()

	reg := registry.New[testFormatter]()
	reg.Register("upper", testUpperFormatter{})
	reg.Register("lower", testLowerFormatter{})

	tests := []struct {
		name      string
		formatter string
		input     string
		wantOut   string
	}{
		{"upper_simple", "upper", "test", "TEST"},
		{"upper_mixed", "upper", "TeSt", "TEST"},
		{"lower_simple", "lower", "TEST", "test"},
		{"lower_mixed", "lower", "TeSt", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt, err := reg.FromString(tt.formatter)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.formatter, err)
			}
			if got := fmt.Format(tt.input); got != tt.wantOut {
				t.Errorf("%s formatter.Format(%q) = %q, want %q", tt.formatter, tt.input, got, tt.wantOut)
			}
		})
	}
}

// TestRegistry_MultipleInstances verifies that multiple registry instances
// are independent and don't interfere with each other.
func TestRegistry_MultipleInstances(t *testing.T) {
	t.Helper()

	tests := []struct {
		name   string
		key    string
		value1 string
		value2 string
	}{
		{"shared_key", "shared", "value1", "value2"},
		{"encoder_key", "encoder", "json", "console"},
		{"formatter_key", "formatter", "upper", "lower"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg1 := registry.New[string]()
			reg2 := registry.New[string]()

			reg1.Register(tt.key, tt.value1)
			reg2.Register(tt.key, tt.value2)

			got1, err := reg1.FromString(tt.key)
			if err != nil {
				t.Fatalf("reg1.FromString(%q) returned error: %v", tt.key, err)
			}
			if got1 != tt.value1 {
				t.Errorf("reg1.FromString(%q) = %q, want %q", tt.key, got1, tt.value1)
			}

			got2, err := reg2.FromString(tt.key)
			if err != nil {
				t.Fatalf("reg2.FromString(%q) returned error: %v", tt.key, err)
			}
			if got2 != tt.value2 {
				t.Errorf("reg2.FromString(%q) = %q, want %q", tt.key, got2, tt.value2)
			}
		})
	}
}

// TestRegistry_WithDifferentTypes verifies that Registry works correctly
// with various Go types.
func TestRegistry_WithDifferentTypes(t *testing.T) {
	t.Helper()

	t.Run("int", func(t *testing.T) {
		reg := registry.New[int]()
		reg.Register("one", 1)
		reg.Register("two", 2)
		reg.Register("hundred", 100)

		tests := []struct {
			key  string
			want int
		}{
			{"one", 1},
			{"two", 2},
			{"hundred", 100},
		}

		for _, tt := range tests {
			got, err := reg.FromString(tt.key)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.key, err)
			}
			if got != tt.want {
				t.Errorf("FromString(%q) = %d, want %d", tt.key, got, tt.want)
			}
		}
	})

	t.Run("bool", func(t *testing.T) {
		reg := registry.New[bool]()
		reg.Register("true", true)
		reg.Register("false", false)

		tests := []struct {
			key  string
			want bool
		}{
			{"true", true},
			{"false", false},
		}

		for _, tt := range tests {
			got, err := reg.FromString(tt.key)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.key, err)
			}
			if got != tt.want {
				t.Errorf("FromString(%q) = %v, want %v", tt.key, got, tt.want)
			}
		}
	})

	t.Run("struct", func(t *testing.T) {
		type Config struct {
			Name  string
			Value int
		}

		reg := registry.New[Config]()
		reg.Register("dev", Config{Name: "development", Value: 1})
		reg.Register("prod", Config{Name: "production", Value: 2})

		tests := []struct {
			key  string
			want Config
		}{
			{"dev", Config{Name: "development", Value: 1}},
			{"prod", Config{Name: "production", Value: 2}},
		}

		for _, tt := range tests {
			got, err := reg.FromString(tt.key)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.key, err)
			}
			if got != tt.want {
				t.Errorf("FromString(%q) = %+v, want %+v", tt.key, got, tt.want)
			}
		}
	})
}
