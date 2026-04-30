package v1alpha1

import (
	"context"
	"fmt"
	"net"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var mlflowserverlog = ctrl.Log.WithName("mlflowserver-resource")

func (r *MLflowServer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-mlops-NotHarshhaa-io-v1alpha1-mlflowserver,mutating=true,failurePolicy=fail,sideEffects=None,groups=mlops.NotHarshhaa.io,resources=mlflowservers,verbs=create;update,versions=v1alpha1,name=mmlflowserver.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &MLflowServer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MLflowServer) Default() {
	mlflowserverlog.Info("default", "name", r.Name)

	// Set default values
	if r.Spec.Tracking.Replicas == 0 {
		r.Spec.Tracking.Replicas = 1
	}

	if r.Spec.Version == "" {
		r.Spec.Version = "2.11.0"
	}

	if r.Spec.Backend.Type == "" {
		r.Spec.Backend.Type = BackendTypeSQLite
	}

	// Set default PostgreSQL port
	if r.Spec.Backend.PostgreSQL != nil && r.Spec.Backend.PostgreSQL.Port == 0 {
		r.Spec.Backend.PostgreSQL.Port = 5432
	}

	// Set default MySQL port
	if r.Spec.Backend.MySQL != nil && r.Spec.Backend.MySQL.Port == 0 {
		r.Spec.Backend.MySQL.Port = 3306
	}

	// Set default database name
	if r.Spec.Backend.PostgreSQL != nil && r.Spec.Backend.PostgreSQL.Database == "" {
		r.Spec.Backend.PostgreSQL.Database = "mlflow"
	}
	if r.Spec.Backend.MySQL != nil && r.Spec.Backend.MySQL.Database == "" {
		r.Spec.Backend.MySQL.Database = "mlflow"
	}

	// Set default artifact store type
	if r.Spec.ArtifactStore.Type == "" {
		r.Spec.ArtifactStore.Type = ArtifactStoreTypePVC
	}

	// Set default PVC size
	if r.Spec.ArtifactStore.PVC != nil && r.Spec.ArtifactStore.PVC.Size == "" {
		r.Spec.ArtifactStore.PVC.Size = "100Gi"
	}
}

//+kubebuilder:webhook:path=/validate-mlops-NotHarshhaa-io-v1alpha1-mlflowserver,mutating=false,failurePolicy=fail,sideEffects=None,groups=mlops.NotHarshhaa.io,resources=mlflowservers,verbs=create;update,versions=v1alpha1,name=vmlflowserver.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MLflowServer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MLflowServer) ValidateCreate() (admission.Warnings, error) {
	mlflowserverlog.Info("validate create", "name", r.Name)
	return nil, r.validateMLflowServer()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MLflowServer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	mlflowserverlog.Info("validate update", "name", r.Name)
	return nil, r.validateMLflowServer()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MLflowServer) ValidateDelete() (admission.Warnings, error) {
	mlflowserverlog.Info("validate delete", "name", r.Name)
	return nil, nil
}

func (r *MLflowServer) validateMLflowServer() error {
	// Validate namespace if multi-tenancy is enabled
	if err := r.validateNamespace(); err != nil {
		return err
	}

	// Validate backend configuration
	if err := r.validateBackend(); err != nil {
		return err
	}

	// Validate artifact store configuration
	if err := r.validateArtifactStore(); err != nil {
		return err
	}

	// Validate ingress configuration
	if err := r.validateIngress(); err != nil {
		return err
	}

	// Validate resources
	if err := r.validateResources(); err != nil {
		return err
	}

	return nil
}

func (r *MLflowServer) validateNamespace() error {
	// This function validates namespace access for multi-tenancy
	// In a real implementation, you would check against a ConfigMap or annotation
	// that defines allowed namespaces

	// For now, we'll allow all namespaces
	// To enable multi-tenancy, you would:
	// 1. Read the AllowedNamespacesAnnotation from the operator's ConfigMap
	// 2. Check if the current namespace is in the allowed list
	// 3. Return an error if the namespace is not allowed

	// Example implementation (commented out):
	/*
		if allowedNamespaces, ok := r.Annotations[AllowedNamespacesAnnotation]; ok {
			allowed := strings.Split(allowedNamespaces, ",")
			allowedMap := make(map[string]bool)
			for _, ns := range allowed {
				allowedMap[strings.TrimSpace(ns)] = true
			}
			if !allowedMap[r.Namespace] {
				return fmt.Errorf("namespace %s is not in the allowed namespaces list", r.Namespace)
			}
		}
	*/

	return nil
}

