# MLflow Kubernetes Operator Helm Chart

This Helm chart installs the MLflow Kubernetes Operator on a Kubernetes cluster.

## Installation

### Add the Helm Repository

```bash
helm repo add mlflow-k8s-operator https://NotHarshhaa.github.io/mlflow-k8s-operator
helm repo update
```

### Install the Operator

```bash
helm install mlflow-operator mlflow-k8s-operator/mlflow-k8s-operator \
  --namespace mlflow-system \
  --create-namespace
```

### Install with Custom Values

```bash
helm install mlflow-operator mlflow-k8s-operator/mlflow-k8s-operator \
  --namespace mlflow-system \
  --create-namespace \
  --set networkPolicy.enabled=true \
  --set podDisruptionBudget.enabled=true \
  --set podDisruptionBudget.minAvailable=1 \
  --set controller.manager.args[0]=--leader-elect=true
```

### Install with Values File

Create a `values.yaml` file:

```yaml
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

Then install:

```bash
helm install mlflow-operator mlflow-k8s-operator/mlflow-k8s-operator \
  --namespace mlflow-system \
  --create-namespace \
  -f values.yaml
```

## Configuration

See [values.yaml](../charts/mlflow-k8s-operator/values.yaml) for all available configuration options.

### Security Features

Enable security features for production deployments:

```yaml
# Enable network policies for network isolation
networkPolicy:
  enabled: true

# Enable Pod Disruption Budgets for high availability
podDisruptionBudget:
  enabled: true
  minAvailable: 1

# Enable leader election for HA deployments
controller:
  manager:
    args:
      - --leader-elect=true
```

### Resource Limits

Configure resource limits:

```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 100m
    memory: 256Mi
```

## Upgrading

```bash
helm upgrade mlflow-operator mlflow-k8s-operator/mlflow-k8s-operator \
  --namespace mlflow-system
```

## Uninstalling

```bash
helm uninstall mlflow-operator --namespace mlflow-system
```

## Creating an MLflow Server

After installing the operator, create an MLflow server:

```yaml
apiVersion: mlops.NotHarshhaa.io/v1alpha1
kind: MLflowServer
metadata:
  name: mlflow-server
  namespace: mlflow-system
spec:
  version: "2.11.0"
  tracking:
    replicas: 1
  backend:
    type: postgresql
    postgresql:
      host: postgres.example.com
      port: 5432
      database: mlflow
      credentialsSecret: postgres-creds
  artifactStore:
    type: s3
    s3:
      bucket: mlflow-artifacts
      region: us-east-1
      credentialsSecret: s3-creds
  ingress:
    enabled: true
    host: mlflow.example.com
```

Apply the manifest:

```bash
kubectl apply -f mlflow-server.yaml
```

## Troubleshooting

Check operator logs:

```bash
kubectl logs -n mlflow-system -l app.kubernetes.io/name=mlflow-k8s-operator -f
```

Check operator status:

```bash
kubectl get mlflowserver -n mlflow-system
kubectl describe mlflowserver mlflow-server -n mlflow-system
```

## Security

For production deployments, see [SECURITY.md](../SECURITY.md) for comprehensive security guidelines.
