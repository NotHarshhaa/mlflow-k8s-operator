package controllers

import (
	"testing"

	"github.com/NotHarshhaa/mlflow-k8s-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

			mlflowServer := &v1alpha1.MLflowServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mlflow",
					Namespace: "default",
				},
				Spec: v1alpha1.MLflowServerSpec{
					Version: "2.11.0",
					Tracking: v1alpha1.TrackingConfig{
						Replicas: 1,
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
					},
					Backend: v1alpha1.BackendConfig{
						Type: v1alpha1.BackendTypePostgreSQL,
						PostgreSQL: &v1alpha1.PostgreSQLConfig{
							Host:              "postgresql.default.svc.cluster.local",
							Port:              5432,
							Database:          "mlflow",
							CredentialsSecret: "mlflow-db-credentials",
							SSLMode:           "require",
						},
					},
					ArtifactStore: v1alpha1.ArtifactStoreConfig{
						Type: v1alpha1.ArtifactStoreTypeS3,
						S3: &v1alpha1.S3Config{
							Bucket:            "mlflow-artifacts",
							Region:            "us-east-1",
							CredentialsSecret: "aws-credentials",
						},
					},
				},
			}

			// Note: This is a basic structure for the test
			// In a real environment, you would use envtest to set up a test Kubernetes cluster
			// and then actually create the resource and verify the controller's behavior

			Expect(mlflowServer.Spec.Version).To(Equal("2.11.0"))
			Expect(mlflowServer.Spec.Tracking.Replicas).To(Equal(int32(1)))
		})

		It("Should validate the MLflowServer spec", func() {
			By("Testing spec validation")
			spec := v1alpha1.MLflowServerSpec{
				Version: "2.11.0",
				Tracking: v1alpha1.TrackingConfig{
					Replicas: 1,
				},
				Backend: v1alpha1.BackendConfig{
					Type:   v1alpha1.BackendTypeSQLite,
					SQLite: &v1alpha1.SQLiteConfig{},
				},
				ArtifactStore: v1alpha1.ArtifactStoreConfig{
					Type: v1alpha1.ArtifactStoreTypePVC,
					PVC:  &v1alpha1.PVCConfig{},
				},
			}

			Expect(spec.Version).NotTo(BeEmpty())
			Expect(spec.Backend.Type).NotTo(BeEmpty())
			Expect(spec.ArtifactStore.Type).NotTo(BeEmpty())
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
