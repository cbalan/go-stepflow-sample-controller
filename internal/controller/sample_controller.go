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

	"github.com/cbalan/go-stepflow"
	samplestepflow "github.com/cbalan/go-stepflow-sample-controller/internal/stepflow/sample"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	samplev1alpha1 "github.com/cbalan/go-stepflow-sample-controller/api/v1alpha1"
)

// SampleReconciler reconciles a Sample object
type SampleReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
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

	sample := &samplev1alpha1.Sample{}
	if err := r.Get(ctx, req.NamespacedName, sample); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	oldState := sample.Status.State

	// stepflow is complete. Job done.
	if r.stepflow.IsCompleted(oldState) {
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling Sample", "oldState", oldState)

	newState, err := r.stepflow.Apply(ctx, oldState)
	if err != nil {
		return ctrl.Result{}, err
	}

	sample.Status.State = newState
	if err := r.Status().Update(ctx, sample); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Applied sample stepflow", "oldState", oldState, "newState", newState)

	return ctrl.Result{}, nil
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

// SetupWithManager sets up the controller with the Manager.
func (r *SampleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.SetupStepFlow(); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&samplev1alpha1.Sample{}).
		Named("sample").
		Complete(r)
}
