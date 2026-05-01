package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// MLflowServerSpec defines the desired state of MLflowServer
type MLflowServerSpec struct {
	// Version is the MLflow image version to deploy
	// +kubebuilder:default="2.11.0"
	Version string `json:"version"`

	// Tracking configuration for the MLflow tracking server
	Tracking TrackingConfig `json:"tracking"`

	// Backend configuration for the tracking database
	Backend BackendConfig `json:"backend"`

	// ArtifactStore configuration for artifact storage
	ArtifactStore ArtifactStoreConfig `json:"artifactStore"`

	// Ingress configuration for external access
	Ingress IngressConfig `json:"ingress,omitempty"`

	// Autoscaling configuration for the MLflow server
	Autoscaling *AutoscalingConfig `json:"autoscaling,omitempty"`

	// Pod scheduling configuration
	Scheduling *SchedulingConfig `json:"scheduling,omitempty"`

	// Pod disruption budget configuration
	PodDisruptionBudget *PodDisruptionBudgetConfig `json:"podDisruptionBudget,omitempty"`

	// ServiceMonitor configuration for Prometheus Operator
	ServiceMonitor *ServiceMonitorConfig `json:"serviceMonitor,omitempty"`

	// Database migration configuration
	Migration *MigrationConfig `json:"migration,omitempty"`

	// Security configuration for enhanced security features
	Security *SecurityConfig `json:"security,omitempty"`
}

// TrackingConfig defines the tracking server configuration
type TrackingConfig struct {
	// Replicas is the number of tracking server pods
	// +kubebuilder:default=1
	// +kubebuilder:minimum=1
	Replicas int32 `json:"replicas,omitempty"`

	// Resources defines CPU and memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// AdditionalArgs are additional arguments to pass to the MLflow server
	AdditionalArgs []string `json:"additionalArgs,omitempty"`

	// Probes configuration for health checks
	Probes *ProbesConfig `json:"probes,omitempty"`

	// Lifecycle hooks for the container
	Lifecycle *corev1.Lifecycle `json:"lifecycle,omitempty"`

	// InitContainers to run before the MLflow container
	InitContainers []corev1.Container `json:"initContainers,omitempty"`

	// SidecarContainers to run alongside the MLflow container
	SidecarContainers []corev1.Container `json:"sidecarContainers,omitempty"`

	// PodAnnotations are annotations to add to the pod
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// PodLabels are labels to add to the pod
	PodLabels map[string]string `json:"podLabels,omitempty"`
}

// BackendConfig defines the tracking database backend configuration
type BackendConfig struct {
	// Type is the database backend type: postgresql, mysql, or sqlite
	Type BackendType `json:"type"`

	// PostgreSQL configuration when type is postgresql
	PostgreSQL *PostgreSQLConfig `json:"postgresql,omitempty"`

	// MySQL configuration when type is mysql
	MySQL *MySQLConfig `json:"mysql,omitempty"`

	// SQLite configuration when type is sqlite
	SQLite *SQLiteConfig `json:"sqlite,omitempty"`
}

// BackendType is the type of backend database
// +kubebuilder:validation:Enum=postgresql;mysql;sqlite
type BackendType string

const (
	BackendTypePostgreSQL BackendType = "postgresql"
	BackendTypeMySQL      BackendType = "mysql"
	BackendTypeSQLite     BackendType = "sqlite"
)

// PostgreSQLConfig defines PostgreSQL connection configuration
type PostgreSQLConfig struct {
	// Host is the PostgreSQL server hostname
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`
	Host string `json:"host"`

	// Port is the PostgreSQL server port
	// +kubebuilder:default=5432
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`

	// Database is the database name
	// +kubebuilder:default=mlflow
	// +kubebuilder:validation:Pattern=`^[a-zA-Z][a-zA-Z0-9_]*$`
	Database string `json:"database,omitempty"`

	// CredentialsSecret is the name of the secret containing database credentials
	// Required keys: username, password
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	CredentialsSecret string `json:"credentialsSecret"`

	// SSLMode is the SSL mode for the connection
	// +kubebuilder:default=require
	// +kubebuilder:validation:Enum=disable;allow;prefer;require;verify-ca;verify-full
	SSLMode string `json:"sslMode,omitempty"`
}

