package sample

import (
	"context"
	"fmt"
	"k8s.io/utils/clock"
	"time"

	"github.com/cbalan/go-stepflow"
	"github.com/cbalan/go-stepflow-sample-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Exchange struct {
	Clock  clock.Clock
	Sample *v1alpha1.Sample
	Result ctrl.Result
}

type contextKey string

const exContextKey = contextKey("ex")

func WithExchange(ctx context.Context, ex *Exchange) context.Context {
	return context.WithValue(ctx, exContextKey, ex)
}

func FromContext(ctx context.Context) (*Exchange, error) {
	ex, ok := ctx.Value(exContextKey).(*Exchange)
	if !ok {
		return nil, fmt.Errorf("failed to get sample Exchange from context")
	}

	return ex, nil
}

func NewExchange(clock clock.Clock, sample *v1alpha1.Sample) *Exchange {
	return &Exchange{Clock: clock, Sample: sample}
}

func step1(ctx context.Context) error {
	log := logf.FromContext(ctx)
	log.Info("step1")

	return nil
}

func someTimeToPass(ctx context.Context) (bool, error) {
	log := logf.FromContext(ctx)
	log.Info("someTimeToPass")

	// Get exchange from context.
	ex, err := FromContext(ctx)
	if err != nil {
		return false, err
	}

	// Done if more than 30 seconds went by after creation time.
	timeoutTime := metav1.NewTime(ex.Clock.Now().Add(-1 * 30 * time.Second))
	creationTime := ex.Sample.ObjectMeta.CreationTimestamp
	isDone := !timeoutTime.Before(&creationTime)

	// Requeue hint.
	ex.Result.RequeueAfter = time.Second

	return isDone, nil
}

func step3(ctx context.Context) error {
	log := logf.FromContext(ctx)
	log.Info("step3")

	ex, err := FromContext(ctx)
	if err != nil {
		return err
	}

	if ex.Sample.Spec.Foo == "error" {
		return fmt.Errorf("error via sample.spec.foo")
	}

	return nil
}

func NewStepFlow() (stepflow.StepFlow, error) {
	return stepflow.NewStepFlow("sample/v1", stepflow.Steps().
		Do("step1", step1).
		WaitFor("someTimeToPass", someTimeToPass).
		Do("step3", step3))
}
