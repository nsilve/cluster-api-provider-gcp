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
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var gcpmachinepoollog = logf.Log.WithName("gcpmachinepool-resource")

func (r *GCPMachinePool) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:verbs=create;update,path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-gcpmachinepool,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=gcpmachinepools,versions=v1beta1,name=default.gcpmachinepool.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1beta1
//+kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-gcpmachinepool,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=gcpmachinepools,versions=v1beta1,name=validation.gcpmachinepool.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &GCPMachinePool{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *GCPMachinePool) Default() {
	gcpmachinepoollog.Info("default", "name", r.Name)
}

var _ webhook.Validator = &GCPMachinePool{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *GCPMachinePool) ValidateCreate() error {
	gcpmachinepoollog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *GCPMachinePool) ValidateUpdate(oldRaw runtime.Object) error {
	old, ok := oldRaw.(*GCPMachinePool)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected an GCPMachinePool but got a %T", oldRaw))
	}

	if !reflect.DeepEqual(r.Spec, old.Spec) {
		return apierrors.NewBadRequest("GCPMachinePool.Spec is immutable")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *GCPMachinePool) ValidateDelete() error {
	gcpmachinepoollog.Info("validate delete", "name", r.Name)
	return nil
}
