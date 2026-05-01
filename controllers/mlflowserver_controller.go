package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mlopsv1alpha1 "github.com/NotHarshhaa/mlflow-k8s-operator/api/v1alpha1"
)

// MLflowServerReconciler reconciles a MLflowServer object
type MLflowServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mlops.NotHarshhaa.io,resources=mlflowservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mlops.NotHarshhaa.io,resources=mlflowservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mlops.NotHarshhaa.io,resources=mlflowservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the main reconciliation loop
func (r *MLflowServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	mlflowServer := &mlopsv1alpha1.MLflowServer{}
	if err := r.Get(ctx, req.NamespacedName, mlflowServer); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("MLflowServer resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get MLflowServer")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !mlflowServer.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, mlflowServer)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(mlflowServer, mlopsv1alpha1.MLflowServerFinalizer) {
		controllerutil.AddFinalizer(mlflowServer, mlopsv1alpha1.MLflowServerFinalizer)
		if err := r.Update(ctx, mlflowServer); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Update observed generation
	mlflowServer.Status.ObservedGeneration = mlflowServer.Generation

	// Validate secrets before reconciliation
	if err := r.validateSecrets(ctx, mlflowServer); err != nil {
		logger.Error(err, "Secret validation failed")
		mlflowServer.SetCondition(metav1.Condition{
			Type:               mlopsv1alpha1.ConditionReady,
			Status:             metav1.ConditionFalse,
			Reason:             "SecretValidationFailed",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})
		if err := r.Status().Update(ctx, mlflowServer); err != nil {
			logger.Error(err, "Failed to update status")
		}
		return ctrl.Result{}, err
	}

	// Reconcile ConfigMap
	if err := r.reconcileConfigMap(ctx, mlflowServer); err != nil {
		logger.Error(err, "Failed to reconcile ConfigMap")
		return ctrl.Result{}, err
	}

	// Reconcile PVC if needed
	if err := r.reconcilePVC(ctx, mlflowServer); err != nil {
		logger.Error(err, "Failed to reconcile PVC")
		return ctrl.Result{}, err
	}

	// Reconcile Deployment
	if err := r.reconcileDeployment(ctx, mlflowServer); err != nil {
		logger.Error(err, "Failed to reconcile Deployment")
		return ctrl.Result{}, err
	}

	// Reconcile Service
	if err := r.reconcileService(ctx, mlflowServer); err != nil {
		logger.Error(err, "Failed to reconcile Service")
		return ctrl.Result{}, err
	}

	// Reconcile Ingress if enabled
	if mlflowServer.Spec.Ingress.Enabled {
		if err := r.reconcileIngress(ctx, mlflowServer); err != nil {
			logger.Error(err, "Failed to reconcile Ingress")
			return ctrl.Result{}, err
		}
	}

	// Reconcile HPA if enabled
	if mlflowServer.Spec.Autoscaling != nil && mlflowServer.Spec.Autoscaling.Enabled {
		if err := r.reconcileHPA(ctx, mlflowServer); err != nil {
			logger.Error(err, "Failed to reconcile HPA")
			return ctrl.Result{}, err
		}
	}

	// Reconcile PodDisruptionBudget if enabled
	if mlflowServer.Spec.PodDisruptionBudget != nil && mlflowServer.Spec.PodDisruptionBudget.Enabled {
		if err := r.reconcilePDB(ctx, mlflowServer); err != nil {
			logger.Error(err, "Failed to reconcile PodDisruptionBudget")
			return ctrl.Result{}, err
		}
	}

	// Reconcile ServiceMonitor if enabled (requires prometheus-operator)
	// This is commented out as it requires the prometheus-operator CRD to be installed
	// if mlflowServer.Spec.ServiceMonitor != nil && mlflowServer.Spec.ServiceMonitor.Enabled {
	// 	if err := r.reconcileServiceMonitor(ctx, mlflowServer); err != nil {
	// 		logger.Error(err, "Failed to reconcile ServiceMonitor")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	// Reconcile migration job if enabled
	if mlflowServer.Spec.Migration != nil && mlflowServer.Spec.Migration.Enabled {
		if err := r.reconcileMigrationJob(ctx, mlflowServer); err != nil {
			logger.Error(err, "Failed to reconcile migration job")
			return ctrl.Result{}, err
		}
	}

	// Update status
	if err := r.updateStatus(ctx, mlflowServer); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// validateSecrets validates that all required secrets exist and contain required keys
func (r *MLflowServerReconciler) validateSecrets(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	// Validate backend database secrets
	switch mlflowServer.Spec.Backend.Type {
	case mlopsv1alpha1.BackendTypePostgreSQL:
		cfg := mlflowServer.Spec.Backend.PostgreSQL
		if cfg != nil {
			if err := r.validateSecretKeys(ctx, mlflowServer.Namespace, cfg.CredentialsSecret, []string{"username", "password"}); err != nil {
				return fmt.Errorf("PostgreSQL secret validation failed: %w", err)
			}
		}
	case mlopsv1alpha1.BackendTypeMySQL:
		cfg := mlflowServer.Spec.Backend.MySQL
		if cfg != nil {
			if err := r.validateSecretKeys(ctx, mlflowServer.Namespace, cfg.CredentialsSecret, []string{"username", "password"}); err != nil {
				return fmt.Errorf("MySQL secret validation failed: %w", err)
			}
		}
	}

	// Validate artifact store secrets
	switch mlflowServer.Spec.ArtifactStore.Type {
	case mlopsv1alpha1.ArtifactStoreTypeS3:
		cfg := mlflowServer.Spec.ArtifactStore.S3
		if cfg != nil {
			if err := r.validateSecretKeys(ctx, mlflowServer.Namespace, cfg.CredentialsSecret, []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}); err != nil {
				return fmt.Errorf("S3 secret validation failed: %w", err)
			}
		}
	case mlopsv1alpha1.ArtifactStoreTypeGCS:
		cfg := mlflowServer.Spec.ArtifactStore.GCS
		if cfg != nil {
			if err := r.validateSecretKeys(ctx, mlflowServer.Namespace, cfg.CredentialsSecret, []string{"service-account.json"}); err != nil {
				return fmt.Errorf("GCS secret validation failed: %w", err)
			}
		}
	case mlopsv1alpha1.ArtifactStoreTypeAzure:
		cfg := mlflowServer.Spec.ArtifactStore.Azure
		if cfg != nil {
			if err := r.validateSecretKeys(ctx, mlflowServer.Namespace, cfg.CredentialsSecret, []string{"account-name", "account-key"}); err != nil {
				return fmt.Errorf("Azure secret validation failed: %w", err)
			}
		}
	}

	logger.Info("Secret validation passed")
	return nil
}

