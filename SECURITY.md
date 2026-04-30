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
- **Default deny**: Network isolation with explicit allow rules
- **Ingress control**: Only allows traffic on required ports (metrics, health probes, webhook)
- **Egress control**: Only allows DNS resolution and API server communication
- **Configurable**: Additional egress rules can be configured via values.yaml

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

### 9. Admission Webhook Validation

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

### 10. Health Checks

#### Backend Connectivity
- **Secret Validation**: Verifies database secrets exist before marking backend as connected
- **Status Conditions**: Updates status conditions based on connectivity checks
- **Error Reporting**: Provides detailed error messages when connectivity fails

#### Artifact Store Connectivity
- **Secret Validation**: Verifies artifact store secrets exist before marking store as connected
- **Status Conditions**: Updates status conditions based on connectivity checks
- **Error Reporting**: Provides detailed error messages when connectivity fails

### 11. Multi-tenancy Framework

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

### Secret Management

1. **Use External Secrets**: Consider using external secret operators (e.g., External Secrets Operator)
2. **Rotate Secrets Regularly**: Implement secret rotation policies
3. **Use Secret Encryption**: Enable encryption at rest for Kubernetes secrets
4. **Limit Secret Access**: Use Kubernetes RBAC to restrict secret access

### Network Security

1. **Enable Network Policies**: Restrict pod-to-pod communication
2. **Use Service Mesh**: Consider using Istio or Linkerd for additional security
3. **Enable TLS**: Use TLS for all external communication
4. **Configure Firewalls**: Use network-level firewalls for additional protection

### Monitoring and Auditing

1. **Enable Audit Logging**: Configure Kubernetes audit logging
2. **Monitor Resource Usage**: Set up alerts for resource usage
3. **Track Secret Access**: Monitor secret access patterns
4. **Log Security Events**: Configure security event logging

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
6. **Secret Encryption**: Integration with external secret management systems
7. **Image Verification**: Signature verification for container images
8. **Policy as Code**: Integration with OPA/Gatekeeper
9. **Comprehensive RBAC**: Further refinement of RBAC rules for least privilege

## References

- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/security-checklist/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [RBAC Best Practices](https://kubernetes.io/docs/concepts/security/rbac-good-practices/)