// MySQLConfig defines MySQL connection configuration
type MySQLConfig struct {
	// Host is the MySQL server hostname
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`
	Host string `json:"host"`

	// Port is the MySQL server port
	// +kubebuilder:default=3306
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`

	// Database is the database name
	// +kubebuilder:default=mlflow
	// +kubebuilder:validation:Pattern=`^[a-zA-Z][a-zA-Z0-9_]*$`
	Database string `json:"database,omitempty"`

	// CredentialsSecret is the name of the secret containing database credentials
	// Required keys: username, password
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	CredentialsSecret string `json:"credentialsSecret"`
}

// SQLiteConfig defines SQLite configuration
type SQLiteConfig struct {
	// PVC is the PVC configuration for SQLite storage
	PVC *PVCConfig `json:"pvc,omitempty"`
}

// ArtifactStoreConfig defines the artifact storage backend configuration
type ArtifactStoreConfig struct {
	// Type is the artifact store type: s3, gcs, azure, or pvc
	Type ArtifactStoreType `json:"type"`

	// S3 configuration when type is s3
	S3 *S3Config `json:"s3,omitempty"`

	// GCS configuration when type is gcs
	GCS *GCSConfig `json:"gcs,omitempty"`

	// Azure configuration when type is azure
	Azure *AzureConfig `json:"azure,omitempty"`

	// PVC configuration when type is pvc
	PVC *PVCConfig `json:"pvc,omitempty"`
}

// ArtifactStoreType is the type of artifact store
// +kubebuilder:validation:Enum=s3;gcs;azure;pvc
type ArtifactStoreType string

const (
	ArtifactStoreTypeS3    ArtifactStoreType = "s3"
	ArtifactStoreTypeGCS   ArtifactStoreType = "gcs"
	ArtifactStoreTypeAzure ArtifactStoreType = "azure"
	ArtifactStoreTypePVC   ArtifactStoreType = "pvc"
)

// S3Config defines S3 artifact store configuration
type S3Config struct {
	// Bucket is the S3 bucket name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`
	Bucket string `json:"bucket"`

	// Region is the AWS region
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z]{2}-[a-z]+-\d{1}$`
	Region string `json:"region"`

	// CredentialsSecret is the name of the secret containing AWS credentials
	// Required keys: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	CredentialsSecret string `json:"credentialsSecret"`

	// PathPrefix is an optional prefix for artifact paths
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_\-/]*$`
	PathPrefix string `json:"pathPrefix,omitempty"`

	// EndpointURL is an optional custom S3 endpoint (for MinIO or other S3-compatible storage)
	// +kubebuilder:validation:Pattern=`^https?://[a-zA-Z0-9\.\-]+(:[0-9]+)?(/.*)?$`
	EndpointURL string `json:"endpointURL,omitempty"`
}

// GCSConfig defines GCS artifact store configuration
type GCSConfig struct {
	// Bucket is the GCS bucket name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`
	Bucket string `json:"bucket"`

	// CredentialsSecret is the name of the secret containing GCP service account key
	// Required key: service-account.json
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	CredentialsSecret string `json:"credentialsSecret"`
}

// AzureConfig defines Azure Blob storage configuration
type AzureConfig struct {
	// StorageAccount is the Azure storage account name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]{3,24}$`
	StorageAccount string `json:"storageAccount"`

	// Container is the blob container name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9][a-z0-9\-]*$`
	Container string `json:"container"`

	// CredentialsSecret is the name of the secret containing Azure storage credentials
	// Required keys: account-name, account-key
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	CredentialsSecret string `json:"credentialsSecret"`
}

// PVCConfig defines PVC configuration
type PVCConfig struct {
	// StorageClass is the storage class for the PVC
	StorageClass string `json:"storageClass,omitempty"`

	// Size is the size of the PVC
	// +kubebuilder:default="100Gi"
	Size string `json:"size,omitempty"`

	// AccessMode is the access mode for the PVC
	// +kubebuilder:default="ReadWriteOnce"
	AccessMode string `json:"accessMode,omitempty"`
}

