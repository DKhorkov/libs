package tracing

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/DKhorkov/libs/middlewares/http/intercepting_response_writer"
	"github.com/DKhorkov/libs/middlewares/http/route"
	"github.com/DKhorkov/libs/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/metadata"
)

// Middleware создаёт корневой спан запроса с именем "{METHOD} {маршрут}" и
// пишет события его начала и конца.
//
// Маршрут — шаблон (WithRoute; по умолчанию — шаблон ServeMux), а не путь:
// с путём каждая операция с идентификатором была бы отдельной операцией в
// Jaeger и отдельной серией в span_metrics.
func Middleware(
	tp tracing.Provider,
	spanConfig tracing.SpanConfig,
	opts ...Option,
) func(next http.Handler) http.Handler {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if o.skip(r) {
				next.ServeHTTP(w, r)

				return
			}

			routeOf := o.route(r)
			method := route.Method(r)

			ctx, span := tp.Span(r.Context(), spanName(method, routeOf), spanConfig.Opts...)
			defer span.End()

			span.SetAttributes(
				attribute.String("http.request.method", method),
				attribute.String("http.route", routeOf),
			)

			span.AddEvent(spanConfig.Events.Start.Name, spanConfig.Events.Start.Opts...)
			defer span.AddEvent(spanConfig.Events.End.Name, spanConfig.Events.End.Opts...)

			traceID := span.SpanContext().TraceID().String()
			ctx = metadata.AppendToOutgoingContext(ctx, tracing.Key, traceID)
			r = r.WithContext(ctx)

			// Паника — ошибка корневого спана: иначе он закрылся бы со статусом
			// Unset, и доля ошибок span_metrics её не видела. Паника не
			// глотается; http.ErrAbortHandler — штатный обрыв ответа.
			defer func() {
				if recovered := recover(); recovered != nil {
					if err, isErr := recovered.(error); !isErr ||
						!errors.Is(err, http.ErrAbortHandler) {
						span.SetStatus(tracing.StatusError, fmt.Sprint("panic: ", recovered))
					}

					panic(recovered)
				}
			}()

			rw := intercepting_response_writer.New(w)
			next.ServeHTTP(rw, r)

			span.SetAttributes(attribute.Int("http.response.status_code", rw.StatusCode))

			if rw.StatusCode >= http.StatusBadRequest {
				span.SetStatus(tracing.StatusError, string(rw.Body))
			}
		})
	}
}

// spanName — "{METHOD} {маршрут}" по семантическим соглашениям OTel; без
// маршрута — только метод. Метод уже нормализован (route.Method).
func spanName(method, routeOf string) string {
	if routeOf == "" || routeOf == route.Other {
		return method
	}

	return method + " " + routeOf
}
