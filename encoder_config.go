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

package rxlog

import (
	"dirpx.dev/rxlog/rxapi/caller"
	"dirpx.dev/rxlog/rxapi/chrono/duration"
	"dirpx.dev/rxlog/rxapi/chrono/time"
	aerr "dirpx.dev/rxlog/rxapi/error"
	"dirpx.dev/rxlog/rxapi/level"
	"dirpx.dev/rxlog/rxapi/name"
	"dirpx.dev/rxlog/rxapi/reflect"
	"dirpx.dev/rxlog/rxapi/stack"
	"dirpx.dev/rxlog/rxcore/field/fields"
)

// Config describes how a log Encoder SHOULD serialize log entries to bytes.
//
// This configuration is organized into logical sections for clarity, covering
// rxcore encoding behavior, reflection and fallback rules, output formatting,
// error handling, and format-specific options.
//
// Unless individual fields specify stricter rules, zero values generally mean
// that the encoder implementation MAY apply its own default behavior or treat
// the corresponding feature as disabled. Implementations SHOULD document their
// concrete defaults, and callers SHOULD refer to NewDefaultEncoderConfig (or an
// equivalent helper) for recommended production settings.
type EncoderConfig struct {
	// MessageKey is the field name used to encode the primary log message
	// text in structured formats.
	//
	// If MessageKey is empty, the encoder implementation MUST fall back to
	// its default key name. Implementations SHOULD document their default
	// (for example, "msg" or "message").
	MessageKey string

	// LevelKey is the field name used to encode the log level.
	//
	// If LevelKey is empty, the encoder implementation MUST fall back to its
	// default key name. Implementations SHOULD document their default (for
	// example, "level", "severity", or "lvl").
	LevelKey string

	// TimeKey is the field name used to encode the timestamp.
	//
	// If TimeKey is empty, the encoder implementation MUST fall back to its
	// default key name. Implementations SHOULD document their default (for
	// example, "time", "ts", "timestamp", or "@timestamp").
	TimeKey string

	// CallerKey is the field name used to encode caller information
	// (file, line, function) in structured formats.
	//
	// If CallerKey is empty, the encoder implementation MUST fall back to its
	// default key name. Implementations SHOULD document their default (for
	// example, "caller", "source", or "location").
	CallerKey string

	// StacktraceKey is the field name used to encode stack traces.
	//
	// If StacktraceKey is empty, the encoder implementation MUST fall back to
	// its default key name. Implementations SHOULD document their default
	// (for example, "stacktrace", "stack", or "trace").
	StacktraceKey string

	// NameKey is the field name used to encode the logger name.
	//
	// If NameKey is empty, the encoder implementation MUST fall back to its
	// default key name. Implementations SHOULD document their default
	// (for example, "logger", "name", or "component").
	NameKey string

	// ErrorKey is the field name used to encode error values.
	//
	// If ErrorKey is empty, the encoder implementation MUST fall back to its
	// default key name. Implementations SHOULD document their default
	// (for example, "error" or "err").
	ErrorKey string

	// OperationKey is the field name used to encode function names when they
	// are emitted separately from caller information.
	//
	// If OperationKey is empty, the encoder implementation MUST fall back to
	// its default key name or omit separate function fields entirely,
	// depending on its documented behavior.
	OperationKey string

	// EncodeLevel encodes log levels (for example, DEBUG, INFO, WARN, ERROR).
	//
	// If EncodeLevel is nil, the encoder MUST use its default level encoder.
	// Implementations SHOULD document their default encoding (for example,
	// symbolic names or numeric codes) and SHOULD keep it stable over time
	// for compatibility with downstream consumers.
	EncodeLevel level.Encoder

	// EncodeTime encodes timestamps.
	//
	// If EncodeTime is nil, the encoder MUST use its default time encoder
	// (commonly RFC3339-based or Unix time). Implementations SHOULD document
	// their default time format and MUST handle zero-valued timestamps in a
	// well-defined way (such as omission or a sentinel value).
	EncodeTime time.Encoder

	// EncodeDuration encodes time.Duration values.
	//
	// If EncodeDuration is nil, the encoder MUST use its default duration
	// encoder (for example, a numeric representation in nanoseconds or a
	// textual duration string). The chosen format SHOULD remain stable over
	// time to avoid breaking downstream parsers.
	EncodeDuration duration.Encoder

	// EncodeCaller encodes caller information (file, line, function).
	//
	// If EncodeCaller is nil, the encoder MUST use its default caller
	// encoder. Implementations SHOULD document whether the default encoding
	// is compact, full path, or some hybrid representation.
	EncodeCaller caller.Encoder

	// EncodeName encodes logger names.
	//
	// If EncodeName is nil, the encoder MUST use its default name encoder,
	// which typically writes the name as-is. Implementations MAY normalize
	// names (for example, trimming whitespace) but SHOULD document such
	// behavior.
	EncodeName name.Encoder

	// EncodeError encodes error values.
	//
	// If EncodeError is nil, the encoder MUST fall back to its default error
	// encoding strategy, which MAY rely on reflection, err.Error(), or a
	// combination of both. Implementations SHOULD document the default
	// representation of errors, including nil values.
	EncodeError aerr.Encoder

	// EncodeStacktrace encodes stack traces represented as strings.
	//
	// If EncodeStacktrace is nil, the encoder MUST use its default stack encoding,
	// which typically emits the stack trace as-is. Implementations MAY
	// reformat or trim stack traces but SHOULD document any such behavior.
	EncodeStacktrace stack.Encoder

	// ReflectedEncoder is used as a fallback for arbitrary Go values when no
	// specialized encoder is available.
	//
	// Typical inputs include structs, maps, slices, and other composite types.
	// If ReflectedEncoder is nil, the encoder MUST use its default
	// reflection-based encoder. Implementations SHOULD document the rules
	// they apply for field naming, zero-value omission, and type handling.
	ReflectedEncoder reflect.Encoder

	// UseJSONTags controls whether the reflected encoder honors `json:"name"`
	// struct tags when deriving field names.
	//
	// When true, the reflected encoder SHOULD treat JSON tags as authoritative
	// for naming, thereby aligning behavior with encoding/json conventions.
	// When false, it MUST ignore JSON tags and derive names directly from
	// struct field names. The default behavior SHOULD be documented.
	UseJSONTags bool

	// RespectStringer controls whether the encoder uses a value's String()
	// method (when it implements fmt.Stringer) before falling back to
	// reflection.
	//
	// When true, values that implement fmt.Stringer SHOULD be encoded via
	// their String() method, which can improve performance and readability.
	// When false, the encoder MAY still use reflection even for Stringer
	// types. Implementations SHOULD document their default behavior.
	RespectStringer bool

	// LineEnding defines the sequence appended after each encoded entry.
	//
	// Typical values include "\n" for Unix-style line endings, "\r\n" for
	// Windows-style, or "" to disable automatic line termination. When empty,
	// the encoder implementation MAY use a default line ending or omit line
	// endings entirely, depending on its documented behavior.
	LineEnding string

	// FieldSeparator is used by human-readable encoders (such as console/text
	// encoders) to separate fields in a single log line.
	//
	// Structured encoders (such as pure JSON) typically ignore this field.
	// When used, it SHOULD be a short, readable separator (for example, a
	// single space or a visually distinct delimiter).
	FieldSeparator string

	// SkipLineEnding controls whether the encoder appends LineEnding after
	// each encoded entry.
	//
	// When SkipLineEnding is true, the encoder MUST ignore LineEnding and MUST
	// NOT append any automatic line terminator. This mode is useful when an
	// upstream or downstream component is responsible for framing or when log
	// entries are embedded into a larger structure (for example, a JSON array
	// or a custom transport protocol).
	//
	// When SkipLineEnding is false, the encoder SHOULD append LineEnding after
	// each entry, subject to the encoder’s documented defaults if LineEnding
	// itself is empty.
	SkipLineEnding bool

	// PrettyPrint controls whether JSON output is formatted with indentation
	// and human-friendly spacing.
	//
	// When PrettyPrint is true, the encoder MUST emit JSON with indentation
	// and line breaks, using IndentString to control the indentation unit.
	// This significantly improves readability for humans but is typically
	// slower and produces larger output, so it SHOULD be enabled primarily
	// for development, debugging, or manually inspected logs.
	//
	// When PrettyPrint is false, the encoder SHOULD emit compact JSON without
	// extraneous whitespace, which is generally preferred for production and
	// high-volume logging.
	PrettyPrint bool

	// IndentString specifies the indentation prefix used when PrettyPrint is
	// true.
	//
	// When PrettyPrint is enabled, the encoder MUST use IndentString as the
	// indentation unit for nested JSON structures. If IndentString is empty,
	// the encoder SHOULD fall back to a documented default (typically two
	// spaces, "  ").
	//
	// When PrettyPrint is false, IndentString MUST be ignored.
	IndentString string

	// DisableHTMLEscape controls whether certain characters in JSON strings
	// are escaped using their Unicode code point escape sequences.
	//
	// When DisableHTMLEscape is false (the typical default), the encoder
	// SHOULD escape characters such as '&', '<', and '>' as "\u0026",
	// "\u003c", and "\u003e" respectively, mirroring encoding/json’s default
	// behavior. This can help avoid accidental HTML/script embedding issues
	// when logs are rendered in browsers or HTML contexts.
	//
	// When DisableHTMLEscape is true, the encoder MUST emit these characters
	// as-is in JSON strings (for example, "&" remains "&"). This improves
	// readability and fidelity of certain payloads but MAY be undesirable if
	// logs are directly embedded into HTML without additional sanitization.
	DisableHTMLEscape bool

	// ErrorHandler defines how unexpected encoding errors are handled.
	//
	// When ErrorHandler is nil, the encoder MUST use a default handler,
	// commonly one that skips failing fields and continues encoding (such as
	// SkipOnErrorHandler). Callers that require stricter behavior (for
	// example, failing the entire entry on any error) SHOULD provide an
	// explicit handler (such as FailOnErrorHandler).
	//
	// Implementations MAY expose built-in handler types (for example,
	// FailOnErrorHandler, SkipOnErrorHandler, ReplaceOnErrorHandler,
	// LogOnErrorHandler), and custom handler MAY implement the Handler
	// interface to integrate with the encoding pipeline.
	ErrorHandler aerr.Handler

	// FormatOptions stores encoder-specific options in an untyped form for
	// extensibility.
	//
	// Concrete encoder implementations MAY assert FormatOptions to an
	// expected type (for example, *JSONOptions, *ConsoleOptions, or
	// *TextOptions) and interpret its fields accordingly. Callers MUST ensure
	// that the type they provide matches the expectations of the chosen
	// encoder; otherwise, behavior is undefined and MAY result in panics.
	//
	// When FormatOptions is nil, the encoder MUST use its internal defaults
	// for all format-specific behavior.
	FormatOptions interface{}
}

