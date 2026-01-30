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

package stack_test

import (
	"testing"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/stack"
	stackpkg "dirpx.dev/rxlog/rxcore/stack"
)

// encodeWith is a helper that runs the given stack encoder on a fresh buffer
// and returns the resulting string.
func encodeWith(t *testing.T, enc stack.Encoder, s string) string {
	t.Helper()

	dst := &buffer.Buffer{}
	out := enc(dst, s)
	if out == nil {
		t.Fatalf("encoder returned nil buffer")
	}

	return string(out.Bytes())
}

// TestFullStackEncoder_NonEmptyStack verifies that FullStackEncoder outputs
// the complete stack trace as-is.
func TestFullStackEncoder_NonEmptyStack(t *testing.T) {
	t.Helper()

	tests := []struct {
		name  string
		stack string
		want  string
	}{
		{
			name:  "single line",
			stack: "goroutine 1 [running]:",
			want:  "goroutine 1 [running]:",
		},
		{
			name: "multi-line stack",
			stack: `goroutine 1 [running]:
main.foo(0x1, 0x2)
  /app/main.go:42 +0x123
main.main()
  /app/main.go:10 +0x456`,
			want: `goroutine 1 [running]:
main.foo(0x1, 0x2)
  /app/main.go:42 +0x123
main.main()
  /app/main.go:10 +0x456`,
		},
		{
			name:  "with tabs",
			stack: "main.go:10\n\tmain.foo()",
			want:  "main.go:10\n\tmain.foo()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeWith(t, stackpkg.FullStackEncoder, tt.stack)
			if got != tt.want {
				t.Errorf("FullStackEncoder = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFullStackEncoder_EmptyStack verifies that FullStackEncoder appends
// nothing when given an empty stack trace.
func TestFullStackEncoder_EmptyStack(t *testing.T) {
	t.Helper()

	got := encodeWith(t, stackpkg.FullStackEncoder, "")
	if got != "" {
		t.Errorf("FullStackEncoder(\"\") = %q, want empty string", got)
	}
}

// TestFullStackEncoder_AppendsToExistingBuffer verifies that the encoder
// appends its output to the existing contents of the buffer.
func TestFullStackEncoder_AppendsToExistingBuffer(t *testing.T) {
	t.Helper()

	stack := "goroutine 1:\nmain.go:10"

	dst := &buffer.Buffer{}
	dst.AppendString("stack=")

	out := stackpkg.FullStackEncoder(dst, stack)
	if out == nil {
		t.Fatal("FullStackEncoder returned nil buffer")
	}

	got := string(out.Bytes())
	want := "stack=goroutine 1:\nmain.go:10"
	if got != want {
		t.Errorf("buffer after encoding = %q, want %q", got, want)
	}
}

// TestFullStackEncoder_CanBeReusedAcrossCalls verifies that a single encoder
// instance is safe to reuse across multiple calls with different stack traces.
func TestFullStackEncoder_CanBeReusedAcrossCalls(t *testing.T) {
	t.Helper()

	enc := stackpkg.FullStackEncoder

	// First call.
	buf1 := &buffer.Buffer{}
	stack1 := "first stack trace"
	out1 := enc(buf1, stack1)
	if out1 == nil {
		t.Fatal("first call to encoder returned nil buffer")
	}
	got1 := string(out1.Bytes())
	want1 := "first stack trace"
	if got1 != want1 {
		t.Errorf("first call: got %q, want %q", got1, want1)
	}

	// Second call with different stack and a fresh buffer.
	buf2 := &buffer.Buffer{}
	stack2 := "second stack trace"
	out2 := enc(buf2, stack2)
	if out2 == nil {
		t.Fatal("second call to encoder returned nil buffer")
	}
	got2 := string(out2.Bytes())
	want2 := "second stack trace"
	if got2 != want2 {
		t.Errorf("second call: got %q, want %q", got2, want2)
	}
}
