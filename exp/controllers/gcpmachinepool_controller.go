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

package controllers

import (
	"context"
	"fmt"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	infrav1 "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud/services/compute/managedinstancegroups"
	infrav1exp "sigs.k8s.io/cluster-api-provider-gcp/exp/api/v1beta1"
	"sigs.k8s.io/cluster-api-provider-gcp/util/reconciler"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterv1exp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/predicates"
	"sigs.k8s.io/cluster-api/util/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"
)

// GCPMachinePoolReconciler reconciles a GCPMachinePool object
type GCPMachinePoolReconciler struct {
	client.Client
	ReconcileTimeout time.Duration
	//Scheme           *runtime.Scheme
	WatchFilterValue string
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpmachinepools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpmachinepools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpmachinepools/finalizers,verbs=update
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinepools;machinepools/status,verbs=get;list;watch

// SetupWithManager sets up the controller with the Manager.
func (r *GCPMachinePoolReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	log := ctrl.LoggerFrom(ctx)

	//var rc reconcile.Reconciler = r
	//if options.Cache != nil {
	//	r = coalescing.NewReconciler(ampr, options.Cache, log)
	//}

	c, err := ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1exp.GCPMachinePool{}).
		//WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue)).
		//Watches(
		//	&source.Kind{Type: &clusterv1exp.MachinePool{}},
		//	handler.EnqueueRequestsFromMapFunc(r.machinePoolToInfrastructureMapFunc(infrav1exp.GroupVersion.WithKind("GCPMachinePool"), log)),
		//).
		//Watches(
		//	&source.Kind{Type: &infrav1.GCPCluster{}},
		//	handler.EnqueueRequestsFromMapFunc(r.GCPClusterToGCPMachinePools(ctx)),
		//).
		//Watches(
		//	&source.Kind{Type: &infrav1.GCPMachineTemplate{}},
		//	handler.EnqueueRequestsFromMapFunc(r.GCPMachineTemplateToGCPMachinePools(ctx)),
		//	//builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		//).
		Build(r)
	if err != nil {
		return errors.Wrap(err, "error creating controller")
	}

	// Add a watch on clusterv1.Cluster object for unpause & ready notifications.
	if err := c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(r.ClusterToGCPMachinePools(ctx)),
		predicates.ClusterUnpausedAndInfrastructureReady(log),
	); err != nil {
		return errors.Wrap(err, "failed adding a watch for ready clusters")
	}

	//if err := c.Watch(
	//	&source.Kind{Type: &infrav1.GCPMachineTemplate{}},
	//	handler.EnqueueRequestsFromMapFunc(r.GCPMachineTemplateToGCPMachinePools(ctx, log)),
	//); err != nil {
	//	return errors.Wrap(err, "failed adding a watch for GCPMachineTemplate")
	//}

	return nil

}

