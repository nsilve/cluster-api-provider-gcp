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
	"sigs.k8s.io/cluster-api-provider-gcp/cloud"
	infrav1exp "sigs.k8s.io/cluster-api-provider-gcp/exp/api/v1beta1"
	clusterv1exp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MachinePoolScopeParams defines the input parameters used to create a new MachinePoolScope.
type MachinePoolScopeParams struct {
	Client         client.Client
	ClusterGetter  cloud.ClusterGetter
	MachinePool    *clusterv1exp.MachinePool
	GCPMachinePool *infrav1exp.GCPMachinePool
}

// NewMachinePoolScope creates a new MachinePoolScope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewMachinePoolScope(params MachinePoolScopeParams) (*MachinePoolScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a MachinePoolScope")
	}
	if params.MachinePool == nil {
		return nil, errors.New("machinepool is required when creating a MachinePoolScope")
	}
	if params.GCPMachinePool == nil {
		return nil, errors.New("gcp machinepool is required when creating a MachinePoolScope")
	}

	helper, err := patch.NewHelper(params.GCPMachinePool, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &MachinePoolScope{
		client:         params.Client,
		MachinePool:    params.MachinePool,
		GCPMachinePool: params.GCPMachinePool,
		ClusterGetter:  params.ClusterGetter,
		patchHelper:    helper,
	}, nil
}

// MachinePoolScope defines a scope defined around a machinepool and its cluster.
type MachinePoolScope struct {
	client         client.Client
	patchHelper    *patch.Helper
	ClusterGetter  cloud.ClusterGetter
	MachinePool    *clusterv1exp.MachinePool
	GCPMachinePool *infrav1exp.GCPMachinePool
}

//
//func (m *MachinePoolScope) Region() string {
//	return m.ClusterGetter.Region()
//}
//
//func (m *MachinePoolScope) NetworkName() string {
//	return m.ClusterGetter.NetworkName()
//}
//
//func (m *MachinePoolScope) Network() *infrav1.Network {
//	return m.ClusterGetter.Network()
//}
//
//func (m *MachinePoolScope) AdditionalLabels() infrav1.Labels {
//	return m.ClusterGetter.AdditionalLabels()
//}

//
//func (m *MachinePoolScope) FailureDomains() clusterv1.FailureDomains {
//	return m.ClusterGetter.FailureDomains()
//}

//
//func (m *MachinePoolScope) TargetTCPProxySpec() *compute.TargetTcpProxy {
//	//TODO implement me
//	panic("implement me")
//}

// ANCHOR: MachineGetter

//// Cloud returns initialized cloud.
//func (m *MachinePoolScope) Cloud() cloud.Cloud {
//	return m.ClusterGetter.Cloud()
//}

//// Zone returns the FailureDomain for the GCPMachinePool.
//func (m *MachinePoolScope) Zone() string {
//	//////////
//	if m.MachinePool.Spec.FailureDomains == nil {
//		fd := m.ClusterGetter.FailureDomains()
//		if len(fd) == 0 {
//			return ""
//		}
//		zones := make([]string, 0, len(fd))
//		for zone := range fd {
//			zones = append(zones, zone)
//		}
//		sort.Strings(zones)
//		return zones[0]
//	}
//	return m.MachinePool.Spec.FailureDomains[0]
//}
//
//// Project return the project for the GCPMachinePool's cluster.
//func (m *MachinePoolScope) Project() string {
//	return m.ClusterGetter.Project()
//}
//
//// Name returns the GCPMachinePool name.
//func (m *MachinePoolScope) Name() string {
//	return m.GCPMachinePool.Name
//}
//
//// Namespace returns the namespace name.
//func (m *MachinePoolScope) Namespace() string {
//	return m.GCPMachinePool.Namespace
//}

