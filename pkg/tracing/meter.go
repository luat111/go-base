package tracing

import (
	"context"
	"go-base/pkg/logger"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// InitMeter sets up the global OpenTelemetry MeterProvider backed by an OTLP
// gRPC exporter. It reads the same OTEL_EXPORTER_OTLP_ENDPOINT used by the
// tracer so only one environment variable is needed.
//
// The returned shutdown function must be called on application exit to flush
// pending metric data points before the process exits.
func initMeter(
	ctx context.Context,
	log logger.ILogger,
	endpoint string,
	res *resource.Resource,
) (*metric.MeterProvider, error) {

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(exporter, metric.WithInterval(15*time.Second)),
		),
		metric.WithResource(res),
	)

	log.Info("OpenTelemetry MeterProvider initialised")

	return mp, nil
}
