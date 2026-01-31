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

package encoder_test

import (
	"testing"

	"dirpx.dev/rxlog/rxcore/encoder"
)

// TestFromString_RegisteredConfigs verifies that all built-in encoder config
// presets can be looked up by their registered names.
func TestFromString_RegisteredConfigs(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
	}{
		{"json"},
		{"console"},
		{"development"},
		{"production"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := encoder.FromString(tt.name)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.name, err)
			}

			// Verify basic fields are set.
			if cfg.MessageKey == "" {
				t.Errorf("config %q has empty MessageKey", tt.name)
			}
			if cfg.LevelKey == "" {
				t.Errorf("config %q has empty LevelKey", tt.name)
			}
			if cfg.TimeKey == "" {
				t.Errorf("config %q has empty TimeKey", tt.name)
			}
		})
	}
}

// TestFromString_UnknownConfig verifies that FromString returns an error
// for unregistered config preset names.
func TestFromString_UnknownConfig(t *testing.T) {
	t.Helper()

	_, err := encoder.FromString("nonexistent")
	if err == nil {
		t.Fatal("FromString(\"nonexistent\") returned nil error, want error")
	}
}

// TestMustFromString_PanicsOnUnknown verifies that MustFromString panics
// when given an unregistered config preset name.
func TestMustFromString_PanicsOnUnknown(t *testing.T) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustFromString(\"unknown\") did not panic")
		}
	}()

	_ = encoder.MustFromString("unknown")
}

// TestMustFromString_ReturnsConfigOnSuccess verifies that MustFromString
// returns a valid config for registered names without panicking.
func TestMustFromString_ReturnsConfigOnSuccess(t *testing.T) {
	t.Helper()

	cfg := encoder.MustFromString("json")

	// Verify it's a valid config.
	if cfg.MessageKey == "" {
		t.Error("json config has empty MessageKey")
	}
	if cfg.EncodeLevel == nil {
		t.Error("json config has nil EncodeLevel")
	}
}

// TestRegister_AddsCustomConfig verifies that Register allows adding
// custom configs that can then be retrieved via FromString.
func TestRegister_AddsCustomConfig(t *testing.T) {
	t.Helper()

	const customName = "test-custom-config"

	// Register a custom config.
	customCfg := encoder.Config{
		MessageKey: "custom_msg",
		LevelKey:   "custom_level",
		TimeKey:    "custom_time",
	}

	encoder.Register(customName, customCfg)

	// Verify the custom config can be retrieved.
	cfg, err := encoder.FromString(customName)
	if err != nil {
		t.Fatalf("FromString(%q) after Register returned error: %v", customName, err)
	}

	if cfg.MessageKey != "custom_msg" {
		t.Errorf("custom config MessageKey = %q, want %q", cfg.MessageKey, "custom_msg")
	}
	if cfg.LevelKey != "custom_level" {
		t.Errorf("custom config LevelKey = %q, want %q", cfg.LevelKey, "custom_level")
	}
}

// TestRegister_OverwritesExistingConfig verifies that Register replaces
// an existing config when the same name is used.
func TestRegister_OverwritesExistingConfig(t *testing.T) {
	t.Helper()

	const testName = "test-overwrite-config"

	// Register initial config.
	initialCfg := encoder.Config{
		MessageKey: "initial",
	}
	encoder.Register(testName, initialCfg)

	// Overwrite with a different config.
	replacementCfg := encoder.Config{
		MessageKey: "replaced",
	}
	encoder.Register(testName, replacementCfg)

	// Verify the replacement config is now used.
	cfg, err := encoder.FromString(testName)
	if err != nil {
		t.Fatalf("FromString(%q) after replacement returned error: %v", testName, err)
	}

	if cfg.MessageKey != "replaced" {
		t.Errorf("config after overwrite MessageKey = %q, want %q", cfg.MessageKey, "replaced")
	}
}

// TestBuiltinConfigs_HaveRequiredEncoders verifies that all built-in configs
// have their encoder functions set.
func TestBuiltinConfigs_HaveRequiredEncoders(t *testing.T) {
	t.Helper()

	presets := []string{"json", "console", "development", "production"}

	for _, preset := range presets {
		t.Run(preset, func(t *testing.T) {
			cfg, err := encoder.FromString(preset)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", preset, err)
			}

			// Verify all encoder functions are set.
			if cfg.EncodeLevel == nil {
				t.Errorf("%q config has nil EncodeLevel", preset)
			}
			if cfg.EncodeTime == nil {
				t.Errorf("%q config has nil EncodeTime", preset)
			}
			if cfg.EncodeDuration == nil {
				t.Errorf("%q config has nil EncodeDuration", preset)
			}
			if cfg.EncodeCaller == nil {
				t.Errorf("%q config has nil EncodeCaller", preset)
			}
			if cfg.EncodeName == nil {
				t.Errorf("%q config has nil EncodeName", preset)
			}
			if cfg.EncodeError == nil {
				t.Errorf("%q config has nil EncodeError", preset)
			}
			if cfg.EncodeStacktrace == nil {
				t.Errorf("%q config has nil EncodeStacktrace", preset)
			}
		})
	}
}

// TestProductionConfig_IsCompact verifies that the production config is
// optimized for compact output (no pretty print, no indent).
func TestProductionConfig_IsCompact(t *testing.T) {
	t.Helper()

	cfg, err := encoder.FromString("production")
	if err != nil {
		t.Fatalf("FromString(\"production\") returned error: %v", err)
	}

	if cfg.PrettyPrint {
		t.Error("production config has PrettyPrint=true, want false")
	}
	if cfg.IndentString != "" {
		t.Errorf("production config has IndentString=%q, want empty", cfg.IndentString)
	}
}

// TestDevelopmentConfig_IsPrettyPrinted verifies that the development config
// enables pretty printing for readability.
func TestDevelopmentConfig_IsPrettyPrinted(t *testing.T) {
	t.Helper()

	cfg, err := encoder.FromString("development")
	if err != nil {
		t.Fatalf("FromString(\"development\") returned error: %v", err)
	}

	if !cfg.PrettyPrint {
		t.Error("development config has PrettyPrint=false, want true")
	}
	if cfg.IndentString == "" {
		t.Error("development config has empty IndentString, want non-empty")
	}
}
