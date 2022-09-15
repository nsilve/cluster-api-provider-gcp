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

package instancetemplates

import (
	"context"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"google.golang.org/api/compute/v1"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud/gcperrors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconcile reconcile cluster instancetemplate components.
func (s *Service) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling instancetemplate resources")

	//instancetemplates, err := s.createOrGetInstanceTemplates(ctx)
	_, err := s.createOrGetInstanceTemplates(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Delete delete cluster control-plane instancetemplate compoenents.
func (s *Service) Delete(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Deleting instancetemplate resources")

	return s.deleteInstanceTemplates(ctx)
}

func (s *Service) createOrGetInstanceTemplates(ctx context.Context) (*compute.InstanceTemplate, error) {
	log := log.FromContext(ctx)
	instancetemplateSpec := s.scope.InstanceTemplateSpec()
	log.V(2).Info("Looking for instancetemplate", "name", instancetemplateSpec.Name)
	instancetemplate, err := s.instancetemplates.Get(ctx, meta.GlobalKey(instancetemplateSpec.Name))
	if err != nil {
		if !gcperrors.IsNotFound(err) {
			log.Error(err, "Error looking for instancetemplate")
			return nil, err
		}

		log.V(2).Info("Creating instancetemplate", "name", instancetemplateSpec.Name)
		if err := s.instancetemplates.Insert(ctx, meta.GlobalKey(instancetemplateSpec.Name), instancetemplateSpec); err != nil {
			log.Error(err, "Error creating instancetemplate", "name", instancetemplateSpec.Name)
			return nil, err
		}

		instancetemplate, err = s.instancetemplates.Get(ctx, meta.GlobalKey(instancetemplateSpec.Name))
		if err != nil {
			return nil, err
		}
	}
	return instancetemplate, nil
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

//func (s *Service) getInstanceTemplateFromInstanceGroupManager(ctx context.Context, instanceGroupManager *compute.InstanceGroupManager) (*compute.InstanceTemplate, error) {
//	log := log.FromContext(ctx)
//
//	name := instanceGroupManager.InstanceGroup
//	log.Info("getInstanceTemplateFromInstanceGroupManager", "name", name)
//
//	managedinstancegroups.New().
//
//}
