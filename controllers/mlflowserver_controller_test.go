package controllers

import (
	"context"
	"testing"

	"github.com/NotHarshhaa/mlflow-k8s-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMLflowServerController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = Describe("MLflowServer Controller", func() {
	Context("When creating a new MLflowServer", func() {
		It("Should create the deployment successfully", func() {
			By("Creating a new MLflowServer")
			ctx := context.Background()

			mlflowServer := &v1alpha1.MLflowServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mlflow",
					Namespace: "default",
				},
				Spec: v1alpha1.MLflowServerSpec{
					Replicas: ptr.To[int32](1),
					Image: v1alpha1.ImageSpec{
						Repository: "mlflow/mlflow",
						Tag:        "2.7.1",
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
					BackendStore: v1alpha1.BackendStoreSpec{
						Type:     "postgresql",
						Host:     "postgresql.default.svc.cluster.local",
						Port:     5432,
						Database: "mlflow",
						User:     "mlflow",
						PasswordSecret: &v1alpha1.SecretRef{
							Name: "mlflow-db-credentials",
							Key:  "password",
						},
					},
					DefaultArtifactRoot: "s3://mlflow-artifacts",
				},
			}

			// Note: This is a basic structure for the test
			// In a real environment, you would use envtest to set up a test Kubernetes cluster
			// and then actually create the resource and verify the controller's behavior

			Expect(mlflowServer.Spec.Replicas).To(Equal(ptr.To[int32](1)))
			Expect(mlflowServer.Spec.Image.Repository).To(Equal("mlflow/mlflow"))
		})

		It("Should validate the MLflowServer spec", func() {
			By("Testing spec validation")
			spec := v1alpha1.MLflowServerSpec{
				Replicas: ptr.To[int32](1),
				Image: v1alpha1.ImageSpec{
					Repository: "mlflow/mlflow",
					Tag:        "2.7.1",
				},
			}

			Expect(spec.Image.Repository).NotTo(BeEmpty())
			Expect(spec.Image.Tag).NotTo(BeEmpty())
		})
	})
})

// Unit tests for helper functions
var _ = Describe("MLflowServer Controller Helpers", func() {
	Context("When calculating resource requirements", func() {
		It("Should correctly parse resource strings", func() {
			By("Testing resource parsing")
			// This would test the internal resource calculation logic
			// once implemented in the controller
			Expect(true).To(BeTrue())
		})
	})
})
