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

package filectl_test

import (
	"testing"

	filectl "dirpx.dev/rxlog/levelctl/file"
	"dirpx.dev/rxlog/rxapi/level"
)

// TestWithPath verifies that WithPath assigns the provided path
// to Handler.Path without additional normalization.
//
// Normalization of empty paths (falling back to DefaultPath) is
// performed by NewHandler or Apply, not by the option itself.
func TestWithPath(t *testing.T) {
	var h filectl.Handler

	// Non-empty path is stored as-is.
	filectl.WithPath("/etc/myapp/log.level")(&h)
	if h.Path != "/etc/myapp/log.level" {
		t.Fatalf("expected Path %q, got %q", "/etc/myapp/log.level", h.Path)
	}

	// Empty path is also stored as-is; fallback to DefaultPath is
	// applied by NewHandler, not by the option.
	filectl.WithPath("")(&h)
	if h.Path != "" {
		t.Fatalf("expected Path to be empty, got %q", h.Path)
	}
}

// TestWithDefaultLevel verifies that WithDefaultLevel sets DefaultLevel
// and marks HasDefault as true.
func TestWithDefaultLevel(t *testing.T) {
	var h filectl.Handler

	if h.HasDefault {
		t.Fatalf("expected HasDefault to be false by default")
	}

	filectl.WithDefaultLevel(level.Warn)(&h)

	if !h.HasDefault {
		t.Fatalf("expected HasDefault to be true after WithDefaultLevel")
	}
	if h.DefaultLevel != level.Warn {
		t.Fatalf("expected DefaultLevel %v, got %v", level.Warn, h.DefaultLevel)
	}
}

// TestWithRequired verifies that WithRequired toggles the Required flag
// exactly to the provided value.
func TestWithRequired(t *testing.T) {
	var h filectl.Handler

	filectl.WithRequired(true)(&h)
	if !h.Required {
		t.Fatalf("expected Required to be true, got false")
	}

	filectl.WithRequired(false)(&h)
	if h.Required {
		t.Fatalf("expected Required to be false, got true")
	}
}

// TestWithReadFile verifies that WithReadFile installs the provided ReadFile
// function into the Handler and that the stored function is actually invoked.
func TestWithReadFile(t *testing.T) {
	var h filectl.Handler

	called := false
	readFile := func(path string) ([]byte, error) {
		called = true
		return []byte("info"), nil
	}

	filectl.WithReadFile(readFile)(&h)

	if h.ReadFile == nil {
		t.Fatalf("expected ReadFile to be set")
	}

	// Invoke the stored function to ensure it points to the provided one.
	if _, err := h.ReadFile("ignored"); err != nil {
		t.Fatalf("unexpected error from ReadFile: %v", err)
	}
	if !called {
		t.Fatalf("expected stored ReadFile to call provided function")
	}
}