// IngressConfig defines ingress configuration
type IngressConfig struct {
	// Enabled determines whether to create an Ingress resource
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Host is the hostname for the MLflow UI
	// +kubebuilder:validation:Pattern=`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`
	Host string `json:"host,omitempty"`

	// TLS configuration
	TLS *TLSConfig `json:"tls,omitempty"`

	// IngressClassName is the ingress class name
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	IngressClassName string `json:"ingressClassName,omitempty"`

	// Annotations are additional annotations for the Ingress resource
	Annotations map[string]string `json:"annotations,omitempty"`
}

// TLSConfig defines TLS configuration
type TLSConfig struct {
	// Enabled determines whether to enable TLS
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Issuer is the cert-manager issuer name
	Issuer string `json:"issuer,omitempty"`

	// SecretName is the name of the TLS secret (if not using cert-manager)
	SecretName string `json:"secretName,omitempty"`
}

// AutoscalingConfig defines Horizontal Pod Autoscaler configuration
type AutoscalingConfig struct {
	// Enabled determines whether to enable HPA
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// MinReplicas is the minimum number of replicas
	// +kubebuilder:default=1
	// +kubebuilder:minimum=1
	MinReplicas int32 `json:"minReplicas,omitempty"`

	// MaxReplicas is the maximum number of replicas
	// +kubebuilder:default=10
	// +kubebuilder:minimum=1
	MaxReplicas int32 `json:"maxReplicas,omitempty"`

	// TargetCPUUtilizationPercentage is the target CPU utilization percentage
	// +kubebuilder:default=80
	// +kubebuilder:minimum=1
	// +kubebuilder:maximum=100
	TargetCPUUtilizationPercentage int32 `json:"targetCPUUtilizationPercentage,omitempty"`

	// TargetMemoryUtilizationPercentage is the target memory utilization percentage
	// +kubebuilder:minimum=1
	// +kubebuilder:maximum=100
	TargetMemoryUtilizationPercentage int32 `json:"targetMemoryUtilizationPercentage,omitempty"`
}

// SchedulingConfig defines pod scheduling configuration
type SchedulingConfig struct {
	// NodeSelector is the node selector for pod assignment
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations are the pod's tolerations
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity is the pod's affinity rules
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// PriorityClassName is the priority class name
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// TopologySpreadConstraints are the pod's topology spread constraints
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
}