//// WorkerGroupName returns the worker instance group name.
//func (m *MachinePoolScope) WorkerGroupName(zone string) string {
//	return fmt.Sprintf("%s-%s-%s", m.ClusterGetter.Name(), infrav1.WorkerRoleTagValue, zone)
//}
//
//// IsControlPlane returns true if the machine is a control plane.
//func (m *MachinePoolScope) IsControlPlane() bool {
//	//return util.IsControlPlaneMachine(m.Machine)
//	return false ////
//}
//
//// Role returns the machine role from the labels.
//func (m *MachinePoolScope) Role() string {
//	//if util.IsControlPlaneMachine(m.Machine) {
//	//	return "control-plane"
//	//}
//
//	return "node"
//}
//
//// GetInstanceID returns the GCPMachine instance id by parsing Spec.ProviderID.
//func (m *MachinePoolScope) GetInstanceID() *string {
//	parsed, err := noderefutil.NewProviderID(m.GetProviderID())
//	if err != nil {
//		return nil
//	}
//
//	return pointer.StringPtr(parsed.ID())
//}
//
//// GetProviderID returns the GCPMachine providerID from the spec.
//func (m *MachinePoolScope) GetProviderID() string {
//	//if m.GCPMachine.Spec.ProviderID != nil {
//	//	return *m.GCPMachine.Spec.ProviderID
//	//}
//
//	return ""
//}

// ANCHOR_END: MachineGetter

// ANCHOR: MachineSetter

// SetProviderID sets the GCPMachine providerID in spec.
//func (m *MachinePoolScope) SetProviderID() {
//	providerID := cloud.ProviderIDPrefix + path.Join(m.ClusterGetter.Project(), m.Zone(), m.Name())
//	m.GCPMachinePool.Spec.ProviderID = pointer.StringPtr(providerID)
//}
//
//// GetInstanceStatus returns the GCPMachinePool instance status.
//func (m *MachinePoolScope) GetInstanceStatus() *infrav1.InstanceStatus {
//	return m.GCPMachinePool.Status.InstanceStatus
//}
//
//// SetInstanceStatus sets the GCPMachinePool instance status.
//func (m *MachinePoolScope) SetInstanceStatus(v infrav1.InstanceStatus) {
//	m.GCPMachinePool.Status.InstanceStatus = &v
//}
//
//// SetReady sets the GCPMachinePool Ready Status.
//func (m *MachinePoolScope) SetReady() {
//	m.GCPMachinePool.Status.Ready = true
//}
//
//// SetFailureMessage sets the GCPMachinePool status failure message.
//func (m *MachinePoolScope) SetFailureMessage(v error) {
//	m.GCPMachinePool.Status.FailureMessage = pointer.StringPtr(v.Error())
//}
//
//// SetFailureReason sets the GCPMachinePool status failure reason.
//func (m *MachinePoolScope) SetFailureReason(v capierrors.MachineStatusError) {
//	m.GCPMachinePool.Status.FailureReason = &v
//}
//
//// SetAnnotation sets a key value annotation on the GCPMachinePool.
//func (m *MachinePoolScope) SetAnnotation(key, value string) {
//	if m.GCPMachinePool.Annotations == nil {
//		m.GCPMachinePool.Annotations = map[string]string{}
//	}
//	m.GCPMachinePool.Annotations[key] = value
//}
//
//// SetAddresses sets the addresses field on the GCPMachinePool.
//func (m *MachinePoolScope) SetAddresses(addressList []corev1.NodeAddress) {
//	m.GCPMachinePool.Status.Addresses = addressList
//}

// ANCHOR_END: MachineSetter

// ANCHOR: MachineInstanceSpec