// validateSecretKeys validates that a secret exists and contains the required keys
func (r *MLflowServerReconciler) validateSecretKeys(ctx context.Context, namespace, secretName string, requiredKeys []string) error {
	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("secret %s not found in namespace %s", secretName, namespace)
		}
		return fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	// Check for required keys
	for _, key := range requiredKeys {
		if _, exists := secret.Data[key]; !exists {
			return fmt.Errorf("secret %s is missing required key: %s", secretName, key)
		}
	}

	return nil
}

// handleDeletion handles the deletion of MLflowServer resources
func (r *MLflowServerReconciler) handleDeletion(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Remove finalizer if cleanup is complete
	if controllerutil.ContainsFinalizer(mlflowServer, mlopsv1alpha1.MLflowServerFinalizer) {
		controllerutil.RemoveFinalizer(mlflowServer, mlopsv1alpha1.MLflowServerFinalizer)
		if err := r.Update(ctx, mlflowServer); err != nil {
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("Finalizer removed successfully")
	}

	return ctrl.Result{}, nil
}

// reconcileConfigMap creates or updates the ConfigMap for MLflow configuration
func (r *MLflowServerReconciler) reconcileConfigMap(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	desiredConfigMap := r.buildConfigMap(mlflowServer)

	if err := controllerutil.SetControllerReference(mlflowServer, desiredConfigMap, r.Scheme); err != nil {
		return err
	}

	existingConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredConfigMap.Name, Namespace: desiredConfigMap.Namespace}, existingConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating ConfigMap", "name", desiredConfigMap.Name)
			return r.Create(ctx, desiredConfigMap)
		}
		return err
	}

	// Update if needed
	existingConfigMap.Data = desiredConfigMap.Data
	logger.Info("Updating ConfigMap", "name", desiredConfigMap.Name)
	return r.Update(ctx, existingConfigMap)
}

// buildConfigMap builds the ConfigMap for MLflow configuration
func (r *MLflowServerReconciler) buildConfigMap(mlflowServer *mlopsv1alpha1.MLflowServer) *corev1.ConfigMap {
	data := make(map[string]string)

	// Build backend database URI
	backendURI := r.buildBackendURI(mlflowServer)
	data["BACKEND_STORE_URI"] = backendURI

	// Build artifact store URI
	artifactURI := r.buildArtifactStoreURI(mlflowServer)
	data["DEFAULT_ARTIFACT_ROOT"] = artifactURI

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-config", mlflowServer.Name),
			Namespace: mlflowServer.Namespace,
			Labels:    r.getLabels(mlflowServer),
		},
		Data: data,
	}
}

// buildBackendURI builds the backend database URI based on the backend type
func (r *MLflowServerReconciler) buildBackendURI(mlflowServer *mlopsv1alpha1.MLflowServer) string {
	switch mlflowServer.Spec.Backend.Type {
	case mlopsv1alpha1.BackendTypePostgreSQL:
		cfg := mlflowServer.Spec.Backend.PostgreSQL
		if cfg == nil {
			return ""
		}
		port := cfg.Port
		if port == 0 {
			port = 5432
		}
		database := cfg.Database
		if database == "" {
			database = "mlflow"
		}
		// Use environment variable substitution at runtime instead of exposing in ConfigMap
		return fmt.Sprintf("postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@%s:%d/%s", cfg.Host, port, database)

	case mlopsv1alpha1.BackendTypeMySQL:
		cfg := mlflowServer.Spec.Backend.MySQL
		if cfg == nil {
			return ""
		}
		port := cfg.Port
		if port == 0 {
			port = 3306
		}
		database := cfg.Database
		if database == "" {
			database = "mlflow"
		}
		// Use environment variable substitution at runtime instead of exposing in ConfigMap
		return fmt.Sprintf("mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@%s:%d/%s", cfg.Host, port, database)

	case mlopsv1alpha1.BackendTypeSQLite:
		return "sqlite:////mlflow/mlflow.db"

	default:
		return ""
	}
}

