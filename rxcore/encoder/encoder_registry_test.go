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
	"errors"
	"testing"
	"time"

	"dirpx.dev/rxlog/rxapi/buffer"
	"dirpx.dev/rxlog/rxapi/core"
	encoderapi "dirpx.dev/rxlog/rxapi/encoder"
	"dirpx.dev/rxlog/rxapi/encoder/base"
	"dirpx.dev/rxlog/rxapi/field"
	"dirpx.dev/rxlog/rxcore/encoder"
)

// mockEncoder is a test encoder that implements the encoder.Encoder interface.
type mockEncoder struct {
	name string
}

func (m *mockEncoder) Encode(dst *buffer.Buffer, entry core.Entry, fields []field.Field) (*buffer.Buffer, error) {
	dst.AppendString(m.name)
	return dst, nil
}

// mockEncoder must implement base.ObjectEncoder interface (all methods).
// For tests, we just implement stubs.
func (m *mockEncoder) BeginObject(dst *buffer.Buffer) *buffer.Buffer { return dst }
func (m *mockEncoder) EndObject(dst *buffer.Buffer) *buffer.Buffer   { return dst }
func (m *mockEncoder) AddKey(dst *buffer.Buffer, key string) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddArray(dst *buffer.Buffer, key string, marshaler base.ArrayMarshaler) (*buffer.Buffer, error) {
	return dst, nil
}
func (m *mockEncoder) AddObject(dst *buffer.Buffer, key string, marshaler base.ObjectMarshaler) (*buffer.Buffer, error) {
	return dst, nil
}
func (m *mockEncoder) AddBinary(dst *buffer.Buffer, key string, value []byte) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddByteString(dst *buffer.Buffer, key string, value []byte) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddBool(dst *buffer.Buffer, key string, value bool) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddComplex128(dst *buffer.Buffer, key string, value complex128) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddComplex64(dst *buffer.Buffer, key string, value complex64) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddDuration(dst *buffer.Buffer, key string, value time.Duration) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddFloat64(dst *buffer.Buffer, key string, value float64) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddFloat32(dst *buffer.Buffer, key string, value float32) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddInt(dst *buffer.Buffer, key string, value int) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddInt64(dst *buffer.Buffer, key string, value int64) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddInt32(dst *buffer.Buffer, key string, value int32) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddInt16(dst *buffer.Buffer, key string, value int16) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddInt8(dst *buffer.Buffer, key string, value int8) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddString(dst *buffer.Buffer, key, value string) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddTime(dst *buffer.Buffer, key string, value time.Time) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddUint(dst *buffer.Buffer, key string, value uint) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddUint64(dst *buffer.Buffer, key string, value uint64) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddUint32(dst *buffer.Buffer, key string, value uint32) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddUint16(dst *buffer.Buffer, key string, value uint16) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddUint8(dst *buffer.Buffer, key string, value uint8) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddUintptr(dst *buffer.Buffer, key string, value uintptr) *buffer.Buffer {
	return dst
}
func (m *mockEncoder) AddReflected(dst *buffer.Buffer, key string, value interface{}) (*buffer.Buffer, error) {
	return dst, nil
}
func (m *mockEncoder) OpenNamespace(dst *buffer.Buffer, key string) *buffer.Buffer {
	return dst
}

// TestFromString_UnknownEncoder verifies that FromString returns an error
// for unregistered encoder types.
func TestFromString_UnknownEncoder(t *testing.T) {
	t.Helper()

	cfg := encoder.NewDefaultConfig()

	_, err := encoder.FromString("nonexistent", cfg)
	if err == nil {
		t.Fatal("FromString(\"nonexistent\", cfg) returned nil error, want error")
	}
}

// TestMustFromString_PanicsOnUnknown verifies that MustFromString panics
// when given an unregistered encoder type.
func TestMustFromString_PanicsOnUnknown(t *testing.T) {
	t.Helper()

	cfg := encoder.NewDefaultConfig()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustFromString(\"unknown\", cfg) did not panic")
		}
	}()

	_ = encoder.MustFromString("unknown", cfg)
}

