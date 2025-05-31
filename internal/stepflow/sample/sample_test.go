package sample

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	samplev1alpha1 "github.com/cbalan/go-stepflow-sample-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clocktesting "k8s.io/utils/clock/testing"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = Describe("Sample stepflow", func() {
	Describe("step1", func() {
		It("returns nil", func() {
			clock := clocktesting.NewFakeClock(time.Now())
			sample := &samplev1alpha1.Sample{}

			ex := NewExchange(clock, sample)

			err := step1(WithExchange(context.TODO(), ex))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("someTimeToPass", func() {
		Context("Not enough time went by", func() {
			It("returns false", func() {
				clock := clocktesting.NewFakeClock(time.Now())
				sample := &samplev1alpha1.Sample{}
				sample.CreationTimestamp = metav1.NewTime(clock.Now())

				ex := NewExchange(clock, sample)

				actual, err := someTimeToPass(WithExchange(context.TODO(), ex))
				Expect(err).NotTo(HaveOccurred())

				Expect(actual).To(BeFalse())
			})
		})

		Context("Enough time went by", func() {
			It("returns true", func() {
				clock := clocktesting.NewFakeClock(time.Now())
				sample := &samplev1alpha1.Sample{}
				sample.CreationTimestamp = metav1.NewTime(clock.Now())

				// Advance clock.
				clock.Step(30 * time.Second)

				ex := NewExchange(clock, sample)

				actual, err := someTimeToPass(WithExchange(context.TODO(), ex))
				Expect(err).NotTo(HaveOccurred())

				Expect(actual).To(BeTrue())
			})
		})
	})

	Describe("step3", func() {
		Context("sample.spec.foo has a non error value", func() {
			It("returns nil", func() {
				clock := clocktesting.NewFakeClock(time.Now())
				sample := &samplev1alpha1.Sample{}
				sample.Spec.Foo = "non error"

				ex := NewExchange(clock, sample)

				err := step3(WithExchange(context.TODO(), ex))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("Error via sample.spec.foo", func() {
			It("returns error", func() {
				clock := clocktesting.NewFakeClock(time.Now())
				sample := &samplev1alpha1.Sample{}
				sample.Spec.Foo = "error"

				ex := NewExchange(clock, sample)

				err := step3(WithExchange(context.TODO(), ex))
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(Equal("error via sample.spec.foo"))
			})
		})
	})

	Describe("NewStepFlow", func() {
		It("returns a new StepFlow", func() {
			stepFlow, err := NewStepFlow()
			Expect(err).ToNot(HaveOccurred())
			Expect(stepFlow).ToNot(BeNil())
		})
	})
})
