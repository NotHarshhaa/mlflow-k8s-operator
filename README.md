# mlflow-k8s-operator ⚗️

> A Kubernetes operator that provisions and manages MLflow tracking servers, model registries, and artifact stores as native CRDs — so data scientists never touch kubectl.

[![Go version](https://img.shields.io/badge/go-1.21%2B-blue)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.25%2B-326CE5?logo=kubernetes&logoColor=white)](https://kubernetes.io/)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Operator SDK](https://img.shields.io/badge/built%20with-Operator%20SDK-red)](https://sdk.operatorframework.io/)
[![Docker Image](https://img.shields.io/badge/Docker%20Hub-harshhaareddy%2Fmlflow--k8s--operator-blue?logo=docker)](https://hub.docker.com/r/harshhaareddy/mlflow-k8s-operator)
[![GHCR Image](https://img.shields.io/badge/GHCR-notharshhaa%2Fmlflow--k8s--operator-blue?logo=github)](https://github.com/notharshhaa/mlflow-k8s-operator/pkgs/container/mlflow-k8s-operator)
[![Helm Chart](https://img.shields.io/badge/Helm%20Chart-1.0.0-blue?logo=helm)](https://NotHarshhaa.github.io/mlflow-k8s-operator)

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
- **Enterprise security** — Pod Security Standards, Network Policies, RBAC, image signing, anti-affinity, external secret management

---

## Quickstart

### Docker Images

The operator is available as Docker images:

```bash
# Docker Hub
docker pull harshhaareddy/mlflow-k8s-operator:latest

# GitHub Container Registry
docker pull ghcr.io/notharshhaa/mlflow-k8s-operator:latest
```

Both images support multi-architecture (linux/amd64, linux/arm64).

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

### Configure and deploy MLflow servers

For detailed configuration options, examples, and deployment guides, see the [Configuration Guide](CONFIGURATION.md).

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

## Roadmap

- [x] **v0.1** — `MLflowServer` CRD, S3/GCS artifact stores, PostgreSQL backend, basic reconciler
- [x] **v0.2** — Ingress + TLS via cert-manager, multi-namespace support, Helm chart
- [x] **v0.3** — Azure Blob support, rolling upgrades, Prometheus metrics
- [x] **v0.4** — HPA, PDB, scheduling controls, custom probes, lifecycle hooks, init/sidecar containers, migration jobs
- [x] **v1.0** — Enterprise security features (PSS, Network Policies, RBAC, image signing, anti-affinity, external secret management), stable API
- [ ] **v1.1** — `MLflowExperiment` CRD for GitOps-driven experiment management
- [ ] **v1.2** — Model Registry HA mode, read replicas for tracking database
- [ ] **v1.3** — OLM (OperatorHub) listing, full e2e test suite

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
