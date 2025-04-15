package interceptors_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/DKhorkov/libs/grpc/interceptors"
	"github.com/DKhorkov/libs/logging"
	mocklogging "github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/requestid"
)

type testRequest struct {
	Username string
	Password string
}

func TestUnaryServerLoggingInterceptor(t *testing.T) {
	t.Run("Without requestID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		logger := mocklogging.NewMockLogger(ctrl)

		req := struct{ Username string }{Username: "user"}
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
		ctx := context.Background()

		logger.
			EXPECT().
			InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1)

		interceptor := interceptors.UnaryServerLoggingInterceptor(logger)

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		resp, err := interceptor(ctx, req, info, handler)
		require.NoError(t, err)
		require.Equal(t, "response", resp)
	})

	t.Run("With error from handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		logger := mocklogging.NewMockLogger(ctrl)

		requestID := "test-request-id"
		req := struct{ Username string }{Username: "user"}
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(strings.ToLower(requestid.Key), requestID))

		logger.
			EXPECT().
			InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1)

		interceptor := interceptors.UnaryServerLoggingInterceptor(logger)

		handler := func(ctx context.Context, req any) (any, error) {
			return nil, errors.New("handler error")
		}

		resp, err := interceptor(ctx, req, info, handler)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("Non-struct request", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		logger := mocklogging.NewMockLogger(ctrl)

		requestID := "test-request-id"
		req := "non-struct-request"
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(strings.ToLower(requestid.Key), requestID))

		interceptor := interceptors.UnaryServerLoggingInterceptor(logger)

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		require.Panics(t, func() {
			interceptor(ctx, req, info, handler)
		})
	})
}

func TestUnaryClientLoggingInterceptor(t *testing.T) {
	t.Run("Log without password field", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		logger := logging.Logger(slogLogger)
		clientLogger := interceptors.UnaryClientLoggingInterceptor(logger)

		ctx := context.Background()
		req := struct{ Username string }{Username: "user"}
		clientLogger.Log(ctx, grpclogging.LevelDebug, "client request", "request", req)

		logOutput := buf.String()
		require.Contains(t, logOutput, `"level":"DEBUG"`)
		require.Contains(t, logOutput, `"msg":"client request"`)
		require.Contains(t, logOutput, `"request":{"Username":"user"}`)
	})

	t.Run("Panic on invalid logger type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		logger := mocklogging.NewMockLogger(ctrl)
		clientLogger := interceptors.UnaryClientLoggingInterceptor(logger)

		ctx := context.Background()
		req := testRequest{Username: "user", Password: "secret"}

		require.Panics(t, func() {
			clientLogger.Log(ctx, grpclogging.LevelDebug, "client request", "request", req)
		})
	})

	t.Run("Non-struct field", func(t *testing.T) {
		var buf bytes.Buffer
		slogLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		logger := logging.Logger(slogLogger)
		clientLogger := interceptors.UnaryClientLoggingInterceptor(logger)

		ctx := context.Background()
		req := "non-struct-field"
		clientLogger.Log(ctx, grpclogging.LevelDebug, "client request", "field", req)

		logOutput := buf.String()
		require.Contains(t, logOutput, `"level":"DEBUG"`)
		require.Contains(t, logOutput, `"msg":"client request"`)
		require.Contains(t, logOutput, `"field":"non-struct-field"`)
	})
}