// buildArtifactStoreURI builds the artifact store URI based on the artifact store type
func (r *MLflowServerReconciler) buildArtifactStoreURI(mlflowServer *mlopsv1alpha1.MLflowServer) string {
	switch mlflowServer.Spec.ArtifactStore.Type {
	case mlopsv1alpha1.ArtifactStoreTypeS3:
		cfg := mlflowServer.Spec.ArtifactStore.S3
		if cfg == nil {
			return ""
		}
		uri := fmt.Sprintf("s3://%s", cfg.Bucket)
		if cfg.PathPrefix != "" {
			uri += "/" + cfg.PathPrefix
		}
		return uri

	case mlopsv1alpha1.ArtifactStoreTypeGCS:
		cfg := mlflowServer.Spec.ArtifactStore.GCS
		if cfg == nil {
			return ""
		}
		return fmt.Sprintf("gs://%s", cfg.Bucket)

	case mlopsv1alpha1.ArtifactStoreTypeAzure:
		cfg := mlflowServer.Spec.ArtifactStore.Azure
		if cfg == nil {
			return ""
		}
		return fmt.Sprintf("wasbs://%s@%s.blob.core.windows.net", cfg.Container, cfg.StorageAccount)

	case mlopsv1alpha1.ArtifactStoreTypePVC:
		return "/mlflow/artifacts"

	default:
		return ""
	}
}

// reconcilePVC creates or updates the PVC for artifact storage if needed
func (r *MLflowServerReconciler) reconcilePVC(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	// Only create PVC if backend is SQLite or artifact store is PVC
	if mlflowServer.Spec.Backend.Type != mlopsv1alpha1.BackendTypeSQLite &&
		mlflowServer.Spec.ArtifactStore.Type != mlopsv1alpha1.ArtifactStoreTypePVC {
		return nil
	}

	var pvcConfig *mlopsv1alpha1.PVCConfig
	if mlflowServer.Spec.Backend.Type == mlopsv1alpha1.BackendTypeSQLite {
		pvcConfig = mlflowServer.Spec.Backend.SQLite.PVC
	} else {
		pvcConfig = mlflowServer.Spec.ArtifactStore.PVC
	}

	if pvcConfig == nil {
		return nil
	}

	desiredPVC := r.buildPVC(mlflowServer, pvcConfig)

	if err := controllerutil.SetControllerReference(mlflowServer, desiredPVC, r.Scheme); err != nil {
		return err
	}

	existingPVC := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredPVC.Name, Namespace: desiredPVC.Namespace}, existingPVC)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating PVC", "name", desiredPVC.Name)
			return r.Create(ctx, desiredPVC)
		}
		return err
	}

	return nil
}

// buildPVC builds the PVC for storage
func (r *MLflowServerReconciler) buildPVC(mlflowServer *mlopsv1alpha1.MLflowServer, pvcConfig *mlopsv1alpha1.PVCConfig) *corev1.PersistentVolumeClaim {
	size := pvcConfig.Size
	if size == "" {
		size = "100Gi"
	}

	accessMode := pvcConfig.AccessMode
	if accessMode == "" {
		accessMode = "ReadWriteOnce"
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-storage", mlflowServer.Name),
			Namespace: mlflowServer.Namespace,
			Labels:    r.getLabels(mlflowServer),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.PersistentVolumeAccessMode(accessMode)},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}

	if pvcConfig.StorageClass != "" {
		pvc.Spec.StorageClassName = &pvcConfig.StorageClass
	}

	return pvc
}

// reconcileDeployment creates or updates the Deployment for MLflow
func (r *MLflowServerReconciler) reconcileDeployment(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	desiredDeployment := r.buildDeployment(mlflowServer)

	if err := controllerutil.SetControllerReference(mlflowServer, desiredDeployment, r.Scheme); err != nil {
		return err
	}

	existingDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredDeployment.Name, Namespace: desiredDeployment.Namespace}, existingDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating Deployment", "name", desiredDeployment.Name)
			return r.Create(ctx, desiredDeployment)
		}
		return err
	}

	// Update if needed
	if existingDeployment.Spec.Replicas != desiredDeployment.Spec.Replicas {
		existingDeployment.Spec.Replicas = desiredDeployment.Spec.Replicas
		logger.Info("Updating Deployment replicas", "name", desiredDeployment.Name)
		return r.Update(ctx, existingDeployment)
	}

	return nil
}

