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
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	infrav1 "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud/services/compute/instancetemplates"
	infrav1exp "sigs.k8s.io/cluster-api-provider-gcp/exp/api/v1beta1"
	infrautil "sigs.k8s.io/cluster-api-provider-gcp/util"
	"sigs.k8s.io/cluster-api-provider-gcp/util/reconciler"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/predicates"
	"sigs.k8s.io/cluster-api/util/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GCPMachineTemplateReconciler reconciles a GCPMachineTemplate object
type GCPMachineTemplateReconciler struct {
	client.Client
	ReconcileTimeout time.Duration
	Scheme           *runtime.Scheme
	WatchFilterValue string
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpmachinetemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpmachinetemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gcpmachinetemplates/finalizers,verbs=update

// SetupWithManager sets up the controller with the Manager.
func (r *GCPMachineTemplateReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	log := ctrl.LoggerFrom(ctx)

	if err := mgr.GetFieldIndexer().IndexField(ctx, &infrav1exp.GCPMachinePool{}, ".spec.machineTemplate.infrastructureRef.name", func(rawObj client.Object) []string {
		GCPMachinePool := rawObj.(*infrav1exp.GCPMachinePool)
		infraRef := GCPMachinePool.Spec.MachineTemplate.InfrastructureRef
		if infraRef == nil {
			return nil
		}

		if infraRef.APIVersion != infrav1.GroupVersion.String() || infraRef.Kind != "GCPMachineTemplate" {
			return nil
		}

		return []string{infraRef.Name}
	}); err != nil {
		return err
	}

	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.GCPMachineTemplate{}).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(log, r.WatchFilterValue)).
		Watches(
			&source.Kind{Type: &infrav1exp.GCPMachinePool{}},
			handler.EnqueueRequestsFromMapFunc(r.GCPMachinePoolToGCPMachineTemplateMapFunc(ctx, log)),
		).
		Build(r)
	if err != nil {
		return errors.Wrap(err, "error creating controller")
	}

	// Add a watch on clusterv1.Cluster object for unpause & ready notifications.
	if err := c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(r.ClusterToGCPMachineTemplates(ctx)),
		predicates.ClusterUnpausedAndInfrastructureReady(log),
		predicates.ResourceNotPausedAndHasFilterLabel(log, r.WatchFilterValue),
	); err != nil {
		return errors.Wrap(err, "failed adding a watch for ready clusters")
	}

	return nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GCPMachineTemplate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *GCPMachineTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx, cancel := context.WithTimeout(ctx, reconciler.DefaultedLoopTimeout(r.ReconcileTimeout))
	defer cancel()

	log := ctrl.LoggerFrom(ctx)
	gcpMachineTemplate := &infrav1.GCPMachineTemplate{}
	err := r.Get(ctx, req.NamespacedName, gcpMachineTemplate)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Fetch the Cluster.

	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, gcpMachineTemplate.ObjectMeta)
	if err != nil {
		log.Info("MachineTemplate is missing cluster label or cluster does not exist")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	//cluster, err := util.GetClusterByName(ctx, r.Client, gcpMachineTemplate.Namespace, gcpMachineTemplate.Spec.ClusterName)
	//if err != nil {
	//	log.Info("MachineTemplate is missing cluster label or cluster does not exist")
	//	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	//}

	if annotations.IsPaused(cluster, gcpMachineTemplate) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)
	gcpCluster := &infrav1.GCPCluster{}
	gcpClusterKey := client.ObjectKey{
		Namespace: gcpMachineTemplate.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, gcpClusterKey, gcpCluster); err != nil {
		log.Info("GCPCluster is not available yet")
		return ctrl.Result{}, nil
	}

	//Create the cluster scope
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:     r.Client,
		Cluster:    cluster,
		GCPCluster: gcpCluster,
	})
	if err != nil {
		return ctrl.Result{}, err
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
	}()

	// Handle deleted machine pools
	if !gcpMachineTemplate.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machineTemplateScope, clusterScope)
	}

	// Handle non-deleted machine pools
	ret, err := r.reconcileNormal(ctx, machineTemplateScope, clusterScope) /// create only if it is needed!!
	if err != nil {
		return ret, err
	}

	//labels := map[string]string{clusterv1.ClusterLabelName: cluster.Name}
	GCPMachinePoolList := &infrav1exp.GCPMachinePoolList{}
	// add extra opts for same cluster and infraRef...!!!
	if err := r.List(ctx, GCPMachinePoolList,
		client.InNamespace(gcpMachineTemplate.Namespace),
		client.MatchingLabels(map[string]string{clusterv1.ClusterLabelName: cluster.Name}),
		client.MatchingFields{".spec.machineTemplate.infrastructureRef.apiVersion": gcpMachineTemplate.APIVersion},
		client.MatchingFields{".spec.machineTemplate.infrastructureRef.kind": "GCPMachineTemplate"},
		client.MatchingFields{".spec.machineTemplate.infrastructureRef.name": gcpMachineTemplate.Name},
	); err != nil {
		log.Error(err, "failed to list GCPMachinePoolList")
		return ctrl.Result{}, err
	}

	//references := gcpMachineTemplate.Status.References.GCPMachinePools
	//if references == nil {
	references := make(map[string]string)
	//}
	for _, mp := range GCPMachinePoolList.Items {
		// add GCP url from status when added!!!
		references[mp.Name] = mp.Name
	}
	gcpMachineTemplate.Status.References.GCPMachinePools = references

	log.Info("GCPMachinePoolList", "list", GCPMachinePoolList)

	return ctrl.Result{}, err
}

