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

package fields

const (
	// Service is the field key that identifies the logical application,
	// service, or component emitting the log entry.
	//
	// Producers SHOULD use a stable, human-readable name (for example,
	// "router", "auth", "billing-rxapi") so that operators and tooling can
	// reliably group and filter logs by service. The value under this key
	// SHOULD remain consistent across instances of the same service.
	Service = "service"

	// ServiceNamespace is the field key that identifies a logical grouping
	// or namespace for the service (for example, "payments", "edge").
	//
	// This is particularly useful in multi-service architectures where the
	// same service name might appear in different domains or namespaces.
	ServiceNamespace = "service_namespace"

	// Version is the field key that records the version of the running
	// service or binary.
	//
	// Typical values include semantic versions (for example, "1.2.3"), git
	// commit hashes, or build numbers. Producers SHOULD ensure that the value
	// is precise enough to uniquely identify the deployed code, enabling
	// troubleshooting and rollback analysis.
	Version = "version"

	// BuildID is the field key that records the build identifier used to
	// produce the binary (for example, a CI build number or artifact ID).
	//
	// This is often used together with Version for precise provenance.
	BuildID = "build_id"

	// Env is the field key that describes the runtime environment in which
	// the service operates.
	//
	// Typical values include "prod", "staging", "qa", or "dev". Producers
	// SHOULD populate this field so that logs from different environments can
	// be easily separated and handled according to different retention or
	// alerting policies.
	Env = "env"

	// Region is the field key that indicates the geographic or logical
	// region in which the service instance is running.
	//
	// Producers SHOULD use this key in multi-region deployments to support
	// region-scoped analysis, latency investigations, and failover audits.
	// The value format (for example, "us-east-1", "europe-west1") is
	// infrastructure-specific.
	Region = "region"

	// Zone is the field key that records the availability zone or similar
	// subdivision within a region.
	//
	// Example values include "us-east-1a" or "europe-west1-b". This is
	// useful for pinpointing zone-local issues.
	Zone = "zone"

	// Cluster is the field key that identifies the logical cluster in which
	// the workload is running (for example, a Kubernetes cluster name).
	Cluster = "cluster"

	// Project is the field key that records the higher-level project or
	// system name under which the service is deployed (for example, a cloud
	// project or account).
	Project = "project"

	// NodeID is the field key that identifies the node, host, or machine on
	// which the process is running.
	//
	// Producers SHOULD set this to a value that correlates with infrastructure
	// identifiers (for example, hostname, VM name, or node ID) so that logs
	// can be joined with system-level metrics and events.
	NodeID = "node_id"

	// HostName is the field key that records the hostname as reported by
	// the operating system.
	//
	// In many environments this is similar to NodeID, but the two may differ
	// depending on naming schemes. Producers MAY record both when useful.
	HostName = "host"

	// InstanceID is the field key that identifies a specific instance of the
	// service.
	//
	// Typical values include pod names, container IDs, or replica IDs. In
	// horizontally scaled deployments, producers SHOULD populate this field
	// to distinguish between different instances of the same service in the
	// same environment and region.
	InstanceID = "instance_id"

	// ProcessID is the field key that records the operating system process
	// identifier (PID) of the running service.
	//
	// This can assist with low-level debugging and correlation with system
	// metrics.
	ProcessID = "pid"

	// ProcessName is the field key that records the executable or logical
	// process name.
	ProcessName = "process"

	// Runtime is the field key that identifies the runtime in use (for
	// example, "go").
	Runtime = "runtime"

	// RuntimeVersion is the field key that records the version of the runtime
	// (for example, "go1.23.0").
	RuntimeVersion = "runtime_version"
)