// buildDeployment builds the Deployment for MLflow
func (r *MLflowServerReconciler) buildDeployment(mlflowServer *mlopsv1alpha1.MLflowServer) *appsv1.Deployment {
	replicas := mlflowServer.Spec.Tracking.Replicas
	if replicas == 0 {
		replicas = 1
	}

	image := fmt.Sprintf("ghcr.io/mlflow/mlflow:%s", mlflowServer.Spec.Version)

	env := []corev1.EnvVar{
		{
			Name: "BACKEND_STORE_URI",
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-config", mlflowServer.Name),
					},
					Key: "BACKEND_STORE_URI",
				},
			},
		},
		{
			Name: "DEFAULT_ARTIFACT_ROOT",
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-config", mlflowServer.Name),
					},
					Key: "DEFAULT_ARTIFACT_ROOT",
				},
			},
		},
	}

	// Add credential environment variables based on backend type
	env = append(env, r.buildCredentialEnvVars(mlflowServer)...)

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	// Add PVC volume if needed
	if mlflowServer.Spec.Backend.Type == mlopsv1alpha1.BackendTypeSQLite ||
		mlflowServer.Spec.ArtifactStore.Type == mlopsv1alpha1.ArtifactStoreTypePVC {
		volumes = append(volumes, corev1.Volume{
			Name: "mlflow-storage",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("%s-storage", mlflowServer.Name),
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "mlflow-storage",
			MountPath: "/mlflow",
		})
	}

	// Add GCS secret volume if using GCS artifact store
	if mlflowServer.Spec.ArtifactStore.Type == mlopsv1alpha1.ArtifactStoreTypeGCS {
		cfg := mlflowServer.Spec.ArtifactStore.GCS
		if cfg != nil && cfg.CredentialsSecret != "" {
			volumes = append(volumes, corev1.Volume{
				Name: "gcs-credentials",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: cfg.CredentialsSecret,
					},
				},
			})
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      "gcs-credentials",
				MountPath: "/var/secrets/google",
				ReadOnly:  true,
			})
		}
	}

	// Build MLflow server args with custom arguments
	args := []string{"server", "--host", "0.0.0.0", "--port", "5000"}
	if len(mlflowServer.Spec.Tracking.AdditionalArgs) > 0 {
		args = append(args, mlflowServer.Spec.Tracking.AdditionalArgs...)
	}

	// Build pod labels
	podLabels := r.getLabels(mlflowServer)
	for k, v := range mlflowServer.Spec.Tracking.PodLabels {
		podLabels[k] = v
	}

	// Build the MLflow container
	mlflowContainer := corev1.Container{
		Name:            "mlflow",
		Image:           image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args:            args,
		Env:             env,
		VolumeMounts: append(volumeMounts, corev1.VolumeMount{
			Name:      "tmp",
			MountPath: "/tmp",
		}),
		Resources: mlflowServer.Spec.Tracking.Resources,
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: boolPtr(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
			ReadOnlyRootFilesystem: boolPtr(true),
			RunAsNonRoot:           boolPtr(true),
			RunAsUser:              int64Ptr(1000),
		},
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: 5000,
				Protocol:      corev1.ProtocolTCP,
			},
		},
	}

	// Apply custom probes if configured
	if mlflowServer.Spec.Tracking.Probes != nil {
		if mlflowServer.Spec.Tracking.Probes.LivenessProbe != nil {
			mlflowContainer.LivenessProbe = mlflowServer.Spec.Tracking.Probes.LivenessProbe
		} else {
			mlflowContainer.LivenessProbe = &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/health",
						Port: intstr.FromInt(5000),
					},
				},
				InitialDelaySeconds: 30,
				PeriodSeconds:       10,
			}
		}
		if mlflowServer.Spec.Tracking.Probes.ReadinessProbe != nil {
			mlflowContainer.ReadinessProbe = mlflowServer.Spec.Tracking.Probes.ReadinessProbe
		} else {
			mlflowContainer.ReadinessProbe = &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/health",
						Port: intstr.FromInt(5000),
					},
				},
				InitialDelaySeconds: 10,
				PeriodSeconds:       5,
			}
		}
		if mlflowServer.Spec.Tracking.Probes.StartupProbe != nil {
			mlflowContainer.StartupProbe = mlflowServer.Spec.Tracking.Probes.StartupProbe
		}
	} else {
		// Default probes
		mlflowContainer.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health",
					Port: intstr.FromInt(5000),
				},
			},
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
		}
		mlflowContainer.ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health",
					Port: intstr.FromInt(5000),
				},
			},
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
		}
	}

	// Apply lifecycle hooks if configured
	if mlflowServer.Spec.Tracking.Lifecycle != nil {
		mlflowContainer.Lifecycle = mlflowServer.Spec.Tracking.Lifecycle
	}

	// Build containers list
	containers := []corev1.Container{mlflowContainer}
	containers = append(containers, mlflowServer.Spec.Tracking.SidecarContainers...)

	// Build pod spec
	podSpec := corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: boolPtr(true),
			RunAsUser:    int64Ptr(1000),
			FSGroup:      int64Ptr(1000),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
		Containers: containers,
		Volumes: append(volumes, corev1.Volume{
			Name: "tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}),
	}

	// Add init containers if configured
	if len(mlflowServer.Spec.Tracking.InitContainers) > 0 {
		podSpec.InitContainers = mlflowServer.Spec.Tracking.InitContainers
	}

	// Apply scheduling configuration
	if mlflowServer.Spec.Scheduling != nil {
		podSpec.NodeSelector = mlflowServer.Spec.Scheduling.NodeSelector
		podSpec.Tolerations = mlflowServer.Spec.Scheduling.Tolerations
		podSpec.Affinity = mlflowServer.Spec.Scheduling.Affinity
		podSpec.PriorityClassName = mlflowServer.Spec.Scheduling.PriorityClassName
		podSpec.TopologySpreadConstraints = mlflowServer.Spec.Scheduling.TopologySpreadConstraints
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mlflowServer.Name,
			Namespace: mlflowServer.Namespace,
			Labels:    r.getLabels(mlflowServer),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.getSelectorLabels(mlflowServer),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: mlflowServer.Spec.Tracking.PodAnnotations,
				},
				Spec: podSpec,
			},
		},
	}

	return deployment
}