func (r *GCPMachineTemplateReconciler) reconcileNormal(ctx context.Context, machineTemplateScope *scope.MachineTemplateScope, clusterScope *scope.ClusterScope) (_ reconcile.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)
	log = log.WithValues("cluster", clusterScope.Name())
	log.Info("Reconciling GCPMachineTemplate")

	// If the AzureMachine is in an error state, return early.
	//if machineTemplateScope.GCPMachineTemplate.Status.FailureReason != nil || machineTemplateScope.AzureMachinePool.Status.FailureMessage != nil {
	//	log.Info("Error state detected, skipping reconciliation")
	//	return reconcile.Result{}, nil
	//}

	// If the GCPMachineTemplate doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(machineTemplateScope.GCPMachineTemplate, infrav1.MachineTemplateFinalizer)

	machineTemplateScope.GCPMachineTemplate.OwnerReferences = util.EnsureOwnerRef(
		machineTemplateScope.GCPMachineTemplate.OwnerReferences,
		v1.OwnerReference{
			APIVersion: clusterv1.GroupVersion.String(),
			Kind:       "Cluster",
			Name:       clusterScope.Name(),
			UID:        clusterScope.Cluster.UID,
			//Controller:         nil,
			//BlockOwnerDeletion: nil,
		})

	// Register the finalizer immediately to avoid orphaning GCP resources on delete
	if err := machineTemplateScope.PatchObject(); err != nil {
		return reconcile.Result{}, err
	}

	//if !clusterScope.Cluster.Status.InfrastructureReady {
	//	log.Info("Cluster infrastructure is not ready yet")
	//	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	//}

	// Make sure bootstrap data is available and populated.
	//if machineTemplateScope.MachinePool.Spec.Template.Spec.Bootstrap.DataSecretName == nil {
	//	log.Info("Bootstrap data secret reference is not yet available")
	//	return reconcile.Result{}, nil
	//}

	if err := instancetemplates.New(machineTemplateScope).Reconcile(ctx); err != nil {
		log.Error(err, "Error reconciling instancetemplates resources")
		record.Warnf(machineTemplateScope.GCPMachineTemplate, "GCPMachineTemplateReconcile", "Reconcile error - %v", err)
		return ctrl.Result{}, err
	}

	machineTemplateScope.GCPMachineTemplate.Status.Ready = true

	//ams, err := r.createGCPMachineTemplateService(machineTemplateScope)
	//if err != nil {
	//	return reconcile.Result{}, errors.Wrap(err, "failed creating a newAzureMachinePoolService")
	//}
	//
	//if err := ams.Reconcile(ctx); err != nil {
	//	// Handle transient and terminal errors
	//	var reconcileError azure.ReconcileError
	//	if errors.As(err, &reconcileError) {
	//		if reconcileError.IsTerminal() {
	//			log.Error(err, "failed to reconcile AzureMachinePool", "name", machineTemplateScope.Name())
	//			return reconcile.Result{}, nil
	//		}
	//
	//		if reconcileError.IsTransient() {
	//			log.Error(err, "failed to reconcile AzureMachinePool", "name", machineTemplateScope.Name())
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
	//	machineTemplateScope.ProviderID(), "state", machineTemplateScope.ProvisioningState())
	//
	//switch machineTemplateScope.ProvisioningState() {
	//case infrav1.Deleting:
	//	log.Info("Unexpected scale set deletion", "id", machineTemplateScope.ProviderID())
	//	ampr.Recorder.Eventf(machineTemplateScope.AzureMachinePool, corev1.EventTypeWarning, "UnexpectedVMDeletion", "Unexpected Azure scale set deletion")
	//case infrav1.Failed:
	//	err := ams.Delete(ctx)
	//	if err != nil {
	//		return reconcile.Result{}, errors.Wrap(err, "failed to delete scale set in a failed state")
	//	}
	//	return reconcile.Result{}, errors.Wrap(err, "Scale set deleted, retry creating in next reconcile")
	//}
	//
	//if machineTemplateScope.NeedsRequeue() {
	//	return reconcile.Result{
	//		RequeueAfter: 30 * time.Second,
	//	}, nil
	//}

	return reconcile.Result{}, nil
}

