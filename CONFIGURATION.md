# Configuration Guide

This guide provides detailed configuration options for deploying and managing MLflow servers using the mlflow-k8s-operator.

## Table of Contents

- [Quickstart Example](#quickstart-example)
- [CRD Reference](#crd-reference)
- [Multi-tenant Setup](#multi-tenant-setup)
- [Artifact Store Backends](#artifact-store-backends)
- [Upgrading MLflow Version](#upgrading-mlflow-version)
- [Advanced Features](#advanced-features)
- [Security Features](#security-features)

---

## Quickstart Example

### Provision your first MLflow server

```yaml
# mlflow-server.yaml
apiVersion: mlops.NotHarshhaa.io/v1alpha1
kind: MLflowServer
metadata:
  name: team-alpha-mlflow
  namespace: ml-team-alpha
spec:
  version: "2.11.0"

  tracking:
    replicas: 2
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      limits:
        cpu: "1"
        memory: "1Gi"

  backend:
    type: postgresql
    postgresql:
      host: postgres.ml-team-alpha.svc.cluster.local
      database: mlflow
      credentialsSecret: mlflow-db-credentials

  artifactStore:
    type: s3
    s3:
      bucket: my-mlflow-artifacts
      region: us-east-1
      credentialsSecret: aws-artifact-credentials

  ingress:
    enabled: true
    host: mlflow.team-alpha.internal
    tls:
      enabled: true
      issuer: letsencrypt-prod
```

```bash
kubectl apply -f mlflow-server.yaml
```

---

## CRD Reference

### `MLflowServer` spec

| Field | Type | Required | Description |
|---|---|---|---|
| `version` | string | ✓ | MLflow image version to deploy |
| `tracking.replicas` | int | | Number of tracking server pods (default: 1) |
| `tracking.resources` | ResourceRequirements | | CPU/memory requests and limits |
| `tracking.additionalArgs` | []string | | Additional MLflow server arguments |
| `tracking.probes` | ProbesConfig | | Custom health check probes |
| `tracking.lifecycle` | Lifecycle | | Container lifecycle hooks |
| `tracking.initContainers` | []Container | | Init containers |
| `tracking.sidecarContainers` | []Container | | Sidecar containers |
| `tracking.podAnnotations` | map[string]string | | Pod annotations |
| `tracking.podLabels` | map[string]string | | Pod labels |
| `backend.type` | enum | ✓ | `postgresql`, `mysql`, `sqlite` |
| `backend.postgresql` | PostgreSQLConfig | | Connection config for PostgreSQL |
| `artifactStore.type` | enum | ✓ | `s3`, `gcs`, `azure`, `pvc` |
| `artifactStore.s3` | S3Config | | S3 bucket and credentials |
| `ingress.enabled` | bool | | Create Ingress resource (default: false) |
| `ingress.host` | string | | Hostname for the MLflow UI |
| `ingress.tls` | TLSConfig | | TLS config via cert-manager |
| `autoscaling` | AutoscalingConfig | | HPA configuration |
| `scheduling` | SchedulingConfig | | Pod scheduling configuration |
| `podDisruptionBudget` | PodDisruptionBudgetConfig | | PDB configuration |
| `serviceMonitor` | ServiceMonitorConfig | | ServiceMonitor configuration |
| `migration` | MigrationConfig | | Database migration configuration |
| `security` | SecurityConfig | | Enterprise security configuration |

### Status conditions

```bash
kubectl get mlflowserver team-alpha-mlflow -n ml-team-alpha
```

```
NAME                  VERSION   READY   ARTIFACT STORE   BACKEND      AGE
team-alpha-mlflow     2.11.0    True    s3               postgresql   3d
```

Conditions reported on the resource:

| Condition | Description |
|---|---|
| `Ready` | All pods running and healthy |
| `ArtifactStoreConnected` | Artifact backend reachable and writable |
| `BackendConnected` | Tracking database reachable |
| `Upgrading` | Version upgrade in progress |

---

## Multi-tenant Setup

Each team gets an isolated MLflow server in their own namespace. RBAC is scoped so teams can only manage their own `MLflowServer` resources:

```yaml
# values.yaml for operator install
multitenancy:
  enabled: true
  allowedNamespaces:
    - ml-team-alpha
    - ml-team-beta
    - ml-platform
```

---

## Artifact Store Backends

### Amazon S3

```yaml
artifactStore:
  type: s3
  s3:
    bucket: my-mlflow-artifacts
    region: us-east-1
    credentialsSecret: aws-credentials   # keys: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
    pathPrefix: team-alpha/              # optional prefix
```

### Google Cloud Storage

```yaml
artifactStore:
  type: gcs
  gcs:
    bucket: my-gcs-mlflow-bucket
    credentialsSecret: gcp-sa-key       # key: service-account.json
```

### Azure Blob Storage

```yaml
artifactStore:
  type: azure
  azure:
    storageAccount: myaccount
    container: mlflow-artifacts
    credentialsSecret: azure-storage-credentials
```

### PVC (local / NFS)

```yaml
artifactStore:
  type: pvc
  pvc:
    storageClass: nfs-client
    size: 100Gi
```

---

## Upgrading MLflow Version

Update the `version` field and apply:

```bash
kubectl patch mlflowserver team-alpha-mlflow \
  -n ml-team-alpha \
  --type merge \
  -p '{"spec":{"version":"2.12.0"}}'
```

The operator performs a rolling update — new pods come up before old ones are terminated. 

To enable automatic database migrations on version upgrade, configure the migration section:

```yaml
migration:
  enabled: true
  backoffLimit: 6
  activeDeadlineSeconds: 600
```

When enabled, the operator creates a migration job that runs `mlflow db upgrade` before the new pods start.

---

## Advanced Features

### Horizontal Pod Autoscaler

Enable auto-scaling based on CPU/memory:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

### Pod Scheduling

Control pod placement with node affinity, tolerations, and priority classes:

```yaml
scheduling:
  nodeSelector:
    nodepool: ml-workload
  tolerations:
    - key: "mlflow"
      operator: "Equal"
      value: "true"
      effect: "NoSchedule"
  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchExpressions:
                - key: app.kubernetes.io/instance
                  operator: In
                  values:
                    - my-mlflow
            topologyKey: kubernetes.io/hostname
  priorityClassName: high-priority
```

### Custom MLflow Arguments

Pass custom flags to the MLflow server:

```yaml
tracking:
  additionalArgs:
    - "--default-artifact-root"
    - "s3://my-bucket"
    - "--serve-artifacts"
```

### Custom Health Probes

Configure custom liveness, readiness, and startup probes:

```yaml
tracking:
  probes:
    livenessProbe:
      httpGet:
        path: /health
        port: 5000
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /health
        port: 5000
      initialDelaySeconds: 10
      periodSeconds: 5
    startupProbe:
      httpGet:
        path: /health
        port: 5000
      failureThreshold: 30
```

### Pod Disruption Budget

Ensure high availability during node maintenance:

```yaml
podDisruptionBudget:
  enabled: true
  minAvailable: 1
  # or
  # maxUnavailable: 1
```

### Lifecycle Hooks

Configure pre-stop and post-start hooks:

```yaml
tracking:
  lifecycle:
    preStop:
      exec:
        command: ["/bin/sh", "-c", "sleep 15"]
    postStart:
      exec:
        command: ["/bin/sh", "-c", "echo 'Started'"]
```

### Init Containers

Run setup tasks before the MLflow container starts:

```yaml
tracking:
  initContainers:
    - name: init-db-check
      image: postgres:15
      command: ['sh', '-c', 'until pg_isready -h postgres; do sleep 2; done']
```

### Sidecar Containers

Add logging, monitoring, or other auxiliary containers:

```yaml
tracking:
  sidecarContainers:
    - name: log-collector
      image: fluent/fluent-bit:2.2
      volumeMounts:
        - name: varlog
          mountPath: /var/log
          readOnly: true
```

---

## Security Features

The operator provides enterprise-grade security features to protect your MLflow deployments in production environments.

### Pod Security Standards

Enforce Kubernetes Pod Security Standards (PSS) compliance:

```yaml
security:
  podSecurityStandard: restricted  # privileged, baseline, restricted
```

- **restricted** - Most secure, follows Kubernetes best practices
- **baseline** - Minimal security controls
- **privileged** - No restrictions (not recommended for production)

### Security Contexts

Configure pod and container-level security contexts:

```yaml
security:
  podSecurityContext:
    runAsNonRoot: true
    runAsUser: 1000
    runAsGroup: 1000
    fsGroup: 1000
    seccompProfile:
      type: RuntimeDefault
    supplementalGroups:
      - 1000
      - 1001

  containerSecurityContext:
    allowPrivilegeEscalation: false
    readOnlyRootFilesystem: true
    runAsNonRoot: true
    runAsUser: 1000
    privileged: false
    capabilities:
      drop:
        - ALL
    seccompProfile:
      type: RuntimeDefault
```

### Network Policies

Control pod-to-pod communication with Network Policies:

```yaml
security:
  networkPolicy:
    enabled: true
    policyType: restrictive  # permissive, restrictive, custom
    allowedNamespaces:
      - ml-production
      - ml-monitoring
    allowedIPRanges:
      - 10.0.0.0/8
      - 192.168.0.0/16
```

**Policy Types:**
- **permissive** - Allows all traffic within cluster
- **restrictive** - Only allows traffic from same namespace and required external services
- **custom** - Define your own ingress/egress rules

### Service Account Security

Configure dedicated service accounts with bound tokens:

```yaml
security:
  serviceAccount:
    create: true
    automountServiceAccountToken: false
    boundServiceAccountToken: true
    tokenExpiration: "3600s"
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/mlflow-role
```

### Image Security

Secure your container images with pull policies and signature verification:

```yaml
security:
  imageSecurity:
    pullPolicy: Always  # Always, IfNotPresent, Never
    imagePullSecrets:
      - name: ghcr-registry-secret
    signatureVerification:
      enabled: true
      keySecret: cosign-public-key
      cosignRepository: ghcr.io/NotHarshhaa/mlflow
    vulnerabilityScan:
      enabled: true
      failOnVulnerabilities: true
      severityThreshold: HIGH  # CRITICAL, HIGH, MEDIUM, LOW
```

### Anti-Affinity

Ensure high availability by spreading pods across nodes and zones:

```yaml
security:
  antiAffinity:
    enabled: true
    topologyKey: kubernetes.io/hostname
    spreadAcrossNodes: true
    spreadAcrossZones: true
```

### RBAC Security

Create custom RBAC resources with resource quotas:

```yaml
security:
  rbac:
    enabled: true
    roleName: mlflow-server-role
    roleBindingName: mlflow-server-rolebinding
    subjects:
      - kind: ServiceAccount
        name: secure-mlflow-sa
        namespace: ml-production
    resourceQuota:
      enabled: true
      cpu: "10"
      memory: "20Gi"
      storage: "100Gi"
      pods: "20"
```

### External Secret Management

Integrate with external secret management systems:

#### HashiCorp Vault

```yaml
security:
  secretManagement:
    type: vault
    vault:
      address: https://vault.internal:8200
      authMethod: kubernetes
      role: mlflow-role
      secretPath: secret/mlflow/production
      namespace: mlflow
```

#### AWS Secrets Manager

```yaml
security:
  secretManagement:
    type: aws-secrets-manager
    awsSecretsManager:
      region: us-east-1
      secretPrefix: mlflow/
      credentialsSecret: aws-vault-credentials
```

#### Azure Key Vault

```yaml
security:
  secretManagement:
    type: azure-key-vault
    azureKeyVault:
      vaultName: mlflow-kv
      tenantId: your-tenant-id
      credentialsSecret: azure-credentials
```

#### GCP Secret Manager

```yaml
security:
  secretManagement:
    type: gcp-secret-manager
    gcpSecretManager:
      project: my-project
      secretPrefix: mlflow-
      credentialsSecret: gcp-credentials
```

### Security-Enhanced Example

See `config/samples/mlflowserver_v1alpha1_security.yaml` for a complete security-hardened MLflow server configuration with all security features enabled.
