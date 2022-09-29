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

package managedinstancegroups

import (
	"context"
	"fmt"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud/gcperrors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

// Reconcile reconcile cluster managedinstancegroup components.
func (s *Service) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)

	log.Info("Reconciling instancetemplate resources")
	_, err := s.createOrRecreateInstanceTemplates(ctx)
	if err != nil {
		return err
	}

	log.Info("Reconciling managedinstancegroup resources")
	_, err = s.createOrUpdateManagedInstanceGroups(ctx)
	if err != nil {
		return err
	}

	//healthcheck, err := s.createOrGetHealthCheck(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//backendsvc, err := s.createOrGetBackendService(ctx, managedinstancegroups, healthcheck)
	//if err != nil {
	//	return err
	//}
	//
	//target, err := s.createOrGetTargetTCPProxy(ctx, backendsvc)
	//if err != nil {
	//	return err
	//}
	//
	//addr, err := s.createOrGetAddress(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//return s.createForwardingRule(ctx, target, addr)

	return nil
}

// Delete delete cluster control-plane managedinstancegroup compoenents.
func (s *Service) Delete(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Deleting managedinstancegroup resources")
	if err := s.deleteManagedInstanceGroups(ctx); err != nil {
		return err
	}

	log.Info("Deleting instancetemplate resources")
	return s.deleteInstanceTemplates(ctx)
}

//
//func (s *Service) createOrRecreateInstanceTemplates(ctx context.Context) (*compute.InstanceTemplate, error) {
//	/// name should be gcpmachinepool_name + hash
//	/// at the end(?) IT housekeeping is needed
//	log := log.FromContext(ctx)
//
//	instancetemplateSpec := s.scope.InstanceTemplateSpec()
//	log.V(2).Info("Looking for instancetemplate", "name", instancetemplateSpec.Name)
//	instancetemplate, err := s.instancetemplates.Get(ctx, meta.GlobalKey(instancetemplateSpec.Name))
//	if err != nil {
//		if !gcperrors.IsNotFound(err) {
//			log.Error(err, "Error looking for instancetemplate")
//			return nil, err
//		}
//
//		log.V(2).Info("Creating instancetemplate", "name", instancetemplateSpec.Name)
//		if err := s.instancetemplates.Insert(ctx, meta.GlobalKey(instancetemplateSpec.Name), instancetemplateSpec); err != nil {
//			log.Error(err, "Error creating instancetemplate", "name", instancetemplateSpec.Name)
//			return nil, err
//		}
//
//		instancetemplate, err = s.instancetemplates.Get(ctx, meta.GlobalKey(instancetemplateSpec.Name))
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	return instancetemplate, nil
//}