// NewDefaultConfig constructs a EncoderConfig populated with recommended defaults
// suitable for most JSON-style structured encoders in rxcore.
//
// The returned configuration:
//
//   - Uses canonical field keys from rxcore/field/fields for common
//     attributes (message, level, timestamp, caller, stacktrace, logger
//     name, error, function), aligning the encoded schema with the
//     shared vocabulary used by other parts of the logging stack.
//
//   - Leaves all Encode* function fields nil so that concrete encoder
//     implementations can apply their own documented defaults for level,
//     time, duration, caller, name, error, and stacktrace encoding.
//
//   - Enables JSON-tag-aware reflection (UseJSONTags = true) and respects
//     fmt.Stringer implementations (RespectStringer = true) to keep the
//     default behavior close to encoding/json and idiomatic Go types.
//
//   - Uses a Unix-style line ending ("\n") and a single space as a field
//     separator, which are commonly expected by line-oriented log
//     collectors and human readers.
//
//   - Leaves PrettyPrint disabled and IndentString empty. Callers that
//     enable PrettyPrint SHOULD either set IndentString explicitly or call
//     Normalized to obtain the default indentation unit.
//
//   - Leaves DisableHTMLEscape, ErrorHandler, and FormatOptions at their
//     zero values so that concrete encoders can apply their own internal
//     defaults.
//
// Callers MAY treat this as a baseline and override individual fields as
// needed before constructing a concrete encoder.
func NewDefaultEncoderConfig() EncoderConfig {
	return EncoderConfig{
		MessageKey:        fields.Message,
		LevelKey:          fields.Level,
		TimeKey:           fields.Timestamp,
		CallerKey:         fields.Caller,
		StacktraceKey:     fields.Stacktrace,
		NameKey:           fields.Logger,
		ErrorKey:          fields.Error,
		OperationKey:      fields.Operation,
		EncodeLevel:       nil,
		EncodeTime:        nil,
		EncodeDuration:    nil,
		EncodeCaller:      nil,
		EncodeName:        nil,
		EncodeError:       nil,
		EncodeStacktrace:  nil,
		ReflectedEncoder:  nil,
		UseJSONTags:       true,
		RespectStringer:   true,
		LineEnding:        "\n",
		FieldSeparator:    " ",
		SkipLineEnding:    false,
		PrettyPrint:       false,
		IndentString:      "",
		DisableHTMLEscape: false,
		ErrorHandler:      nil,
		FormatOptions:     nil,
	}
}