func (r *MLflowServer) validateBackend() error {
	switch r.Spec.Backend.Type {
	case BackendTypePostgreSQL:
		cfg := r.Spec.Backend.PostgreSQL
		if cfg == nil {
			return fmt.Errorf("PostgreSQL configuration is required when backend type is PostgreSQL")
		}
		if cfg.Host == "" {
			return fmt.Errorf("PostgreSQL host is required")
		}
		if cfg.CredentialsSecret == "" {
			return fmt.Errorf("PostgreSQL credentials secret is required")
		}
		if err := validateHostname(cfg.Host); err != nil {
			return fmt.Errorf("invalid PostgreSQL host: %w", err)
		}
	case BackendTypeMySQL:
		cfg := r.Spec.Backend.MySQL
		if cfg == nil {
			return fmt.Errorf("MySQL configuration is required when backend type is MySQL")
		}
		if cfg.Host == "" {
			return fmt.Errorf("MySQL host is required")
		}
		if cfg.CredentialsSecret == "" {
			return fmt.Errorf("MySQL credentials secret is required")
		}
		if err := validateHostname(cfg.Host); err != nil {
			return fmt.Errorf("invalid MySQL host: %w", err)
		}
	}
	return nil
}

func (r *MLflowServer) validateArtifactStore() error {
	switch r.Spec.ArtifactStore.Type {
	case ArtifactStoreTypeS3:
		cfg := r.Spec.ArtifactStore.S3
		if cfg == nil {
			return fmt.Errorf("S3 configuration is required when artifact store type is S3")
		}
		if cfg.Bucket == "" {
			return fmt.Errorf("S3 bucket name is required")
		}
		if cfg.Region == "" {
			return fmt.Errorf("S3 region is required")
		}
		if cfg.CredentialsSecret == "" {
			return fmt.Errorf("S3 credentials secret is required")
		}
		if err := validateS3BucketName(cfg.Bucket); err != nil {
			return fmt.Errorf("invalid S3 bucket name: %w", err)
		}
		if err := validateAWSRegion(cfg.Region); err != nil {
			return fmt.Errorf("invalid AWS region: %w", err)
		}
	case ArtifactStoreTypeGCS:
		cfg := r.Spec.ArtifactStore.GCS
		if cfg == nil {
			return fmt.Errorf("GCS configuration is required when artifact store type is GCS")
		}
		if cfg.Bucket == "" {
			return fmt.Errorf("GCS bucket name is required")
		}
		if cfg.CredentialsSecret == "" {
			return fmt.Errorf("GCS credentials secret is required")
		}
		if err := validateGCSBucketName(cfg.Bucket); err != nil {
			return fmt.Errorf("invalid GCS bucket name: %w", err)
		}
	case ArtifactStoreTypeAzure:
		cfg := r.Spec.ArtifactStore.Azure
		if cfg == nil {
			return fmt.Errorf("Azure configuration is required when artifact store type is Azure")
		}
		if cfg.StorageAccount == "" {
			return fmt.Errorf("Azure storage account name is required")
		}
		if cfg.Container == "" {
			return fmt.Errorf("Azure container name is required")
		}
		if cfg.CredentialsSecret == "" {
			return fmt.Errorf("Azure credentials secret is required")
		}
		if err := validateAzureStorageAccount(cfg.StorageAccount); err != nil {
			return fmt.Errorf("invalid Azure storage account name: %w", err)
		}
	}
	return nil
}

func (r *MLflowServer) validateIngress() error {
	if !r.Spec.Ingress.Enabled {
		return nil
	}

	if r.Spec.Ingress.Host == "" {
		return fmt.Errorf("ingress host is required when ingress is enabled")
	}

	if err := validateHostname(r.Spec.Ingress.Host); err != nil {
		return fmt.Errorf("invalid ingress host: %w", err)
	}

	return nil
}

func (r *MLflowServer) validateResources() error {
	resources := r.Spec.Tracking.Resources
	if resources.Limits.Cpu().IsZero() && resources.Limits.Memory().IsZero() &&
		resources.Requests.Cpu().IsZero() && resources.Requests.Memory().IsZero() {
		return fmt.Errorf("resource limits or requests should be specified")
	}

	// Validate that requests don't exceed limits
	if !resources.Requests.Cpu().IsZero() && !resources.Limits.Cpu().IsZero() {
		if resources.Requests.Cpu().Cmp(*resources.Limits.Cpu()) > 0 {
			return fmt.Errorf("CPU request cannot exceed CPU limit")
		}
	}

	if !resources.Requests.Memory().IsZero() && !resources.Limits.Memory().IsZero() {
		if resources.Requests.Memory().Cmp(*resources.Limits.Memory()) > 0 {
			return fmt.Errorf("memory request cannot exceed memory limit")
		}
	}

	return nil
}

