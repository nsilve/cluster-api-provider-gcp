/*
Copyright 2020 The Kubernetes Authors.

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

// Package util implements utilities.
package util

import (
	"context"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	infrav1 "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	infrav1exp "sigs.k8s.io/cluster-api-provider-gcp/exp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetGCPMachineTemplateFromGCPMachinePool finds and return a GCPMachineTemplate object from a GCPMachinePool.
func GetGCPMachineTemplateFromGCPMachinePool(ctx context.Context, c client.Client, gcpMachinePool infrav1exp.GCPMachinePool) (*infrav1.GCPMachineTemplate, error) {
	gcpMachineTemplate := &infrav1.GCPMachineTemplate{}

	infraRef := gcpMachinePool.Spec.MachineTemplate.InfrastructureRef

	if infraRef.Kind != "GCPMachineTemplate" {
		return nil, nil
	}
	gv, err := schema.ParseGroupVersion(infraRef.APIVersion)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if gv.Group == infrav1.GroupVersion.Group {
		namespace := gcpMachinePool.Namespace
		if infraRef.Namespace != "" {
			namespace = infraRef.Namespace
		}
		key := client.ObjectKey{
			Namespace: namespace,
			Name:      infraRef.Name,
		}
		if err := c.Get(ctx, key, gcpMachineTemplate); err != nil {
			return nil, errors.Wrapf(err, "failed to get GCPMachineTemplate/%s", key.Name)
		}

		return gcpMachineTemplate, nil
	}

	return nil, nil
}
