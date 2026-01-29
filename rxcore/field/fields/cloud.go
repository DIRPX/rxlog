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
	// CloudProvider is the field key that records the cloud provider
	// (for example, "aws", "gcp", "azure").
	CloudProvider = "cloud_provider"

	// CloudRegion is the field key that records the cloud region in a form
	// aligned with the provider's naming (for example, "us-east-1").
	CloudRegion = "cloud_region"

	// CloudZone is the field key that records the cloud availability zone.
	CloudZone = "cloud_zone"

	// K8sCluster is the field key that records the Kubernetes cluster name.
	K8sCluster = "k8s_cluster"

	// K8sNamespace is the field key that records the Kubernetes namespace.
	K8sNamespace = "k8s_namespace"

	// K8sPod is the field key that records the Kubernetes pod name.
	K8sPod = "k8s_pod"

	// K8sContainer is the field key that records the Kubernetes container
	// name within a pod.
	K8sContainer = "k8s_container"

	// K8sNode is the field key that records the Kubernetes node name.
	K8sNode = "k8s_node"
)