//// InstanceImageSpec returns compute instance image attched-disk spec.
//func (m *MachinePoolScope) InstanceImageSpec() *compute.AttachedDisk {
//	version := ""
//	if m.Machine.Spec.Version != nil {
//		version = *m.Machine.Spec.Version
//	}
//	image := "capi-ubuntu-1804-k8s-" + strings.ReplaceAll(semver.MajorMinor(version), ".", "-")
//	sourceImage := path.Join("projects", m.ClusterGetter.Project(), "global", "images", "family", image)
//	if m.GCPMachine.Spec.Image != nil {
//		sourceImage = *m.GCPMachine.Spec.Image
//	} else if m.GCPMachine.Spec.ImageFamily != nil {
//		sourceImage = *m.GCPMachine.Spec.ImageFamily
//	}
//
//	diskType := infrav1.PdStandardDiskType
//	if t := m.GCPMachine.Spec.RootDeviceType; t != nil {
//		diskType = *t
//	}
//
//	return &compute.AttachedDisk{
//		AutoDelete: true,
//		Boot:       true,
//		InitializeParams: &compute.AttachedDiskInitializeParams{
//			DiskSizeGb:  m.GCPMachine.Spec.RootDeviceSize,
//			DiskType:    path.Join("zones", m.Zone(), "diskTypes", string(diskType)),
//			SourceImage: sourceImage,
//		},
//	}
//}
//
//// InstanceAdditionalDiskSpec returns compute instance additional attched-disk spec.
//func (m *MachinePoolScope) InstanceAdditionalDiskSpec() []*compute.AttachedDisk {
//	additionalDisks := make([]*compute.AttachedDisk, 0, len(m.GCPMachine.Spec.AdditionalDisks))
//	for _, disk := range m.GCPMachine.Spec.AdditionalDisks {
//		additionalDisk := &compute.AttachedDisk{
//			AutoDelete: true,
//			InitializeParams: &compute.AttachedDiskInitializeParams{
//				DiskSizeGb: pointer.Int64PtrDerefOr(disk.Size, 30),
//				DiskType:   path.Join("zones", m.Zone(), "diskTypes", string(*disk.DeviceType)),
//			},
//		}
//		if additionalDisk.InitializeParams.DiskType == string(infrav1.LocalSsdDiskType) {
//			additionalDisk.Type = "SCRATCH" // Default is PERSISTENT.
//			// Override the Disk size
//			additionalDisk.InitializeParams.DiskSizeGb = 375
//			// For local SSDs set interface to NVME (instead of default SCSI) which is faster.
//			// Most OS images would work with both NVME and SCSI disks but some may work
//			// considerably faster with NVME.
//			// https://cloud.google.com/compute/docs/disks/local-ssd#choose_an_interface
//			additionalDisk.Interface = "NVME"
//		}
//		additionalDisks = append(additionalDisks, additionalDisk)
//	}
//
//	return additionalDisks
//}
//
//// InstanceNetworkInterfaceSpec returns compute network interface spec.
//func (m *MachinePoolScope) InstanceNetworkInterfaceSpec() *compute.NetworkInterface {
//	networkInterface := &compute.NetworkInterface{
//		Network: path.Join("projects", m.ClusterGetter.Project(), "global", "networks", m.ClusterGetter.NetworkName()),
//	}
//
//	if m.GCPMachine.Spec.PublicIP != nil && *m.GCPMachine.Spec.PublicIP {
//		networkInterface.AccessConfigs = []*compute.AccessConfig{
//			{
//				Type: "ONE_TO_ONE_NAT",
//				Name: "External NAT",
//			},
//		}
//	}
//
//	if m.GCPMachine.Spec.Subnet != nil {
//		networkInterface.Subnetwork = path.Join("regions", m.ClusterGetter.Region(), "subnetworks", *m.GCPMachine.Spec.Subnet)
//	}
//
//	return networkInterface
//}
//
//// InstanceServiceAccountsSpec returns service-account spec.
//func (m *MachinePoolScope) InstanceServiceAccountsSpec() *compute.ServiceAccount {
//	serviceAccount := &compute.ServiceAccount{
//		Email: "default",
//		Scopes: []string{
//			compute.CloudPlatformScope,
//		},
//	}
//
//	if m.GCPMachine.Spec.ServiceAccount != nil {
//		serviceAccount.Email = m.GCPMachine.Spec.ServiceAccount.Email
//		serviceAccount.Scopes = m.GCPMachine.Spec.ServiceAccount.Scopes
//	}
//
//	return serviceAccount
//}
//
//// InstanceAdditionalMetadataSpec returns additional metadata spec.
//func (m *MachinePoolScope) InstanceAdditionalMetadataSpec() *compute.Metadata {
//	metadata := new(compute.Metadata)
//	for _, additionalMetadata := range m.GCPMachine.Spec.AdditionalMetadata {
//		metadata.Items = append(metadata.Items, &compute.MetadataItems{
//			Key:   additionalMetadata.Key,
//			Value: additionalMetadata.Value,
//		})
//	}
//
//	return metadata
//}
//
//// InstanceSpec returns instance spec.
//func (m *MachinePoolScope) InstanceSpec() *compute.Instance {
//	instance := &compute.Instance{
//		Name:        m.Name(),
//		Zone:        m.Zone(),
//		MachineType: path.Join("zones", m.Zone(), "machineTypes", m.GCPMachine.Spec.InstanceType),
//		Tags: &compute.Tags{
//			Items: append(
//				m.GCPMachine.Spec.AdditionalNetworkTags,
//				fmt.Sprintf("%s-%s", m.ClusterGetter.Name(), m.Role()),
//				m.ClusterGetter.Name(),
//			),
//		},
//		Labels: infrav1.Build(infrav1.BuildParams{
//			ClusterName: m.ClusterGetter.Name(),
//			Lifecycle:   infrav1.ResourceLifecycleOwned,
//			Role:        pointer.StringPtr(m.Role()),
//			// TODO(vincepri): Check what needs to be added for the cloud provider label.
//			Additional: m.ClusterGetter.AdditionalLabels().AddLabels(m.GCPMachine.Spec.AdditionalLabels),
//		}),
//		Scheduling: &compute.Scheduling{
//			Preemptible: m.GCPMachine.Spec.Preemptible,
//		},
//	}
//
//	instance.CanIpForward = true
//	if m.GCPMachine.Spec.IPForwarding != nil && *m.GCPMachine.Spec.IPForwarding == infrav1.IPForwardingDisabled {
//		instance.CanIpForward = false
//	}
//
//	instance.Disks = append(instance.Disks, m.InstanceImageSpec())
//	instance.Disks = append(instance.Disks, m.InstanceAdditionalDiskSpec()...)
//	instance.Metadata = m.InstanceAdditionalMetadataSpec()
//	instance.ServiceAccounts = append(instance.ServiceAccounts, m.InstanceServiceAccountsSpec())
//	instance.NetworkInterfaces = append(instance.NetworkInterfaces, m.InstanceNetworkInterfaceSpec())
//	return instance
//}
//
//// ANCHOR_END: MachineInstanceSpec
//
//// GetBootstrapData returns the bootstrap data from the secret in the Machine's bootstrap.dataSecretName.
//func (m *MachinePoolScope) GetBootstrapData() (string, error) {
//	if m.Machine.Spec.Bootstrap.DataSecretName == nil {
//		return "", errors.New("error retrieving bootstrap data: linked Machine's bootstrap.dataSecretName is nil")
//	}
//
//	secret := &corev1.Secret{}
//	key := types.NamespacedName{Namespace: m.Namespace(), Name: *m.Machine.Spec.Bootstrap.DataSecretName}
//	if err := m.client.Get(context.TODO(), key, secret); err != nil {
//		return "", errors.Wrapf(err, "failed to retrieve bootstrap data secret for GCPMachine %s/%s", m.Namespace(), m.Name())
//	}
//
//	value, ok := secret.Data["value"]
//	if !ok {
//		return "", errors.New("error retrieving bootstrap data: secret value key is missing")
//	}
//
//	return string(value), nil
//}

// PatchObject persists the cluster configuration and status.
func (m *MachinePoolScope) PatchObject() error {
	return m.patchHelper.Patch(context.TODO(), m.GCPMachinePool)
}

// Close closes the current scope persisting the cluster configuration and status.
func (m *MachinePoolScope) Close() error {
	//	ctx, log, done := tele.StartSpanWithLogger(ctx, "scope.MachinePoolScope.Close")
	//	defer done()
	//
	//	if m.vmssState != nil {
	//		if err := m.applyAzureMachinePoolMachines(ctx); err != nil {
	//			log.Error(err, "failed to apply changes to the AzureMachinePoolMachines")
	//			return errors.Wrap(err, "failed to apply changes to AzureMachinePoolMachines")
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
