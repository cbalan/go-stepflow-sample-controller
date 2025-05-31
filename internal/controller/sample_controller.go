/*
Copyright 2025.

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

package controller

import (
	"context"
	"k8s.io/utils/clock"
	"time"

	"github.com/cbalan/go-stepflow"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	samplev1alpha1 "github.com/cbalan/go-stepflow-sample-controller/api/v1alpha1"
	samplestepflow "github.com/cbalan/go-stepflow-sample-controller/internal/stepflow/sample"
)

// SampleReconciler reconciles a Sample object
type SampleReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Clock    clock.Clock
	stepflow stepflow.StepFlow
}

// +kubebuilder:rbac:groups=sample.stepflow,resources=samples,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sample.stepflow,resources=samples/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=sample.stepflow,resources=samples/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Sample object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *SampleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling sample")

	// Load sample resource.
	sample := &samplev1alpha1.Sample{}
	if err := r.Get(ctx, req.NamespacedName, sample); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// No action required for completed resources.
	if meta.IsStatusConditionTrue(sample.Status.Conditions, samplev1alpha1.CompletedCondition) {
		return ctrl.Result{}, nil
	}

	// Create sample exchange.
	ex := samplestepflow.NewExchange(r.Clock, sample)

	// Apply sample stepflow against the old state and exchange.
	var errApply error
	sample.Status.State, errApply = r.stepflow.Apply(samplestepflow.WithExchange(ctx, ex), sample.Status.State)

	// On error, complete the sample resource with error.
	if errApply != nil {
		meta.SetStatusCondition(&sample.Status.Conditions, metav1.Condition{
			Type:    samplev1alpha1.CompletedCondition,
			Status:  metav1.ConditionTrue,
			Reason:  samplev1alpha1.ErrorReason,
			Message: errApply.Error(),
		})

		log.Error(errApply, "Failed to apply sample stepflow")

		return ctrl.Result{}, r.Status().Update(ctx, sample)
	}

	// If completed, complete the sample resource with success.
	if r.stepflow.IsCompleted(sample.Status.State) {
		meta.SetStatusCondition(&sample.Status.Conditions, metav1.Condition{
			Type:    samplev1alpha1.CompletedCondition,
			Status:  metav1.ConditionTrue,
			Reason:  samplev1alpha1.SuccessReason,
			Message: "Sample stepflow completed successfully",
		})

		log.Info("Completed sample stepflow")

		return ctrl.Result{}, r.Status().Update(ctx, sample)
	}

	// Default reconcile result.
	reconcileResult := ctrl.Result{RequeueAfter: 500 * time.Millisecond}

	// Apply exchange requeue hint, if set.
	if ex.Result.RequeueAfter > 0 {
		reconcileResult.RequeueAfter = ex.Result.RequeueAfter
	}

	return reconcileResult, r.Status().Update(ctx, sample)
}

// SetupStepFlow sets up the controller sample stepflow.
func (r *SampleReconciler) SetupStepFlow() error {
	if r.stepflow != nil {
		return nil
	}

	flow, err := samplestepflow.NewStepFlow()
	if err != nil {
		return err
	}
	r.stepflow = flow

	return nil
}

func (r *SampleReconciler) SetupClock() error {
	if r.Clock != nil {
		return nil
	}

	r.Clock = clock.RealClock{}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SampleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.SetupClock(); err != nil {
		return err
	}

	if err := r.SetupStepFlow(); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&samplev1alpha1.Sample{}).
		Named("sample").
		Complete(r)
}
