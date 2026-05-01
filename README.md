# mlflow-k8s-operator ⚗️

> A Kubernetes operator that provisions and manages MLflow tracking servers, model registries, and artifact stores as native CRDs — so data scientists never touch kubectl.

[![Go version](https://img.shields.io/badge/go-1.21%2B-blue)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.25%2B-326CE5?logo=kubernetes&logoColor=white)](https://kubernetes.io/)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Operator SDK](https://img.shields.io/badge/built%20with-Operator%20SDK-red)](https://sdk.operatorframework.io/)

---

## The problem

Every team doing ML on Kubernetes ends up manually deploying MLflow — writing Deployments, Services, PVCs, Ingress rules, and wiring up S3 or GCS artifact stores by hand. When you have five teams, you have five slightly different setups, each owned by a different person, none of them production-hardened.

**mlflow-k8s-operator solves this** by giving you a single `MLflowServer` CRD. Define what you want, apply it, and the operator handles the rest — provisioning, upgrades, storage backends, ingress, TLS, and lifecycle management.

---

## Features

- **`MLflowServer` CRD** — declare a full MLflow stack in a single manifest
- **Artifact store backends** — S3, GCS, Azure Blob, and local PVC out of the box
- **PostgreSQL / MySQL integration** — auto-provisions or connects to existing databases for the tracking backend
- **Ingress management** — generates Ingress or HTTPRoute resources with optional TLS via cert-manager
- **Multi-tenancy** — namespace-scoped servers with RBAC isolation per team
- **Upgrades and rollbacks** — controlled MLflow version upgrades with zero-downtime rolling restarts
- **Prometheus metrics** — exposes operator health and MLflow server status as metrics
- **Helm chart included** — install the operator in one command
- **Horizontal Pod Autoscaler** — auto-scale MLflow servers based on CPU/memory utilization
- **Pod scheduling controls** — node affinity, tolerations, priority classes, and topology spread constraints
- **Custom MLflow arguments** — pass custom flags to the MLflow server
- **Customizable health probes** — configure liveness, readiness, and startup probes
- **Pod Disruption Budget** — ensure high availability during node maintenance
- **Lifecycle hooks** — pre-stop and post-start hooks for graceful shutdown and initialization
- **Init containers** — run setup tasks before the MLflow container starts
- **Sidecar containers** — add logging, monitoring, or other auxiliary containers
- **Database migration jobs** — automatic schema migrations on version upgrades
- **ServiceMonitor support** — optional Prometheus Operator integration (requires prometheus-operator)

---

## Quickstart

### Install the operator

> **Note:** The Helm chart is hosted on GitHub Pages. The GitHub Actions workflow automatically publishes the chart to the `gh-pages` branch when changes are pushed to `main`. Enable GitHub Pages in your repository settings (Settings > Pages > Source: Deploy from a branch > Branch: gh-pages).

```bash
helm repo add mlflow-k8s-operator https://NotHarshhaa.github.io/mlflow-k8s-operator
helm install mlflow-operator mlflow-k8s-operator/mlflow-k8s-operator \
  --namespace mlflow-system \
  --create-namespace
```

Or install from local directory:

```bash
helm install mlflow-operator ./charts/mlflow-k8s-operator \
  --namespace mlflow-system \
  --create-namespace
```

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

That's it. The operator provisions the Deployment, Service, PVC, ConfigMap, Ingress, and wires up the artifact store — fully reconciled and self-healing.

---

## CRD reference

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

## Multi-tenant setup

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

## Artifact store backends

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

## Kubernetes deployment architecture

```
mlflow-system/
  └── mlflow-operator (Deployment)          ← watches all namespaces for MLflowServer CRDs

ml-team-alpha/
  ├── MLflowServer/team-alpha-mlflow        ← your CRD
  ├── Deployment/team-alpha-mlflow          ← managed by operator
  ├── Service/team-alpha-mlflow             ← ClusterIP
  ├── Ingress/team-alpha-mlflow             ← optional
  ├── ConfigMap/team-alpha-mlflow-config    ← generated config
  └── Secret/team-alpha-mlflow-env          ← injected credentials
```

---

## Prometheus metrics

The operator exposes metrics at `:8080/metrics`:

| Metric | Type | Description |
|---|---|---|
| `mlflow_operator_servers_total` | Gauge | Total MLflowServer resources managed |
| `mlflow_operator_servers_ready` | Gauge | Servers in Ready state |
| `mlflow_operator_reconcile_duration_seconds` | Histogram | Reconcile loop duration |
| `mlflow_operator_reconcile_errors_total` | Counter | Failed reconciliations |
| `mlflow_server_artifact_store_up` | Gauge | Artifact backend reachability (per server) |
| `mlflow_server_backend_up` | Gauge | Tracking DB reachability (per server) |

---

## Upgrading MLflow version

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

## Roadmap

- [x] **v0.1** — `MLflowServer` CRD, S3/GCS artifact stores, PostgreSQL backend, basic reconciler
- [x] **v0.2** — Ingress + TLS via cert-manager, multi-namespace support, Helm chart
- [x] **v0.3** — Azure Blob support, rolling upgrades, Prometheus metrics
- [x] **v0.4** — HPA, PDB, scheduling controls, custom probes, lifecycle hooks, init/sidecar containers, migration jobs
- [ ] **v0.5** — `MLflowExperiment` CRD for GitOps-driven experiment management
- [ ] **v0.6** — Model Registry HA mode, read replicas for tracking database
- [ ] **v1.0** — OLM (OperatorHub) listing, stable API, full e2e test suite

---

## Local development

Prerequisites: `go 1.21+`, `kubectl`, `kind` or `k3d`, `operator-sdk`

```bash
git clone https://github.com/NotHarshhaa/mlflow-k8s-operator
cd mlflow-k8s-operator

# spin up a local cluster
kind create cluster --name mlflow-dev

# install CRDs
make install

# run operator locally (outside cluster)
make run

# in another terminal — apply a test server
kubectl apply -f config/samples/mlflowserver_v1alpha1.yaml
```

Run tests:

```bash
make test          # unit tests
make e2e-test      # end-to-end tests against kind cluster
```

---

## Contributing

Issues and PRs are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting.

Areas actively looking for help:
- Additional artifact store backends (MinIO, HDFS)
- OLM bundle packaging for OperatorHub submission
- E2E test coverage
- Documentation improvements

---

## License

Apache 2.0 — see [LICENSE](LICENSE).
