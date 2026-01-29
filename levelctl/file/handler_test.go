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
	"errors"
	"io/fs"
	"testing"

	filectl "dirpx.dev/rxlog/levelctl/file"
	"dirpx.dev/rxlog/rxapi/level"
)

// fakeThreshold is a minimal in-memory implementation of level.Threshold
// used exclusively for tests. It is NOT concurrency-safe and must only be
// used from a single goroutine within a test.
type fakeThreshold struct {
	cur level.Level
}

func (f *fakeThreshold) Enabled(l level.Level) bool {
	return l >= f.cur
}

func (f *fakeThreshold) Level() level.Level {
	return f.cur
}

func (f *fakeThreshold) SetLevel(l level.Level) {
	f.cur = l
}

// TestHandler_ApplyNilLevel verifies that Apply fails when the Handler's
// Level field is nil.
func TestHandler_ApplyNilLevel(t *testing.T) {
	h := filectl.NewHandler(nil, "")

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error when Level is nil, got nil")
	}
}

// TestHandler_ApplyMissingNoDefault ensures that when the configuration
// file is missing, no default is configured, and Required is false,
// Apply leaves the threshold unchanged and returns nil.
func TestHandler_ApplyMissingNoDefault(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := filectl.NewHandler(th, "ignored",
		filectl.WithReadFile(func(path string) ([]byte, error) {
			return nil, fs.ErrNotExist
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to remain %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyMissingWithDefault ensures that when the configuration
// file is missing but a default level is configured, Apply uses the default.
func TestHandler_ApplyMissingWithDefault(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := filectl.NewHandler(th, "ignored",
		filectl.WithDefaultLevel(level.Warn),
		filectl.WithReadFile(func(path string) ([]byte, error) {
			return nil, fs.ErrNotExist
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Warn {
		t.Fatalf("expected level to be updated to %v, got %v", level.Warn, got)
	}
}

// TestHandler_ApplyMissingRequired ensures that when Required is true and
// the configuration file is missing, Apply returns an error and does not
// change the current threshold.
func TestHandler_ApplyMissingRequired(t *testing.T) {
	th := &fakeThreshold{cur: level.Error}

	h := filectl.NewHandler(th, "ignored",
		filectl.WithRequired(true),
		filectl.WithReadFile(func(path string) ([]byte, error) {
			return nil, fs.ErrNotExist
		}),
	)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error for missing required file, got nil")
	}

	if got := th.Level(); got != level.Error {
		t.Fatalf("expected level to remain %v, got %v", level.Error, got)
	}
}

// TestHandler_ApplyReadError ensures that non-NotExist read errors are
// treated as hard failures regardless of Required/Default configuration.
func TestHandler_ApplyReadError(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := filectl.NewHandler(th, "ignored",
		filectl.WithDefaultLevel(level.Warn),
		filectl.WithReadFile(func(path string) ([]byte, error) {
			return nil, errors.New("I/O failure")
		}),
	)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error for read failure, got nil")
	}

	// Threshold MUST remain unchanged.
	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to remain %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyValidValue ensures that a valid level string in the
// configuration file is parsed, normalized, and applied.
func TestHandler_ApplyValidValue(t *testing.T) {
	th := &fakeThreshold{cur: level.Warn}

	const path = "/etc/myapp/log.level"

	h := filectl.NewHandler(th, path,
		filectl.WithReadFile(func(p string) ([]byte, error) {
			if p != path {
				t.Fatalf("expected read from %q, got %q", path, p)
			}
			// Trailing newline is expected to be trimmed.
			return []byte("info\n"), nil
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to be updated to %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyEmptyWithDefault ensures that an existing file whose
// contents are effectively empty triggers use of the default level when
// configured.
func TestHandler_ApplyEmptyWithDefault(t *testing.T) {
	th := &fakeThreshold{cur: level.Error}

	h := filectl.NewHandler(th, "ignored",
		filectl.WithDefaultLevel(level.Info),
		filectl.WithReadFile(func(path string) ([]byte, error) {
			return []byte("   \n\t"), nil // effectively empty after trimming
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to be updated to %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyEmptyRequired ensures that an existing but empty file
// results in an error when Required is true, leaving the threshold
// unchanged.
func TestHandler_ApplyEmptyRequired(t *testing.T) {
	th := &fakeThreshold{cur: level.Warn}

	h := filectl.NewHandler(th, "ignored",
		filectl.WithRequired(true),
		filectl.WithReadFile(func(path string) ([]byte, error) {
			return []byte("   "), nil
		}),
	)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error for empty required file, got nil")
	}

	if got := th.Level(); got != level.Warn {
		t.Fatalf("expected level to remain %v, got %v", level.Warn, got)
	}
}

// TestHandler_ApplyInvalidValue ensures that an invalid level string in
// the configuration file results in an error and does not change the
// current threshold.
func TestHandler_ApplyInvalidValue(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := filectl.NewHandler(th, "ignored",
		filectl.WithReadFile(func(path string) ([]byte, error) {
			return []byte("not-a-level"), nil
		}),
	)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error for invalid level value, got nil")
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to remain %v, got %v", level.Info, got)
	}
}

// TestHandler_MustApplyPanicsOnError verifies that MustApply panics when
// Apply would return an error.
func TestHandler_MustApplyPanicsOnError(t *testing.T) {
	h := filectl.NewHandler(nil, "")

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic from MustApply, got none")
		}
	}()

	h.MustApply()
}