func (s *Service) createOrUpdateManagedInstanceGroups(ctx context.Context) (*compute.InstanceGroupManager, error) {
	log := log.FromContext(ctx)
	///// commented out to speed-up testing
	//fd := s.scope.FailureDomains()
	//zones := make([]string, 0, len(fd))
	//for zone := range fd {
	//	zones = append(zones, zone)
	//}
	//
	//groups := make([]*compute.InstanceGroupManager, 0, len(zones))
	//groupsMap := s.scope.Network().WorkerInstanceGroups
	//if groupsMap == nil {
	//	groupsMap = make(map[string]string)
	//}

	//zones := []string{"us-central1-a"}
	//
	//groups := make([]*compute.InstanceGroupManager, 0, len(zones))
	//groupsMap := s.scope.Network().WorkerInstanceGroups
	//if groupsMap == nil {
	//	groupsMap = make(map[string]string)
	//}
	//
	//for _, zone := range zones {

	//instancegroupSpec := s.scope.ManagedInstanceGroupSpec()
	//zone := instancegroupSpec.Zone[strings.LastIndex(instancegroupSpec.Zone, "/")+1:]
	//log.V(2).Info("Looking for managedinstancegroup in zone", "zone", zone, "name", instancegroupSpec.Name)
	//instancegroup, err := s.managedinstancegroups.Get(ctx, meta.ZonalKey(instancegroupSpec.Name, zone))
	//if err != nil {
	//	if !gcperrors.IsNotFound(err) {
	//		log.Error(err, "Error looking for managedinstancegroup in zone", "zone", zone)
	//		return instancegroup, err
	//	}
	//
	//	log.V(2).Info("Creating managedinstancegroup in zone", "zone", zone, "name", instancegroupSpec.Name)
	//	if err := s.managedinstancegroups.Insert(ctx, meta.ZonalKey(instancegroupSpec.Name, zone), instancegroupSpec); err != nil {
	//		log.Error(err, "Error creating managedinstancegroup", "name", instancegroupSpec.Name)
	//		return instancegroup, err
	//	}
	//
	//	instancegroup, err = s.managedinstancegroups.Get(ctx, meta.ZonalKey(instancegroupSpec.Name, zone))
	//	if err != nil {
	//		return instancegroup, err
	//	}
	//}
	//
	//return instancegroup, nil

	//////////////////////

	//log.V(2).Info("Getting bootstrap data for machinepool")
	//bootstrapData, err := s.scope.GetBootstrapData()
	//if err != nil {
	//	log.Error(err, "Error getting bootstrap data for machinepool")
	//	return nil, errors.Wrap(err, "failed to retrieve bootstrap data")
	//}

	regionManagedInstanceGroupSpec := s.scope.RegionManagedInstanceGroupSpec()

	log.V(2).Info("Looking for regionmanagedinstancegroup in region", "region", regionManagedInstanceGroupSpec.Region, "name", regionManagedInstanceGroupSpec.Name)
	regioninstancegroup, err := s.regionmanagedinstancegroups.Get(ctx, meta.RegionalKey(regionManagedInstanceGroupSpec.Name, regionManagedInstanceGroupSpec.Region))
	if err != nil {
		if !gcperrors.IsNotFound(err) {
			log.Error(err, "Error looking for regionmanagedinstancegroup in region", "region", regionManagedInstanceGroupSpec.Region)
			return regioninstancegroup, err
		}
		log.V(2).Info("Creating regionmanagedinstancegroup in region", "region", regionManagedInstanceGroupSpec.Region, "name", regionManagedInstanceGroupSpec.Name)
		if err = s.regionmanagedinstancegroups.Insert(ctx, meta.RegionalKey(regionManagedInstanceGroupSpec.Name, regionManagedInstanceGroupSpec.Region), regionManagedInstanceGroupSpec); err != nil {
			log.Error(err, "Error creating regionmanagedinstancegroup", "name", regionManagedInstanceGroupSpec.Name)
			return regioninstancegroup, err
		}
		regioninstancegroup, err = s.regionmanagedinstancegroups.Get(ctx, meta.RegionalKey(regionManagedInstanceGroupSpec.Name, regionManagedInstanceGroupSpec.Region))
		if err != nil {
			return regioninstancegroup, err
		}
	}

	if s.isRegionManagedInstanceGroupEqual(ctx, regionManagedInstanceGroupSpec, regioninstancegroup) {
		log.Info("mig equal")
	} else {
		log.Info("mig non equal")

		//s.scope.SetInstanceTemplateName(regionManagedInstanceGroupSpec.Versions[0].InstanceTemplate)
		if err = s.regionmanagedinstancegroups.SetInstanceTemplate(ctx, meta.RegionalKey(
			regionManagedInstanceGroupSpec.Name,
			regionManagedInstanceGroupSpec.Region),
			&compute.RegionInstanceGroupManagersSetTemplateRequest{
				InstanceTemplate: s.scope.GetFullInstanceTemplateName(),
			}); err != nil {
			log.Error(err, "Error setting instance template to regionmanagedinstancegroup", "mig", regionManagedInstanceGroupSpec.Name, "instance template", s.scope.GetInstanceTemplateName())
			return regioninstancegroup, err
		}
	}

	return regioninstancegroup, nil
	//if err = s.regionmanagedinstancegroups.Insert(ctx, meta.RegionalKey(regioninstancegroupSpec.Name, regioninstancegroupSpec.Region), regioninstancegroupSpec); err != nil {
	//	log.Error(err, "Error creating regionmanagedinstancegroup", "name", regioninstancegroupSpec.Name)
	//}
	//////////////////////

	//groups = append(groups, instancegroup)
	//groupsMap[zone] = instancegroup.SelfLink
	//}
	//
	//s.scope.Network().WorkerInstanceGroups = groupsMap
	//
	//return groups, nil
}

