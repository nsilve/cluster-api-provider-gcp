/*
Copyright The Kubernetes Authors.

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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GCPMachinePoolSpec defines the desired state of GCPMachinePool
type GCPMachinePoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster

	// Foo is an example fiGCPMachine instance is runningeld of GCPMachinePool. Edit gcpmachinepool_types.go to remove/update
	Foo string `json:"foo,omitempty"`

	// +optional
	InfrastructureRef *corev1.ObjectReference `json:"infrastructureRef,omitempty"`

	// ProviderID is the identification ID of the Managed Instance Group
	// +optional
	ProviderID string `json:"providerID,omitempty"`

	// ProviderIDList are the identification IDs of machine instances provided by the provider.
	// This field must match the provider IDs as seen on the node objects corresponding to a machine pool's machine instances.
	// +optional
	ProviderIDList []string `json:"providerIDList,omitempty"`

	// Region is the region of the Managed Instance Group
	// +optional
	Region string `json:"region,omitempty"`

	// Zone is the zone of the Managed Instance Group
	// +optional
	Zone string `json:"zone,omitempty"`
}

// GCPMachinePoolStatus defines the observed state of GCPMachinePool
type GCPMachinePoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Replicas is the most recently observed number of replicas
	// +optional
	Replicas int32 `json:"replicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=gcpmachinepools,scope=Namespaced,categories=cluster-api,shortName=gcpmp
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Replicas",type="string",JSONPath=".status.replicas",description="GCPMachinePool replicas count"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="GCPMachinePool replicas count"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.provisioningState",description="GCP Instance Group provisioning state"
// +kubebuilder:printcolumn:name="Cluster",type="string",priority=1,JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this GCPMachinePool belongs"
// +kubebuilder:printcolumn:name="MachinePool",type="string",priority=1,JSONPath=".metadata.ownerReferences[?(@.kind==\"MachinePool\")].name",description="MachinePool object to which this GCPMachinePool belongs"
// +kubebuilder:printcolumn:name="Instance Group ID",type="string",priority=1,JSONPath=".spec.providerID",description="GCP Instance Group ID"
// +kubebuilder:printcolumn:name="VM Size",type="string",priority=1,JSONPath=".spec.template.vmSize",description="GCP VM Size"

// GCPMachinePool is the Schema for the gcpmachinepools API
type GCPMachinePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPMachinePoolSpec   `json:"spec,omitempty"`
	Status GCPMachinePoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GCPMachinePoolList contains a list of GCPMachinePool
type GCPMachinePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPMachinePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPMachinePool{}, &GCPMachinePoolList{})
}
