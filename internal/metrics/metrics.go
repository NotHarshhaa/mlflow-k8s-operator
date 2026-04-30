package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// Registry is the prometheus registry
	Registry = metrics.Registry

	// ServersTotal is the total number of MLflowServer resources managed
	ServersTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mlflow_operator_servers_total",
			Help: "Total MLflowServer resources managed",
		},
		[]string{"namespace"},
	)

	// ServersReady is the number of servers in Ready state
	ServersReady = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mlflow_operator_servers_ready",
			Help: "Servers in Ready state",
		},
		[]string{"namespace"},
	)

	// ReconcileDuration is the duration of reconcile loops
	ReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mlflow_operator_reconcile_duration_seconds",
			Help:    "Reconcile loop duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace", "name"},
	)

	// ReconcileErrors is the total number of failed reconciliations
	ReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mlflow_operator_reconcile_errors_total",
			Help: "Failed reconciliations",
		},
		[]string{"namespace", "name"},
	)

	// ArtifactStoreUp indicates artifact backend reachability per server
	ArtifactStoreUp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mlflow_server_artifact_store_up",
			Help: "Artifact backend reachability (per server)",
		},
		[]string{"namespace", "name", "type"},
	)

	// BackendUp indicates tracking DB reachability per server
	BackendUp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mlflow_server_backend_up",
			Help: "Tracking DB reachability (per server)",
		},
		[]string{"namespace", "name", "type"},
	)
)

// RegisterCustomMetrics registers custom metrics with the registry
func RegisterCustomMetrics(r prometheus.Registerer) {
	r.MustRegister(ServersTotal)
	r.MustRegister(ServersReady)
	r.MustRegister(ReconcileDuration)
	r.MustRegister(ReconcileErrors)
	r.MustRegister(ArtifactStoreUp)
	r.MustRegister(BackendUp)
}
