package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestInstrumentHTTPFiltersOperationalEndpoints(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)))
	defer provider.Shutdown(context.Background())
	otel.SetTracerProvider(provider)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/orders", func(http.ResponseWriter, *http.Request) {})
	mux.HandleFunc("GET /api/v1/snapshot", func(http.ResponseWriter, *http.Request) {})
	mux.HandleFunc("GET /health/ready", func(http.ResponseWriter, *http.Request) {})
	mux.HandleFunc("GET /metrics", func(http.ResponseWriter, *http.Request) {})
	mux.HandleFunc("GET /api/v1/events", func(http.ResponseWriter, *http.Request) {})
	handler := InstrumentHTTP(mux)

	for _, path := range []string{"/api/v1/orders", "/api/v1/snapshot", "/health/ready", "/metrics", "/api/v1/events"} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want only the business endpoint traced", len(spans))
	}
	if spans[0].Name != "GET /api/v1/orders" {
		t.Fatalf("span name = %q, want stable route name", spans[0].Name)
	}
}

func TestTracerProviderUsesConfiguredResource(t *testing.T) {
	previous, hadPrevious := os.LookupEnv("POD_NAME")
	if err := os.Setenv("POD_NAME", "order-worker-7"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if hadPrevious {
			_ = os.Setenv("POD_NAME", previous)
		} else {
			_ = os.Unsetenv("POD_NAME")
		}
	}()

	exporter := tracetest.NewInMemoryExporter()
	provider := newTracerProvider("order-worker", exporter)
	defer provider.Shutdown(context.Background())
	ctx := context.Background()
	_, span := provider.Tracer("test").Start(ctx, "database projection")
	span.End()
	if err := provider.ForceFlush(ctx); err != nil {
		t.Fatal(err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	attributes := spans[0].Resource.Attributes()
	want := map[string]string{
		"service.name":                "order-worker",
		"service.namespace":           "food-delivery",
		"service.instance.id":         "order-worker-7",
		"deployment.environment.name": "local",
	}
	for key, value := range want {
		found := false
		for _, attribute := range attributes {
			if string(attribute.Key) == key && attribute.Value.AsString() == value {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("resource attribute %s=%q not found", key, value)
		}
	}
}
