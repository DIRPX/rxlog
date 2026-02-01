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
	// Duration is the field key that records the elapsed time for an
	// operation associated with the log entry.
	//
	// Encoders SHOULD use a well-defined unit or textual format for the
	// stored value (for example, a time.Duration string or a numeric value in
	// milliseconds) and SHOULD document that choice. Producers typically set
	// this for request/response latency, background job runtime, or other
	// measured operations.
	Duration = "duration"

	// Latency is the field key that records the end-to-end latency for a
	// user-visible operation or request.
	//
	// While similar to Duration, Latency is often reserved for higher-level
	// round-trip measurements.
	Latency = "latency"

	// Attempt is the field key that records the current attempt number for
	// retrying operations.
	Attempt = "attempt"

	// Retry is the field key that records whether an operation is being
	// retried.
	Retry = "retry"

	// SampleRate is the field key that records the effective sampling rate
	// for the log entry (for example, 1.0 for "always", 0.1 for 10%).
	SampleRate = "sample_rate"

	// SamplingDecision is the field key that records whether the event was
	// sampled in or out, and why.
	SamplingDecision = "sampling_decision"
)
