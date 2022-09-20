/*
Copyright 2021 The Kubernetes Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// MachineTemplateFinalizer is used to ensure deletion of dependencies (nodes, infra).
	MachineTemplateFinalizer = "gcpmachinetemplate.infrastructure.cluster.x-k8s.io"
)

// GCPMachineTemplateSpec defines the desired state of GCPMachineTemplate.
type GCPMachineTemplateSpec struct {
	Template    GCPMachineTemplateResource `json:"template"`
	ClusterName string                     `json:"clusterName"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=gcpmachinetemplates,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status

// GCPMachineTemplate is the Schema for the gcpmachinetemplates API.
type GCPMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPMachineTemplateSpec   `json:"spec,omitempty"`
	Status GCPMachineTemplateStatus `json:"status,omitempty"`
}

// GCPMachineTemplateStatus defines the observed state of GCPMachineTemplate.
type GCPMachineTemplateStatus struct {
	References References `json:"references,omitempty"`
	Ready      bool       `json:"ready"`
}

// GCPMachineTemplateList contains a list of GCPMachineTemplate.
type GCPMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPMachineTemplate `json:"items"`
}

type References struct {
	GCPMachinePools map[string]string `json:"gcpMachinePools,omitempty"`
}

func init() {
	SchemeBuilder.Register(&GCPMachineTemplate{}, &GCPMachineTemplateList{})
}
