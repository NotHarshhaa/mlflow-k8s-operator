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

---

## Quickstart

### Install the operator

```bash
helm repo add mlflow-k8s-operator https://your-org.github.io/mlflow-k8s-operator
helm install mlflow-operator mlflow-k8s-operator/mlflow-k8s-operator \
  --namespace mlflow-system \
  --create-namespace
```

### Provision your first MLflow server

```yaml
# mlflow-server.yaml
apiVersion: mlops.your-org.io/v1alpha1
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
| `backend.type` | enum | ✓ | `postgresql`, `mysql`, `sqlite` |
| `backend.postgresql` | PostgreSQLConfig | | Connection config for PostgreSQL |
| `artifactStore.type` | enum | ✓ | `s3`, `gcs`, `azure`, `pvc` |
| `artifactStore.s3` | S3Config | | S3 bucket and credentials |
| `ingress.enabled` | bool | | Create Ingress resource (default: false) |
| `ingress.host` | string | | Hostname for the MLflow UI |
| `ingress.tls` | TLSConfig | | TLS config via cert-manager |

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

The operator performs a rolling update — new pods come up before old ones are terminated. A `db migrate` job runs automatically if schema changes are required.

---

## Roadmap

- [ ] **v0.1** — `MLflowServer` CRD, S3/GCS artifact stores, PostgreSQL backend, basic reconciler
- [ ] **v0.2** — Ingress + TLS via cert-manager, multi-namespace support, Helm chart
- [ ] **v0.3** — Azure Blob support, rolling upgrades, Prometheus metrics
- [ ] **v0.4** — `MLflowExperiment` CRD for GitOps-driven experiment management
- [ ] **v0.5** — Model Registry HA mode, auto-scaling tracking servers
- [ ] **v1.0** — OLM (OperatorHub) listing, stable API, full e2e test suite

---

## Local development

Prerequisites: `go 1.21+`, `kubectl`, `kind` or `k3d`, `operator-sdk`

```bash
git clone https://github.com/your-org/mlflow-k8s-operator
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
