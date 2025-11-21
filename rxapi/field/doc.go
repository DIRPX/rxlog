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

// Package field defines the core representation of structured attributes used
// in rxlog log entries.
//
// A Field is a compact, typed key–value pair that describes a single logical
// attribute of a log event (for example, "user_id", "request_id", "duration").
// Fields are designed to:
//
//   - avoid reflection in hot paths by normalizing values into a small,
//     fixed set of storage slots;
//
//   - preserve type information via an explicit discriminator; and
//
//   - provide a stable contract between field producers (callers and helper
//     functions) and field consumers (encoders and cores).
//
// # Overview
//
// Conceptually, a Field is a tagged union:
//
//   - Key is the logical name of the attribute as it will appear in the
//     encoded output;
//
//   - Type is a discriminator of type ftype.Type (defined in the
//     dirpx.dev/rxlog/rxapi/field/type package) that describes which kind of
//     value the field holds; and
//
//   - a small set of storage slots (Integer, String, Interface) hold the
//     actual data in a normalized representation.
//
// This design allows encoders to recover the original logical value without
// using reflection: they switch on Type and read the appropriate storage
// slot(s) according to the rules defined for each Type variant.
//
// # Producer responsibilities
//
// Code that constructs Field values (for example, helper functions in this
// package or application-level adapters) MUST:
//
//   - choose a Type that accurately reflects both the logical kind of the
//     value and the storage slot(s) being used;
//
//   - populate the storage slot(s) in a way that matches the chosen Type
//     (for example, storing integer-like values in the Integer slot, string
//     data in the String slot, and more complex values in the Interface
//     slot);
//
//   - avoid using the uninitialized or invalid Type sentinel for any Field
//     that might be observed by encoders or other consumers;
//
//   - use the Skip sentinel Type only when the intent is to produce a field
//     that encoders will completely ignore; and
//
//   - for Inline fields, ensure that Interface holds a value implementing the
//     appropriate object marshalling interface expected by encoders (as
//     documented by the corresponding ftype.Type and encoder/base contracts).
//
// In typical usage, callers SHOULD construct Field values via well-typed
// helper functions provided by this package rather than filling the struct
// by hand. These helpers encode the mapping from concrete Go types to
// normalized storage and Type values, reducing the risk of inconsistencies.
//
// # Consumer responsibilities
//
// Code that consumes Field values (most notably encoders) MUST treat Type as
// the authoritative indicator of how to interpret the underlying storage:
//
//   - Consumers MUST NOT attempt to infer the meaning of Integer, String, or
//     Interface from their raw contents without consulting Type.
//
//   - For each Type, consumers MUST follow the documented rules for how to
//     read and interpret the storage slots (for example, treating Integer as
//     a nanosecond count for duration-like Types, or as IEEE-754 bits for
//     floating-point Types).
//
//   - For complex Types that rely on auxiliary interfaces (such as array or
//     object marshalers, reflective encoders, or error encoders), consumers
//     MUST type-assert Interface to the expected interface type and handle
//     assertion failures as programming errors.
//
//   - For Skip, consumers MUST emit nothing; such fields represent explicit
//     no-ops in the field list.
//
//   - For Inline, consumers MUST expand the nested object's fields directly
//     into the current encoding scope rather than emitting an additional
//     nested object keyed by Field.Key.
//
// If a Field is observed whose Type does not match the way its storage slots
// are populated, that indicates a bug in the code that created the Field.
// Consumers MAY treat such inconsistencies as programmer errors and SHOULD
// fail loudly (for example, by panicking or recording an explicit encoding
// error) rather than silently emitting incorrect data.
//
// # Encoding integration
//
// The Field type sits at the boundary between high-level logging APIs and
// low-level encoders. The central operation provided by this package is a
// method that materializes a Field into an object encoder:
//
//   - The AddTo method translates the Field's compact internal representation
//     (Type plus storage slots) back into the high-level value expected by an
//     object encoder and appends it to the current encoding buffer.
//
//   - AddTo is performance-sensitive and assumes that Field instances were
//     constructed consistently with their Type. When this assumption is
//     violated, type assertions inside AddTo MAY panic or produce undefined
//     output; callers MUST NOT rely on any behavior in the presence of such
//     mismatches.
//
//   - Some object-encoder operations (for example, those that handle arrays,
//     nested objects, reflected values, or errors) can fail. Instead of
//     propagating such errors directly to callers of AddTo, this package
//     records them as additional fields (for example, by emitting a separate
//     error-description attribute) so that operators can still see that
//     serialization failed for a particular attribute.
//
// This contract allows higher-level components to remain agnostic about the
// details of field storage and encoding while still benefiting from
// predictable, reflection-free serialization.
//
// # Equality and diagnostics
//
// For testing and diagnostic tooling, the Field type provides an Equals
// method that checks whether two Field values represent the same logical
// attribute. Its semantics are carefully defined:
//
//   - Type and Key MUST match, otherwise the fields cannot represent the
//     same attribute.
//
//   - For some complex Types (for example, arrays, objects, errors, or
//     reflect-based values), equality is defined in terms of structural
//     equality of the underlying Interface value (typically via deep
//     comparison), rather than pointer identity.
//
//   - For all remaining Types, equality falls back to full struct equality,
//     comparing Key, Type, and all storage slots.
//
// Because deep structural comparison can be relatively expensive, Equals is
// primarily intended for tests and introspection tools, not for hot paths in
// the logging pipeline.
//
// # Usage context
//
// The field package is the core of rxlog's structured logging model. It does
// not perform I/O and does not depend on higher-level logging concepts such
// as cores or writers. Instead, it provides:
//
//   - the Field type as a compact, typed representation of a single log
//     attribute;
//
//   - a contract for how Fields are constructed and consumed, expressed via
//     the combination of Type and storage slots; and
//
//   - integration points (such as AddTo and Equals) that bridge between
//     field producers, encoders, and diagnostic tooling.
//
// By centralizing attribute representation in this package, rxlog enables
// encoders to operate without reflection, keeps field invariants explicit,
// and provides a stable foundation for structured logging across the
// codebase.
package field
