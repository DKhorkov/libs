package interceptors

import (
	"context"
	"log/slog"
	"reflect"
	"strings"

	"github.com/DKhorkov/libs/contextlib"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/requestid"
	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	passwordFieldName = "Password"
)

// UnaryServerLoggingInterceptor intercepts gRPC handler, logs request with provided request ID and calls handler.
func UnaryServerLoggingInterceptor(logger logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var requestID string

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			requestIDKey := strings.ToLower(requestid.Key) // metadata sends all keys in lowercase
			if _, ok = md[requestIDKey]; ok {
				requestID = md[requestIDKey][0] // md is a map[string][]string
			}

			ctx = contextlib.WithValue(
				ctx,
				requestid.Key,
				requestID,
			) // setting to context value for inner usage
		}

		// Making password field empty not to store in logs:
		var passwordField reflect.Value

		var passwordCopy string

		if reflectValue := reflect.ValueOf(req); reflectValue.IsValid() && !reflectValue.IsZero() {
			if reflectValue.Kind() == reflect.Ptr {
				passwordField = reflectValue.Elem().FieldByName(passwordFieldName)
			} else {
				passwordField = reflectValue.FieldByName(passwordFieldName)
			}

			if passwordField.IsValid() && !passwordField.IsZero() {
				passwordCopy = passwordField.String()
				passwordField.SetString("")
			}
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

		// Returning password field for correct logics due to a pointer request:
		if passwordField.IsValid() {
			passwordField.SetString(passwordCopy)
		}

		return handler(ctx, req)
	}
}

// UnaryClientLoggingInterceptor adapts logging.logger to interceptor logger.
func UnaryClientLoggingInterceptor(logger logging.Logger) grpclogging.Logger {
	return grpclogging.LoggerFunc(
		func(
			ctx context.Context,
			logLevel grpclogging.Level,
			msg string,
			fields ...any,
		) {
			// Making password field empty not to store in logs:
			var passwordField reflect.Value

			var passwordCopy string

			for _, field := range fields {
				if reflectValue := reflect.ValueOf(field); reflectValue.IsValid() &&
					!reflectValue.IsZero() {
					if reflectValue.Kind() == reflect.Ptr {
						reflectValue = reflectValue.Elem()
					}

					if reflectValue.Kind() == reflect.Struct {
						passwordField = reflectValue.FieldByName(passwordFieldName)
						if passwordField.IsValid() && !passwordField.IsZero() {
							passwordCopy = passwordField.String()
							passwordField.SetString("")
						}
					}
				}
			}

			logger, ok := logger.(*slog.Logger)
			if !ok {
				panic("error during conversion to grpc logger")
			}

			logger.Log(
				ctx,
				slog.Level(logLevel),
				msg,
				fields...,
			)

			// Returning password field for correct logics due to a pointer request:
			if passwordField.IsValid() {
				passwordField.SetString(passwordCopy)
			}
		},
	)
}
