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

package api

// LogSchemaVersion identifies the canonical schema version for log records
// and configuration associated with this api module.
//
// This constant is intended as a coarse-grained compatibility marker for
// tools that consume rxlog output (collectors, indexers, dashboards) and
// for configuration expressed in terms of this module's types.
//
// The version string is NOT a Go module or semantic version; it is an
// opaque tag that:
//
//   - MUST remain stable for a given released schema;
//   - MUST be bumped whenever field names, meanings, or required defaults
//     change in a breaking or backward-incompatible way;
//   - MAY remain unchanged across internal or purely additive changes that
//     do not break existing consumers.
//
// Downstream systems MAY use this constant to assert compatibility or to
// select parsing and normalization behavior for different schema versions.
// Implementations that expose schema metadata SHOULD surface this value so
// that operators can reason about log and config compatibility.
const LogSchemaVersion = "rxlog.api.v1"
