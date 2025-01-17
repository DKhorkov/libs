package tracing

import "go.opentelemetry.io/otel/trace"

// Config represents tracing setup config.
type Config struct {
	ServiceName    string
	ServiceVersion string
	JaegerURL      string
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
