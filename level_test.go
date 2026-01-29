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

package rxlog_test

import (
	"testing"

	"dirpx.dev/rxlog"
	"dirpx.dev/rxlog/rxapi/level"
)

func TestLevelConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rxlog    rxlog.Level
		rxapiLvl level.Level
	}{
		{"Trace", rxlog.TraceLevel, level.Trace},
		{"Debug", rxlog.DebugLevel, level.Debug},
		{"Info", rxlog.InfoLevel, level.Info},
		{"Notice", rxlog.NoticeLevel, level.Notice},
		{"Warn", rxlog.WarnLevel, level.Warn},
		{"Error", rxlog.ErrorLevel, level.Error},
		{"Critical", rxlog.CriticalLevel, level.Critical},
		{"Fatal", rxlog.FatalLevel, level.Fatal},
		{"Invalid", rxlog.InvalidLevel, level.Invalid},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.rxlog != tt.rxapiLvl {
				t.Errorf("%s: rxlog=%v, rxapi/level=%v, want equal",
					tt.name, tt.rxlog, tt.rxapiLvl)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  rxlog.Level
		error bool
	}{
		{"trace", rxlog.TraceLevel, false},
		{"debug", rxlog.DebugLevel, false},
		{"info", rxlog.InfoLevel, false},
		{"warn", rxlog.WarnLevel, false},
		{"error", rxlog.ErrorLevel, false},
		{"invalid-level", rxlog.InvalidLevel, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got, err := rxlog.ParseLevel(tt.input)
			if tt.error {
				if err == nil {
					t.Errorf("ParseLevel(%q) error = nil, want error", tt.input)
				}
				if got != rxlog.InvalidLevel {
					t.Errorf("ParseLevel(%q) = %v, want InvalidLevel", tt.input, got)
				}
			} else {
				if err != nil {
					t.Errorf("ParseLevel(%q) error = %v, want nil", tt.input, err)
				}
				if got != tt.want {
					t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
				}
			}
		})
	}
}

func TestMustParseLevel_Success(t *testing.T) {
	t.Parallel()

	lvl := rxlog.MustParseLevel("info")
	if lvl != rxlog.InfoLevel {
		t.Errorf("MustParseLevel(\"info\") = %v, want %v", lvl, rxlog.InfoLevel)
	}
}

func TestMustParseLevel_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseLevel did not panic on invalid input")
		}
	}()

	rxlog.MustParseLevel("invalid-level")
}

func TestLevelTypeAlias(t *testing.T) {
	t.Parallel()

	// Verify that rxlog.Level is compatible with rxapi/level.Level
	var rxlogLvl rxlog.Level = rxlog.InfoLevel
	var rxapiLvl level.Level = rxlogLvl

	if rxapiLvl != level.Info {
		t.Errorf("Type alias failed: rxapiLvl = %v, want %v", rxapiLvl, level.Info)
	}
}
