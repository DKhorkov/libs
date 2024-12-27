package grpc

import (
	"context"
	"log/slog"

	"github.com/DKhorkov/libs/contextlib"
	"github.com/DKhorkov/libs/requestid"
	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
)

// UnaryServerLoggingInterceptor intercepts gRPC handler, logs request with provided request ID and calls handler.
func UnaryServerLoggingInterceptor(
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

// UnaryClientLoggingInterceptor adapts slog logger to interceptor logger.
func UnaryClientLoggingInterceptor(logger *slog.Logger) grpclogging.Logger {
	return grpclogging.LoggerFunc(
		func(
			ctx context.Context,
			logLevel grpclogging.Level,
			msg string,
			fields ...any,
		) {
			logger.Log(
				ctx,
				slog.Level(logLevel),
				msg,
				fields...,
			)
		},
	)
}