// Clone returns a shallow copy of the configuration.
//
// All value-typed fields are copied by value. The Encode* function fields,
// ErrorHandler, and FormatOptions are copied as-is; in particular:
//
//   - Function fields (EncodeLevel, EncodeTime, EncodeDuration, EncodeCaller,
//     EncodeName, EncodeError, EncodeStacktrace, ReflectedEncoder) are
//     immutable references to functions and are therefore safe to reuse.
//
//   - ErrorHandler and FormatOptions are copied as interface values. If they
//     contain pointers to mutable data structures, those structures are NOT
//     deep-cloned; both the original and the clone will reference the same
//     underlying value.
//
// This behavior is appropriate for most encoder configuration scenarios,
// where EncoderConfig values are treated as immutable templates. Callers that need
// deep copies of FormatOptions MUST perform such copying themselves before
// assigning to the cloned configuration.
func (c EncoderConfig) Clone() EncoderConfig {
	return c
}

// Normalized returns a copy of the configuration with implied, purely local
// defaults applied.
//
// Currently, Normalized performs only one normalization step:
//
//   - If PrettyPrint is true and IndentString is empty, it sets
//     IndentString to a canonical two-space indentation ("  ").
//
// This mirrors the typical default used by encoding/json and many human-
// friendly JSON formatters. No other fields are altered; in particular,
// encoder implementations remain responsible for applying their own
// documented defaults when Encode* function fields are nil or when key
// names are left empty.
//
// Normalized is intended as a convenience for encoders that want to honor
// the PrettyPrint/IndentString semantics without duplicating the same
// indentation logic in multiple places. Callers MAY ignore this helper
// entirely and implement their own normalization if they require different
// behavior.
func (c EncoderConfig) Normalized() EncoderConfig {
	if c.PrettyPrint && c.IndentString == "" {
		// Use a widely expected JSON indentation unit.
		c.IndentString = "  "
	}
	return c
}