// PodDisruptionBudgetConfig defines Pod Disruption Budget configuration
type PodDisruptionBudgetConfig struct {
	// Enabled determines whether to create a PDB
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// MinAvailable is the minimum number of pods that must be available
	MinAvailable *int32 `json:"minAvailable,omitempty"`

	// MaxUnavailable is the maximum number of pods that can be unavailable
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`
}

// ServiceMonitorConfig defines ServiceMonitor configuration for Prometheus Operator
type ServiceMonitorConfig struct {
	// Enabled determines whether to create a ServiceMonitor
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Interval is the scrape interval
	// +kubebuilder:default="30s"
	Interval string `json:"interval,omitempty"`

	// ScrapeTimeout is the scrape timeout
	// +kubebuilder:default="10s"
	ScrapeTimeout string `json:"scrapeTimeout,omitempty"`

	// Labels are additional labels for the ServiceMonitor
	Labels map[string]string `json:"labels,omitempty"`
}

// ProbesConfig defines health check probes configuration
type ProbesConfig struct {
	// LivenessProbe configuration for the container
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`

	// ReadinessProbe configuration for the container
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`

	// StartupProbe configuration for the container
	StartupProbe *corev1.Probe `json:"startupProbe,omitempty"`
}

// MigrationConfig defines database migration configuration
type MigrationConfig struct {
	// Enabled determines whether to enable automatic database migrations on version upgrade
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// JobAnnotations are annotations to add to the migration job
	JobAnnotations map[string]string `json:"jobAnnotations,omitempty"`

	// BackoffLimit is the number of retries before marking the job as failed
	// +kubebuilder:default=6
	BackoffLimit int32 `json:"backoffLimit,omitempty"`

	// ActiveDeadlineSeconds is the duration the job may be active before the system tries to terminate it
	ActiveDeadlineSeconds int64 `json:"activeDeadlineSeconds,omitempty"`
}

// SecurityConfig defines security configuration for the MLflow server
type SecurityConfig struct {
	// PodSecurityContext defines pod-level security attributes
	PodSecurityContext *PodSecurityContextConfig `json:"podSecurityContext,omitempty"`

	// ContainerSecurityContext defines container-level security attributes
	ContainerSecurityContext *ContainerSecurityContextConfig `json:"containerSecurityContext,omitempty"`

	// NetworkPolicy defines network isolation rules
	NetworkPolicy *NetworkPolicyConfig `json:"networkPolicy,omitempty"`

	// PodSecurityStandard defines the Pod Security Standard level
	// +kubebuilder:default=restricted
	// +kubebuilder:validation:Enum=privileged;baseline;restricted
	PodSecurityStandard string `json:"podSecurityStandard,omitempty"`

	// ServiceAccount defines service account security settings
	ServiceAccount *ServiceAccountConfig `json:"serviceAccount,omitempty"`

	// ImageSecurity defines image security settings
	ImageSecurity *ImageSecurityConfig `json:"imageSecurity,omitempty"`

	// SecretManagement defines external secret management configuration
	SecretManagement *SecretManagementConfig `json:"secretManagement,omitempty"`

	// AntiAffinity defines security-focused anti-affinity rules
	AntiAffinity *AntiAffinityConfig `json:"antiAffinity,omitempty"`

	// RBAC defines RBAC security settings
	RBAC *RBACSecurityConfig `json:"rbac,omitempty"`
}

// PodSecurityContextConfig defines pod-level security context configuration
type PodSecurityContextConfig struct {
	// RunAsNonRoot indicates that the container must run as a non-root user
	// +kubebuilder:default=true
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`

	// RunAsUser defines the UID to run the entrypoint of the container process
	RunAsUser *int64 `json:"runAsUser,omitempty"`

	// RunAsGroup defines the GID to run the entrypoint of the container process
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`

	// FSGroup defines the FSGroup to apply to the container
	FSGroup *int64 `json:"fsGroup,omitempty"`

	// SeccompProfile defines the seccomp profile to use
	SeccompProfile *SeccompProfileConfig `json:"seccompProfile,omitempty"`

	// SupplementalGroups defines the supplemental groups for the container
	SupplementalGroups []int64 `json:"supplementalGroups,omitempty"`

	// Sysctls defines sysctl parameters for the container
	Sysctls []corev1.Sysctl `json:"sysctls,omitempty"`
}

// SeccompProfileConfig defines seccomp profile configuration
type SeccompProfileConfig struct {
	// Type indicates which kind of seccomp profile will be applied
	// +kubebuilder:validation:Enum=Localhost;RuntimeDefault;Unconfined
	Type string `json:"type,omitempty"`

	// LocalhostProfile indicates a profile defined in a file on the node
	LocalhostProfile string `json:"localhostProfile,omitempty"`
}

// ContainerSecurityContextConfig defines container-level security context configuration
type ContainerSecurityContextConfig struct {
	// AllowPrivilegeEscalation controls whether a process can gain more privileges
	// +kubebuilder:default=false
	AllowPrivilegeEscalation *bool `json:"allowPrivilegeEscalation,omitempty"`

	// ReadOnlyRootFilesystem mounts the container's root filesystem as read-only
	// +kubebuilder:default=true
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`

	// RunAsNonRoot indicates that the container must run as a non-root user
	// +kubebuilder:default=true
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`

	// RunAsUser defines the UID to run the entrypoint of the container process
	RunAsUser *int64 `json:"runAsUser,omitempty"`

	// Capabilities defines Linux capabilities
	Capabilities *CapabilitiesConfig `json:"capabilities,omitempty"`

	// Privileged determines if the container runs in privileged mode
	// +kubebuilder:default=false
	Privileged *bool `json:"privileged,omitempty"`

	// SeccompProfile defines the seccomp profile to use
	SeccompProfile *SeccompProfileConfig `json:"seccompProfile,omitempty"`

	// NoNewPrivileges prevents the process from gaining additional privileges
	// +kubebuilder:default=true
	NoNewPrivileges *bool `json:"noNewPrivileges,omitempty"`
}

// CapabilitiesConfig defines Linux capabilities configuration
type CapabilitiesConfig struct {
	// Add adds Linux capabilities
	Add []corev1.Capability `json:"add,omitempty"`

	// Drop drops Linux capabilities
	// +kubebuilder:default=["ALL"]
	Drop []corev1.Capability `json:"drop,omitempty"`
}

// NetworkPolicyConfig defines network policy configuration
type NetworkPolicyConfig struct {
	// Enabled determines whether to create a NetworkPolicy
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// PolicyType defines the type of network policy
	// +kubebuilder:default=restrictive
	// +kubebuilder:validation:Enum=permissive;restrictive;custom
	PolicyType string `json:"policyType,omitempty"`

	// IngressRules defines custom ingress rules when PolicyType is custom
	IngressRules []NetworkPolicyIngressRule `json:"ingressRules,omitempty"`

	// EgressRules defines custom egress rules when PolicyType is custom
	EgressRules []NetworkPolicyEgressRule `json:"egressRules,omitempty"`

	// AllowedNamespaces defines namespaces allowed to communicate with the MLflow server
	AllowedNamespaces []string `json:"allowedNamespaces,omitempty"`

	// AllowedIPRanges defines IP ranges allowed to communicate with the MLflow server
	AllowedIPRanges []string `json:"allowedIPRanges,omitempty"`
}

// NetworkPolicyIngressRule defines an ingress rule for network policy
type NetworkPolicyIngressRule struct {
	// Ports defines the ports for this rule
	Ports []NetworkPolicyPort `json:"ports,omitempty"`

	// From defines the sources for this rule
	From []NetworkPolicyPeer `json:"from,omitempty"`
}

// NetworkPolicyEgressRule defines an egress rule for network policy
type NetworkPolicyEgressRule struct {
	// Ports defines the ports for this rule
	Ports []NetworkPolicyPort `json:"ports,omitempty"`

	// To defines the destinations for this rule
	To []NetworkPolicyPeer `json:"to,omitempty"`
}

// NetworkPolicyPort defines a port for network policy
type NetworkPolicyPort struct {
	// Protocol defines the protocol (TCP, UDP, or SCTP)
	// +kubebuilder:validation:Enum=TCP;UDP;SCTP
	Protocol string `json:"protocol,omitempty"`

	// Port defines the port number or name
	Port *intstr.IntOrString `json:"port,omitempty"`

	// EndPort defines the end port for a range
	EndPort int32 `json:"endPort,omitempty"`
}

// NetworkPolicyPeer defines a peer for network policy
type NetworkPolicyPeer struct {
	// PodSelector defines a pod selector
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`

	// NamespaceSelector defines a namespace selector
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// IPBlock defines an IP block
	IPBlock *IPBlock `json:"ipBlock,omitempty"`
}

