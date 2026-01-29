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

// Package name defines abstractions for logical names used in the rxlog
// logging API.
//
// A "name" in this context is a stable, human-readable identifier associated
// with some entity in the logging system: for example, a logger instance, a
// subsystem, a component, or a service. Names are used to distinguish and
// group log records, but this package deliberately does not impose any
// particular naming scheme or uniqueness policy.
//
// # Overview
//
// The package provides two primary abstractions:
//
//   - Namer: an interface implemented by types that can expose a logical
//     name, via a Name() string method.
//
//   - Encoder: a function type responsible for serializing a name (a string)
//     into a buffer.Buffer used by log encoders.
//
// Together these abstractions decouple three concerns:
//
//   - how names are chosen and managed by application code,
//
//   - how names are retrieved from objects participating in the logging
//     pipeline, and
//
//   - how names are rendered into the serialized form of a log entry.
//
// # Name and Namer semantics
//
// The Namer interface is intentionally minimal:
//
//	type Namer interface {
//	    Name() string
//	}
//
// It represents any value that can provide a logical name for itself or for
// some associated entity (such as an underlying logger, component, or
// subsystem). Typical examples include:
//
//   - logger types that expose their configured name,
//
//   - wrapper types that delegate to an underlying named component,
//
//   - configuration objects that carry a label describing their scope.
//
// The Name method has the following semantics:
//
//   - It SHOULD return a concise, human-readable identifier. Names are
//     primarily intended to be read and reasoned about by humans.
//
//   - Implementations MAY return an empty string if no meaningful name is
//     available. Callers MUST be prepared to handle this case gracefully,
//     for example by omitting the name field or substituting a default.
//
//   - Implementations SHOULD return a deterministic value for a given
//     instance and SHOULD NOT change the returned name over the lifetime
//     of that instance, unless such changes are explicitly part of the
//     design.
//
// This package does not define any requirements around global uniqueness or
// naming conventions. Applications are free to adopt their own naming
// strategies (for example, hierarchical names, dot-separated paths, or
// service/component identifiers) as long as they are represented as strings.
//
// # Encoding names
//
// The Encoder type describes how a name string is rendered into bytes:
//
//	type Encoder func(dst *buffer.Buffer, name string) *buffer.Buffer
//
// An Encoder is responsible for taking a name (which may be empty) and
// appending an appropriate serialized representation of that name to the
// provided buffer. Common strategies include:
//
//   - emitting the name as-is (for example, as a JSON string),
//
//   - normalizing or transforming the name (for example, lowercasing or
//     applying a consistent delimiter style),
//
//   - omitting the name entirely when it is empty or otherwise considered
//     inapplicable.
//
// Encoder implementations MUST obey the ownership and lifetime rules of the
// buffer package:
//
//   - The dst argument is owned by the caller for the duration of the call.
//     The Encoder MAY append directly to dst or MAY obtain and return a
//     different *buffer.Buffer if that is more efficient.
//
//   - The Encoder MUST return the *buffer.Buffer that now holds the encoded
//     representation. If this differs from dst, the original dst MUST remain
//     in a valid state according to its own contract.
//
//   - The Encoder MUST NOT call Free on any buffer it receives or returns,
//     and it MUST NOT retain references to those buffers or to slices derived
//     from them beyond the end of the call.
//
// Within these constraints, implementations are free to choose any stable,
// documented mapping from input names to their serialized forms.
//
// # Usage context
//
// The name package is intentionally small and self-contained. It does not
// perform I/O and does not depend on other parts of the logging subsystem.
// Instead, it provides a shared vocabulary for:
//
//   - attaching logical names to objects via the Namer interface, and
//
//   - converting those names into bytes via Encoder for inclusion in
//     serialized log entries.
//
// Higher-level components in the rxlog stack MAY use Namer to obtain names
// for loggers, cores, or other entities, and MAY rely on a configured
// Encoder when including those names in the final log output. By centralizing
// these responsibilities in a dedicated package, rxlog keeps naming concerns
// explicit and composable while allowing applications to define their own
// naming schemes and conventions.
package name
