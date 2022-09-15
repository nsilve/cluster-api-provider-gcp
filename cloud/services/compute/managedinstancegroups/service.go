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
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/filter"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"google.golang.org/api/compute/v1"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud"
)

type managedinstancegroupsInterface interface {
	Get(ctx context.Context, key *meta.Key) (*compute.InstanceGroupManager, error)
	List(ctx context.Context, zone string, fl *filter.F) ([]*compute.InstanceGroupManager, error)
	Insert(ctx context.Context, key *meta.Key, obj *compute.InstanceGroupManager) error
	Delete(ctx context.Context, key *meta.Key) error
}

// Scope is an interfaces that hold used methods.
type Scope interface {
	cloud.ClusterGetter
	//cloud.MachineTemplate
	ManagedInstanceGroupSpec() *compute.InstanceGroupManager
}

// Service implements instancegroups reconciler.
type Service struct {
	scope                 Scope
	managedinstancegroups managedinstancegroupsInterface
}

var _ cloud.Reconciler = &Service{}

// New returns Service from given scope.
func New(scope Scope) *Service {
	return &Service{
		scope:                 scope,
		managedinstancegroups: scope.Cloud().InstanceGroupManagers(),
	}
}