// IPBlock defines an IP block for network policy
type IPBlock struct {
	// CIDR defines the CIDR block
	CIDR string `json:"cidr,omitempty"`

	// Except defines CIDR ranges to exclude
	Except []string `json:"except,omitempty"`
}

// ServiceAccountConfig defines service account security configuration
type ServiceAccountConfig struct {
	// Create determines whether to create a dedicated service account
	// +kubebuilder:default=true
	Create bool `json:"create,omitempty"`

	// Annotations are annotations to add to the service account
	Annotations map[string]string `json:"annotations,omitempty"`

	// AutomountServiceAccountToken determines whether to mount the service account token
	// +kubebuilder:default=false
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty"`

	// BoundServiceAccountToken determines whether to use bound service account tokens
	// +kubebuilder:default=true
	BoundServiceAccountToken bool `json:"boundServiceAccountToken,omitempty"`

	// TokenExpiration defines the expiration duration for bound tokens
	// +kubebuilder:default="3600s"
	TokenExpiration string `json:"tokenExpiration,omitempty"`
}

// ImageSecurityConfig defines image security configuration
type ImageSecurityConfig struct {
	// PullPolicy defines the image pull policy
	// +kubebuilder:default=Always
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	PullPolicy string `json:"pullPolicy,omitempty"`

	// ImagePullSecrets defines secrets for pulling images
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// SignatureVerification defines image signature verification settings
	SignatureVerification *SignatureVerificationConfig `json:"signatureVerification,omitempty"`

	// VulnerabilityScan defines vulnerability scanning settings
	VulnerabilityScan *VulnerabilityScanConfig `json:"vulnerabilityScan,omitempty"`
}

