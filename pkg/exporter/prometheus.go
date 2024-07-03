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
	Reg     prometheus.Registry
}

/*
  Prometheus module based on code from github user: jessicalins
  https://github.com/jessicalins/instrumentation-practices-examples/blob/main/middleware/httpmiddleware/httpmiddleware.go
*/

func NewPrometheusExporter(port string) *Prometheus {
	prom := Prometheus{
		Port:    port,
		Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
		Reg:     *prometheus.NewRegistry(),
	}
	return &prom
}

func (p *Prometheus) Wrapper(handlerName string, handlerFunc func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	registry := prometheus.NewRegistry()

	// Add go runtime metrics and process collectors
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	// reg := prometheus.WrapRegistererWith(prometheus.Labels{"handler": handlerName}, registry)
	requestsTotal := promauto.With(&p.Reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Tracks the number of HTTP requests.",
		}, []string{"method", "code"},
	)
	requestDuration := promauto.With(&p.Reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: p.Buckets,
		},
		[]string{"method", "code"},
	)
	requestSize := promauto.With(&p.Reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Tracks the size of HTTP requests.",
		},
		[]string{"method", "code"},
	)
	responseSize := promauto.With(&p.Reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Tracks the size of HTTP responses.",
		},
		[]string{"method", "code"},
	)

	handler := http.HandlerFunc(handlerFunc)
	base := promhttp.InstrumentHandlerCounter(
		requestsTotal,
		promhttp.InstrumentHandlerDuration(
			requestDuration,
			promhttp.InstrumentHandlerRequestSize(
				requestSize,
				promhttp.InstrumentHandlerResponseSize(
					responseSize,
					handler,
				),
			),
		),
	)

	return base.ServeHTTP
}

func (p *Prometheus) Export() http.Handler {
	return promhttp.HandlerFor(&p.Reg, promhttp.HandlerOpts{})
}
