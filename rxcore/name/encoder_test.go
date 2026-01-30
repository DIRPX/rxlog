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
	"dirpx.dev/rxlog/rxapi/name"
	namepkg "dirpx.dev/rxlog/rxcore/name"
)

// encodeWith is a helper that runs the given name encoder on a fresh buffer
// and returns the resulting string.
func encodeWith(t *testing.T, enc name.Encoder, n string) string {
	t.Helper()

	dst := &buffer.Buffer{}
	out := enc(dst, n)
	if out == nil {
		t.Fatalf("encoder returned nil buffer")
	}

	return string(out.Bytes())
}

// TestFullNameEncoder_NonEmptyName verifies that FullNameEncoder outputs
// the complete name as-is.
func TestFullNameEncoder_NonEmptyName(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		want string
	}{
		{"simple", "simple"},
		{"service.handler", "service.handler"},
		{"app.module.component.subcomponent", "app.module.component.subcomponent"},
		{"with-dash", "with-dash"},
		{"with_underscore", "with_underscore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeWith(t, namepkg.FullNameEncoder, tt.name)
			if got != tt.want {
				t.Errorf("FullNameEncoder(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// TestFullNameEncoder_EmptyName verifies that FullNameEncoder appends
// nothing when given an empty name.
func TestFullNameEncoder_EmptyName(t *testing.T) {
	t.Helper()

	got := encodeWith(t, namepkg.FullNameEncoder, "")
	if got != "" {
		t.Errorf("FullNameEncoder(\"\") = %q, want empty string", got)
	}
}

// TestFullNameEncoder_AppendsToExistingBuffer verifies that the encoder
// appends its output to the existing contents of the buffer.
func TestFullNameEncoder_AppendsToExistingBuffer(t *testing.T) {
	t.Helper()

	dst := &buffer.Buffer{}
	dst.AppendString("logger=")

	out := namepkg.FullNameEncoder(dst, "app.service")
	if out == nil {
		t.Fatal("FullNameEncoder returned nil buffer")
	}

	got := string(out.Bytes())
	want := "logger=app.service"
	if got != want {
		t.Errorf("buffer after encoding = %q, want %q", got, want)
	}
}

// TestFullNameEncoder_CanBeReusedAcrossCalls verifies that a single encoder
// instance is safe to reuse across multiple calls with different names.
func TestFullNameEncoder_CanBeReusedAcrossCalls(t *testing.T) {
	t.Helper()

	enc := namepkg.FullNameEncoder

	// First call.
	buf1 := &buffer.Buffer{}
	out1 := enc(buf1, "first.logger")
	if out1 == nil {
		t.Fatal("first call to encoder returned nil buffer")
	}
	got1 := string(out1.Bytes())
	want1 := "first.logger"
	if got1 != want1 {
		t.Errorf("first call: got %q, want %q", got1, want1)
	}

	// Second call with different name and a fresh buffer.
	buf2 := &buffer.Buffer{}
	out2 := enc(buf2, "second.logger")
	if out2 == nil {
		t.Fatal("second call to encoder returned nil buffer")
	}
	got2 := string(out2.Bytes())
	want2 := "second.logger"
	if got2 != want2 {
		t.Errorf("second call: got %q, want %q", got2, want2)
	}
}
