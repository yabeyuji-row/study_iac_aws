package observability

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type HTTPMetrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func NewHTTPMetrics() (metrics *HTTPMetrics) {
	metrics = &HTTPMetrics{
		registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "todo_api", Name: "http_requests_total",
			Help: "Completed HTTP requests.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "todo_api", Name: "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		}, []string{"method", "route"}),
	}
	metrics.registry.MustRegister(metrics.requests, metrics.duration)
	return metrics
}

func (metrics *HTTPMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(metrics.registry, promhttp.HandlerOpts{})
}

func HTTPMiddleware(metrics *HTTPMetrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoContext echo.Context) (err error) {
			started := time.Now()
			request := echoContext.Request()
			err = next(echoContext)

			route := echoContext.Path()
			if route == "" {
				route = "unmatched"
			}
			status := echoContext.Response().Status
			if httpError, ok := err.(*echo.HTTPError); ok {
				status = httpError.Code
			}
			if status == 0 {
				status = http.StatusOK
			}
			metrics.requests.WithLabelValues(request.Method, route, strconv.Itoa(status)).Inc()
			metrics.duration.WithLabelValues(request.Method, route).Observe(time.Since(started).Seconds())
			return err
		}
	}
}

func ConfigureTracing(ctx context.Context, endpoint string) (shutdown func(context.Context) error, err error) {
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint+"/v1/traces"))
	if err != nil {
		return nil, fmt.Errorf("create OTLP HTTP exporter: %w", err)
	}
	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("todo-api"),
		)),
	)
	otel.SetTracerProvider(tracerProvider)
	return tracerProvider.Shutdown, nil
}
