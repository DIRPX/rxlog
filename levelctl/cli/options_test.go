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
	"reflect"
	"testing"

	clictl "dirpx.dev/rxlog/levelctl/cli"
	"dirpx.dev/rxlog/rxapi/level"
)

// TestWithFlagName verifies that WithFlagName assigns the provided name
// to Handler.FlagName without additional normalization.
//
// Normalization of empty names (falling back to DefaultFlagName) is
// performed by NewHandler, not by the option itself.
func TestWithFlagName(t *testing.T) {
	var h clictl.Handler

	clictl.WithFlagName("verbosity")(&h)
	if h.FlagName != "verbosity" {
		t.Fatalf("expected FlagName %q, got %q", "verbosity", h.FlagName)
	}

	clictl.WithFlagName("")(&h)
	if h.FlagName != "" {
		t.Fatalf("expected FlagName to be empty, got %q", h.FlagName)
	}
}

// TestWithShortFlag verifies that WithShortFlag assigns the provided
// short flag name to Handler.ShortFlag.
func TestWithShortFlag(t *testing.T) {
	var h clictl.Handler

	clictl.WithShortFlag("L")(&h)
	if h.ShortFlag != "L" {
		t.Fatalf("expected ShortFlag %q, got %q", "L", h.ShortFlag)
	}

	clictl.WithShortFlag("")(&h)
	if h.ShortFlag != "" {
		t.Fatalf("expected ShortFlag to be empty, got %q", h.ShortFlag)
	}
}

// TestWithDefaultLevel verifies that WithDefaultLevel sets DefaultLevel
// and marks HasDefault as true.
func TestWithDefaultLevel(t *testing.T) {
	var h clictl.Handler

	if h.HasDefault {
		t.Fatalf("expected HasDefault to be false by default")
	}

	clictl.WithDefaultLevel(level.Warn)(&h)

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
	var h clictl.Handler

	clictl.WithRequired(true)(&h)
	if !h.Required {
		t.Fatalf("expected Required to be true, got false")
	}

	clictl.WithRequired(false)(&h)
	if h.Required {
		t.Fatalf("expected Required to be false, got true")
	}
}

// TestWithArgs verifies that WithArgs stores a copy of the provided
// slice and does not retain aliases to the caller's slice, so subsequent
// mutations by the caller do not affect Handler.Args.
func TestWithArgs(t *testing.T) {
	orig := []string{"--log-level=info"}
	h := clictl.Handler{}

	clictl.WithArgs(orig)(&h)

	if h.Args == nil {
		t.Fatalf("expected Args to be non-nil after WithArgs")
	}
	if &h.Args[0] == &orig[0] {
		t.Fatalf("expected Args to be a copy, but underlying arrays appear to alias")
	}
	if !reflect.DeepEqual(h.Args, orig) {
		t.Fatalf("expected Args %v, got %v", orig, h.Args)
	}

	// Mutate original slice and ensure Handler.Args does not change.
	orig[0] = "--log-level=debug"
	if h.Args[0] != "--log-level=info" {
		t.Fatalf("expected Args[0] to remain %q, got %q", "--log-level=info", h.Args[0])
	}

	// Nil input should reset Args to nil.
	clictl.WithArgs(nil)(&h)
	if h.Args != nil {
		t.Fatalf("expected Args to be nil after WithArgs(nil), got %v", h.Args)
	}
}