func (ampr *GCPMachineTemplateReconciler) reconcileDelete(ctx context.Context, machineTemplateScope *scope.MachineTemplateScope, clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log = log.WithValues("cluster", clusterScope.Name())
	log.Info("Reconciling delete GCPMachineTemplate")

	if err := instancetemplates.New(machineTemplateScope).Delete(ctx); err != nil {
		log.Error(err, "Error reconciling delete instancetemplate resources")
		record.Warnf(machineTemplateScope.GCPMachineTemplate, "GCPMachineTemplateReconcile", "Reconcile delete error - %v", err)
		return ctrl.Result{}, err
	}

	////clusterScope.SetReady()

	//ctx, log, done := tele.StartSpanWithLogger(ctx, "controllers.AzureMachinePoolReconciler.reconcileDelete")
	//defer done()
	//
	//log.V(2).Info("handling deleted AzureMachinePool")
	//
	//if infracontroller.ShouldDeleteIndividualResources(ctx, clusterScope) {
	//	amps, err := ampr.createAzureMachinePoolService(machineTemplateScope)
	//	if err != nil {
	//		return reconcile.Result{}, errors.Wrap(err, "failed creating a new AzureMachinePoolService")
	//	}
	//
	//	log.V(4).Info("deleting AzureMachinePool resource individually")
	//	if err := amps.Delete(ctx); err != nil {
	//		return reconcile.Result{}, errors.Wrapf(err, "error deleting AzureMachinePool %s/%s", clusterScope.Namespace(), machineTemplateScope.Name())
	//	}
	//}
	//
	//// Delete succeeded, remove finalizer
	//log.V(4).Info("removing finalizer for AzureMachinePool")

	log.Info("Removing finalizer for GCPMachineTemplate")
	controllerutil.RemoveFinalizer(machineTemplateScope.GCPMachineTemplate, infrav1.MachineTemplateFinalizer)

	return reconcile.Result{}, nil
}

// ClusterToGCPMachineTemplates is a handler.ToRequestsFunc to be used to enqeue requests for reconciliation
// of ClusterToGCPMachineTemplates.
func (r *GCPMachineTemplateReconciler) ClusterToGCPMachineTemplates(ctx context.Context) handler.MapFunc {
	log := ctrl.LoggerFrom(ctx)
	log.Info(">>> Reconcile GCPMachineTemplate due to Cluster")
	return func(o client.Object) []ctrl.Request {
		result := []ctrl.Request{}

		c, ok := o.(*clusterv1.Cluster)
		if !ok {
			log.Error(errors.Errorf("expected a Cluster but got a %T", o), "failed to get GCPMachineTemplate for Cluster")
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
		machineTemplateList := &infrav1.GCPMachineTemplateList{}
		if err := r.List(ctx, machineTemplateList, client.InNamespace(c.Namespace), client.MatchingLabels(labels)); err != nil {
			log.Error(err, "failed to list MachineTemplates")
			return nil
		}
		for _, m := range machineTemplateList.Items {
			//if m.Spec.Template.Spec.Template.ObjectMeta.Spec.InfrastructureRef.Name == "" {
			if m.Spec.Template.Spec.InstanceType == "" { //// to fix
				log.Info("ClusterToGCPMachineTemplates empty spec.Template.Spec.machineTemplate.infrastructureRef.Name")
				continue
			}
			name := client.ObjectKey{Namespace: m.Namespace, Name: m.Name}
			result = append(result, ctrl.Request{NamespacedName: name})
		}

		return result
	}
}

func (r GCPMachineTemplateReconciler) GCPMachinePoolToGCPMachineTemplateMapFunc(ctx context.Context, log logr.Logger) handler.MapFunc {
	return func(o client.Object) []ctrl.Request {
		result := []ctrl.Request{}
		c, ok := o.(*infrav1exp.GCPMachinePool)
		if !ok {
			log.Error(errors.Errorf("expected a GCPMachinePool but got a %T", o), "failed to get GCPMachinePool for GCPMachineTemplate")
			return nil
		}

		gcpMachineTemplate, err := infrautil.GetGCPMachineTemplateFromGCPMachinePool(ctx, r.Client, *c)
		switch {
		case apierrors.IsNotFound(err) || gcpMachineTemplate == nil:
			return result
		case err != nil:
			log.Error(err, "failed to get ref GCPMachineTemplate")
			return result
		}

		return []ctrl.Request{
			{
				NamespacedName: client.ObjectKey{
					Namespace: gcpMachineTemplate.Namespace,
					Name:      gcpMachineTemplate.Name,
				},
			},
		}
	}
}
