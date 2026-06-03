package grpc

import (
	"context"
	"fmt"
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	GrpcTrackingId = "x-correlation-id"
	Request        = "REQUEST"
	Consumer       = "CONSUMER"
)

type gRPCLog struct {
	CorrelationId string        `json:"correlationId"`
	StartTime     string        `json:"startTime"`
	ResponseTime  time.Duration `json:"responseTime"`
	Method        string        `json:"method"`
	StatusCode    int32         `json:"statusCode"`
	Data          any           `json:"data"`
}

func ObservabilityInterceptor(logger logger.ILogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		md, _ := metadata.FromIncomingContext(ctx)

		// correlationId := getMetadataValue(md, GrpcTrackingId)
		// Prefer the OTel trace ID so logs and traces share the same identifier.
		// Fall back to the incoming correlation-id metadata header.
		correlationId := tracing.TraceIDFromContext(ctx)
		if correlationId == "" {
			correlationId = getMetadataValue(md, GrpcTrackingId)
		}

		resp, err := handler(ctx, req)
		if err != nil {
			errMsg := fmt.Sprintf("error while handling gRPC request to method %q: %q", info.FullMethod, err)
			logger.Error(ctx, errMsg)
		}

		// Record gRPC status on the active OTel span.
		setGRPCSpanStatus(ctx, err)

		logRPC(Consumer, logger, correlationId, start, err, info.FullMethod, req)

		return resp, err
	}
}

func logRPC(logType string, logger logger.ILogger, correlationId string, start time.Time, err error, method string, data any) {
	logMsg := gRPCLog{
		CorrelationId: correlationId,
		StartTime:     start.Format("2006-01-02T15:04:05"),
		ResponseTime:  time.Since(start),
		Method:        method,
		Data:          data,
	}

	if err != nil {
		statusErr, _ := status.FromError(err)
		logMsg.StatusCode = int32(statusErr.Code())
	} else {
		logMsg.StatusCode = int32(codes.OK)
	}

	if logger == nil {
		return
	}

	logger.Info("GRPC", "Type", logType, "Message", logMsg)
}

// setGRPCSpanStatus annotates the active span with the gRPC status code and
// marks it as an error span when the handler returned an error.
func setGRPCSpanStatus(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	st, _ := status.FromError(err)
	grpcCode := st.Code()

	span.SetAttributes(attribute.Int("rpc.grpc.status_code", int(grpcCode)))

	if err != nil {
		span.SetStatus(otelcodes.Error, st.Message())
	} else {
		span.SetStatus(otelcodes.Ok, "")
	}
}

func getMetadataValue(md metadata.MD, key string) string {
	if values, ok := md[key]; ok && len(values) > 0 {
		return values[0]
	}

	return ""
}
