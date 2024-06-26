package exporter

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var ()

type Prometheus struct {
	Port    string
	Buckets []float64
}

/*
  Prometheus module based on code from github user: jessicalins
  https://github.com/jessicalins/instrumentation-practices-examples/blob/main/middleware/httpmiddleware/httpmiddleware.go
*/

func NewPrometheusExporter(port string) Prometheus {
	prom := Prometheus{
		Port:    port,
		Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
	}
	return prom
}

func (p Prometheus) Wrapper(handlerName string) http.HandlerFunc {
	registry := prometheus.NewRegistry()

	// Add go runtime metrics and process collectors
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	reg := prometheus.WrapRegistererWith(prometheus.Labels{"handler": handlerName}, registry)
	requestsTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Tracks the number of HTTP requests.",
		}, []string{"method", "code"},
	)
	requestDuration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: p.Buckets,
		},
		[]string{"method", "code", "path"},
	)
	requestSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Tracks the size of HTTP requests.",
		},
		[]string{"method", "code"},
	)
	responseSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Tracks the size of HTTP responses.",
		},
		[]string{"method", "code"},
	)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	base := promhttp.InstrumentHandlerCounter(
		requestsTotal,
		promhttp.InstrumentHandlerDuration(
			requestDuration,
			promhttp.InstrumentHandlerRequestSize(
				requestSize,
				promhttp.InstrumentHandlerResponseSize(
					responseSize,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						handler.ServeHTTP(w, r)
					}),
				),
			),
		),
	)

	return base.ServeHTTP
}
