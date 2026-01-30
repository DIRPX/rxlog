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

package level_test

import (
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/level"
	levelpkg "dirpx.dev/rxlog/rxcore/level"
)

// TestFromString_RegisteredEncoders verifies that all built-in level encoders
// can be looked up by their registered names.
func TestFromString_RegisteredEncoders(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		want string // expected output for level.Info
	}{
		{"lowercase", "info"},
		{"lower", "info"},
		{"capital", "INFO"},
		{"uppercase", "INFO"},
		{"upper", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := levelpkg.FromString(tt.name)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.name, err)
			}
			if enc == nil {
				t.Fatalf("FromString(%q) returned nil encoder", tt.name)
			}

			// Verify the encoder produces expected output.
			dst := &buffer.Buffer{}
			out := enc(dst, level.Info)
			got := string(out.Bytes())
			if got != tt.want {
				t.Errorf("encoder(%q) for Info = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// TestFromString_UnknownEncoder verifies that FromString returns an error
// for unregistered encoder names.
func TestFromString_UnknownEncoder(t *testing.T) {
	t.Helper()

	_, err := levelpkg.FromString("nonexistent")
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

	_ = levelpkg.MustFromString("unknown")
}

// TestMustFromString_ReturnsEncoderOnSuccess verifies that MustFromString
// returns a valid encoder for registered names without panicking.
func TestMustFromString_ReturnsEncoderOnSuccess(t *testing.T) {
	t.Helper()

	enc := levelpkg.MustFromString("lowercase")
	if enc == nil {
		t.Fatal("MustFromString(\"lowercase\") returned nil encoder")
	}

	dst := &buffer.Buffer{}
	out := enc(dst, level.Error)
	got := string(out.Bytes())
	want := "error"
	if got != want {
		t.Errorf("encoder from MustFromString(\"lowercase\") = %q, want %q", got, want)
	}
}

// TestRegister_AddsCustomEncoder verifies that Register allows adding
// custom encoders that can then be retrieved via FromString.
func TestRegister_AddsCustomEncoder(t *testing.T) {
	t.Helper()

	const customName = "test-custom-level"

	// Register a custom encoder that prefixes level with "custom:".
	customEnc := func(dst *buffer.Buffer, l level.Level) *buffer.Buffer {
		dst.AppendString("custom:")
		dst.AppendString(l.String())
		return dst
	}

	levelpkg.Register(customName, customEnc)

	// Verify the custom encoder can be retrieved.
	enc, err := levelpkg.FromString(customName)
	if err != nil {
		t.Fatalf("FromString(%q) after Register returned error: %v", customName, err)
	}

	dst := &buffer.Buffer{}
	out := enc(dst, level.Warn)
	got := string(out.Bytes())
	want := "custom:warn"
	if got != want {
		t.Errorf("custom encoder output = %q, want %q", got, want)
	}
}

// TestRegister_OverwritesExistingEncoder verifies that Register replaces
// an existing encoder when the same name is used.
func TestRegister_OverwritesExistingEncoder(t *testing.T) {
	t.Helper()

	const testName = "test-overwrite-level"

	// Register initial encoder.
	initialEnc := func(dst *buffer.Buffer, l level.Level) *buffer.Buffer {
		dst.AppendString("initial:")
		dst.AppendString(l.String())
		return dst
	}
	levelpkg.Register(testName, initialEnc)

	// Overwrite with a different encoder.
	replacementEnc := func(dst *buffer.Buffer, l level.Level) *buffer.Buffer {
		dst.AppendString("replaced:")
		dst.AppendString(l.String())
		return dst
	}
	levelpkg.Register(testName, replacementEnc)

	// Verify the replacement encoder is now used.
	enc, err := levelpkg.FromString(testName)
	if err != nil {
		t.Fatalf("FromString(%q) after replacement returned error: %v", testName, err)
	}

	dst := &buffer.Buffer{}
	out := enc(dst, level.Debug)
	got := string(out.Bytes())
	want := "replaced:debug"
	if got != want {
		t.Errorf("encoder after overwrite = %q, want %q", got, want)
	}
}