func (s *Service) isRegionManagedInstanceGroupEqual(ctx context.Context, spec *compute.InstanceGroupManager, instanceGroupManager *compute.InstanceGroupManager) bool {
	//log := log.FromContext(ctx)

	specInstanceTemplate := spec.Versions[0].InstanceTemplate
	instanceGroupManagerInstanceTemplate := instanceGroupManager.Versions[0].InstanceTemplate[strings.Index(instanceGroupManager.Versions[0].InstanceTemplate, "/global/")+1:]

	//log.Info("IT diff", "specit", specInstanceTemplate, "igmit", instanceGroupManagerInstanceTemplate)

	return specInstanceTemplate == instanceGroupManagerInstanceTemplate
}

//func (s *Service) GetManagedInstanceGroups(ctx context.Context, key *meta.Key) (*compute.InstanceGroupManager, error) {
//	return s.managedinstancegroups.Get(ctx, key)
//}

func (s *Service) GetRegionManagedInstanceGroups(ctx context.Context, key *meta.Key) (*compute.InstanceGroupManager, error) {
	return s.regionmanagedinstancegroups.Get(ctx, key)
}

//
//func (s *Service) registerWorkerInstance(ctx context.Context, instancegroup *compute.InstanceGroup) error {
//	log := log.FromContext(ctx)
//	instancegroupName := s.scope.WorkerGroupName()
//	//log.V(2).Info("Ensuring instance already registered in the instancegroup", "name", instance.Name, "instancegroup", instancegroupName)
//	//instancegroupKey := meta.ZonalKey(instancegroupName, s.scope.Zone())
//	//instanceList, err := s.instancegroups.ListInstances(ctx, instancegroupKey, &compute.InstanceGroupsListInstancesRequest{
//	//	InstanceState: "RUNNING",
//	//}, filter.None)
//	//if err != nil {
//	//	log.Error(err, "Error retrieving list of instances in the instancegroup", "instancegroup", instancegroupName)
//	//	return err
//	//}
//
//	instancegroupSets := sets.NewString()
//	instancegroupSets.Insert(*instancegroup.
//	//defer instancegroupSets.Delete()
//	//for _, i := range instanceList {
//	//	instanceSets.Insert(i.Instance)
//	//}
//
//	if !instanceSets.Has(instance.SelfLink) && instance.Status == string(infrav1.InstanceStatusRunning) {
//		log.V(2).Info("Registering instance in the instancegroup", "name", instance.Name, "instancegroup", instancegroupName)
//		if err := s.instancegroups.AddInstances(ctx, instancegroupKey, &compute.InstanceGroupsAddInstancesRequest{
//			Instances: []*compute.InstanceReference{
//				{
//					Instance: instance.SelfLink,
//				},
//			},
//		}); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (s *Service) deregisterWorkerInstance(ctx context.Context, instancegroup *compute.InstanceGroup) error {
//	log := log.FromContext(ctx)
//	instancegroupName := s.scope.WorkerGroupName()
//	log.V(2).Info("Ensuring instance already registered in the instancegroup", "name", instance.Name, "instancegroup", instancegroupName)
//	instancegroupKey := meta.ZonalKey(instancegroupName, s.scope.Zone())
//	instanceList, err := s.instancegroups.ListInstances(ctx, instancegroupKey, &compute.InstanceGroupsListInstancesRequest{
//		InstanceState: "RUNNING",
//	}, filter.None)
//	if err != nil {
//		return gcperrors.IgnoreNotFound(err)
//	}
//
//	instanceSets := sets.NewString()
//	defer instanceSets.Delete()
//	for _, i := range instanceList {
//		instanceSets.Insert(i.Instance)
//	}
//
//	if len(instanceSets.List()) > 0 && instanceSets.Has(instance.SelfLink) {
//		log.V(2).Info("Deregistering instance in the instancegroup", "name", instance.Name, "instancegroup", instancegroupName)
//		if err := s.instancegroups.RemoveInstances(ctx, instancegroupKey, &compute.InstanceGroupsRemoveInstancesRequest{
//			Instances: []*compute.InstanceReference{
//				{
//					Instance: instance.SelfLink,
//				},
//			},
//		}); err != nil {
//			return gcperrors.IgnoreNotFound(err)
//		}
//	}
//
//	return nil
//}

