package interceptors

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/DKhorkov/libs/tracing"
)

// UnaryServerTracingInterceptor creates span on base of existing span and logs its Start and End events.
func UnaryServerTracingInterceptor(
	tp *tracing.TraceProvider,
	spanConfig tracing.SpanConfig,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var span trace.Span
		defer func() {
			if span != nil {
				span.End()
			}
		}()

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			traceKey := strings.ToLower(tracing.Key) // metadata sends all keys in lowercase
			if _, ok = md[traceKey]; ok {
				traceHex := md[traceKey][0] // md is a map[string][]string
				traceID, err := tp.TraceIDFromHex(traceHex)
				if err != nil {
					return nil, err
				}

				ctx, span = tp.SpanFromTraceID(ctx, traceID, info.FullMethod, spanConfig.Opts...)
				span.AddEvent(spanConfig.Events.Start.Name, spanConfig.Events.Start.Opts...)
			}
		}

		result, err := handler(ctx, req)

		if span != nil {
			span.AddEvent(spanConfig.Events.End.Name, spanConfig.Events.End.Opts...)
		}

		return result, err
	}
}

// UnaryClientTracingInterceptor creates span on base of existing span and logs its Start and End events.
func UnaryClientTracingInterceptor(
	tp *tracing.TraceProvider,
	spanConfig tracing.SpanConfig,
) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx, span := tp.Span(ctx, method, spanConfig.Opts...)
		defer span.End()

		span.AddEvent(spanConfig.Events.Start.Name, spanConfig.Events.Start.Opts...)
		err := invoker(ctx, method, req, reply, cc, opts...)
		span.AddEvent(spanConfig.Events.End.Name, spanConfig.Events.End.Opts...)

		return err
	}
}
