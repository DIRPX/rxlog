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
	// GRPCService is the field key that records the gRPC service name
	// handling the request.
	GRPCService = "grpc_service"

	// GRPCMethod is the field key that records the gRPC method name.
	GRPCMethod = "grpc_method"

	// GRPCCode is the field key that records the gRPC status code for the
	// RPC (for example, "OK", "Unavailable").
	GRPCCode = "grpc_code"
)
