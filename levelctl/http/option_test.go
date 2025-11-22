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

package httpctl_test

import (
	"testing"

	httpctl "dirpx.dev/rxlog/levelctl/http"
)

// TestWithParamName verifies that WithParamName assigns the provided name
// to Handler.ParamName without performing any additional normalization.
// Normalization of empty names is handled by NewHandler, not by the option.
func TestWithParamName(t *testing.T) {
	var h httpctl.Handler

	// Non-empty name.
	httpctl.WithParamName("custom")(&h)
	if h.ParamName != "custom" {
		t.Fatalf("expected ParamName %q, got %q", "custom", h.ParamName)
	}

	// Empty name should be stored as-is; fallback to DefaultLevelParamName
	// is applied by NewHandler, not by the option.
	httpctl.WithParamName("")(&h)
	if h.ParamName != "" {
		t.Fatalf("expected ParamName to be empty, got %q", h.ParamName)
	}
}

// TestWithMaxBodyBytes verifies that WithMaxBodyBytes assigns the provided
// value directly to Handler.MaxBodyBytes. Interpretation of zero or
// negative values is handled by NewHandler, not by the option itself.
func TestWithMaxBodyBytes(t *testing.T) {
	var h httpctl.Handler

	// Positive value.
	httpctl.WithMaxBodyBytes(1234)(&h)
	if h.MaxBodyBytes != 1234 {
		t.Fatalf("expected MaxBodyBytes %d, got %d", 1234, h.MaxBodyBytes)
	}

	// Zero and negative values are accepted and stored as-is; the handler
	// constructor is responsible for applying DefaultMaxBodyBytes if needed.
	httpctl.WithMaxBodyBytes(0)(&h)
	if h.MaxBodyBytes != 0 {
		t.Fatalf("expected MaxBodyBytes %d, got %d", 0, h.MaxBodyBytes)
	}

	httpctl.WithMaxBodyBytes(-1)(&h)
	if h.MaxBodyBytes != -1 {
		t.Fatalf("expected MaxBodyBytes %d, got %d", -1, h.MaxBodyBytes)
	}
}

// TestWithReadOnly verifies that WithReadOnly toggles Handler.ReadOnly
// exactly to the provided value.
func TestWithReadOnly(t *testing.T) {
	var h httpctl.Handler

	// Enable read-only mode.
	httpctl.WithReadOnly(true)(&h)
	if !h.ReadOnly {
		t.Fatalf("expected ReadOnly to be true, got false")
	}

	// Disable read-only mode.
	httpctl.WithReadOnly(false)(&h)
	if h.ReadOnly {
		t.Fatalf("expected ReadOnly to be false, got true")
	}
}
