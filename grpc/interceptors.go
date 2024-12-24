package grpc

import (
	"context"
	"log/slog"

	"github.com/DKhorkov/libs/contextlib"
	"github.com/DKhorkov/libs/requestid"
	"google.golang.org/grpc"
)

// ServerLoggingUnaryInterceptor intercepts gRPC handler, logs request with provided request ID and calls handler.
func ServerLoggingUnaryInterceptor(
	logger *slog.Logger,
) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		requestIDer, ok := req.(requestid.RequestIDer)
		if ok {
			requestID := requestIDer.GetRequestID()
			ctx = contextlib.SetValue(ctx, requestid.Key, requestID)

			logger.InfoContext(
				ctx,
				"Received new request",
				"Request ID",
				requestID,
				"Request",
				req,
				"Handler",
				info.FullMethod,
			)
		}

		return handler(ctx, req)
	}
}
