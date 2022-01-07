package metric

import (
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Create a metrics registry.
	Registry = prometheus.NewRegistry()

	// Create some standard server metrics.
	GrpcMetrics = grpcprometheus.NewServerMetrics(
		func(o *prometheus.CounterOpts) {
			o.Namespace = "phalanx"
		},
	)
)

func init() {
	// Register standard server metrics and customized metrics to registry.
	Registry.MustRegister(
		GrpcMetrics,
	)
	GrpcMetrics.EnableHandlingTimeHistogram(
		func(o *prometheus.HistogramOpts) {
			o.Namespace = "phalanx"
		},
	)
}
