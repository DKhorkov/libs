package tracing_test

import (
	"context"
	"testing"

	"github.com/DKhorkov/libs/tracing"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

// testConfig is a config of a collector, which is not running.
//
// It is intentional: see TestNewDoesNotDialCollector.
func testConfig() tracing.Config {
	return tracing.Config{
		CollectorURL:   "127.0.0.1:4317",
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Insecure:       true,
	}
}

// TestNewDoesNotDialCollector pins the laziness of the exporter.
//
// The collector starts slower than the application, and on a developer machine
// it may not start at all. An exporter, which dials on creation, would fail the
// whole server for a missing trace collector — that is, would make observability
// a reason to be unobservable.
//
// otlptracegrpc dials lazily by default. The test pins that property, so that
// adding WithDialOption(grpc.WithBlock()) becomes a red test instead of a
// production incident.
func TestNewDoesNotDialCollector(t *testing.T) {
	t.Parallel()

	provider, err := tracing.New(testConfig())
	require.NoError(t, err)
	require.NotNil(t, provider)

	require.NoError(t, provider.Shutdown(context.Background()))
}

func TestNew(t *testing.T) {
	t.Parallel()

	provider, err := tracing.New(testConfig())
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestShutdown(t *testing.T) {
	t.Parallel()

	provider, err := tracing.New(testConfig())
	require.NoError(t, err)

	require.NoError(t, provider.Shutdown(context.Background()))
}

func TestSpan(t *testing.T) {
	t.Parallel()

	provider, err := tracing.New(testConfig())
	require.NoError(t, err)

	ctx, span := provider.Span(context.Background(), "test-span")
	require.NotNil(t, span)
	require.NotEqual(t, trace.TraceID{}, trace.SpanContextFromContext(ctx).TraceID())
	span.End()
}

func TestSpanFromTraceID(t *testing.T) {
	t.Parallel()

	provider, err := tracing.New(testConfig())
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

	provider, err := tracing.New(testConfig())
	require.NoError(t, err)

	traceID, err := provider.TraceIDFromHex("1234567890abcdef1234567890abcdef")
	require.NoError(t, err)
	require.NotEqual(t, trace.TraceID{}, traceID)
}

func TestTraceIDFromHexInvalid(t *testing.T) {
	t.Parallel()

	provider, err := tracing.New(testConfig())
	require.NoError(t, err)

	_, err = provider.TraceIDFromHex("invalid-hex")
	require.Error(t, err)
}
