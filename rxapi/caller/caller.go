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

package caller

// Caller represents source-code location metadata associated with a log
// entry. It captures enough information to identify the originating call
// site (program counter, file, line, function) at the time the log entry
// was created.
//
// A Caller value MAY be attached to a log entry to improve debuggability,
// stack trace correlation, or integration with external tooling (e.g. IDE
// navigation, error aggregators). When Caller.Defined is false, all other
// fields MUST be treated as unspecified and MUST NOT be relied upon.
type Caller struct {
	// Defined reports whether this Caller instance contains valid location
	// information.
	//
	// When Defined is false, the remaining fields (PC, File, Line, Function)
	// MUST be considered undefined and MUST NOT be used for display, routing,
	// or analysis. Callers MAY still carry a zero-valued Caller in this case
	// to explicitly indicate that call-site capture was disabled or unavailable.
	//
	// When Defined is true, the implementation SHOULD ensure that PC, File,
	// and Line are populated consistently, and Function SHOULD be populated
	// when that information is available.
	Defined bool

	// PC is the program counter value associated with the call site that
	// produced the log entry.
	//
	// When Defined is true, PC SHOULD correspond to an address that can be
	// resolved by runtime or symbolization tools into a concrete function and
	// source location. When Defined is false, PC MAY be zero and MUST NOT be
	// inspected or symbolized by consumers.
	//
	// Implementations MAY choose to omit PC (for example, by leaving it at 0)
	// in environments where capturing it is too expensive or not supported.
	PC uintptr

	// File is the source file path corresponding to the call site.
	//
	// When Defined is true and File is non-empty, it SHOULD identify the file
	// that contains the call that produced the log entry. The exact format
	// (absolute vs relative, trimmed prefixes, path separator style) is
	// implementation-defined and SHOULD be documented by the runtime package.
	//
	// When Defined is false, File SHOULD be the empty string and MUST NOT be
	// relied upon by consumers.
	File string

	// Line is the 1-based line number within File corresponding to the call
	// site that produced the log entry.
	//
	// When Defined is true and Line > 0, it SHOULD indicate the source line
	// associated with the call. A value of 0 or a negative value MUST be
	// treated as "unknown line" even if Defined is true.
	//
	// When Defined is false, Line SHOULD be 0 and MUST NOT be interpreted as
	// a valid source line.
	Line int

	// Function is a best-effort, human-readable name of the function that
	// contains the call site.
	//
	// When Defined is true, Function MAY contain a fully-qualified name
	// (including package and receiver) or another implementation-defined
	// representation. Consumers MUST NOT rely on a particular naming format
	// beyond what is explicitly documented by the implementation.
	//
	// When Defined is false, Function SHOULD be the empty string and MUST be
	// treated as unspecified. Even when Defined is true, Function MAY be empty
	// if name resolution is not available in the current environment.
	Function string
}
