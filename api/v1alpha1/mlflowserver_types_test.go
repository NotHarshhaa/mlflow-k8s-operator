package v1alpha1

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
					Version: "2.11.0",
					Tracking: TrackingConfig{
						Replicas: 1,
					},
					Backend: BackendConfig{
						Type:   BackendTypeSQLite,
						SQLite: &SQLiteConfig{},
					},
					ArtifactStore: ArtifactStoreConfig{
						Type: ArtifactStoreTypePVC,
						PVC:  &PVCConfig{},
					},
				},
			}

			Expect(mlflowServer.Name).To(Equal("test-mlflow"))
			Expect(mlflowServer.Namespace).To(Equal("default"))
			Expect(mlflowServer.Spec.Version).To(Equal("2.11.0"))
			Expect(mlflowServer.Spec.Tracking.Replicas).To(Equal(int32(1)))
		})

		It("Should accept a full specification", func() {
			By("Creating a fully specified MLflowServer")
			mlflowServer := &MLflowServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "full-mlflow",
					Namespace: "mlflow-system",
				},
				Spec: MLflowServerSpec{
					Version: "2.11.0",
					Tracking: TrackingConfig{
						Replicas: 3,
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
					},
					Backend: BackendConfig{
						Type: BackendTypePostgreSQL,
						PostgreSQL: &PostgreSQLConfig{
							Host:              "postgres.example.com",
							Port:              5432,
							Database:          "mlflowdb",
							CredentialsSecret: "db-secret",
							SSLMode:           "require",
						},
					},
					ArtifactStore: ArtifactStoreConfig{
						Type: ArtifactStoreTypeS3,
						S3: &S3Config{
							Bucket:            "my-bucket",
							Region:            "us-east-1",
							CredentialsSecret: "aws-credentials",
							PathPrefix:        "mlflow",
						},
					},
					Ingress: IngressConfig{
						Enabled:          true,
						Host:             "mlflow.example.com",
						IngressClassName: "nginx",
						TLS: &TLSConfig{
							Enabled: true,
							Issuer:  "letsencrypt-prod",
						},
					},
				},
			}

			Expect(mlflowServer.Spec.Version).To(Equal("2.11.0"))
			Expect(mlflowServer.Spec.Tracking.Replicas).To(Equal(int32(3)))
			Expect(mlflowServer.Spec.Backend.Type).To(Equal(BackendTypePostgreSQL))
			Expect(mlflowServer.Spec.Backend.PostgreSQL.Host).To(Equal("postgres.example.com"))
			Expect(mlflowServer.Spec.ArtifactStore.Type).To(Equal(ArtifactStoreTypeS3))
			Expect(mlflowServer.Spec.ArtifactStore.S3.Bucket).To(Equal("my-bucket"))
			Expect(mlflowServer.Spec.Ingress.Enabled).To(BeTrue())
		})
	})

	Context("When defining backend configurations", func() {
		It("Should support PostgreSQL backend", func() {
			backend := BackendConfig{
				Type: BackendTypePostgreSQL,
				PostgreSQL: &PostgreSQLConfig{
					Host:              "localhost",
					Port:              5432,
					Database:          "mlflow",
					CredentialsSecret: "db-secret",
				},
			}

			Expect(backend.Type).To(Equal(BackendTypePostgreSQL))
			Expect(backend.PostgreSQL.Port).To(Equal(int32(5432)))
		})

		It("Should support MySQL backend", func() {
			backend := BackendConfig{
				Type: BackendTypeMySQL,
				MySQL: &MySQLConfig{
					Host:              "localhost",
					Port:              3306,
					Database:          "mlflow",
					CredentialsSecret: "db-secret",
				},
			}

			Expect(backend.Type).To(Equal(BackendTypeMySQL))
			Expect(backend.MySQL.Port).To(Equal(int32(3306)))
		})

		It("Should support SQLite backend", func() {
			backend := BackendConfig{
				Type: BackendTypeSQLite,
				SQLite: &SQLiteConfig{
					PVC: &PVCConfig{
						Size: "10Gi",
					},
				},
			}

			Expect(backend.Type).To(Equal(BackendTypeSQLite))
			Expect(backend.SQLite.PVC.Size).To(Equal("10Gi"))
		})
	})

	Context("When defining artifact store configurations", func() {
		It("Should support S3 artifact store", func() {
			store := ArtifactStoreConfig{
				Type: ArtifactStoreTypeS3,
				S3: &S3Config{
					Bucket:            "my-bucket",
					Region:            "us-east-1",
					CredentialsSecret: "aws-secret",
				},
			}

			Expect(store.Type).To(Equal(ArtifactStoreTypeS3))
			Expect(store.S3.Bucket).To(Equal("my-bucket"))
		})

		It("Should support GCS artifact store", func() {
			store := ArtifactStoreConfig{
				Type: ArtifactStoreTypeGCS,
				GCS: &GCSConfig{
					Bucket:            "my-gcs-bucket",
					CredentialsSecret: "gcp-secret",
				},
			}

			Expect(store.Type).To(Equal(ArtifactStoreTypeGCS))
			Expect(store.GCS.Bucket).To(Equal("my-gcs-bucket"))
		})

		It("Should support PVC artifact store", func() {
			store := ArtifactStoreConfig{
				Type: ArtifactStoreTypePVC,
				PVC: &PVCConfig{
					StorageClass: "standard",
					Size:         "100Gi",
				},
			}

			Expect(store.Type).To(Equal(ArtifactStoreTypePVC))
			Expect(store.PVC.Size).To(Equal("100Gi"))
		})
	})

	Context("When defining ingress configurations", func() {
		It("Should support disabled ingress", func() {
			ingress := IngressConfig{
				Enabled: false,
			}

			Expect(ingress.Enabled).To(BeFalse())
		})

		It("Should support enabled ingress with TLS", func() {
			ingress := IngressConfig{
				Enabled:          true,
				Host:             "mlflow.example.com",
				IngressClassName: "nginx",
				TLS: &TLSConfig{
					Enabled:    true,
					Issuer:     "letsencrypt-prod",
					SecretName: "mlflow-tls",
				},
			}

			Expect(ingress.Enabled).To(BeTrue())
			Expect(ingress.Host).To(Equal("mlflow.example.com"))
			Expect(ingress.TLS.Enabled).To(BeTrue())
		})
	})
})
