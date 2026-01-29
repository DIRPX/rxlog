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

package clictl_test

import (
	"os"
	"testing"

	clictl "dirpx.dev/rxlog/levelctl/cli"
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
	h := clictl.NewHandler(nil)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error when Level is nil, got nil")
	}
}

// TestHandler_ApplyNoFlagNoDefault ensures that when the CLI flag is
// absent, no default is configured, and Required is false, Apply leaves
// the threshold unchanged and returns nil.
func TestHandler_ApplyNoFlagNoDefault(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := clictl.NewHandler(th,
		clictl.WithArgs([]string{
			"--other-flag", "value", // unrelated flags must be ignored
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

// TestHandler_ApplyNoFlagWithDefault ensures that when the CLI flag is
// absent but a default level is configured, Apply uses the default.
func TestHandler_ApplyNoFlagWithDefault(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := clictl.NewHandler(th,
		clictl.WithDefaultLevel(level.Warn),
		clictl.WithArgs([]string{
			"--unrelated", "x",
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Warn {
		t.Fatalf("expected level to be updated to %v, got %v", level.Warn, got)
	}
}

// TestHandler_ApplyNoFlagRequired ensures that when Required is true and
// the CLI flag is absent, Apply returns an error and does not change the
// current threshold.
func TestHandler_ApplyNoFlagRequired(t *testing.T) {
	th := &fakeThreshold{cur: level.Error}

	h := clictl.NewHandler(th,
		clictl.WithRequired(true),
		clictl.WithArgs([]string{
			"--another-flag=info",
		}),
	)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error for missing required flag, got nil")
	}

	if got := th.Level(); got != level.Error {
		t.Fatalf("expected level to remain %v, got %v", level.Error, got)
	}
}

// TestHandler_ApplyLongFlagEquals ensures that a valid level specified as
// "--log-level=info" is parsed and applied.
func TestHandler_ApplyLongFlagEquals(t *testing.T) {
	th := &fakeThreshold{cur: level.Warn}

	h := clictl.NewHandler(th,
		clictl.WithFlagName("log-level"),
		clictl.WithArgs([]string{
			"--log-level=info",
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to be updated to %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyLongFlagSeparate ensures that a valid level specified as
// "--log-level info" is parsed and applied.
func TestHandler_ApplyLongFlagSeparate(t *testing.T) {
	th := &fakeThreshold{cur: level.Error}

	h := clictl.NewHandler(th,
		clictl.WithFlagName("log-level"),
		clictl.WithArgs([]string{
			"--log-level", "debug",
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Debug {
		t.Fatalf("expected level to be updated to %v, got %v", level.Debug, got)
	}
}

// TestHandler_ApplyShortFlagEquals ensures that a valid level specified as
// "-L=info" is parsed and applied when ShortFlag is configured.
func TestHandler_ApplyShortFlagEquals(t *testing.T) {
	th := &fakeThreshold{cur: level.Warn}

	h := clictl.NewHandler(th,
		clictl.WithFlagName("log-level"),
		clictl.WithShortFlag("L"),
		clictl.WithArgs([]string{
			"-L=info",
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to be updated to %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyShortFlagSeparate ensures that a valid level specified as
// "-L info" is parsed and applied when ShortFlag is configured.
func TestHandler_ApplyShortFlagSeparate(t *testing.T) {
	th := &fakeThreshold{cur: level.Warn}

	h := clictl.NewHandler(th,
		clictl.WithFlagName("log-level"),
		clictl.WithShortFlag("L"),
		clictl.WithArgs([]string{
			"-L", "error",
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Error {
		t.Fatalf("expected level to be updated to %v, got %v", level.Error, got)
	}
}

// TestHandler_ApplyMultipleOccurrences ensures that when the flag appears
// multiple times, the last occurrence wins (pflag semantics).
func TestHandler_ApplyMultipleOccurrences(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := clictl.NewHandler(th,
		clictl.WithArgs([]string{
			"--log-level=debug",
			"--log-level=error",
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Error {
		t.Fatalf("expected level to be updated to %v, got %v", level.Error, got)
	}
}

// TestHandler_ApplyEmptyValueWithDefault ensures that an explicitly set but
// empty value (for example, "--log-level=") is treated as empty and triggers
// default-level behavior when configured.
func TestHandler_ApplyEmptyValueWithDefault(t *testing.T) {
	th := &fakeThreshold{cur: level.Debug}

	h := clictl.NewHandler(th,
		clictl.WithDefaultLevel(level.Info),
		clictl.WithArgs([]string{
			"--log-level=",
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to be updated to %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyEmptyValueRequired ensures that an explicitly set but
// empty value causes an error when Required is true and does not change
// the current threshold.
func TestHandler_ApplyEmptyValueRequired(t *testing.T) {
	th := &fakeThreshold{cur: level.Warn}

	h := clictl.NewHandler(th,
		clictl.WithRequired(true),
		clictl.WithArgs([]string{
			"--log-level=",
		}),
	)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error for empty required flag, got nil")
	}

	if got := th.Level(); got != level.Warn {
		t.Fatalf("expected level to remain %v, got %v", level.Warn, got)
	}
}

// TestHandler_ApplyInvalidValue ensures that an invalid level string in
// the flag value results in an error and does not change the current
// threshold.
func TestHandler_ApplyInvalidValue(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := clictl.NewHandler(th,
		clictl.WithArgs([]string{
			"--log-level=not-a-level",
		}),
	)

	if err := h.Apply(); err == nil {
		t.Fatalf("expected error for invalid level value, got nil")
	}

	if got := th.Level(); got != level.Info {
		t.Fatalf("expected level to remain %v, got %v", level.Info, got)
	}
}

// TestHandler_ApplyUnknownFlags ensures that unknown flags do not break
// parsing and that the handler still honors the configured flag.
func TestHandler_ApplyUnknownFlags(t *testing.T) {
	th := &fakeThreshold{cur: level.Info}

	h := clictl.NewHandler(th,
		clictl.WithArgs([]string{
			"--unknown=42",
			"--log-level=warn",
		}),
	)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Warn {
		t.Fatalf("expected level to be updated to %v, got %v", level.Warn, got)
	}
}

// TestHandler_ApplyUsesOsArgsWhenNil ensures that when Args is nil, Apply
// uses os.Args[1:].
func TestHandler_ApplyUsesOsArgsWhenNil(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"prog", "--log-level=error"}

	th := &fakeThreshold{cur: level.Info}
	h := clictl.NewHandler(th)

	if err := h.Apply(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := th.Level(); got != level.Error {
		t.Fatalf("expected level to be updated to %v, got %v", level.Error, got)
	}
}

// TestHandler_MustApplyPanicsOnError verifies that MustApply panics when
// Apply would return an error.
func TestHandler_MustApplyPanicsOnError(t *testing.T) {
	h := clictl.NewHandler(nil)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic from MustApply, got none")
		}
	}()

	h.MustApply()
}
