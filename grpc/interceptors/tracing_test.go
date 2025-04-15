package interceptors_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/DKhorkov/libs/grpc/interceptors"
	"github.com/DKhorkov/libs/tracing"
	mocktracing "github.com/DKhorkov/libs/tracing/mocks"
)

func TestUnaryServerTracingInterceptor(t *testing.T) {
	t.Run("With valid traceID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()

		traceID := trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
		traceHex := traceID.String()
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(strings.ToLower(tracing.Key), traceHex))
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
		spanConfig := tracing.SpanConfig{
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start", Opts: []trace.EventOption{}},
				End:   tracing.SpanEventConfig{Name: "end", Opts: []trace.EventOption{}},
			},
			Opts: []trace.SpanStartOption{},
		}

		// Настраиваем моки
		provider.
			EXPECT().
			TraceIDFromHex(traceHex).
			Return(traceID, nil).
			Times(1)

		provider.
			EXPECT().
			SpanFromTraceID(gomock.Any(), traceID, info.FullMethod, spanConfig.Opts).
			Return(ctx, span).
			Times(1)

		interceptor := interceptors.UnaryServerTracingInterceptor(provider, spanConfig)

		// Моделируем handler
		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		resp, err := interceptor(ctx, nil, info, handler)
		require.NoError(t, err)
		require.Equal(t, "response", resp)
	})

	t.Run("With invalid traceID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(strings.ToLower(tracing.Key), "invalid-trace"))
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
		spanConfig := tracing.SpanConfig{
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start", Opts: []trace.EventOption{}},
				End:   tracing.SpanEventConfig{Name: "end", Opts: []trace.EventOption{}},
			},
		}

		provider.
			EXPECT().
			TraceIDFromHex("invalid-trace").
			Return(trace.TraceID{}, errors.New("invalid trace ID")).
			Times(1)

		interceptor := interceptors.UnaryServerTracingInterceptor(provider, spanConfig)

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		resp, err := interceptor(ctx, nil, info, handler)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Equal(t, "invalid trace ID", err.Error())
	})

	t.Run("Without traceID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		ctx := context.Background()
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
		spanConfig := tracing.SpanConfig{
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start", Opts: []trace.EventOption{}},
				End:   tracing.SpanEventConfig{Name: "end", Opts: []trace.EventOption{}},
			},
		}

		interceptor := interceptors.UnaryServerTracingInterceptor(provider, spanConfig)

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		resp, err := interceptor(ctx, nil, info, handler)
		require.NoError(t, err)
		require.Equal(t, "response", resp)
	})

	t.Run("With handler error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()

		traceID := trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
		traceHex := traceID.String()
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(strings.ToLower(tracing.Key), traceHex))
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
		spanConfig := tracing.SpanConfig{
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start", Opts: []trace.EventOption{}},
				End:   tracing.SpanEventConfig{Name: "end", Opts: []trace.EventOption{}},
			},
			Opts: []trace.SpanStartOption{},
		}

		// Настраиваем моки
		provider.
			EXPECT().
			TraceIDFromHex(traceHex).
			Return(traceID, nil).
			Times(1)

		provider.
			EXPECT().
			SpanFromTraceID(gomock.Any(), traceID, info.FullMethod, spanConfig.Opts).
			Return(ctx, span).
			Times(1)

		interceptor := interceptors.UnaryServerTracingInterceptor(provider, spanConfig)

		handler := func(ctx context.Context, req any) (any, error) {
			return nil, errors.New("handler error")
		}

		resp, err := interceptor(ctx, nil, info, handler)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Equal(t, "handler error", err.Error())
	})
}

func TestUnaryClientTracingInterceptor(t *testing.T) {
	t.Run("Successful invoker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()

		ctx := context.Background()
		method := "/test.Service/Method"
		spanConfig := tracing.SpanConfig{
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start", Opts: []trace.EventOption{}},
				End:   tracing.SpanEventConfig{Name: "end", Opts: []trace.EventOption{}},
			},
			Opts: []trace.SpanStartOption{},
		}

		provider.
			EXPECT().
			Span(ctx, method, gomock.Any()).
			Return(ctx, span).
			Times(1)

		interceptor := interceptors.UnaryClientTracingInterceptor(provider, spanConfig)

		invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			// Присваиваем значение через указатель
			*reply.(*string) = "response"
			return nil
		}

		var reply string
		err := interceptor(ctx, method, nil, &reply, nil, invoker)
		require.NoError(t, err)
		require.Equal(t, "response", reply)
	})

	t.Run("Invoker error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := mocktracing.NewMockProvider(ctrl)
		span := mocktracing.NewMockSpan()

		ctx := context.Background()
		method := "/test.Service/Method"
		spanConfig := tracing.SpanConfig{
			Events: tracing.SpanEventsConfig{
				Start: tracing.SpanEventConfig{Name: "start", Opts: []trace.EventOption{}},
				End:   tracing.SpanEventConfig{Name: "end", Opts: []trace.EventOption{}},
			},
		}

		provider.
			EXPECT().
			Span(ctx, method, gomock.Any()).
			Return(ctx, span).
			Times(1)

		interceptor := interceptors.UnaryClientTracingInterceptor(provider, spanConfig)

		invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return errors.New("invoker error")
		}

		var reply string
		err := interceptor(ctx, method, nil, &reply, nil, invoker)
		require.Error(t, err)
		require.Equal(t, "invoker error", err.Error())
	})
}
