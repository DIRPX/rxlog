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
	"testing"

	envctl "dirpx.dev/rxlog/levelctl/env"
	"dirpx.dev/rxlog/rxapi/level"
)

// TestWithVarName verifies that WithVarName assigns the provided name
// to Handler.VarName without additional normalization. Normalization of
// empty names is performed by NewHandler, not by the option itself.
func TestWithVarName(t *testing.T) {
	var h envctl.Handler

	envctl.WithVarName("MYAPP_LOG_LEVEL")(&h)
	if h.VarName != "MYAPP_LOG_LEVEL" {
		t.Fatalf("expected VarName %q, got %q", "MYAPP_LOG_LEVEL", h.VarName)
	}

	// Empty name is stored as-is; fallback to DefaultVarName is applied
	// by NewHandler, not by the option.
	envctl.WithVarName("")(&h)
	if h.VarName != "" {
		t.Fatalf("expected VarName to be empty, got %q", h.VarName)
	}
}

// TestWithDefaultLevel verifies that WithDefaultLevel sets DefaultLevel
// and marks HasDefault as true.
func TestWithDefaultLevel(t *testing.T) {
	var h envctl.Handler

	if h.HasDefault {
		t.Fatalf("expected HasDefault to be false by default")
	}

	envctl.WithDefaultLevel(level.Warn)(&h)

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
	var h envctl.Handler

	envctl.WithRequired(true)(&h)
	if !h.Required {
		t.Fatalf("expected Required to be true, got false")
	}

	envctl.WithRequired(false)(&h)
	if h.Required {
		t.Fatalf("expected Required to be false, got true")
	}
}

// TestWithLookup verifies that WithLookup installs the provided lookup
// function into the Handler.
func TestWithLookup(t *testing.T) {
	var h envctl.Handler

	called := false
	lookup := func(name string) (string, bool) {
		called = true
		return "info", true
	}

	envctl.WithLookup(lookup)(&h)

	if h.Lookup == nil {
		t.Fatalf("expected Lookup to be set")
	}

	// Invoke the stored Lookup to ensure it points to the provided function.
	if _, _ = h.Lookup("SOME_VAR"); !called {
		t.Fatalf("expected stored Lookup to call provided function")
	}
}