// SignatureVerificationConfig defines image signature verification configuration
type SignatureVerificationConfig struct {
	// Enabled determines whether to verify image signatures
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// KeySecret is the secret containing the public key for verification
	KeySecret string `json:"keySecret,omitempty"`

	// CosignRepository is the cosign repository for signature verification
	CosignRepository string `json:"cosignRepository,omitempty"`
}

// VulnerabilityScanConfig defines vulnerability scanning configuration
type VulnerabilityScanConfig struct {
	// Enabled determines whether to enable vulnerability scanning
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// FailOnVulnerabilities determines whether to fail on found vulnerabilities
	// +kubebuilder:default=true
	FailOnVulnerabilities bool `json:"failOnVulnerabilities,omitempty"`

	// SeverityThreshold defines the minimum severity to fail on
	// +kubebuilder:default=HIGH
	// +kubebuilder:validation:Enum=CRITICAL;HIGH;MEDIUM;LOW
	SeverityThreshold string `json:"severityThreshold,omitempty"`
}

// SecretManagementConfig defines external secret management configuration
type SecretManagementConfig struct {
	// Type defines the external secret management type
	// +kubebuilder:validation:Enum=vault;aws-secrets-manager;azure-key-vault;gcp-secret-manager
	Type string `json:"type,omitempty"`

	// Vault configuration when type is vault
	Vault *VaultConfig `json:"vault,omitempty"`

	// AWSSecretsManager configuration when type is aws-secrets-manager
	AWSSecretsManager *AWSSecretsManagerConfig `json:"awsSecretsManager,omitempty"`

	// AzureKeyVault configuration when type is azure-key-vault
	AzureKeyVault *AzureKeyVaultConfig `json:"azureKeyVault,omitempty"`

	// GCPSecretManager configuration when type is gcp-secret-manager
	GCPSecretManager *GCPSecretManagerConfig `json:"gcpSecretManager,omitempty"`
}

// VaultConfig defines HashiCorp Vault configuration
type VaultConfig struct {
	// Address is the Vault server address
	Address string `json:"address,omitempty"`

	// AuthMethod is the authentication method
	// +kubebuilder:default=kubernetes
	// +kubebuilder:validation:Enum=kubernetes;token;approle
	AuthMethod string `json:"authMethod,omitempty"`

	// Role is the Vault role for Kubernetes auth
	Role string `json:"role,omitempty"`

	// SecretPath is the path to the secret in Vault
	SecretPath string `json:"secretPath,omitempty"`

	// Namespace is the Vault namespace (for Vault Enterprise)
	Namespace string `json:"namespace,omitempty"`
}

// AWSSecretsManagerConfig defines AWS Secrets Manager configuration
type AWSSecretsManagerConfig struct {
	// Region is the AWS region
	Region string `json:"region,omitempty"`

	// SecretPrefix is the prefix for secret names
	SecretPrefix string `json:"secretPrefix,omitempty"`

	// CredentialsSecret is the secret containing AWS credentials
	CredentialsSecret string `json:"credentialsSecret,omitempty"`
}

// AzureKeyVaultConfig defines Azure Key Vault configuration
type AzureKeyVaultConfig struct {
	// VaultName is the Azure Key Vault name
	VaultName string `json:"vaultName,omitempty"`

	// TenantID is the Azure tenant ID
	TenantID string `json:"tenantId,omitempty"`

	// CredentialsSecret is the secret containing Azure credentials
	CredentialsSecret string `json:"credentialsSecret,omitempty"`
}

// GCPSecretManagerConfig defines GCP Secret Manager configuration
type GCPSecretManagerConfig struct {
	// Project is the GCP project ID
	Project string `json:"project,omitempty"`

	// SecretPrefix is the prefix for secret names
	SecretPrefix string `json:"secretPrefix,omitempty"`

	// CredentialsSecret is the secret containing GCP credentials
	CredentialsSecret string `json:"credentialsSecret,omitempty"`
}