// buildCredentialEnvVars builds environment variables for credentials
func (r *MLflowServerReconciler) buildCredentialEnvVars(mlflowServer *mlopsv1alpha1.MLflowServer) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	switch mlflowServer.Spec.Backend.Type {
	case mlopsv1alpha1.BackendTypePostgreSQL:
		cfg := mlflowServer.Spec.Backend.PostgreSQL
		if cfg != nil {
			env = append(env,
				corev1.EnvVar{
					Name: "POSTGRES_USER",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: cfg.CredentialsSecret},
							Key:                  "username",
						},
					},
				},
				corev1.EnvVar{
					Name: "POSTGRES_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: cfg.CredentialsSecret},
							Key:                  "password",
						},
					},
				},
			)
		}
	case mlopsv1alpha1.BackendTypeMySQL:
		cfg := mlflowServer.Spec.Backend.MySQL
		if cfg != nil {
			env = append(env,
				corev1.EnvVar{
					Name: "MYSQL_USER",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: cfg.CredentialsSecret},
							Key:                  "username",
						},
					},
				},
				corev1.EnvVar{
					Name: "MYSQL_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: cfg.CredentialsSecret},
							Key:                  "password",
						},
					},
				},
			)
		}
	}

	switch mlflowServer.Spec.ArtifactStore.Type {
	case mlopsv1alpha1.ArtifactStoreTypeS3:
		cfg := mlflowServer.Spec.ArtifactStore.S3
		if cfg != nil {
			env = append(env,
				corev1.EnvVar{
					Name: "AWS_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: cfg.CredentialsSecret},
							Key:                  "AWS_ACCESS_KEY_ID",
						},
					},
				},
				corev1.EnvVar{
					Name: "AWS_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: cfg.CredentialsSecret},
							Key:                  "AWS_SECRET_ACCESS_KEY",
						},
					},
				},
			)
			if cfg.EndpointURL != "" {
				env = append(env, corev1.EnvVar{
					Name:  "MLFLOW_S3_ENDPOINT_URL",
					Value: cfg.EndpointURL,
				})
			}
		}
	case mlopsv1alpha1.ArtifactStoreTypeGCS:
		cfg := mlflowServer.Spec.ArtifactStore.GCS
		if cfg != nil {
			env = append(env, corev1.EnvVar{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "/var/secrets/google/service-account.json",
			})
		}
	case mlopsv1alpha1.ArtifactStoreTypeAzure:
		cfg := mlflowServer.Spec.ArtifactStore.Azure
		if cfg != nil {
			env = append(env,
				corev1.EnvVar{
					Name: "AZURE_STORAGE_ACCOUNT",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: cfg.CredentialsSecret},
							Key:                  "account-name",
						},
					},
				},
				corev1.EnvVar{
					Name: "AZURE_STORAGE_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: cfg.CredentialsSecret},
							Key:                  "account-key",
						},
					},
				},
			)
		}
	}

	return env
}

// reconcileService creates or updates the Service for MLflow
func (r *MLflowServerReconciler) reconcileService(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	desiredService := r.buildService(mlflowServer)

	if err := controllerutil.SetControllerReference(mlflowServer, desiredService, r.Scheme); err != nil {
		return err
	}

	existingService := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredService.Name, Namespace: desiredService.Namespace}, existingService)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating Service", "name", desiredService.Name)
			return r.Create(ctx, desiredService)
		}
		return err
	}

	return nil
}

// buildService builds the Service for MLflow
func (r *MLflowServerReconciler) buildService(mlflowServer *mlopsv1alpha1.MLflowServer) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mlflowServer.Name,
			Namespace: mlflowServer.Namespace,
			Labels:    r.getLabels(mlflowServer),
		},
		Spec: corev1.ServiceSpec{
			Selector: r.getSelectorLabels(mlflowServer),
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       5000,
					TargetPort: intstr.FromInt(5000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// reconcileIngress creates or updates the Ingress for MLflow
func (r *MLflowServerReconciler) reconcileIngress(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	desiredIngress := r.buildIngress(mlflowServer)

	if err := controllerutil.SetControllerReference(mlflowServer, desiredIngress, r.Scheme); err != nil {
		return err
	}

	existingIngress := &networkingv1.Ingress{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredIngress.Name, Namespace: desiredIngress.Namespace}, existingIngress)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating Ingress", "name", desiredIngress.Name)
			return r.Create(ctx, desiredIngress)
		}
		return err
	}

	return nil
}

// buildIngress builds the Ingress for MLflow
func (r *MLflowServerReconciler) buildIngress(mlflowServer *mlopsv1alpha1.MLflowServer) *networkingv1.Ingress {
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        mlflowServer.Name,
			Namespace:   mlflowServer.Namespace,
			Labels:      r.getLabels(mlflowServer),
			Annotations: mlflowServer.Spec.Ingress.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &mlflowServer.Spec.Ingress.IngressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: mlflowServer.Spec.Ingress.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: mlflowServer.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: 5000,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Add TLS configuration
	if mlflowServer.Spec.Ingress.TLS != nil && mlflowServer.Spec.Ingress.TLS.Enabled {
		if mlflowServer.Spec.Ingress.TLS.Issuer != "" {
			// Use cert-manager annotation
			if ingress.Annotations == nil {
				ingress.Annotations = make(map[string]string)
			}
			ingress.Annotations["cert-manager.io/cluster-issuer"] = mlflowServer.Spec.Ingress.TLS.Issuer
		}

		secretName := mlflowServer.Spec.Ingress.TLS.SecretName
		if secretName == "" {
			secretName = fmt.Sprintf("%s-tls", mlflowServer.Name)
		}

		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{mlflowServer.Spec.Ingress.Host},
				SecretName: secretName,
			},
		}
	}

	return ingress
}

