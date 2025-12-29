package tracing_test

import (
	"context"
	"testing"

	"github.com/DKhorkov/libs/tracing"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestNew(t *testing.T) {
	t.Parallel()

	config := tracing.Config{
		JaegerURL:      "http://localhost:14268/api/traces",
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	provider, err := tracing.New(config)
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestShutdown(t *testing.T) {
	t.Parallel()

	config := tracing.Config{
		JaegerURL:      "http://localhost:14268/api/traces",
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	provider, err := tracing.New(config)
	require.NoError(t, err)

	err = provider.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestSpan(t *testing.T) {
	t.Parallel()

	config := tracing.Config{
		JaegerURL:      "http://localhost:14268/api/traces",
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	provider, err := tracing.New(config)
	require.NoError(t, err)

	ctx, span := provider.Span(context.Background(), "test-span")
	require.NotNil(t, span)
	require.NotEqual(t, trace.TraceID{}, trace.SpanContextFromContext(ctx).TraceID())
	span.End()
}

func TestSpanFromTraceID(t *testing.T) {
	t.Parallel()

	config := tracing.Config{
		JaegerURL:      "http://localhost:14268/api/traces",
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	provider, err := tracing.New(config)
	require.NoError(t, err)

	traceID, err := provider.TraceIDFromHex("1234567890abcdef1234567890abcdef")
	require.NoError(t, err)

	ctx, span := provider.SpanFromTraceID(context.Background(), traceID, "test-span")
	require.NotNil(t, span)
	require.Equal(t, traceID, trace.SpanContextFromContext(ctx).TraceID())
	span.End()
}

func TestTraceIDFromHexValid(t *testing.T) {
	t.Parallel()

	config := tracing.Config{
		JaegerURL:      "http://localhost:14268/api/traces",
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	provider, err := tracing.New(config)
	require.NoError(t, err)

	traceID, err := provider.TraceIDFromHex("1234567890abcdef1234567890abcdef")
	require.NoError(t, err)
	require.NotEqual(t, trace.TraceID{}, traceID)
}

func TestTraceIDFromHexInvalid(t *testing.T) {
	t.Parallel()

	config := tracing.Config{
		JaegerURL:      "http://localhost:14268/api/traces",
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	provider, err := tracing.New(config)
	require.NoError(t, err)

	_, err = provider.TraceIDFromHex("invalid-hex")
	require.Error(t, err)
}
