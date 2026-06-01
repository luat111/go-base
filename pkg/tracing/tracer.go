package tracing

import (
	"go.opentelemetry.io/auto/sdk"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Init sets up the global OpenTelemetry TracerProvider using the Go
// Instrumentation Auto SDK. The Auto SDK integrates with the eBPF
// auto-instrumentation agent – no manual exporter is required. When no
// agent is attached the SDK is a no-op and produces no telemetry.
//
// Call Init() once – as early as possible – before any server or client
// is started.
func Init() {
	tp := sdk.TracerProvider()

	otel.SetTracerProvider(tp)

	// W3C TraceContext + Baggage propagators allow trace context to be
	// forwarded across HTTP headers, gRPC metadata, and MQ message headers.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
}

// GetTracer returns a named tracer from the global TracerProvider.
// Use this helper to create manual spans inside application code.
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
