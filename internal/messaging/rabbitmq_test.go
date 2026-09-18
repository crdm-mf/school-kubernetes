package messaging

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestAMQPTraceContextRoundTrip(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)))
	defer provider.Shutdown(context.Background())
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	parentContext, parent := provider.Tracer("test").Start(context.Background(), "parent")
	headers := amqp.Table{}
	otel.GetTextMapPropagator().Inject(parentContext, amqpTableCarrier(headers))
	parent.End()

	if got := headers["traceparent"]; got == nil {
		t.Fatal("traceparent was not injected")
	}
	extracted := otel.GetTextMapPropagator().Extract(context.Background(), amqpTableCarrier(headers))
	_, child := provider.Tracer("test").Start(extracted, "child")
	child.End()

	spans := exporter.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("got %d spans, want 2", len(spans))
	}
	if spans[1].Parent.TraceID() != spans[0].SpanContext.TraceID() {
		t.Fatalf("child trace ID %s does not match parent trace ID %s", spans[1].SpanContext.TraceID(), spans[0].SpanContext.TraceID())
	}
}

func TestAMQPInvalidTraceContextStartsNewRoot(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)))
	defer provider.Shutdown(context.Background())
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	ctx := otel.GetTextMapPropagator().Extract(context.Background(), amqpTableCarrier(amqp.Table{
		"traceparent": "not-a-traceparent",
	}))
	_, span := provider.Tracer("test").Start(ctx, "root")
	span.End()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Parent.IsValid() {
		t.Fatalf("invalid remote context became a parent: %v", spans[0].Parent)
	}
	if !spans[0].SpanContext.IsValid() {
		t.Fatal("new root span is invalid")
	}
}
