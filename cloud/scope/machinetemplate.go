/*
Copyright 2018 The Kubernetes Authors.

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

// Package scope implements scope types.
package scope

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	infrav1 "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineTemplateScopeParams defines the input parameters used to create a new MachineTemplateScope.
type MachineTemplateScopeParams struct {
	Client             client.Client
	ClusterGetter      cloud.ClusterGetter
	GCPMachineTemplate *infrav1.GCPMachineTemplate
}

// NewMachineTemplateScope creates a new MachineTemplateScope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewMachineTemplateScope(params MachineTemplateScopeParams) (*MachineTemplateScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a MachineTemplateScope")
	}
	if params.GCPMachineTemplate == nil {
		return nil, errors.New("gcp machinetemplate is required when creating a MachineTemplateScope")
	}

	helper, err := patch.NewHelper(params.GCPMachineTemplate, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &MachineTemplateScope{
		client:             params.Client,
		GCPMachineTemplate: params.GCPMachineTemplate,
		ClusterGetter:      params.ClusterGetter,
		patchHelper:        helper,
	}, nil
}

// MachineTemplateScope defines a scope defined around a machinetemplate and its cluster.
type MachineTemplateScope struct {
	client             client.Client
	patchHelper        *patch.Helper
	ClusterGetter      cloud.ClusterGetter
	GCPMachineTemplate *infrav1.GCPMachineTemplate
}

func (m *MachineTemplateScope) Cloud() cloud.Cloud {
	return m.ClusterGetter.Cloud()
}

func (m *MachineTemplateScope) Name() string {
	return m.GCPMachineTemplate.Name
}

func (m *MachineTemplateScope) Namespace() string {
	return m.GCPMachineTemplate.Namespace
}

//func (m *MachineTemplateScope) Project() string {
//	return m.ClusterGetter.Project()
//}

func (m *MachineTemplateScope) InstanceTemplateSpec() *compute.InstanceTemplate {
	return &compute.InstanceTemplate{
		Name: m.Name(),
		Properties: &compute.InstanceProperties{
			Disks: []*compute.AttachedDisk{
				&compute.AttachedDisk{
					//DiskSizeGb: 10,
					InitializeParams: &compute.AttachedDiskInitializeParams{
						SourceImage: *m.GCPMachineTemplate.Spec.Template.Spec.Image,
					},
					AutoDelete: true,
					Boot:       true,
				},
			},
			MachineType: m.GCPMachineTemplate.Spec.Template.Spec.InstanceType,
			NetworkInterfaces: []*compute.NetworkInterface{
				&compute.NetworkInterface{
					AccessConfigs: []*compute.AccessConfig{
						&compute.AccessConfig{
							NetworkTier: "PREMIUM",
							Type:        "ONE_TO_ONE_NAT",
						},
					},
				},
			},
		},
	}
}

// PatchObject persists the cluster configuration and status.
func (m *MachineTemplateScope) PatchObject() error {
	return m.patchHelper.Patch(context.TODO(), m.GCPMachineTemplate)
}

// Close closes the current scope persisting the cluster configuration and status.
func (m *MachineTemplateScope) Close() error {
	//	ctx, log, done := tele.StartSpanWithLogger(ctx, "scope.MachineTemplateScope.Close")
	//	defer done()
	//
	//	if m.vmssState != nil {
	//		if err := m.applyAzureMachineTemplateMachines(ctx); err != nil {
	//			log.Error(err, "failed to apply changes to the AzureMachineTemplateMachines")
	//			return errors.Wrap(err, "failed to apply changes to AzureMachineTemplateMachines")
	//		}
	//
	//		m.setProvisioningStateAndConditions(m.vmssState.State)
	//		if err := m.updateReplicasAndProviderIDs(ctx); err != nil {
	//			return errors.Wrap(err, "failed to update replicas and providerIDs")
	//		}
	//	}
	//
	//	return m.PatchObject(ctx)

	return m.PatchObject()
}
