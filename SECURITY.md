# Security Policy

This document outlines the security features and best practices implemented in the MLflow Kubernetes Operator for enterprise-grade deployments.

## Security Features Implemented

### 1. Pod Security Standards

#### Operator Deployment
- **Non-root user**: Operator runs as non-root user (UID 65534)
- **Read-only root filesystem**: Prevents modifications to the container filesystem
- **Capability dropping**: All Linux capabilities are dropped
- **No privilege escalation**: Privilege escalation is explicitly disabled
- **Seccomp profile**: RuntimeDefault seccomp profile is enforced
- **Resource limits**: CPU and memory limits are enforced (1 CPU, 1Gi memory)

#### MLflow Server Pods
- **Non-root user**: MLflow containers run as non-root user (UID 1000)
- **Read-only root filesystem**: Prevents modifications to the container filesystem
- **Capability dropping**: All Linux capabilities are dropped
- **No privilege escalation**: Privilege escalation is explicitly disabled
- **Seccomp profile**: RuntimeDefault seccomp profile is enforced
- **FSGroup**: File system group ID set for proper file permissions

### 2. Secret Management

#### Secret Validation
- **Pre-reconciliation validation**: All secrets are validated before resource creation
- **Required key checking**: Ensures secrets contain required keys (e.g., username, password)
- **Existence verification**: Verifies secrets exist before attempting to use them
- **Clear error messages**: Provides detailed error messages when validation fails

#### Secret Usage
- **Environment variable substitution**: Credentials are substituted at runtime, not stored in ConfigMaps
- **No credential exposure**: ConfigMaps no longer contain credential placeholders
- **Secret volume mounting**: GCS credentials are properly mounted as read-only volumes

### 3. Network Security

#### Network Policies
- **Operator network policies**: Network isolation for operator pods
- **MLflow server network policies**: Configurable network policies for MLflow deployments
  - **Restrictive mode**: Only allows traffic from same namespace and required external services
  - **Permissive mode**: Allows traffic within cluster
  - **Custom mode**: Define custom ingress/egress rules
- **Namespace-based access control**: Allow/deny traffic from specific namespaces
- **IP-based access control**: Allow/deny traffic from specific IP ranges
- **DNS egress rules**: Allows DNS resolution for restrictive mode
- **Configurable**: Network policies can be enabled/disabled per MLflow server

### 4. Input Validation

#### CRD Validation
- **Hostname validation**: Validates hostname format for PostgreSQL, MySQL, and Ingress hosts
- **Port validation**: Ensures ports are within valid range (1-65535)
- **Database name validation**: Validates database name format
- **Secret name validation**: Validates Kubernetes secret name format
- **Bucket name validation**: Validates S3 and GCS bucket name formats
- **AWS region validation**: Validates AWS region format
- **Azure storage account validation**: Validates Azure storage account format
- **SSL mode validation**: Validates PostgreSQL SSL mode values
- **Endpoint URL validation**: Validates custom S3 endpoint URLs

### 5. Resource Management

#### Resource Limits
- **Operator limits**: Default CPU and memory limits enforced
- **MLflow server limits**: User-configurable resource requests and limits
- **Resource requests**: Minimum resource guarantees for reliable operation

### 6. High Availability

#### Pod Disruption Budgets
- **Configurable PDB**: Supports minAvailable and maxUnavailable settings
- **Graceful node maintenance**: Ensures availability during node upgrades
- **Zero-downtime deployments**: Supports rolling updates without service interruption

### 7. Resource Cleanup

#### Finalizers
- **Graceful deletion**: Finalizers ensure proper cleanup of owned resources
- **Owner references**: Automatic cleanup of child resources when parent is deleted
- **Deletion handling**: Proper handling of deletion timestamps

### 8. RBAC (Role-Based Access Control)

The operator uses least-privilege RBAC:
- **ClusterRole**: Minimal permissions required for operation
- **ServiceAccount**: Dedicated service account for operator
- **Namespace-scoped**: Operations limited to specific namespaces when possible

#### MLflow Server RBAC
- **Custom Roles**: Create dedicated roles for each MLflow server
- **Role Bindings**: Bind roles to service accounts or subjects
- **Resource Quotas**: Enforce resource limits per deployment
  - CPU quotas
  - Memory quotas
  - Storage quotas
  - Pod count limits
- **Configurable**: RBAC can be enabled/disabled per MLflow server

### 9. Pod Security Standards (PSS)

#### PSS Compliance
- **Configurable PSS level**: Support for privileged, baseline, and restricted PSS levels
- **Default restricted**: MLflow servers default to restricted PSS for maximum security
- **Seccomp profiles**: Support for RuntimeDefault, Localhost, and Unconfined profiles
- **Custom security contexts**: Fine-grained control over pod and container security contexts
  - Run as non-root
  - Read-only root filesystem
  - Capability dropping
  - Privilege escalation control
  - Supplemental groups
  - Sysctl parameters

