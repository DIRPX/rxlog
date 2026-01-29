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

// Package ftype defines the enumeration used by the rxlog field subsystem
// to describe which logical kind of value a field holds.
//
// The field API represents values using a compact union-like structure
// (a key, a Type discriminator, and a small set of storage slots such as
// Integer, String, and Interface). The Type defined in this package is the
// discriminator for that union: it tells encoders and other consumers how
// the stored bits MUST be interpreted.
//
// # Overview
//
// At a high level, the Type enumeration serves three purposes:
//
//   - It identifies which member of the field's internal union is active
//     (for example, whether the value is stored in the Integer, String, or
//     Interface slot).
//
//   - It describes the abstract kind of the value (for example, string,
//     numeric, time, duration, structured object, array, error, or a
//     reflective "any" value).
//
//   - It specifies the contract that producers and consumers of fields MUST
//     follow for each kind (for example, which Go type is expected in the
//     Interface slot and how encoders should treat it).
//
// Code that constructs fields is responsible for setting Type consistently
// with how the value is actually stored. Code that consumes fields (for
// example, encoders) MUST rely on Type as the authoritative indicator of
// how to interpret the underlying storage and MUST NOT guess based on the
// raw storage slots alone.
//
// # Relationship to the field package
//
// The ftype package is an internal part of the field subsystem's API surface.
// It does not define the Field type itself; instead, it specifies the set of
// allowed Type values and the invariants associated with each of them.
//
// Typical usage looks like this:
//
//   - High-level helper functions construct fields from concrete Go values
//     (strings, numbers, times, errors, custom marshalers, and so on).
//
//   - Each helper chooses an appropriate Type and populates the corresponding
//     storage slot(s) on the field struct.
//
//   - Encoders and other consumers perform a switch on Type to recover the
//     original logical value and serialize it without reflection.
//
// If a field is observed with a Type that does not match the way its storage
// slots are populated, that indicates a bug in the code that created the
// field. Consumers MAY treat such inconsistencies as programmer errors and
// SHOULD fail loudly (for example, by panicking or recording an explicit
// encoding error) rather than silently emitting incorrect data.
//
// # Type categories and sentinels
//
// The Type enumeration covers a range of logical value categories, including:
//
//   - structured values backed by custom marshalers (for example, array and
//     object kinds that delegate to explicit marshaler interfaces),
//
//   - scalar primitives (strings, byte strings, booleans, signed and unsigned
//     integers, floating-point numbers, complex numbers),
//
//   - time-related values (durations and timestamps stored either in an
//     integer representation or as full Go values),
//
//   - error values, which carry a value implementing the error interface,
//     allowing encoders to serialize error messages and associated metadata,
//
//   - generic or reflective values that defer type-specific handling to
//     higher-level or reflection-based encoders, and
//
//   - sentinel values used for special control flow.
//
// In particular, the enumeration includes special sentinel kinds:
//
//   - A zero-value kind used to represent an uninitialized or invalid field
//     type. Fields with this kind MUST NOT be produced in normal operation;
//     encountering such a field SHOULD be treated as a programming error.
//
//   - A "skip" kind used to indicate that a field should be ignored entirely
//     by encoders. This is useful for helper functions that sometimes decide
//     not to emit anything and want to signal a no-op field.
//
//   - An "inline" kind used to indicate that the value is a structured object
//     whose fields should be merged directly into the surrounding object
//     rather than nested under a separate key.
//
// The exact set of constants and their detailed semantics are documented
// alongside the Type declaration in this package. The package-level
// documentation focuses on the overall model and invariants.
//
// # Producer responsibilities
//
// Code that constructs fields (for example, helper functions in the field
// package or application-level adapters) MUST:
//
//   - choose a Type that accurately reflects both the logical kind of the
//     value and the storage slot(s) being used;
//
//   - populate the storage slot(s) in a way that matches the chosen Type
//     (for example, storing an int64 in the Integer slot for integer kinds,
//     or a specific Go type in the Interface slot for marshaler-based kinds);
//
//   - avoid using the uninitialized/invalid kind for any field that may be
//     observed by encoders or other consumers;
//
//   - use the skip kind only when the intent is to produce a field that
//     encoders will completely ignore.
//
// When constructing fields conditionally, helpers MAY return skip rather
// than omitting the field entirely, as long as all consumers understand that
// skip-typed fields are no-ops at encoding time.
//
// # Consumer responsibilities
//
// Code that consumes fields (most notably encoders) MUST:
//
//   - switch on Type to determine how to interpret the field;
//
//   - apply the semantics documented for each Type constant, including which
//     storage slot(s) to read and which Go types to expect;
//
//   - treat mismatched Type/storage combinations as bugs in the producing
//     code and, where reasonable, fail loudly rather than silently;
//
//   - respect the skip sentinel by emitting nothing for such fields;
//
//   - respect the inline sentinel by expanding the nested object's fields
//     into the current encoding context rather than introducing an extra
//     layer of nesting.
//
// Encoders MUST NOT attempt to reinterpret the raw storage slots in a way
// that contradicts the Type value (for example, treating an integer as a
// timestamp without the corresponding time-related Type).
//
// # Extensibility and stability
//
// The Type enumeration is designed to be extended over time with additional
// kinds when new use cases arise. When new kinds are introduced, they MUST
// be appended in a way that preserves the existing numeric values, so that
// persisted logs and configuration that refer to existing kinds remain
// valid.
//
// Callers and consumers MUST NOT rely on specific numeric values for Type
// constants. They SHOULD instead work with the named constants or, where
// necessary, treat unknown or future kinds conservatively (for example, by
// falling back to a generic representation or treating them as errors).
//
// In summary, the ftype package defines the type system for field values in
// rxlog. It provides a closed set of well-documented kinds that describe how
// each field's underlying storage should be interpreted and serialized,
// enabling encoders to avoid reflection and maintain predictable behavior
// across the logging pipeline.
package ftype
