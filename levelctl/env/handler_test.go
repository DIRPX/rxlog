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

package envctl_test

import (
	"strings"
	"testing"

	envctl "dirpx.dev/rxlog/levelctl/env"
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
	h := envctl.NewHandler(nil)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error when Level is nil, got nil")
	}
}

// TestHandler_ApplyMissingNoDefault ensures that when the environment
// variable is missing, no default is configured, and Required is false,
// Apply leaves the threshold unchanged and returns nil.
func TestHandler_ApplyMissingNoDefault(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := envctl.NewHandler(th,
		envctl.WithLookup(func(name string) (string, bool) {
			return "", false // simulate missing variable
		}),
		// No WithDefaultLevel, no WithRequired.
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to remain %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyMissingWithDefault ensures that when the environment
// variable is missing but a default level is configured, Apply uses the
// default.
func TestHandler_ApplyMissingWithDefault(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := envctl.NewHandler(th,
		envctl.WithDefaultLevel(level.Warn),
		envctl.WithLookup(func(name string) (string, bool) {
			return "", false // simulate missing variable
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
// the environment variable is missing, Apply returns an error and does not
// change the current threshold.
func TestHandler_ApplyMissingRequired(t *testing.T) {
	th := &fakeThreshold{cur: level.Error}

	h := envctl.NewHandler(th,
		envctl.WithRequired(true),
		envctl.WithLookup(func(name string) (string, bool) {
			return "", false // simulate missing variable
		}),
	)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error for missing required env var, got nil")
	}

	if got := th.Level(); got != level.Error {
		t.Fatalf("expected level to remain %v, got %v", level.Error, got)
	}
}

// TestHandler_ApplyValidValue ensures that a valid level string in the
// environment variable is parsed, normalized, and applied.
func TestHandler_ApplyValidValue(t *testing.T) {
	th := &fakeThreshold{cur: level.Warn}

	h := envctl.NewHandler(th,
		envctl.WithLookup(func(name string) (string, bool) {
			if name != envctl.DefaultVarName {
				t.Fatalf("expected lookup to use %q, got %q", envctl.DefaultVarName, name)
			}
			// Leading/trailing whitespace should be trimmed.
			return "  info\n", true
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to be updated to %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyInvalidValue ensures that an invalid level string in
// the environment variable results in an error and does not change the
// current threshold.
func TestHandler_ApplyInvalidValue(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := envctl.NewHandler(th,
		envctl.WithLookup(func(name string) (string, bool) {
			return "not-a-level", true
		}),
	)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error for invalid level string, got nil")
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to remain %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyUsesVarName verifies that WithVarName overrides the
// environment variable name used by Apply.
func TestHandler_ApplyUsesVarName(t *testing.T) {
	const customName = "MYAPP_LOG_LEVEL"

	th := &fakeThreshold{cur: level.Warn}
	var (
		lookupCalled bool
		seenName     string
	)

	h := envctl.NewHandler(th,
		envctl.WithVarName(customName),
		envctl.WithLookup(func(name string) (string, bool) {
			lookupCalled = true
			seenName = name
			return "debug", true
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !lookupCalled {
		t.Fatalf("expected Lookup to be called")
	}
	if seenName != customName {
		t.Fatalf("expected Lookup to be called with %q, got %q", customName, seenName)
	}
	if got := th.Level(); got != level.Debug {
		t.Fatalf("expected level to be updated to %v, got %v", level.Debug, got)
	}
}

// TestHandler_ApplyEmptyValueWithDefault ensures that an explicitly set
// but empty/whitespace-only environment value is treated as "missing"
// and therefore uses the default level when configured.
func TestHandler_ApplyEmptyValueWithDefault(t *testing.T) {
	th := &fakeThreshold{cur: level.Error}

	h := envctl.NewHandler(th,
		envctl.WithDefaultLevel(level.Info),
		envctl.WithLookup(func(name string) (string, bool) {
			return "   \n", true // effectively empty
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to be updated to %v, got %v", level.Info, got)
	}
}

// TestHandler_MustApplyPanicsOnError verifies that MustApply panics when
// Apply would return an error.
func TestHandler_MustApplyPanicsOnError(t *testing.T) {
	h := envctl.NewHandler(nil) // Level is nil → Apply will fail

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic from MustApply, got none")
		}
	}()

	h.MustApply()
}

// TestHandler_ApplyMultipleCalls ensures that repeated calls to Apply
// re-evaluate the environment and update the threshold accordingly.
func TestHandler_ApplyMultipleCalls(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	values := []string{"warn", "error"}
	var idx int

	h := envctl.NewHandler(th,
		envctl.WithLookup(func(name string) (string, bool) {
			if idx >= len(values) {
				return "", false
			}
			v := values[idx]
			idx++
			return v, true
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("first Apply returned error: %v", err)
	}
	if got := th.Level(); got != level.Warn {
		t.Fatalf("after first Apply expected %v, got %v", level.Warn, got)
	}

	if err := h.Apply(); err != nil {
		t.Fatalf("second Apply returned error: %v", err)
	}
	if got := th.Level(); got != level.Error {
		t.Fatalf("after second Apply expected %v, got %v", level.Error, got)
	}
}

// TestHandler_ApplyTrimCase ensures that Apply trims whitespace and
// accepts case-insensitive level names.
func TestHandler_ApplyTrimCase(t *testing.T) {
	th := &fakeThreshold{cur: level.Warn}

	h := envctl.NewHandler(th,
		envctl.WithLookup(func(name string) (string, bool) {
			return strings.ToUpper("info"), true
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to be updated to %v, got %v", level.Info, got)
	}
}
