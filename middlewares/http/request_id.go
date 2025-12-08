package http

import (
	"net/http"

	"google.golang.org/grpc/metadata"

	"github.com/DKhorkov/libs/contextlib"
	"github.com/DKhorkov/libs/requestid"
)

// RequestIDMiddleware generates request ID and paste it to provided context for later usage.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := requestid.New()
		ctx = contextlib.WithValue(
			ctx,
			requestid.Key,
			requestID,
		) // setting for inner usage
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			requestid.Key,
			requestID,
		) // setting for cross-service usage
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
