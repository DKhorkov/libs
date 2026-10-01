// Package tracing provides tools for comfort usage of tracing.
//
// Spans are exported over OTLP gRPC (port 4317 by default), which every modern
// collector speaks — Jaeger v2, the OpenTelemetry Collector, Tempo. The former
// Jaeger v1 exporter is gone: it was removed from OTel Go, and the v1 branch of
// Jaeger is out of support.
//
// The exporter dials lazily. A collector, which is not up yet, does not fail
// the application start.
package tracing
