/*
   Copyright 2026 The DIRPX Authors.

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

package caller_test

import (
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/caller"
	callerpkg "dirpx.dev/rxlog/rxcore/caller"
)

// TestFromString_RegisteredEncoders verifies that all built-in caller encoders
// can be looked up by their registered names.
func TestFromString_RegisteredEncoders(t *testing.T) {
	t.Helper()

	testCaller := caller.Caller{
		Defined: true,
		File:    "/app/handlers/user.go",
		Line:    42,
	}

	tests := []struct {
		name string
		want string // expected output for testCaller
	}{
		{"short", "user.go:42"},
		{"full", "/app/handlers/user.go:42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := callerpkg.FromString(tt.name)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.name, err)
			}
			if enc == nil {
				t.Fatalf("FromString(%q) returned nil encoder", tt.name)
			}

			// Verify the encoder produces expected output.
			dst := &buffer.Buffer{}
			out := enc(dst, testCaller)
			got := string(out.Bytes())
			if got != tt.want {
				t.Errorf("encoder(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// TestFromString_UnknownEncoder verifies that FromString returns an error
// for unregistered encoder names.
func TestFromString_UnknownEncoder(t *testing.T) {
	t.Helper()

	_, err := callerpkg.FromString("nonexistent")
	if err == nil {
		t.Fatal("FromString(\"nonexistent\") returned nil error, want error")
	}
}

// TestMustFromString_PanicsOnUnknown verifies that MustFromString panics
// when given an unregistered encoder name.
func TestMustFromString_PanicsOnUnknown(t *testing.T) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustFromString(\"unknown\") did not panic")
		}
	}()

	_ = callerpkg.MustFromString("unknown")
}

// TestMustFromString_ReturnsEncoderOnSuccess verifies that MustFromString
// returns a valid encoder for registered names without panicking.
func TestMustFromString_ReturnsEncoderOnSuccess(t *testing.T) {
	t.Helper()

	enc := callerpkg.MustFromString("short")
	if enc == nil {
		t.Fatal("MustFromString(\"short\") returned nil encoder")
	}

	testCaller := caller.Caller{
		Defined: true,
		File:    "/src/main.go",
		Line:    10,
	}

	dst := &buffer.Buffer{}
	out := enc(dst, testCaller)
	got := string(out.Bytes())
	want := "main.go:10"
	if got != want {
		t.Errorf("encoder from MustFromString(\"short\") = %q, want %q", got, want)
	}
}

// TestRegister_AddsCustomEncoder verifies that Register allows adding
// custom encoders that can then be retrieved via FromString.
func TestRegister_AddsCustomEncoder(t *testing.T) {
	t.Helper()

	const customName = "test-custom-caller"

	// Register a custom encoder that prefixes with "CALLER:".
	customEnc := func(dst *buffer.Buffer, c caller.Caller) *buffer.Buffer {
		dst.AppendString("CALLER:")
		if c.Defined {
			dst.AppendString(c.File)
		}
		return dst
	}

	callerpkg.Register(customName, customEnc)

	// Verify the custom encoder can be retrieved.
	enc, err := callerpkg.FromString(customName)
	if err != nil {
		t.Fatalf("FromString(%q) after Register returned error: %v", customName, err)
	}

	testCaller := caller.Caller{
		Defined: true,
		File:    "/test.go",
		Line:    5,
	}

	dst := &buffer.Buffer{}
	out := enc(dst, testCaller)
	got := string(out.Bytes())
	want := "CALLER:/test.go"
	if got != want {
		t.Errorf("custom encoder output = %q, want %q", got, want)
	}
}

// TestRegister_OverwritesExistingEncoder verifies that Register replaces
// an existing encoder when the same name is used.
func TestRegister_OverwritesExistingEncoder(t *testing.T) {
	t.Helper()

	const testName = "test-overwrite-caller"

	testCaller := caller.Caller{
		Defined: true,
		File:    "/app/main.go",
		Line:    99,
	}

	// Register initial encoder.
	initialEnc := func(dst *buffer.Buffer, c caller.Caller) *buffer.Buffer {
		dst.AppendString("initial")
		return dst
	}
	callerpkg.Register(testName, initialEnc)

	// Overwrite with a different encoder.
	replacementEnc := func(dst *buffer.Buffer, c caller.Caller) *buffer.Buffer {
		dst.AppendString("replaced")
		return dst
	}
	callerpkg.Register(testName, replacementEnc)

	// Verify the replacement encoder is now used.
	enc, err := callerpkg.FromString(testName)
	if err != nil {
		t.Fatalf("FromString(%q) after replacement returned error: %v", testName, err)
	}

	dst := &buffer.Buffer{}
	out := enc(dst, testCaller)
	got := string(out.Bytes())
	want := "replaced"
	if got != want {
		t.Errorf("encoder after overwrite = %q, want %q", got, want)
	}
}