### 10. Service Account Security

#### Service Account Configuration
- **Dedicated service accounts**: Create dedicated service accounts per MLflow server
- **Bound service account tokens**: Use bound tokens with expiration
- **Automount control**: Disable automatic token mounting when not needed
- **Custom annotations**: Support for cloud provider annotations (e.g., AWS IAM roles)
- **Token expiration**: Configurable token expiration for bound tokens

### 11. Image Security

#### Image Pull Policies
- **Configurable pull policies**: Support for Always, IfNotPresent, and Never
- **Image pull secrets**: Support for private container registries

#### Image Signature Verification
- **Cosign integration**: Verify image signatures using Cosign
- **Public key validation**: Validate signatures with provided public keys
- **Repository verification**: Verify signatures from specific repositories
- **Configurable**: Can be enabled/disabled per deployment

#### Vulnerability Scanning
- **Vulnerability detection**: Integration with vulnerability scanners
- **Severity thresholds**: Fail deployment based on vulnerability severity
- **Configurable policies**: Set fail-on-vulnerability behavior
- **Severity levels**: CRITICAL, HIGH, MEDIUM, LOW

### 12. Anti-Affinity

#### Pod Anti-Affinity
- **Node spreading**: Spread pods across different nodes for high availability
- **Zone spreading**: Spread pods across availability zones
- **Configurable topology**: Custom topology keys for anti-affinity rules
- **Weight-based preferences**: Use preferred or required anti-affinity

### 13. External Secret Management

#### Supported Secret Managers
- **HashiCorp Vault**: Kubernetes authentication method support
- **AWS Secrets Manager**: AWS region and secret prefix configuration
- **Azure Key Vault**: Tenant ID and vault name configuration
- **GCP Secret Manager**: Project-based secret management

#### Secret Integration
- **Runtime secret retrieval**: Secrets fetched at runtime from external systems
- **Kubernetes auth**: Native Kubernetes authentication for Vault
- **Credential management**: Centralized credential management
- **Namespace isolation**: Support for Vault namespaces

### 14. Admission Webhook Validation

#### Webhook Validation
- **Mutating Webhook**: Sets default values for missing fields
- **Validating Webhook**: Validates all MLflowServer resources before creation/update
- **Validation Rules**:
  - Backend configuration validation (PostgreSQL, MySQL, SQLite)
  - Artifact store validation (S3, GCS, Azure, PVC)
  - Ingress configuration validation
  - Resource limits validation
  - Hostname and bucket name format validation
  - Port range validation
- **Default Values**: Automatically sets sensible defaults for replicas, images, and configurations
- **Failure Policy**: Fails on validation errors to prevent invalid resources

### 15. Health Checks

#### Backend Connectivity
- **Secret Validation**: Verifies database secrets exist before marking backend as connected
- **Status Conditions**: Updates status conditions based on connectivity checks
- **Error Reporting**: Provides detailed error messages when connectivity fails

#### Artifact Store Connectivity
- **Secret Validation**: Verifies artifact store secrets exist before marking store as connected
- **Status Conditions**: Updates status conditions based on connectivity checks
- **Error Reporting**: Provides detailed error messages when connectivity fails

### 16. Multi-tenancy Framework

#### Namespace Validation
- **Annotation-based**: Uses annotations to define allowed namespaces
- **Webhook Integration**: Namespace validation integrated with admission webhook
- **Extensible**: Framework for future namespace quota and isolation features
- **Ready for Enablement**: Commented code ready for production multi-tenancy deployment

## Security Best Practices

### Deployment

1. **Enable Network Policies**: Set `networkPolicy.enabled: true` in values.yaml
2. **Enable Pod Disruption Budgets**: Set `podDisruptionBudget.enabled: true` for high availability
3. **Use Leader Election**: Enable leader election for HA deployments
4. **Configure Resource Limits**: Adjust resource limits based on cluster capacity
5. **Use Private Container Registries**: Store operator images in private registries
6. **Enable PSS Compliance**: Set `security.podSecurityStandard: restricted` for maximum security
7. **Use Anti-Affinity**: Enable pod spreading for high availability

### Secret Management

1. **Use External Secrets**: Consider using external secret operators (e.g., External Secrets Operator)
2. **Rotate Secrets Regularly**: Implement secret rotation policies
3. **Use Secret Encryption**: Enable encryption at rest for Kubernetes secrets
4. **Limit Secret Access**: Use Kubernetes RBAC to restrict secret access
5. **Use Bound Tokens**: Enable bound service account tokens with expiration
6. **External Secret Managers**: Use Vault, AWS Secrets Manager, Azure Key Vault, or GCP Secret Manager