// TestRegister_AddsCustomEncoder verifies that Register allows adding
// custom encoder factories that can then be retrieved via FromString.
func TestRegister_AddsCustomEncoder(t *testing.T) {
	t.Helper()

	const customName = "test-custom-encoder"

	// Register a custom encoder factory.
	customFactory := func(cfg encoder.Config) (encoderapi.Encoder, error) {
		return &mockEncoder{name: "custom"}, nil
	}

	encoder.Register(customName, customFactory)

	// Verify the custom encoder can be created.
	cfg := encoder.NewDefaultConfig()
	enc, err := encoder.FromString(customName, cfg)
	if err != nil {
		t.Fatalf("FromString(%q, cfg) after Register returned error: %v", customName, err)
	}
	if enc == nil {
		t.Fatalf("FromString(%q, cfg) returned nil encoder", customName)
	}

	// Verify it's the correct encoder.
	dst := &buffer.Buffer{}
	out, err := enc.Encode(dst, core.Entry{}, nil)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	got := string(out.Bytes())
	want := "custom"
	if got != want {
		t.Errorf("encoder output = %q, want %q", got, want)
	}
}

// TestRegister_OverwritesExistingEncoder verifies that Register replaces
// an existing encoder when the same name is used.
func TestRegister_OverwritesExistingEncoder(t *testing.T) {
	t.Helper()

	const testName = "test-overwrite-encoder"

	cfg := encoder.NewDefaultConfig()

	// Register initial encoder.
	initialFactory := func(cfg encoder.Config) (encoderapi.Encoder, error) {
		return &mockEncoder{name: "initial"}, nil
	}
	encoder.Register(testName, initialFactory)

	// Overwrite with a different encoder.
	replacementFactory := func(cfg encoder.Config) (encoderapi.Encoder, error) {
		return &mockEncoder{name: "replaced"}, nil
	}
	encoder.Register(testName, replacementFactory)

	// Verify the replacement encoder is now used.
	enc, err := encoder.FromString(testName, cfg)
	if err != nil {
		t.Fatalf("FromString(%q, cfg) after replacement returned error: %v", testName, err)
	}

	dst := &buffer.Buffer{}
	out, err := enc.Encode(dst, core.Entry{}, nil)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	got := string(out.Bytes())
	want := "replaced"
	if got != want {
		t.Errorf("encoder output after overwrite = %q, want %q", got, want)
	}
}

// TestMustFromString_ReturnsEncoderOnSuccess verifies that MustFromString
// returns a valid encoder for registered names without panicking.
func TestMustFromString_ReturnsEncoderOnSuccess(t *testing.T) {
	t.Helper()

	const testName = "test-must-success"

	// Register a test encoder.
	encoder.Register(testName, func(cfg encoder.Config) (encoderapi.Encoder, error) {
		return &mockEncoder{name: "success"}, nil
	})

	cfg := encoder.NewDefaultConfig()
	enc := encoder.MustFromString(testName, cfg)
	if enc == nil {
		t.Fatal("MustFromString returned nil encoder")
	}

	// Verify it works.
	dst := &buffer.Buffer{}
	out, err := enc.Encode(dst, core.Entry{}, nil)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	got := string(out.Bytes())
	want := "success"
	if got != want {
		t.Errorf("encoder output = %q, want %q", got, want)
	}
}

// TestFromString_PropagatesFactoryError verifies that FromString propagates
// errors from the factory function.
func TestFromString_PropagatesFactoryError(t *testing.T) {
	t.Helper()

	const testName = "test-factory-error"

	// Register a factory that returns an error.
	expectedErr := errors.New("factory creation failed")
	encoder.Register(testName, func(cfg encoder.Config) (encoderapi.Encoder, error) {
		return nil, expectedErr
	})

	cfg := encoder.NewDefaultConfig()
	enc, err := encoder.FromString(testName, cfg)
	if err == nil {
		t.Fatal("FromString returned nil error, want error from factory")
	}
	if enc != nil {
		t.Errorf("FromString returned non-nil encoder on error: %v", enc)
	}
	if err != expectedErr {
		t.Errorf("FromString error = %v, want %v", err, expectedErr)
	}
}

// TestMustFromString_PanicsOnFactoryError verifies that MustFromString panics
// when the factory function returns an error.
func TestMustFromString_PanicsOnFactoryError(t *testing.T) {
	t.Helper()

	const testName = "test-must-factory-error"

	// Register a factory that returns an error.
	encoder.Register(testName, func(cfg encoder.Config) (encoderapi.Encoder, error) {
		return nil, errors.New("factory error")
	})

	cfg := encoder.NewDefaultConfig()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustFromString did not panic on factory error")
		}
	}()

	_ = encoder.MustFromString(testName, cfg)
}
