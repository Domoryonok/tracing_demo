package tracing

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(next http.Handler, tracer trace.Tracer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

		// if there is a context inside carrier we have to extract it and update our local context
		otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// start new span
		ctx, span := tracer.Start(r.Context(), spanName)
		defer span.End()

		// add the new context into the request
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
