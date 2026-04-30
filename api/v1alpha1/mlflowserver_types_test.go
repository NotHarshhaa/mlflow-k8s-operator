package v1alpha1

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestMLflowServerTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MLflowServer Types Suite")
}

var _ = Describe("MLflowServer", func() {
	Context("When creating a new MLflowServer", func() {
		It("Should have the correct default values", func() {
			By("Creating a minimal MLflowServer")
			mlflowServer := &MLflowServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mlflow",
					Namespace: "default",
				},
				Spec: MLflowServerSpec{
					Replicas: ptr.To[int32](1),
					Image: ImageSpec{
						Repository: "mlflow/mlflow",
						Tag:        "2.7.1",
					},
				},
			}

			Expect(mlflowServer.Name).To(Equal("test-mlflow"))
			Expect(mlflowServer.Namespace).To(Equal("default"))
			Expect(mlflowServer.Spec.Replicas).To(Equal(ptr.To[int32](1)))
			Expect(mlflowServer.Spec.Image.Repository).To(Equal("mlflow/mlflow"))
		})

		It("Should accept a full specification", func() {
			By("Creating a fully specified MLflowServer")
			mlflowServer := &MLflowServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "full-mlflow",
					Namespace: "mlflow-system",
				},
				Spec: MLflowServerSpec{
					Replicas: ptr.To[int32](3),
					Image: ImageSpec{
						Repository: "custom/mlflow",
						Tag:        "v1.0.0",
						PullPolicy: corev1.PullIfNotPresent,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1000m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
					BackendStore: BackendStoreSpec{
						Type:     "postgresql",
						Host:     "postgres.example.com",
						Port:     5432,
						Database: "mlflowdb",
						User:     "mlflowuser",
						PasswordSecret: &SecretRef{
							Name: "db-secret",
							Key:  "password",
						},
					},
					DefaultArtifactRoot: "s3://my-bucket/mlflow",
					Service: ServiceSpec{
						Type: corev1.ServiceTypeLoadBalancer,
						Port: 5000,
					},
					Ingress: &IngressSpec{
						Enabled: true,
						Host:    "mlflow.example.com",
						TLS:     true,
					},
				},
			}

			Expect(mlflowServer.Spec.Replicas).To(Equal(ptr.To[int32](3)))
			Expect(mlflowServer.Spec.Image.Repository).To(Equal("custom/mlflow"))
			Expect(mlflowServer.Spec.BackendStore.Type).To(Equal("postgresql"))
			Expect(mlflowServer.Spec.BackendStore.Host).To(Equal("postgres.example.com"))
			Expect(mlflowServer.Spec.DefaultArtifactRoot).To(Equal("s3://my-bucket/mlflow"))
			Expect(mlflowServer.Spec.Ingress.Enabled).To(BeTrue())
		})
	})

	Context("When defining backend store configurations", func() {
		It("Should support PostgreSQL backend", func() {
			backend := BackendStoreSpec{
				Type:     "postgresql",
				Host:     "localhost",
				Port:     5432,
				Database: "mlflow",
				User:     "mlflow",
			}

			Expect(backend.Type).To(Equal("postgresql"))
			Expect(backend.Port).To(Equal(5432))
		})

		It("Should support MySQL backend", func() {
			backend := BackendStoreSpec{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "mlflow",
				User:     "mlflow",
			}

			Expect(backend.Type).To(Equal("mysql"))
			Expect(backend.Port).To(Equal(3306))
		})

		It("Should support SQLite backend", func() {
			backend := BackendStoreSpec{
				Type: "sqlite",
				Path: "/mlflow/mlflow.db",
			}

			Expect(backend.Type).To(Equal("sqlite"))
			Expect(backend.Path).To(Equal("/mlflow/mlflow.db"))
		})
	})

	Context("When defining service configurations", func() {
		It("Should support ClusterIP service", func() {
			service := ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Port: 5000,
			}

			Expect(service.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(service.Port).To(Equal(5000))
		})

		It("Should support LoadBalancer service", func() {
			service := ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Port: 5000,
			}

			Expect(service.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
		})

		It("Should support NodePort service", func() {
			service := ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Port:     5000,
				NodePort: ptr.To[int32](30000),
			}

			Expect(service.Type).To(Equal(corev1.ServiceTypeNodePort))
			Expect(service.NodePort).To(Equal(ptr.To[int32](30000)))
		})
	})

	Context("When defining ingress configurations", func() {
		It("Should support disabled ingress", func() {
			ingress := &IngressSpec{
				Enabled: false,
			}

			Expect(ingress.Enabled).To(BeFalse())
		})

		It("Should support enabled ingress with TLS", func() {
			ingress := &IngressSpec{
				Enabled: true,
				Host:    "mlflow.example.com",
				TLS:     true,
				TLSSecret: &SecretRef{
					Name: "mlflow-tls",
					Key:  "tls.crt",
				},
			}

			Expect(ingress.Enabled).To(BeTrue())
			Expect(ingress.Host).To(Equal("mlflow.example.com"))
			Expect(ingress.TLS).To(BeTrue())
		})
	})
})
