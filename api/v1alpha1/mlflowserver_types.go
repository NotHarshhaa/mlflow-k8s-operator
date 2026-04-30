package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}

// TrackingConfig defines the tracking server configuration
type TrackingConfig struct {
	// Replicas is the number of tracking server pods
	// +kubebuilder:default=1
	// +kubebuilder:minimum=1
	Replicas int32 `json:"replicas,omitempty"`

	// Resources defines CPU and memory requests and limits
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
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
	Host string `json:"host"`

	// Port is the PostgreSQL server port
	// +kubebuilder:default=5432
	Port int32 `json:"port,omitempty"`

	// Database is the database name
	// +kubebuilder:default=mlflow
	Database string `json:"database,omitempty"`

	// CredentialsSecret is the name of the secret containing database credentials
	// Required keys: username, password
	CredentialsSecret string `json:"credentialsSecret"`

	// SSLMode is the SSL mode for the connection
	// +kubebuilder:default=require
	SSLMode string `json:"sslMode,omitempty"`
}

// MySQLConfig defines MySQL connection configuration
type MySQLConfig struct {
	// Host is the MySQL server hostname
	Host string `json:"host"`

	// Port is the MySQL server port
	// +kubebuilder:default=3306
	Port int32 `json:"port,omitempty"`

	// Database is the database name
	// +kubebuilder:default=mlflow
	Database string `json:"database,omitempty"`

	// CredentialsSecret is the name of the secret containing database credentials
	// Required keys: username, password
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
	Bucket string `json:"bucket"`

	// Region is the AWS region
	Region string `json:"region"`

	// CredentialsSecret is the name of the secret containing AWS credentials
	// Required keys: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
	CredentialsSecret string `json:"credentialsSecret"`

	// PathPrefix is an optional prefix for artifact paths
	PathPrefix string `json:"pathPrefix,omitempty"`

	// EndpointURL is an optional custom S3 endpoint (for MinIO or other S3-compatible storage)
	EndpointURL string `json:"endpointURL,omitempty"`
}

// GCSConfig defines GCS artifact store configuration
type GCSConfig struct {
	// Bucket is the GCS bucket name
	Bucket string `json:"bucket"`

	// CredentialsSecret is the name of the secret containing GCP service account key
	// Required key: service-account.json
	CredentialsSecret string `json:"credentialsSecret"`
}

// AzureConfig defines Azure Blob storage configuration
type AzureConfig struct {
	// StorageAccount is the Azure storage account name
	StorageAccount string `json:"storageAccount"`

	// Container is the blob container name
	Container string `json:"container"`

	// CredentialsSecret is the name of the secret containing Azure storage credentials
	// Required keys: account-name, account-key
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
	Host string `json:"host,omitempty"`

	// TLS configuration
	TLS *TLSConfig `json:"tls,omitempty"`

	// IngressClassName is the ingress class name
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