func getOwnerMachinePool(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*clusterv1exp.MachinePool, error) {
	for _, ref := range obj.OwnerReferences {
		if ref.Kind != "MachinePool" {
			continue
		}
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if gv.Group == clusterv1exp.GroupVersion.Group {
			return getMachinePoolByName(ctx, c, obj.Namespace, ref.Name)
		}
	}
	return nil, nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GCPMachinePool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *GCPMachinePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx, cancel := context.WithTimeout(ctx, reconciler.DefaultedLoopTimeout(r.ReconcileTimeout))
	defer cancel()

	log := ctrl.LoggerFrom(ctx)
	gcpMachinePool := &infrav1exp.GCPMachinePool{}
	err := r.Get(ctx, req.NamespacedName, gcpMachinePool)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Fetch the CAPI MachinePool
	machinePool, err := getOwnerMachinePool(ctx, r.Client, gcpMachinePool.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if machinePool == nil {
		log.Info("MachinePool Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}
	log = log.WithValues("machinePool", machinePool.Name)

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machinePool.ObjectMeta)
	if err != nil {
		log.Info("MachinePool is missing cluster label or cluster does not exist")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, gcpMachinePool) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)
	gcpCluster := &infrav1.GCPCluster{}
	gcpClusterKey := client.ObjectKey{
		Namespace: gcpMachinePool.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, gcpClusterKey, gcpCluster); err != nil {
		log.Info("GCPCluster is not available yet")
		return ctrl.Result{}, nil
	}

	// Create the cluster scope
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:     r.Client,
		Cluster:    cluster,
		GCPCluster: gcpCluster,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	if gcpMachinePool.Spec.MachineTemplate.InfrastructureRef == nil {
		log.Info(fmt.Sprintf("GCPMachinePool [%s] has empty .spec.machineTemplate.infrastructureRef", gcpMachinePool.Name))
		return ctrl.Result{}, nil
	}

	if gcpMachinePool.Spec.MachineTemplate.InfrastructureRef.Kind != "GCPMachineTemplate" {
		log.Info(fmt.Sprintf("GCPMachinePool [%s] has a .spec.machineTemplate.infrastructureRef of type [%s] and not the expected [GCPMachineTemplate]", gcpMachinePool.Name, gcpMachinePool.Spec.MachineTemplate.InfrastructureRef.GroupVersionKind()))
		return ctrl.Result{}, nil
	}

	gcpMachineTemplateNamespacedName := client.ObjectKey{
		//Namespace: gcpMachinePool.Spec.InfrastructureRef.Namespace,
		Namespace: gcpMachinePool.Namespace, // add default value via webhook???
		Name:      gcpMachinePool.Spec.MachineTemplate.InfrastructureRef.Name,
	}

	gcpMachineTemplate := &infrav1.GCPMachineTemplate{}
	err = r.Get(ctx, gcpMachineTemplateNamespacedName, gcpMachineTemplate)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Create the machine pool scope
	machinePoolScope, err := scope.NewMachinePoolScope(scope.MachinePoolScopeParams{
		Client:             r.Client,
		ClusterGetter:      clusterScope,
		MachinePool:        machinePool,
		GCPMachinePool:     gcpMachinePool,
		GCPMachineTemplate: gcpMachineTemplate,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Create the machine template scope
	machineTemplateScope, err := scope.NewMachineTemplateScope(scope.MachineTemplateScopeParams{
		Client:             r.Client,
		ClusterGetter:      clusterScope,
		GCPMachineTemplate: gcpMachineTemplate,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any GCPMachine changes.
	defer func() {
		if err := machineTemplateScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
		if err := machinePoolScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
		if err := clusterScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted machine pools
	if !gcpMachinePool.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machinePoolScope, machineTemplateScope, clusterScope)
	}

	// Handle non-deleted machine pools
	return r.reconcileNormal(ctx, machinePoolScope, machineTemplateScope, clusterScope)
}

// getMachinePoolByName finds and return a Machine object using the specified params.
func getMachinePoolByName(ctx context.Context, c client.Client, namespace, name string) (*clusterv1exp.MachinePool, error) {
	m := &clusterv1exp.MachinePool{}
	key := client.ObjectKey{Name: name, Namespace: namespace}
	if err := c.Get(ctx, key, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *GCPMachinePoolReconciler) reconcileNormal(ctx context.Context, machinePoolScope *scope.MachinePoolScope, machineTemplateScope *scope.MachineTemplateScope, clusterScope *scope.ClusterScope) (_ reconcile.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)
	log = log.WithValues("cluster", clusterScope.Name())
	log.Info("Reconciling GCPMachinePool")

	// If the AzureMachine is in an error state, return early.
	//if machinePoolScope.GCPMachinePool.Status.FailureReason != nil || machinePoolScope.AzureMachinePool.Status.FailureMessage != nil {
	//	log.Info("Error state detected, skipping reconciliation")
	//	return reconcile.Result{}, nil
	//}

	// If the GCPMachinePool doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(machinePoolScope.GCPMachinePool, clusterv1exp.MachinePoolFinalizer)

	// Register the finalizer immediately to avoid orphaning GCP resources on delete
	if err := machinePoolScope.PatchObject(); err != nil {
		return reconcile.Result{}, err
	}

	if !clusterScope.Cluster.Status.InfrastructureReady {
		log.Info("Cluster infrastructure is not ready yet")
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Make sure bootstrap data is available and populated.
	//if machinePoolScope.MachinePool.Spec.Template.Spec.Bootstrap.DataSecretName == nil {
	//	log.Info("Bootstrap data secret reference is not yet available")
	//	return reconcile.Result{}, nil
	//}

	//if err := instancetemplates.New(machineTemplateScope).Reconcile(ctx); err != nil {
	//	log.Error(err, "Error reconciling instancetemplate resources")
	//	record.Warnf(machinePoolScope.GCPMachinePool.Spec.InfrastructureRef, "GCPMachineTemplateReconcile", "Reconcile error - %v", err)
	//	return ctrl.Result{}, err
	//}

	if err := managedinstancegroups.New(machinePoolScope).Reconcile(ctx); err != nil {
		log.Error(err, "Error reconciling managedinstancegroup resources")
		record.Warnf(machinePoolScope.GCPMachinePool, "GCPMachinePoolReconcile", "Reconcile error - %v", err)
		return ctrl.Result{}, err
	}

	//ams, err := r.createGCPMachinePoolService(machinePoolScope)
	//if err != nil {
	//	return reconcile.Result{}, errors.Wrap(err, "failed creating a newAzureMachinePoolService")
	//}
	//
	//if err := ams.Reconcile(ctx); err != nil {
	//	// Handle transient and terminal errors
	//	var reconcileError azure.ReconcileError
	//	if errors.As(err, &reconcileError) {
	//		if reconcileError.IsTerminal() {
	//			log.Error(err, "failed to reconcile AzureMachinePool", "name", machinePoolScope.Name())
	//			return reconcile.Result{}, nil
	//		}
	//
	//		if reconcileError.IsTransient() {
	//			log.Error(err, "failed to reconcile AzureMachinePool", "name", machinePoolScope.Name())
	//			return reconcile.Result{RequeueAfter: reconcileError.RequeueAfter()}, nil
	//		}
	//
	//		return reconcile.Result{}, errors.Wrap(err, "failed to reconcile AzureMachinePool")
	//	}
	//
	//	return reconcile.Result{}, err
	//}
	//
	//log.V(2).Info("Scale Set reconciled", "id",
	//	machinePoolScope.ProviderID(), "state", machinePoolScope.ProvisioningState())
	//
	//switch machinePoolScope.ProvisioningState() {
	//case infrav1.Deleting:
	//	log.Info("Unexpected scale set deletion", "id", machinePoolScope.ProviderID())
	//	ampr.Recorder.Eventf(machinePoolScope.AzureMachinePool, corev1.EventTypeWarning, "UnexpectedVMDeletion", "Unexpected Azure scale set deletion")
	//case infrav1.Failed:
	//	err := ams.Delete(ctx)
	//	if err != nil {
	//		return reconcile.Result{}, errors.Wrap(err, "failed to delete scale set in a failed state")
	//	}
	//	return reconcile.Result{}, errors.Wrap(err, "Scale set deleted, retry creating in next reconcile")
	//}
	//
	//if machinePoolScope.NeedsRequeue() {
	//	return reconcile.Result{
	//		RequeueAfter: 30 * time.Second,
	//	}, nil
	//}

	// instancegroupSpec := s.scope.ManagedInstanceGroupSpec(zone)

	zone := machinePoolScope.ManagedInstanceGroupSpec().Zone[strings.LastIndex(machinePoolScope.ManagedInstanceGroupSpec().Zone, "/")+1:]

	mig, err := managedinstancegroups.New(machinePoolScope).GetManagedInstanceGroups(ctx, meta.ZonalKey(machinePoolScope.ManagedInstanceGroupSpec().Name, zone))
	if err != nil {
		return ctrl.Result{}, err
	}

	//machinePoolScope.GCPMachinePool.Spec.ProviderID = mig.SelfLink
	//fill ProviderIDList!!!

	machinePoolScope.GCPMachinePool.Status.Ready = mig.Status.IsStable
	//machinePoolScope.GCPMachinePool.Status.Replicas = int32(mig.TargetSize)
	machinePoolScope.GCPMachinePool.Status.Replicas = 0

	if !mig.Status.IsStable {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

func (r *GCPMachinePoolReconciler) reconcileDelete(ctx context.Context, machinePoolScope *scope.MachinePoolScope, machineTemplateScope *scope.MachineTemplateScope, clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log = log.WithValues("cluster", clusterScope.Name())
	log.Info("Reconciling delete GCPMachinePool")

	if err := managedinstancegroups.New(machinePoolScope).Delete(ctx); err != nil {
		log.Error(err, "Error reconciling delete managedinstancegroup resources")
		record.Warnf(machinePoolScope.GCPMachinePool, "GCPMachinePoolReconcile", "Reconcile delete error - %v", err)
		return ctrl.Result{}, err
	}

	//if err := instancetemplates.New(machineTemplateScope).Delete(ctx); err != nil {
	//	log.Error(err, "Error reconciling delete instancetemplate resources")
	//	record.Warnf(machineTemplateScope.GCPMachineTemplate, "GCPMachineTemplateReconcile", "Reconcile delete error - %v", err)
	//	return ctrl.Result{}, err
	//}

	////clusterScope.SetReady()

	//ctx, log, done := tele.StartSpanWithLogger(ctx, "controllers.AzureMachinePoolReconciler.reconcileDelete")
	//defer done()
	//
	//log.V(2).Info("handling deleted AzureMachinePool")
	//
	//if infracontroller.ShouldDeleteIndividualResources(ctx, clusterScope) {
	//	amps, err := ampr.createAzureMachinePoolService(machinePoolScope)
	//	if err != nil {
	//		return reconcile.Result{}, errors.Wrap(err, "failed creating a new AzureMachinePoolService")
	//	}
	//
	//	log.V(4).Info("deleting AzureMachinePool resource individually")
	//	if err := amps.Delete(ctx); err != nil {
	//		return reconcile.Result{}, errors.Wrapf(err, "error deleting AzureMachinePool %s/%s", clusterScope.Namespace(), machinePoolScope.Name())
	//	}
	//}
	//
	//// Delete succeeded, remove finalizer
	//log.V(4).Info("removing finalizer for AzureMachinePool")

	log.Info("Removing finalizer for GCPMachinePool")
	controllerutil.RemoveFinalizer(machinePoolScope.GCPMachinePool, clusterv1exp.MachinePoolFinalizer)

	return reconcile.Result{}, nil
}

// machinePoolToInfrastructureMapFunc returns a handler.MapFunc that watches for
// MachinePool events and returns reconciliation requests for an infrastructure provider object.
func (r *GCPMachinePoolReconciler) machinePoolToInfrastructureMapFunc(gvk schema.GroupVersionKind, log logr.Logger) handler.MapFunc {
	return func(o client.Object) []reconcile.Request {
		m, ok := o.(*clusterv1exp.MachinePool)
		if !ok {
			log.V(4).Info("attempt to map incorrect type", "type", fmt.Sprintf("%T", o))
			return nil
		}

		gk := gvk.GroupKind()
		ref := m.Spec.Template.Spec.InfrastructureRef
		// Return early if the GroupKind doesn't match what we expect.
		infraGK := ref.GroupVersionKind().GroupKind()
		if gk != infraGK {
			log.V(4).Info("gk does not match", "gk", gk, "infraGK", infraGK)
			return nil
		}

		log.Info(fmt.Sprintf("machinePoolToInfrastructureMapFunc reconcile %s.%s", ref.Name, m.Namespace))

		return []reconcile.Request{
			{
				NamespacedName: client.ObjectKey{
					Namespace: m.Namespace,
					Name:      ref.Name,
				},
			},
		}
	}
}

// GCPClusterToGCPMachinePools is a handler.ToRequestsFunc to be used to enqeue requests for reconciliation
// of GCPMachinePools.
func (r *GCPMachinePoolReconciler) GCPClusterToGCPMachinePools(ctx context.Context, log logr.Logger) handler.MapFunc {
	//log := ctrl.LoggerFrom(ctx)
	log.Info(">>> Reconcile GCPMachinePool due to GCPCluster")
	return func(o client.Object) []ctrl.Request {
		result := []ctrl.Request{}

		c, ok := o.(*infrav1.GCPCluster)
		if !ok {
			log.Error(errors.Errorf("expected a GCPCluster but got a %T", o), "failed to get GCPMachinePool for GCPCluster")
			return nil
		}

		cluster, err := util.GetOwnerCluster(ctx, r.Client, c.ObjectMeta)
		switch {
		case apierrors.IsNotFound(err) || cluster == nil:
			return result
		case err != nil:
			log.Error(err, "failed to get owning cluster")
			return result
		}

		labels := map[string]string{clusterv1.ClusterLabelName: cluster.Name}
		machinePoolList := &clusterv1exp.MachinePoolList{}
		if err := r.List(ctx, machinePoolList, client.InNamespace(c.Namespace), client.MatchingLabels(labels)); err != nil {
			log.Error(err, "failed to list MachinePools")
			return nil
		}
		for _, m := range machinePoolList.Items {
			if m.Spec.Template.Spec.InfrastructureRef.Name == "" {
				log.Info("GCPClusterToGCPMachinePools empty spec.Template.Spec.InfrastructureRef.Name")
				continue
			}
			name := client.ObjectKey{Namespace: m.Namespace, Name: m.Name}
			result = append(result, ctrl.Request{NamespacedName: name})
		}

		return result
	}
}

// GCPMachineTemplateToGCPMachinePools is a handler.ToRequestsFunc to be used to enqeue requests for reconciliation
// of GCPMachinePools.
func (r *GCPMachinePoolReconciler) GCPMachineTemplateToGCPMachinePools(ctx context.Context, log logr.Logger) handler.MapFunc {
	return func(o client.Object) []ctrl.Request {
		result := []ctrl.Request{}
		c, ok := o.(*infrav1.GCPMachineTemplate)
		if !ok {
			log.Error(errors.Errorf("expected a GCPMachineTemplate but got a %T", o), "failed to get GCPMachinePool for GCPMachineTemplate")
			return nil
		}

		cluster, err := util.GetOwnerCluster(ctx, r.Client, c.ObjectMeta)
		switch {
		case apierrors.IsNotFound(err) || cluster == nil:
			return result
		case err != nil:
			log.Error(err, "failed to get owning cluster")
			return result
		}

		labels := map[string]string{clusterv1.ClusterLabelName: cluster.Name}
		machinePoolList := &clusterv1exp.MachinePoolList{}
		if err := r.List(ctx, machinePoolList, client.InNamespace(c.Namespace), client.MatchingLabels(labels)); err != nil {
			log.Error(err, "failed to list MachinePools")
			return nil
		}
		for _, m := range machinePoolList.Items {
			if m.Spec.Template.Spec.InfrastructureRef.Name == "" {
				log.Info("GCPMachineTemplateToGCPMachinePools empty spec.Template.Spec.InfrastructureRef.Name")
				continue
			}
			name := client.ObjectKey{Namespace: m.Namespace, Name: m.Name}
			result = append(result, ctrl.Request{NamespacedName: name})
		}

		return result
	}
}

// ClusterToGCPMachinePools is a handler.ToRequestsFunc to be used to enqeue requests for reconciliation
// of GCPMachinePools.
func (r *GCPMachinePoolReconciler) ClusterToGCPMachinePools(ctx context.Context) handler.MapFunc {
	log := ctrl.LoggerFrom(ctx)

	return func(o client.Object) []ctrl.Request {
		result := []ctrl.Request{}

		c, ok := o.(*clusterv1.Cluster)
		if !ok {
			log.Error(errors.Errorf("expected a Cluster but got a %T", o), "failed to get GCPMachinePool for Cluster")
			return nil
		}

		cluster, err := util.GetOwnerCluster(ctx, r.Client, c.ObjectMeta)
		switch {
		case apierrors.IsNotFound(err) || cluster == nil:
			return result
		case err != nil:
			log.Error(err, "failed to get owning cluster")
			return result
		}

		labels := map[string]string{clusterv1.ClusterLabelName: cluster.Name}
		machinePoolList := &clusterv1exp.MachinePoolList{}
		if err := r.List(ctx, machinePoolList, client.InNamespace(c.Namespace), client.MatchingLabels(labels)); err != nil {
			log.Error(err, "failed to list MachinePools")
			return nil
		}
		for _, m := range machinePoolList.Items {
			if m.Spec.Template.Spec.InfrastructureRef.Name == "" {
				log.Info("ClusterToGCPMachinePools empty spec.Template.Spec.InfrastructureRef.Name")
				continue
			}
			name := client.ObjectKey{Namespace: m.Namespace, Name: m.Name}
			result = append(result, ctrl.Request{NamespacedName: name})
		}

		return result
	}
}

//
//func (r *GCPMachinePoolReconciler) getInstanceTemplateFromInstanceGroupManager(ctx context.Context, instanceGroupManager *compute.InstanceGroupManager) (*compute.InstanceTemplate, error) {
//	log := log.FromContext(ctx)
//
//	name := instanceGroupManager.InstanceGroup
//	log.Info("getInstanceTemplateFromInstanceGroupManager", "name", name)
//
//	//managedinstancegroups.New(r.
//
//}