//
//func (s *Service) createOrGetHealthCheck(ctx context.Context) (*compute.HealthCheck, error) {
//	log := log.FromContext(ctx)
//	healthcheckSpec := s.scope.HealthCheckSpec()
//	log.V(2).Info("Looking for healthcheck", "name", healthcheckSpec.Name)
//	healthcheck, err := s.healthchecks.Get(ctx, meta.GlobalKey(healthcheckSpec.Name))
//	if err != nil {
//		if !gcperrors.IsNotFound(err) {
//			log.Error(err, "Error looking for healthcheck", "name", healthcheckSpec.Name)
//			return nil, err
//		}
//
//		log.V(2).Info("Creating a healthcheck", "name", healthcheckSpec.Name)
//		if err := s.healthchecks.Insert(ctx, meta.GlobalKey(healthcheckSpec.Name), healthcheckSpec); err != nil {
//			log.Error(err, "Error creating a healthcheck", "name", healthcheckSpec.Name)
//			return nil, err
//		}
//
//		healthcheck, err = s.healthchecks.Get(ctx, meta.GlobalKey(healthcheckSpec.Name))
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	s.scope.Network().APIServerHealthCheck = pointer.String(healthcheck.SelfLink)
//	return healthcheck, nil
//}
//
//func (s *Service) createOrGetBackendService(ctx context.Context, instancegroups []*compute.InstanceGroup, healthcheck *compute.HealthCheck) (*compute.BackendService, error) {
//	log := log.FromContext(ctx)
//	backends := make([]*compute.Backend, 0, len(instancegroups))
//	for _, group := range instancegroups {
//		backends = append(backends, &compute.Backend{
//			BalancingMode: "UTILIZATION",
//			Group:         group.SelfLink,
//		})
//	}
//
//	backendsvcSpec := s.scope.BackendServiceSpec()
//	backendsvcSpec.Backends = backends
//	backendsvcSpec.HealthChecks = []string{healthcheck.SelfLink}
//	backendsvc, err := s.backendservices.Get(ctx, meta.GlobalKey(backendsvcSpec.Name))
//	if err != nil {
//		if !gcperrors.IsNotFound(err) {
//			log.Error(err, "Error looking for backendservice", "name", backendsvcSpec.Name)
//			return nil, err
//		}
//
//		log.V(2).Info("Creating a backendservice", "name", backendsvcSpec.Name)
//		if err := s.backendservices.Insert(ctx, meta.GlobalKey(backendsvcSpec.Name), backendsvcSpec); err != nil {
//			log.Error(err, "Error creating a backendservice", "name", backendsvcSpec.Name)
//			return nil, err
//		}
//
//		backendsvc, err = s.backendservices.Get(ctx, meta.GlobalKey(backendsvcSpec.Name))
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	if len(backendsvc.Backends) != len(backendsvcSpec.Backends) {
//		log.V(2).Info("Updating a backendservice", "name", backendsvcSpec.Name)
//		backendsvc.Backends = backendsvcSpec.Backends
//		if err := s.backendservices.Update(ctx, meta.GlobalKey(backendsvcSpec.Name), backendsvc); err != nil {
//			log.Error(err, "Error updating a backendservice", "name", backendsvcSpec.Name)
//			return nil, err
//		}
//	}
//
//	s.scope.Network().APIServerBackendService = pointer.String(backendsvc.SelfLink)
//	return backendsvc, nil
//}
//
//func (s *Service) createOrGetTargetTCPProxy(ctx context.Context, service *compute.BackendService) (*compute.TargetTcpProxy, error) {
//	log := log.FromContext(ctx)
//	targetSpec := s.scope.TargetTCPProxySpec()
//	targetSpec.Service = service.SelfLink
//	target, err := s.targettcpproxies.Get(ctx, meta.GlobalKey(targetSpec.Name))
//	if err != nil {
//		if !gcperrors.IsNotFound(err) {
//			log.Error(err, "Error looking for targettcpproxy", "name", targetSpec.Name)
//			return nil, err
//		}
//
//		log.V(2).Info("Creating a targettcpproxy", "name", targetSpec.Name)
//		if err := s.targettcpproxies.Insert(ctx, meta.GlobalKey(targetSpec.Name), targetSpec); err != nil {
//			log.Error(err, "Error creating a targettcpproxy", "name", targetSpec.Name)
//			return nil, err
//		}
//
//		target, err = s.targettcpproxies.Get(ctx, meta.GlobalKey(targetSpec.Name))
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	s.scope.Network().APIServerTargetProxy = pointer.String(target.SelfLink)
//	return target, nil
//}
//
//func (s *Service) createOrGetAddress(ctx context.Context) (*compute.Address, error) {
//	log := log.FromContext(ctx)
//	addrSpec := s.scope.AddressSpec()
//	log.V(2).Info("Looking for address", "name", addrSpec.Name)
//	addr, err := s.addresses.Get(ctx, meta.GlobalKey(addrSpec.Name))
//	if err != nil {
//		if !gcperrors.IsNotFound(err) {
//			log.Error(err, "Error looking for address", "name", addrSpec.Name)
//			return nil, err
//		}
//
//		log.V(2).Info("Creating an address", "name", addrSpec.Name)
//		if err := s.addresses.Insert(ctx, meta.GlobalKey(addrSpec.Name), addrSpec); err != nil {
//			log.Error(err, "Error creating an address", "name", addrSpec.Name)
//			return nil, err
//		}
//
//		addr, err = s.addresses.Get(ctx, meta.GlobalKey(addrSpec.Name))
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	s.scope.Network().APIServerAddress = pointer.String(addr.SelfLink)
//	endpoint := s.scope.ControlPlaneEndpoint()
//	endpoint.Host = addr.Address
//	s.scope.SetControlPlaneEndpoint(endpoint)
//	return addr, nil
//}
//
//func (s *Service) createForwardingRule(ctx context.Context, target *compute.TargetTcpProxy, addr *compute.Address) error {
//	log := log.FromContext(ctx)
//	spec := s.scope.ForwardingRuleSpec()
//	key := meta.GlobalKey(spec.Name)
//	spec.IPAddress = addr.SelfLink
//	spec.Target = target.SelfLink
//	log.V(2).Info("Looking for forwardingrule", "name", spec.Name)
//	forwarding, err := s.forwardingrules.Get(ctx, key)
//	if err != nil {
//		if !gcperrors.IsNotFound(err) {
//			log.Error(err, "Error looking for forwardingrule", "name", spec.Name)
//			return err
//		}
//
//		log.V(2).Info("Creating a forwardingrule", "name", spec.Name)
//		if err := s.forwardingrules.Insert(ctx, key, spec); err != nil {
//			log.Error(err, "Error creating a forwardingrule", "name", spec.Name)
//			return err
//		}
//
//		forwarding, err = s.forwardingrules.Get(ctx, key)
//		if err != nil {
//			return err
//		}
//	}
//
//	s.scope.Network().APIServerForwardingRule = pointer.String(forwarding.SelfLink)
//	return nil
//}
//
//func (s *Service) deleteForwardingRule(ctx context.Context) error {
//	log := log.FromContext(ctx)
//	spec := s.scope.ForwardingRuleSpec()
//	key := meta.GlobalKey(spec.Name)
//	log.V(2).Info("Deleting a forwardingrule", "name", spec.Name)
//	if err := s.forwardingrules.Delete(ctx, key); err != nil && !gcperrors.IsNotFound(err) {
//		log.Error(err, "Error updating a forwardingrule", "name", spec.Name)
//		return err
//	}
//
//	s.scope.Network().APIServerForwardingRule = nil
//	return nil
//}
//
//func (s *Service) deleteAddress(ctx context.Context) error {
//	log := log.FromContext(ctx)
//	spec := s.scope.AddressSpec()
//	key := meta.GlobalKey(spec.Name)
//	log.V(2).Info("Deleting a address", "name", spec.Name)
//	if err := s.addresses.Delete(ctx, key); err != nil && !gcperrors.IsNotFound(err) {
//		return err
//	}
//
//	s.scope.Network().APIServerAddress = nil
//	return nil
//}
//
//func (s *Service) deleteTargetTCPProxy(ctx context.Context) error {
//	log := log.FromContext(ctx)
//	spec := s.scope.TargetTCPProxySpec()
//	key := meta.GlobalKey(spec.Name)
//	log.V(2).Info("Deleting a targettcpproxy", "name", spec.Name)
//	if err := s.targettcpproxies.Delete(ctx, key); err != nil && !gcperrors.IsNotFound(err) {
//		log.Error(err, "Error deleting a targettcpproxy", "name", spec.Name)
//		return err
//	}
//
//	s.scope.Network().APIServerTargetProxy = nil
//	return nil
//}
//
//func (s *Service) deleteBackendService(ctx context.Context) error {
//	log := log.FromContext(ctx)
//	spec := s.scope.BackendServiceSpec()
//	key := meta.GlobalKey(spec.Name)
//	log.V(2).Info("Deleting a backendservice", "name", spec.Name)
//	if err := s.backendservices.Delete(ctx, key); err != nil && !gcperrors.IsNotFound(err) {
//		log.Error(err, "Error deleting a backendservice", "name", spec.Name)
//		return err
//	}
//
//	s.scope.Network().APIServerBackendService = nil
//	return nil
//}
//
//func (s *Service) deleteHealthCheck(ctx context.Context) error {
//	log := log.FromContext(ctx)
//	spec := s.scope.HealthCheckSpec()
//	key := meta.GlobalKey(spec.Name)
//	log.V(2).Info("Deleting a healthcheck", "name", spec.Name)
//	if err := s.healthchecks.Delete(ctx, key); err != nil && !gcperrors.IsNotFound(err) {
//		log.Error(err, "Error deleting a healthcheck", "name", spec.Name)
//		return err
//	}
//
//	s.scope.Network().APIServerHealthCheck = nil
//	return nil
//}

func (s *Service) deleteManagedInstanceGroups(ctx context.Context) error {
	log := log.FromContext(ctx)
	//for zone := range s.scope.Network().WorkerInstanceGroups {

	//spec := s.scope.ManagedInstanceGroupSpec()
	//
	//key := meta.ZonalKey(spec.Name, spec.Zone[strings.LastIndex(spec.Zone, "/")+1:])
	//log.V(2).Info("Deleting a managedinstancegroup", "name", spec.Name)
	//if err := s.managedinstancegroups.Delete(ctx, key); err != nil {
	//	if !gcperrors.IsNotFound(err) {
	//		log.Error(err, "Error deleting a managedinstancegroup", "name", spec.Name)
	//		return err
	//	}
	//}

	spec := s.scope.RegionManagedInstanceGroupSpec()

	key := meta.RegionalKey(spec.Name, spec.Region)
	log.V(2).Info("Deleting a regionmanagedinstancegroup", "name", spec.Name)
	if err := s.regionmanagedinstancegroups.Delete(ctx, key); err != nil {
		if !gcperrors.IsNotFound(err) {
			log.Error(err, "Error deleting a regionmanagedinstancegroup", "name", spec.Name)
			return err
		}
	}

	//delete(s.scope.Network().WorkerInstanceGroups, zone)
	//}

	return nil
}

func (s *Service) createInstanceTemplate(ctx context.Context, instancetemplateSpec *compute.InstanceTemplate) (*compute.InstanceTemplate, error) {
	log := log.FromContext(ctx)
	log.V(2).Info("Creating instancetemplate", "name", instancetemplateSpec.Name)
	if err := s.instancetemplates.Insert(ctx, meta.GlobalKey(instancetemplateSpec.Name), instancetemplateSpec); err != nil {
		log.Error(err, "Error creating instancetemplate", "name", instancetemplateSpec.Name)
		return nil, err
	}

	instancetemplate, err := s.instancetemplates.Get(ctx, meta.GlobalKey(instancetemplateSpec.Name))
	if err != nil {
		return nil, err
	}

	return instancetemplate, nil
}

func (s *Service) createOrRecreateInstanceTemplates(ctx context.Context) (*compute.InstanceTemplate, error) {
	/// name should be gcpmachinepool_name + hash
	/// at the end(?) IT housekeeping is needed
	log := log.FromContext(ctx)

	log.V(2).Info("Getting bootstrap data for machinepool")
	bootstrapData, err := s.scope.GetBootstrapData()
	if err != nil {
		log.Error(err, "Error getting bootstrap data for machine")
		return nil, errors.Wrap(err, "failed to retrieve bootstrap data")
	}

	//if s.scope.GetInstanceTemplateName() == "" {
	//	s.scope.SetInstanceTemplateName(s.scope.GenerateName(bootstrapData, 0))
	//}

	var instancetemplate *compute.InstanceTemplate
	instancetemplateSpec := s.scope.InstanceTemplateSpec()

	//if instancetemplateSpec.Properties.Metadata == nil {
	//	instancetemplateSpec.Properties.Metadata = new(compute.Metadata)
	//}
	//instancetemplateSpec.Properties.Metadata.Items = append(instancetemplateSpec.Properties.Metadata.Items, &compute.MetadataItems{
	//	Key:   "user-data",
	//	Value: pointer.StringPtr(bootstrapData),
	//})

	// delete old
	oldInstanceTemplate := s.scope.GetOldInstanceTemplateName()
	if oldInstanceTemplate != "" {
		key := meta.GlobalKey(oldInstanceTemplate)
		log.V(2).Info("Deleting an instancetemplate", "name", oldInstanceTemplate)
		if err := s.instancetemplates.Delete(ctx, key); err != nil {
			if !gcperrors.IsNotFound(err) {
				log.Error(err, "Error deleting an instancetemplate", "name", oldInstanceTemplate)
				return nil, err
			}
			log.Info("Error deleting instancetemplate, not found", "name", oldInstanceTemplate)
		}
		s.scope.SetOldInstanceTemplateName("")
	}

	if instancetemplateSpec.Name == "" {
		instancetemplateSpec.Name = s.scope.GenerateName(bootstrapData, 0)
	} else {
		log.V(2).Info("Looking for instancetemplate", "name", instancetemplateSpec.Name)
		instancetemplate, err = s.instancetemplates.Get(ctx, meta.GlobalKey(instancetemplateSpec.Name))
		if err != nil {
			if !gcperrors.IsNotFound(err) {
				log.Error(err, "Error looking for instancetemplate")
				return nil, err
			}
		}
	}

	if instancetemplate == nil {
		instancetemplate, err = s.createInstanceTemplate(ctx, instancetemplateSpec)
		if err != nil {
			return nil, err
		}

		s.scope.SetInstanceTemplateName(instancetemplate.Name)
	}

	if s.isInstanceTemplateEqual(ctx, instancetemplateSpec, instancetemplate) {
		log.Info("it equal")
	} else {
		log.Info("it non equal")
		var newName string
		// recreate!!
		for i := 0; i < 10; i++ {
			newName = s.scope.GenerateName(bootstrapData, i)
			if newName != instancetemplateSpec.Name {
				// conflict
				break
			} else {
				log.Info("it name conflict, retrying")
			}
		}

		instancetemplateSpec.Name = newName

		if s.scope.GetOldInstanceTemplateName() != "" {
			log.Error(errors.New("Instance Template leak"), fmt.Sprintf("%s instance template should be deleted manually", s.scope.GetOldInstanceTemplateName()))
		}

		// create new
		instancetemplate, err = s.createInstanceTemplate(ctx, instancetemplateSpec)
		if err != nil {
			return nil, err
		}

		// set new
		s.scope.SetOldInstanceTemplateName(s.scope.GetFullInstanceTemplateName())
		s.scope.SetInstanceTemplateName(instancetemplateSpec.Name)
	}

	return instancetemplate, nil
}

func (s *Service) isInstanceTemplateEqual(ctx context.Context, spec *compute.InstanceTemplate, instancetemplate *compute.InstanceTemplate) bool {
	//log := log.FromContext(ctx)

	specBootstrapData := s.scope.GetBootstrapDataFromTemplate(spec)
	instancetemplateBootstrapData := s.scope.GetBootstrapDataFromTemplate(instancetemplate)

	//log.Info("BD diff", "specbd", specBootstrapData, "itbd", instancetemplateBootstrapData)

	return specBootstrapData == instancetemplateBootstrapData
}

func (s *Service) deleteInstanceTemplates(ctx context.Context) error {
	log := log.FromContext(ctx)
	instancetemplateSpec := s.scope.InstanceTemplateSpec()

	key := meta.GlobalKey(instancetemplateSpec.Name)
	log.V(2).Info("Deleting an instancetemplate", "name", instancetemplateSpec.Name)
	if err := s.instancetemplates.Delete(ctx, key); err != nil {
		if !gcperrors.IsNotFound(err) {
			log.Error(err, "Error deleting an instancetemplate", "name", instancetemplateSpec.Name)
			return err
		}
	}

	return nil
}