// Validation helper functions
func validateHostname(hostname string) error {
	// Basic hostname validation
	if len(hostname) > 253 {
		return fmt.Errorf("hostname too long (max 253 characters)")
	}

	// Validate hostname format
	hostnameRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)
	if !hostnameRegex.MatchString(hostname) {
		return fmt.Errorf("invalid hostname format")
	}

	// Validate it's not an IP address
	if net.ParseIP(hostname) != nil {
		return fmt.Errorf("hostname cannot be an IP address")
	}

	return nil
}

func validateS3BucketName(bucket string) error {
	if len(bucket) < 3 || len(bucket) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters")
	}

	bucketRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`)
	if !bucketRegex.MatchString(bucket) {
		return fmt.Errorf("invalid bucket name format")
	}

	return nil
}

func validateGCSBucketName(bucket string) error {
	if len(bucket) < 3 || len(bucket) > 222 {
		return fmt.Errorf("bucket name must be between 3 and 222 characters")
	}

	bucketRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`)
	if !bucketRegex.MatchString(bucket) {
		return fmt.Errorf("invalid bucket name format")
	}

	return nil
}

func validateAWSRegion(region string) error {
	regionRegex := regexp.MustCompile(`^[a-z]{2}-[a-z]+-\d{1}$`)
	if !regionRegex.MatchString(region) {
		return fmt.Errorf("invalid AWS region format")
	}

	return nil
}

func validateAzureStorageAccount(account string) error {
	if len(account) < 3 || len(account) > 24 {
		return fmt.Errorf("storage account name must be between 3 and 24 characters")
	}

	accountRegex := regexp.MustCompile(`^[a-z0-9]{3,24}$`)
	if !accountRegex.MatchString(account) {
		return fmt.Errorf("invalid storage account name format")
	}

	return nil
}

// ValidateSecrets validates that referenced secrets exist (called from controller)
func (r *MLflowServer) ValidateSecrets(ctx context.Context, k8sClient client.Client, namespace string) error {
	// Validate backend database secrets
	switch r.Spec.Backend.Type {
	case BackendTypePostgreSQL:
		cfg := r.Spec.Backend.PostgreSQL
		if cfg != nil && cfg.CredentialsSecret != "" {
			if err := validateSecretExists(ctx, k8sClient, namespace, cfg.CredentialsSecret, []string{"username", "password"}); err != nil {
				return fmt.Errorf("PostgreSQL secret validation failed: %w", err)
			}
		}
	case BackendTypeMySQL:
		cfg := r.Spec.Backend.MySQL
		if cfg != nil && cfg.CredentialsSecret != "" {
			if err := validateSecretExists(ctx, k8sClient, namespace, cfg.CredentialsSecret, []string{"username", "password"}); err != nil {
				return fmt.Errorf("MySQL secret validation failed: %w", err)
			}
		}
	}

	// Validate artifact store secrets
	switch r.Spec.ArtifactStore.Type {
	case ArtifactStoreTypeS3:
		cfg := r.Spec.ArtifactStore.S3
		if cfg != nil && cfg.CredentialsSecret != "" {
			if err := validateSecretExists(ctx, k8sClient, namespace, cfg.CredentialsSecret, []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"}); err != nil {
				return fmt.Errorf("S3 secret validation failed: %w", err)
			}
		}
	case ArtifactStoreTypeGCS:
		cfg := r.Spec.ArtifactStore.GCS
		if cfg != nil && cfg.CredentialsSecret != "" {
			if err := validateSecretExists(ctx, k8sClient, namespace, cfg.CredentialsSecret, []string{"service-account.json"}); err != nil {
				return fmt.Errorf("GCS secret validation failed: %w", err)
			}
		}
	case ArtifactStoreTypeAzure:
		cfg := r.Spec.ArtifactStore.Azure
		if cfg != nil && cfg.CredentialsSecret != "" {
			if err := validateSecretExists(ctx, k8sClient, namespace, cfg.CredentialsSecret, []string{"account-name", "account-key"}); err != nil {
				return fmt.Errorf("Azure secret validation failed: %w", err)
			}
		}
	}

	return nil
}

func validateSecretExists(ctx context.Context, k8sClient client.Client, namespace, secretName string, requiredKeys []string) error {
	secret := &corev1.Secret{}
	err := k8sClient.Get(ctx, client.ObjectKey{Name: secretName, Namespace: namespace}, secret)
	if err != nil {
		return fmt.Errorf("secret %s not found in namespace %s: %w", secretName, namespace, err)
	}

	for _, key := range requiredKeys {
		if _, exists := secret.Data[key]; !exists {
			return fmt.Errorf("secret %s is missing required key: %s", secretName, key)
		}
	}

	return nil
}