// AntiAffinityConfig defines security-focused anti-affinity configuration
type AntiAffinityConfig struct {
	// Enabled determines whether to enable anti-affinity
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// TopologyKey defines the topology key for anti-affinity
	// +kubebuilder:default=kubernetes.io/hostname
	TopologyKey string `json:"topologyKey,omitempty"`

	// SpreadAcrossNodes determines if pods should spread across nodes
	// +kubebuilder:default=true
	SpreadAcrossNodes bool `json:"spreadAcrossNodes,omitempty"`

	// SpreadAcrossZones determines if pods should spread across zones
	SpreadAcrossZones bool `json:"spreadAcrossZones,omitempty"`
}

// RBACSecurityConfig defines RBAC security configuration
type RBACSecurityConfig struct {
	// Enabled determines whether to create custom RBAC resources
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// RoleName is the name of the custom role
	RoleName string `json:"roleName,omitempty"`

	// RoleBindingName is the name of the custom role binding
	RoleBindingName string `json:"roleBindingName,omitempty"`

	// Subjects defines the subjects for the role binding
	Subjects []rbacv1.Subject `json:"subjects,omitempty"`

	// ResourceQuota defines resource quota limits
	ResourceQuota *ResourceQuotaConfig `json:"resourceQuota,omitempty"`
}

// ResourceQuotaConfig defines resource quota configuration
type ResourceQuotaConfig struct {
	// Enabled determines whether to create a resource quota
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// CPU defines CPU limits
	CPU string `json:"cpu,omitempty"`

	// Memory defines memory limits
	Memory string `json:"memory,omitempty"`

	// Storage defines storage limits
	Storage string `json:"storage,omitempty"`

	// Pods defines pod count limits
	Pods string `json:"pods,omitempty"`
}

// MLflowServerStatus defines the observed state of MLflowServer
type MLflowServerStatus struct {
	// Conditions represents the latest available observations of the MLflowServer's current state
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Ready indicates whether the MLflow server is ready
	Ready bool `json:"ready,omitempty"`

	// ArtifactStoreConnected indicates whether the artifact store is reachable
	ArtifactStoreConnected bool `json:"artifactStoreConnected,omitempty"`

	// BackendConnected indicates whether the backend database is reachable
	BackendConnected bool `json:"backendConnected,omitempty"`

	// ObservedGeneration is the most recent generation observed by the controller
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// MLflowServer is the Schema for the mlflowservers API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mlflow
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Artifact Store",type=string,JSONPath=`.spec.artifactStore.type`
// +kubebuilder:printcolumn:name="Backend",type=string,JSONPath=`.spec.backend.type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type MLflowServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MLflowServerSpec   `json:"spec,omitempty"`
	Status MLflowServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MLflowServerList contains a list of MLflowServer
type MLflowServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MLflowServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MLflowServer{}, &MLflowServerList{})
}

const (
	// ConditionReady indicates the MLflow server is ready
	ConditionReady = "Ready"

	// ConditionArtifactStoreConnected indicates the artifact store is connected
	ConditionArtifactStoreConnected = "ArtifactStoreConnected"

	// ConditionBackendConnected indicates the backend is connected
	ConditionBackendConnected = "BackendConnected"

	// ConditionUpgrading indicates a version upgrade is in progress
	ConditionUpgrading = "Upgrading"

	// ConditionMigrationInProgress indicates a database migration is in progress
	ConditionMigrationInProgress = "MigrationInProgress"

	// MLflowServerFinalizer is the finalizer for MLflowServer resources
	MLflowServerFinalizer = "mlflowservers.mlops.NotHarshhaa.io/finalizer"

	// AllowedNamespacesAnnotation is the annotation key for allowed namespaces
	AllowedNamespacesAnnotation = "mlflow-operator.mlops.NotHarshhaa.io/allowed-namespaces"
)

// SetCondition sets a condition on the MLflowServer status
func (m *MLflowServer) SetCondition(condition metav1.Condition) {
	conditions := m.Status.Conditions
	for i, c := range conditions {
		if c.Type == condition.Type {
			if c.Status != condition.Status || c.Reason != condition.Reason {
				conditions[i] = condition
				m.Status.Conditions = conditions
				return
			}
			return
		}
	}
	m.Status.Conditions = append(conditions, condition)
}

// GetCondition returns the condition with the given type
func (m *MLflowServer) GetCondition(conditionType string) *metav1.Condition {
	for _, c := range m.Status.Conditions {
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}
