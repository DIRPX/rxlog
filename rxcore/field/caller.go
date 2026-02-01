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

package field

const (
	// Caller is the field key that encodes caller information (source file,
	// line number, and optionally function name) for the log site.
	//
	// Encoders MAY represent this in a compact form (for example,
	// "file.go:123") or in a more structured form, depending on the chosen
	// format. When caller capture is enabled, producers SHOULD set this field
	// to improve debuggability and source-level navigation.
	Caller = "caller"

	// File is the field key that records the source file name associated
	// with the log site, when represented separately from Caller.
	File = "file"

	// Line is the field key that records the source line number associated
	// with the log site, when represented separately from Caller.
	Line = "line"

	// Function is the field key that records the function or method name
	// associated with the log site.
	//
	// Note: this key is separate from Operation/Op, which describes the
	// logical operation being performed, whereas Function typically reflects
	// the implementation-level symbol name.
	Function = "function"
)