// reconcileHPA creates or updates the HorizontalPodAutoscaler for MLflow
func (r *MLflowServerReconciler) reconcileHPA(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	desiredHPA := r.buildHPA(mlflowServer)

	if err := controllerutil.SetControllerReference(mlflowServer, desiredHPA, r.Scheme); err != nil {
		return err
	}

	existingHPA := &autoscalingv2.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredHPA.Name, Namespace: desiredHPA.Namespace}, existingHPA)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating HPA", "name", desiredHPA.Name)
			return r.Create(ctx, desiredHPA)
		}
		return err
	}

	// Update if needed
	existingHPA.Spec = desiredHPA.Spec
	logger.Info("Updating HPA", "name", desiredHPA.Name)
	return r.Update(ctx, existingHPA)
}

// buildHPA builds the HorizontalPodAutoscaler for MLflow
func (r *MLflowServerReconciler) buildHPA(mlflowServer *mlopsv1alpha1.MLflowServer) *autoscalingv2.HorizontalPodAutoscaler {
	cfg := mlflowServer.Spec.Autoscaling
	minReplicas := int32(1)
	if cfg.MinReplicas > 0 {
		minReplicas = cfg.MinReplicas
	}
	maxReplicas := int32(10)
	if cfg.MaxReplicas > 0 {
		maxReplicas = cfg.MaxReplicas
	}

	metrics := []autoscalingv2.MetricSpec{}
	if cfg.TargetCPUUtilizationPercentage > 0 {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: &cfg.TargetCPUUtilizationPercentage,
				},
			},
		})
	}
	if cfg.TargetMemoryUtilizationPercentage > 0 {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: &cfg.TargetMemoryUtilizationPercentage,
				},
			},
		})
	}

	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mlflowServer.Name,
			Namespace: mlflowServer.Namespace,
			Labels:    r.getLabels(mlflowServer),
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       mlflowServer.Name,
			},
			MinReplicas: &minReplicas,
			MaxReplicas: maxReplicas,
			Metrics:     metrics,
		},
	}
}

// reconcilePDB creates or updates the PodDisruptionBudget for MLflow
func (r *MLflowServerReconciler) reconcilePDB(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	desiredPDB := r.buildPDB(mlflowServer)

	if err := controllerutil.SetControllerReference(mlflowServer, desiredPDB, r.Scheme); err != nil {
		return err
	}

	existingPDB := &policyv1.PodDisruptionBudget{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredPDB.Name, Namespace: desiredPDB.Namespace}, existingPDB)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating PodDisruptionBudget", "name", desiredPDB.Name)
			return r.Create(ctx, desiredPDB)
		}
		return err
	}

	// Update if needed
	existingPDB.Spec = desiredPDB.Spec
	logger.Info("Updating PodDisruptionBudget", "name", desiredPDB.Name)
	return r.Update(ctx, existingPDB)
}

// buildPDB builds the PodDisruptionBudget for MLflow
func (r *MLflowServerReconciler) buildPDB(mlflowServer *mlopsv1alpha1.MLflowServer) *policyv1.PodDisruptionBudget {
	cfg := mlflowServer.Spec.PodDisruptionBudget
	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mlflowServer.Name,
			Namespace: mlflowServer.Namespace,
			Labels:    r.getLabels(mlflowServer),
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: r.getSelectorLabels(mlflowServer),
			},
		},
	}

	if cfg.MinAvailable != nil {
		minAvailable := intstr.FromInt(int(*cfg.MinAvailable))
		pdb.Spec.MinAvailable = &minAvailable
	}
	if cfg.MaxUnavailable != nil {
		maxUnavailable := intstr.FromInt(int(*cfg.MaxUnavailable))
		pdb.Spec.MaxUnavailable = &maxUnavailable
	}

	return pdb
}

// reconcileServiceMonitor creates or updates the ServiceMonitor for MLflow
// This function is commented out as it requires the prometheus-operator CRD to be installed
// To enable, add the monitoringv1 import and uncomment this function
/*
func (r *MLflowServerReconciler) reconcileServiceMonitor(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	desiredSM := r.buildServiceMonitor(mlflowServer)

	if err := controllerutil.SetControllerReference(mlflowServer, desiredSM, r.Scheme); err != nil {
		return err
	}

	existingSM := &monitoringv1.ServiceMonitor{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredSM.Name, Namespace: desiredSM.Namespace}, existingSM)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating ServiceMonitor", "name", desiredSM.Name)
			return r.Create(ctx, desiredSM)
		}
		return err
	}

	// Update if needed
	existingSM.Spec = desiredSM.Spec
	existingSM.Labels = desiredSM.Labels
	logger.Info("Updating ServiceMonitor", "name", desiredSM.Name)
	return r.Update(ctx, existingSM)
}

// buildServiceMonitor builds the ServiceMonitor for MLflow
func (r *MLflowServerReconciler) buildServiceMonitor(mlflowServer *mlopsv1alpha1.MLflowServer) *monitoringv1.ServiceMonitor {
	cfg := mlflowServer.Spec.ServiceMonitor
	interval := "30s"
	if cfg.Interval != "" {
		interval = cfg.Interval
	}
	scrapeTimeout := "10s"
	if cfg.ScrapeTimeout != "" {
		scrapeTimeout = cfg.ScrapeTimeout
	}

	labels := r.getLabels(mlflowServer)
	for k, v := range cfg.Labels {
		labels[k] = v
	}

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mlflowServer.Name,
			Namespace: mlflowServer.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: r.getSelectorLabels(mlflowServer),
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:          "http",
					Interval:      monitoringv1.Duration(interval),
					ScrapeTimeout: monitoringv1.Duration(scrapeTimeout),
					Path:          "/metrics",
				},
			},
		},
	}
}
*/

