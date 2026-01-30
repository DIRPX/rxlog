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

package name_test

import (
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
	namepkg "dirpx.dev/rxlog/rxcore/name"
)

// TestFromString_RegisteredEncoders verifies that all built-in name encoders
// can be looked up by their registered names.
func TestFromString_RegisteredEncoders(t *testing.T) {
	t.Helper()

	testName := "app.service.handler"

	tests := []struct {
		name string
		want string // expected output for testName
	}{
		{"full", "app.service.handler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := namepkg.FromString(tt.name)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.name, err)
			}
			if enc == nil {
				t.Fatalf("FromString(%q) returned nil encoder", tt.name)
			}

			// Verify the encoder produces expected output.
			dst := &buffer.Buffer{}
			out := enc(dst, testName)
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

	_, err := namepkg.FromString("nonexistent")
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

	_ = namepkg.MustFromString("unknown")
}

// TestMustFromString_ReturnsEncoderOnSuccess verifies that MustFromString
// returns a valid encoder for registered names without panicking.
func TestMustFromString_ReturnsEncoderOnSuccess(t *testing.T) {
	t.Helper()

	enc := namepkg.MustFromString("full")
	if enc == nil {
		t.Fatal("MustFromString(\"full\") returned nil encoder")
	}

	dst := &buffer.Buffer{}
	out := enc(dst, "test.logger")
	got := string(out.Bytes())
	want := "test.logger"
	if got != want {
		t.Errorf("encoder from MustFromString(\"full\") = %q, want %q", got, want)
	}
}

// TestRegister_AddsCustomEncoder verifies that Register allows adding
// custom encoders that can then be retrieved via FromString.
func TestRegister_AddsCustomEncoder(t *testing.T) {
	t.Helper()

	const customName = "test-custom-name"

	// Register a custom encoder that prefixes with "LOGGER:".
	customEnc := func(dst *buffer.Buffer, n string) *buffer.Buffer {
		dst.AppendString("LOGGER:")
		dst.AppendString(n)
		return dst
	}

	namepkg.Register(customName, customEnc)

	// Verify the custom encoder can be retrieved.
	enc, err := namepkg.FromString(customName)
	if err != nil {
		t.Fatalf("FromString(%q) after Register returned error: %v", customName, err)
	}

	dst := &buffer.Buffer{}
	out := enc(dst, "myapp")
	got := string(out.Bytes())
	want := "LOGGER:myapp"
	if got != want {
		t.Errorf("custom encoder output = %q, want %q", got, want)
	}
}

// TestRegister_OverwritesExistingEncoder verifies that Register replaces
// an existing encoder when the same name is used.
func TestRegister_OverwritesExistingEncoder(t *testing.T) {
	t.Helper()

	const testName = "test-overwrite-name"

	// Register initial encoder.
	initialEnc := func(dst *buffer.Buffer, n string) *buffer.Buffer {
		dst.AppendString("initial:")
		dst.AppendString(n)
		return dst
	}
	namepkg.Register(testName, initialEnc)

	// Overwrite with a different encoder.
	replacementEnc := func(dst *buffer.Buffer, n string) *buffer.Buffer {
		dst.AppendString("replaced:")
		dst.AppendString(n)
		return dst
	}
	namepkg.Register(testName, replacementEnc)

	// Verify the replacement encoder is now used.
	enc, err := namepkg.FromString(testName)
	if err != nil {
		t.Fatalf("FromString(%q) after replacement returned error: %v", testName, err)
	}

	dst := &buffer.Buffer{}
	out := enc(dst, "app")
	got := string(out.Bytes())
	want := "replaced:app"
	if got != want {
		t.Errorf("encoder after overwrite = %q, want %q", got, want)
	}
}
