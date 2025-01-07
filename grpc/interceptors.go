package grpc

import (
	"context"
	"log/slog"
	"strings"

	"google.golang.org/grpc/metadata"

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
		var requestID string

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			requestIDKey := strings.ToLower(requestid.Key) // metadata sends all keys in lowercase
			if _, ok = md[requestIDKey]; ok {
				requestID = md[requestIDKey][0] // md is a map[string][]string
			}

			ctx = contextlib.SetValue(ctx, requestid.Key, requestID) // setting to context value for inner usage
		}

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
