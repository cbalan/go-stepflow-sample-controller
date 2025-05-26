package sample

import (
	"context"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/cbalan/go-stepflow"
)

func step1(ctx context.Context) error {
	log := logf.FromContext(ctx)
	log.Info("step1")

	return nil
}

func step2(ctx context.Context) error {
	log := logf.FromContext(ctx)
	log.Info("step2")

	return nil
}

func step3(ctx context.Context) error {
	log := logf.FromContext(ctx)
	log.Info("step3")

	return nil
}

func NewStepFlow() (stepflow.StepFlow, error) {
	return stepflow.NewStepFlow("sample/v1", stepflow.Steps().
		Do("step1", step1).
		Do("step2", step2).
		Do("step3", step3))
}