### Network Security

1. **Enable Network Policies**: Restrict pod-to-pod communication
2. **Use Service Mesh**: Consider using Istio or Linkerd for additional security
3. **Enable TLS**: Use TLS for all external communication
4. **Configure Firewalls**: Use network-level firewalls for additional protection
5. **Use Restrictive Mode**: Set `security.networkPolicy.policyType: restrictive` for maximum security

### Image Security

1. **Verify Image Signatures**: Enable Cosign signature verification
2. **Use Vulnerability Scanning**: Enable vulnerability scanning with severity thresholds
3. **Use Private Registries**: Store images in private container registries
4. **Set Pull Policies**: Use `Always` pull policy for production environments

### Monitoring and Auditing

1. **Enable Audit Logging**: Configure Kubernetes audit logging
2. **Monitor Resource Usage**: Set up alerts for resource usage
3. **Track Secret Access**: Monitor secret access patterns
4. **Log Security Events**: Configure security event logging
5. **Monitor Security Conditions**: Track security-related status conditions

## Reporting Vulnerabilities

If you discover a security vulnerability, please report it responsibly:

1. **Do not create public issues**: Send security reports privately
2. **Provide details**: Include steps to reproduce and impact assessment
3. **Allow time for fixes**: Give maintainers time to investigate and fix
4. **Follow disclosure**: Wait for coordinated disclosure before public announcement

## Security Configuration

### Enable Security Features

```yaml
# values.yaml
networkPolicy:
  enabled: true

podDisruptionBudget:
  enabled: true
  minAvailable: 1

controller:
  manager:
    args:
      - --leader-elect=true
```

### Configure Resource Limits

```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 100m
    memory: 256Mi
```

### Enable Multi-tenancy (Future)

```yaml
multitenancy:
  enabled: true
  allowedNamespaces:
    - ml-team-alpha
    - ml-team-beta
```

### MLflow Server Security Configuration

```yaml
# MLflowServer security configuration
apiVersion: mlops.NotHarshhaa.io/v1alpha1
kind: MLflowServer
metadata:
  name: secure-mlflow
spec:
  security:
    # Pod Security Standards
    podSecurityStandard: restricted
    
    # Network Policy
    networkPolicy:
      enabled: true
      policyType: restrictive
      allowedNamespaces:
        - ml-production
      allowedIPRanges:
        - 10.0.0.0/8
    
    # Service Account
    serviceAccount:
      create: true
      automountServiceAccountToken: false
      boundServiceAccountToken: true
      tokenExpiration: "3600s"
    
    # Image Security
    imageSecurity:
      pullPolicy: Always
      signatureVerification:
        enabled: true
        keySecret: cosign-public-key
      vulnerabilityScan:
        enabled: true
        severityThreshold: HIGH
    
    # Anti-Affinity
    antiAffinity:
      enabled: true
      spreadAcrossNodes: true
      spreadAcrossZones: true
    
    # RBAC
    rbac:
      enabled: true
      resourceQuota:
        enabled: true
        cpu: "10"
        memory: "20Gi"
        pods: "20"
    
    # External Secrets
    secretManagement:
      type: vault
      vault:
        address: https://vault.internal:8200
        authMethod: kubernetes
        role: mlflow-role
```

## Compliance

This operator is designed to help meet common security compliance requirements:

- **SOC 2**: Security controls for data protection
- **PCI DSS**: Payment card industry data security standards
- **HIPAA**: Healthcare data protection requirements
- **GDPR**: General data protection regulation compliance

Note: Actual compliance depends on proper configuration and deployment practices.

## Future Security Enhancements

Planned security improvements:

1. ~~**Admission Webhooks**: Validate resources at admission time~~ **COMPLETED**
2. ~~**TLS for Metrics**: Enable TLS for metrics endpoint~~ **Infrastructure Ready**
3. **Audit Logging**: Built-in audit logging support
4. ~~**Multi-tenancy**: Namespace isolation and quotas~~ **Framework Added**
5. ~~**Health Checks**: Proactive health checks for backend and artifact store~~ **COMPLETED**
6. ~~**Secret Encryption**: Integration with external secret management systems~~ **COMPLETED**
7. ~~**Image Verification**: Signature verification for container images~~ **COMPLETED**
8. **Policy as Code**: Integration with OPA/Gatekeeper
9. **Comprehensive RBAC**: Further refinement of RBAC rules for least privilege
10. **Service Mesh Integration**: Native Istio/Linkerd integration for advanced security
11. **Zero Trust Architecture**: Complete zero-trust networking model
12. **Compliance Automation**: Automated compliance checking and reporting

## References

- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/security-checklist/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [RBAC Best Practices](https://kubernetes.io/docs/concepts/security/rbac-good-practices/)