// reconcileMigrationJob creates or updates the migration job for MLflow
func (r *MLflowServerReconciler) reconcileMigrationJob(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	// Check if migration is needed by comparing current version with previous
	// For now, we'll create a job if the migration config is enabled
	desiredJob := r.buildMigrationJob(mlflowServer)

	if err := controllerutil.SetControllerReference(mlflowServer, desiredJob, r.Scheme); err != nil {
		return err
	}

	existingJob := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredJob.Name, Namespace: desiredJob.Namespace}, existingJob)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating migration job", "name", desiredJob.Name)
			return r.Create(ctx, desiredJob)
		}
		return err
	}

	// Don't update if job is already completed
	if existingJob.Status.Succeeded > 0 {
		logger.Info("Migration job already completed", "name", desiredJob.Name)
		return nil
	}

	return nil
}

// buildMigrationJob builds the migration job for MLflow
func (r *MLflowServerReconciler) buildMigrationJob(mlflowServer *mlopsv1alpha1.MLflowServer) *batchv1.Job {
	cfg := mlflowServer.Spec.Migration
	backoffLimit := int32(6)
	if cfg.BackoffLimit > 0 {
		backoffLimit = cfg.BackoffLimit
	}

	image := fmt.Sprintf("ghcr.io/mlflow/mlflow:%s", mlflowServer.Spec.Version)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-migration", mlflowServer.Name),
			Namespace:   mlflowServer.Namespace,
			Labels:      r.getLabels(mlflowServer),
			Annotations: cfg.JobAnnotations,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.getLabels(mlflowServer),
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:            "mlflow-migrate",
							Image:           image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"mlflow", "db", "upgrade"},
							Env:             r.buildCredentialEnvVars(mlflowServer),
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: fmt.Sprintf("%s-config", mlflowServer.Name),
										},
									},
								},
							},
						},
					},
					Volumes: r.buildMigrationVolumes(mlflowServer),
				},
			},
		},
	}

	if cfg.ActiveDeadlineSeconds > 0 {
		job.Spec.ActiveDeadlineSeconds = &cfg.ActiveDeadlineSeconds
	}

	return job
}

// buildMigrationVolumes builds the volumes for the migration job
func (r *MLflowServerReconciler) buildMigrationVolumes(mlflowServer *mlopsv1alpha1.MLflowServer) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	// Add PVC volume if needed
	if mlflowServer.Spec.Backend.Type == mlopsv1alpha1.BackendTypeSQLite ||
		mlflowServer.Spec.ArtifactStore.Type == mlopsv1alpha1.ArtifactStoreTypePVC {
		volumes = append(volumes, corev1.Volume{
			Name: "mlflow-storage",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("%s-storage", mlflowServer.Name),
				},
			},
		})
	}

	// Add GCS secret volume if using GCS artifact store
	if mlflowServer.Spec.ArtifactStore.Type == mlopsv1alpha1.ArtifactStoreTypeGCS {
		cfg := mlflowServer.Spec.ArtifactStore.GCS
		if cfg != nil && cfg.CredentialsSecret != "" {
			volumes = append(volumes, corev1.Volume{
				Name: "gcs-credentials",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: cfg.CredentialsSecret,
					},
				},
			})
		}
	}

	return volumes
}

// checkBackendConnectivity checks if the backend database is reachable
func (r *MLflowServerReconciler) checkBackendConnectivity(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	// For now, we'll implement a basic check based on the backend type
	// In a production environment, you would want to actually connect to the database
	// and verify connectivity with proper credentials

	switch mlflowServer.Spec.Backend.Type {
	case mlopsv1alpha1.BackendTypePostgreSQL, mlopsv1alpha1.BackendTypeMySQL:
		// Validate that the secret exists
		var secretName string
		if mlflowServer.Spec.Backend.PostgreSQL != nil {
			secretName = mlflowServer.Spec.Backend.PostgreSQL.CredentialsSecret
		} else if mlflowServer.Spec.Backend.MySQL != nil {
			secretName = mlflowServer.Spec.Backend.MySQL.CredentialsSecret
		}

		if secretName != "" {
			secret := &corev1.Secret{}
			err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: mlflowServer.Namespace}, secret)
			if err != nil {
				logger.Error(err, "Backend secret not found")
				return err
			}
		}
		logger.Info("Backend connectivity check passed")
	case mlopsv1alpha1.BackendTypeSQLite:
		logger.Info("SQLite backend - no connectivity check needed")
	}

	return nil
}

