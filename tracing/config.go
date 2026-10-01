package tracing

import "go.opentelemetry.io/otel/trace"

// Config represents tracing setup config.
type Config struct {
	// ServiceName is a name, under which spans are grouped in a collector UI.
	ServiceName string

	// ServiceVersion is a version of the service, attached to every span.
	ServiceVersion string

	// CollectorURL is a host:port of an OTLP gRPC endpoint (4317 by default in
	// every collector).
	//
	// No scheme and no path: unlike the removed Jaeger exporter, which took a
	// full HTTP URL of a v1 collector, OTLP over gRPC takes an address. A value
	// with "http://" in it fails at the first export and nowhere earlier.
	CollectorURL string

	// Insecure disables TLS on the connection to a collector.
	//
	// True for a collector inside the same docker network or on localhost,
	// false for anything reachable from outside.
	Insecure bool
}

// SpanConfig is needed to configure creation of new span.
type SpanConfig struct {
	Name   string
	Opts   []trace.SpanStartOption
	Events SpanEventsConfig
}

// SpanEventsConfig is needed to configure creation of span events.
type SpanEventsConfig struct {
	Start SpanEventConfig
	End   SpanEventConfig
}

// SpanEventConfig is needed to configure creation of single span event.
type SpanEventConfig struct {
	Name string
	Opts []trace.EventOption
}
