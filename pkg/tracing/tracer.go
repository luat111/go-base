package tracing

import (
	"context"
	"go-base/pkg/config"
	"go-base/pkg/logger"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Init sets up the global OpenTelemetry TracerProvider backed by an OTLP gRPC
// exporter. Configuration is read from the standard OTEL_* environment variables:
//
//	OTEL_EXPORTER_OTLP_ENDPOINT  – collector gRPC address (default: localhost:4317)
//	OTEL_RESOURCE_ATTRIBUTES     – e.g. service.name=go-base
//
// The returned shutdown function must be called on application exit to flush and
// close the exporter gracefully.
//
// Call Init() once — as early as possible — before any server or client is started.
func Init(config config.Config, logger logger.ILogger) (shutdown func(), err error) {
	ctx := context.Background()

	// Determine the OTLP endpoint (strip http:// prefix if present – grpc dialer
	// does not understand URI schemes in all versions).
	endpoint := config.Get("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		logger.Warn("OTEL_EXPORTER_OTLP_ENDPOINT is not set")
	}
	// Strip scheme so the gRPC dialer gets a plain host:port.
	for _, prefix := range []string{"http://", "https://"} {
		if len(endpoint) > len(prefix) && endpoint[:len(prefix)] == prefix {
			endpoint = endpoint[len(prefix):]
			break
		}
	}

	// Build the OTLP gRPC exporter. WithInsecure matches the plain HTTP endpoint
	// in the .env file. TLS can be enabled by switching to WithTLSClientConfig.
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Service resource carries the service.name and any attributes from
	// OTEL_RESOURCE_ATTRIBUTES.
	serviceName := config.GetOrDefault("APP_NAME", "go-base")

	res, err := resource.New(ctx,
		resource.WithFromEnv(), // picks up OTEL_RESOURCE_ATTRIBUTES
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)

	if err != nil {
		// Non-fatal – proceed with a minimal resource.
		res = resource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	mp, err := initMeter(ctx, logger, endpoint, res)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	// W3C TraceContext + Baggage propagators allow trace context to be
	// forwarded across HTTP headers, gRPC metadata, and MQ message headers.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	shutdown = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
	}

	return shutdown, nil
}

// GetTracer returns a named tracer from the global TracerProvider.
// Use this helper to create manual spans inside application code.
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
