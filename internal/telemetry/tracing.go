package telemetry

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	traceServiceNamespace = "food-delivery"
	traceEnvironment      = "local"
)

// InitTracing configures W3C propagation for every process and enables OTLP
// export when an endpoint is configured. Export setup is deliberately
// fail-open so telemetry cannot prevent the application from starting.
func InitTracing(ctx context.Context, serviceName string, logger *slog.Logger) func(context.Context) error {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) == "" &&
		strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")) == "" {
		return noopShutdown
	}

	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		if logger != nil {
			logger.Warn("OpenTelemetry exporter unavailable; tracing disabled", "error", err)
		}
		return noopShutdown
	}

	provider := newTracerProvider(serviceName, exporter)
	otel.SetTracerProvider(provider)
	return provider.Shutdown
}

// ShutdownTracing flushes pending spans without making process shutdown fail.
func ShutdownTracing(shutdown func(context.Context) error, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil && logger != nil {
		logger.Warn("OpenTelemetry shutdown failed", "error", err)
	}
}

// InstrumentHTTP adds low-cardinality server spans and propagates W3C context.
// Health, metrics, snapshot polling, and the long-lived event stream are intentionally excluded.
func InstrumentHTTP(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "HTTP",
		otelhttp.WithFilter(func(request *http.Request) bool {
			path := request.URL.Path
			return path != "/metrics" &&
				!strings.HasPrefix(path, "/health/") &&
				path != "/api/v1/snapshot" &&
				path != "/api/v1/events"
		}),
		otelhttp.WithSpanNameFormatter(func(operation string, request *http.Request) string {
			if request.Pattern != "" {
				return request.Pattern
			}
			return request.Method + " " + request.URL.Path
		}),
	)
}

func newTracerProvider(serviceName string, exporter sdktrace.SpanExporter) *sdktrace.TracerProvider {
	instanceID := strings.TrimSpace(os.Getenv("POD_NAME"))
	if instanceID == "" {
		instanceID = "local"
	}
	resource := sdkresource.NewWithAttributes(
		"",
		attribute.String("service.name", serviceName),
		attribute.String("service.namespace", traceServiceNamespace),
		attribute.String("service.instance.id", instanceID),
		attribute.String("deployment.environment.name", traceEnvironment),
	)
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(200*time.Millisecond),
			sdktrace.WithMaxExportBatchSize(128),
		),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
}

func noopShutdown(context.Context) error {
	return nil
}
