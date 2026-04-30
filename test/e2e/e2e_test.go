package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	mlflowv1alpha1 "github.com/your-org/mlflow-k8s-operator/api/v1alpha1"
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.Background())

	By("bootstrapping test environment")
	useKind := os.Getenv("USE_KIND") == "true"

	if useKind {
		By("using Kind cluster for e2e tests")
		testEnv = &envtest.Environment{
			UseExistingCluster: ptr.To(true),
			Config:             config.GetConfigOrDie(),
		}
	} else {
		By("using envtest for e2e tests")
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
			ErrorIfCRDPathMissing: true,
		}
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = mlflowv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("MLflowServer E2E Tests", func() {
	Context("When deploying a basic MLflowServer", func() {
		It("Should successfully deploy and become ready", func() {
			By("Creating a new MLflowServer")
			mlflowServer := &mlflowv1alpha1.MLflowServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "e2e-test-mlflow",
					Namespace: "default",
				},
				Spec: mlflowv1alpha1.MLflowServerSpec{
					Replicas: ptr.To[int32](1),
					Image: mlflowv1alpha1.ImageSpec{
						Repository: "mlflow/mlflow",
						Tag:        "2.7.1",
					},
					BackendStore: mlflowv1alpha1.BackendStoreSpec{
						Type: "sqlite",
						Path: "/mlflow/mlflow.db",
					},
				},
			}

			err := k8sClient.Create(ctx, mlflowServer)
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for the deployment to be created")
			deployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "e2e-test-mlflow",
					Namespace: "default",
				}, deployment)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Waiting for the deployment to become ready")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "e2e-test-mlflow",
					Namespace: "default",
				}, deployment)
				if err != nil {
					return false
				}
				return deployment.Status.ReadyReplicas == *mlflowServer.Spec.Replicas
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up the MLflowServer")
			err = k8sClient.Delete(ctx, mlflowServer)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When deploying MLflowServer with PostgreSQL backend", func() {
		It("Should successfully deploy with external database", func() {
			By("Creating a secret for database credentials")
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mlflow-db-credentials",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"password": []byte("testpassword"),
				},
			}

			err := k8sClient.Create(ctx, secret)
			Expect(err).NotTo(HaveOccurred())

			By("Creating MLflowServer with PostgreSQL backend")
			mlflowServer := &mlflowv1alpha1.MLflowServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "e2e-mlflow-postgres",
					Namespace: "default",
				},
				Spec: mlflowv1alpha1.MLflowServerSpec{
					Replicas: ptr.To[int32](1),
					Image: mlflowv1alpha1.ImageSpec{
						Repository: "mlflow/mlflow",
						Tag:        "2.7.1",
					},
					BackendStore: mlflowv1alpha1.BackendStoreSpec{
						Type:     "postgresql",
						Host:     "postgres.default.svc.cluster.local",
						Port:     5432,
						Database: "mlflow",
						User:     "mlflow",
						PasswordSecret: &mlflowv1alpha1.SecretRef{
							Name: "mlflow-db-credentials",
							Key:  "password",
						},
					},
				},
			}

			err = k8sClient.Create(ctx, mlflowServer)
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the MLflowServer is created")
			createdMLflow := &mlflowv1alpha1.MLflowServer{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "e2e-mlflow-postgres",
					Namespace: "default",
				}, createdMLflow)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdMLflow.Spec.BackendStore.Type).To(Equal("postgresql"))

			By("Cleaning up")
			err = k8sClient.Delete(ctx, mlflowServer)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Delete(ctx, secret)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When deploying MLflowServer with custom resources", func() {
		It("Should apply custom resource limits", func() {
			By("Creating MLflowServer with custom resources")
			mlflowServer := &mlflowv1alpha1.MLflowServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "e2e-mlflow-resources",
					Namespace: "default",
				},
				Spec: mlflowv1alpha1.MLflowServerSpec{
					Replicas: ptr.To[int32](2),
					Image: mlflowv1alpha1.ImageSpec{
						Repository: "mlflow/mlflow",
						Tag:        "2.7.1",
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
					BackendStore: mlflowv1alpha1.BackendStoreSpec{
						Type: "sqlite",
						Path: "/mlflow/mlflow.db",
					},
				},
			}

			err := k8sClient.Create(ctx, mlflowServer)
			Expect(err).NotTo(HaveOccurred())

			By("Verifying deployment has correct resource requirements")
			deployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "e2e-mlflow-resources",
					Namespace: "default",
				}, deployment)
				if err != nil {
					return false
				}
				return len(deployment.Spec.Template.Spec.Containers) > 0
			}, timeout, interval).Should(BeTrue())

			container := deployment.Spec.Template.Spec.Containers[0]
			Expect(container.Resources.Requests.Cpu().String()).To(Equal("200m"))
			Expect(container.Resources.Requests.Memory().String()).To(Equal("512Mi"))

			By("Cleaning up")
			err = k8sClient.Delete(ctx, mlflowServer)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

const (
	timeout  = 5 * time.Minute
	interval = 10 * time.Second
)