// checkArtifactStoreConnectivity checks if the artifact store is reachable
func (r *MLflowServerReconciler) checkArtifactStoreConnectivity(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	// For now, we'll implement a basic check based on the artifact store type
	// In a production environment, you would want to actually connect to the store
	// and verify connectivity with proper credentials

	switch mlflowServer.Spec.ArtifactStore.Type {
	case mlopsv1alpha1.ArtifactStoreTypeS3, mlopsv1alpha1.ArtifactStoreTypeGCS, mlopsv1alpha1.ArtifactStoreTypeAzure:
		// Validate that the secret exists
		var secretName string
		if mlflowServer.Spec.ArtifactStore.S3 != nil {
			secretName = mlflowServer.Spec.ArtifactStore.S3.CredentialsSecret
		} else if mlflowServer.Spec.ArtifactStore.GCS != nil {
			secretName = mlflowServer.Spec.ArtifactStore.GCS.CredentialsSecret
		} else if mlflowServer.Spec.ArtifactStore.Azure != nil {
			secretName = mlflowServer.Spec.ArtifactStore.Azure.CredentialsSecret
		}

		if secretName != "" {
			secret := &corev1.Secret{}
			err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: mlflowServer.Namespace}, secret)
			if err != nil {
				logger.Error(err, "Artifact store secret not found")
				return err
			}
		}
		logger.Info("Artifact store connectivity check passed")
	case mlopsv1alpha1.ArtifactStoreTypePVC:
		logger.Info("PVC artifact store - no connectivity check needed")
	}

	return nil
}

// updateStatus updates the status of the MLflowServer
func (r *MLflowServerReconciler) updateStatus(ctx context.Context, mlflowServer *mlopsv1alpha1.MLflowServer) error {
	logger := log.FromContext(ctx)

	// Check if Deployment is ready
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: mlflowServer.Name, Namespace: mlflowServer.Namespace}, deployment)
	if err != nil {
		mlflowServer.Status.Ready = false
		mlflowServer.SetCondition(metav1.Condition{
			Type:               mlopsv1alpha1.ConditionReady,
			Status:             metav1.ConditionFalse,
			Reason:             "DeploymentNotFound",
			Message:            "Deployment not found",
			LastTransitionTime: metav1.Now(),
		})
	} else {
		if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
			mlflowServer.Status.Ready = true
			mlflowServer.SetCondition(metav1.Condition{
				Type:               mlopsv1alpha1.ConditionReady,
				Status:             metav1.ConditionTrue,
				Reason:             "DeploymentReady",
				Message:            "All replicas are ready",
				LastTransitionTime: metav1.Now(),
			})
		} else {
			mlflowServer.Status.Ready = false
			mlflowServer.SetCondition(metav1.Condition{
				Type:               mlopsv1alpha1.ConditionReady,
				Status:             metav1.ConditionFalse,
				Reason:             "DeploymentNotReady",
				Message:            fmt.Sprintf("Expected %d replicas, got %d ready", *deployment.Spec.Replicas, deployment.Status.ReadyReplicas),
				LastTransitionTime: metav1.Now(),
			})
		}
	}

	// Check artifact store connectivity
	if err := r.checkArtifactStoreConnectivity(ctx, mlflowServer); err != nil {
		mlflowServer.Status.ArtifactStoreConnected = false
		mlflowServer.SetCondition(metav1.Condition{
			Type:               mlopsv1alpha1.ConditionArtifactStoreConnected,
			Status:             metav1.ConditionFalse,
			Reason:             "ConnectivityCheckFailed",
			Message:            fmt.Sprintf("Artifact store connectivity check failed: %v", err),
			LastTransitionTime: metav1.Now(),
		})
	} else {
		mlflowServer.Status.ArtifactStoreConnected = true
		mlflowServer.SetCondition(metav1.Condition{
			Type:               mlopsv1alpha1.ConditionArtifactStoreConnected,
			Status:             metav1.ConditionTrue,
			Reason:             "Connected",
			Message:            "Artifact store is connected",
			LastTransitionTime: metav1.Now(),
		})
	}

	// Check backend connectivity
	if err := r.checkBackendConnectivity(ctx, mlflowServer); err != nil {
		mlflowServer.Status.BackendConnected = false
		mlflowServer.SetCondition(metav1.Condition{
			Type:               mlopsv1alpha1.ConditionBackendConnected,
			Status:             metav1.ConditionFalse,
			Reason:             "ConnectivityCheckFailed",
			Message:            fmt.Sprintf("Backend connectivity check failed: %v", err),
			LastTransitionTime: metav1.Now(),
		})
	} else {
		mlflowServer.Status.BackendConnected = true
		mlflowServer.SetCondition(metav1.Condition{
			Type:               mlopsv1alpha1.ConditionBackendConnected,
			Status:             metav1.ConditionTrue,
			Reason:             "Connected",
			Message:            "Backend database is connected",
			LastTransitionTime: metav1.Now(),
		})
	}

	logger.Info("Updating status")
	return r.Status().Update(ctx, mlflowServer)
}

// getLabels returns the labels for resources
func (r *MLflowServerReconciler) getLabels(mlflowServer *mlopsv1alpha1.MLflowServer) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "mlflow",
		"app.kubernetes.io/instance":   mlflowServer.Name,
		"app.kubernetes.io/managed-by": "mlflow-k8s-operator",
		"app.kubernetes.io/component":  "tracking-server",
	}
}

// getSelectorLabels returns the selector labels for resources
func (r *MLflowServerReconciler) getSelectorLabels(mlflowServer *mlopsv1alpha1.MLflowServer) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "mlflow",
		"app.kubernetes.io/instance": mlflowServer.Name,
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *MLflowServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.MLflowServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&networkingv1.Ingress{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&policyv1.PodDisruptionBudget{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

var (
	pathTypePrefix = networkingv1.PathTypePrefix
)

// Helper functions for pointer conversions
func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}
