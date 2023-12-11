package middleware

import (
	"context"
	"fmt"
	"net/http"

	cmotel "github.com/coordimap/cm-otel-go"
	"go.opentelemetry.io/otel"
)

// CoordimapMiddleware initiates the cmOtel object and creates the first span that holds information about the endpoint being called.
func CoordimapMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var cmOtel cmotel.CMOtel

		prefix := cmotel.GetEnvWithPrefix("", cmotel.EnvCmPrefix)
		cmOtel = cmotel.New(
			otel.Tracer(cmotel.GetEnvWithPrefix(prefix, cmotel.EnvTracerName)),
			cmotel.GetEnvWithPrefix(prefix, cmotel.EnvServiceName),
		)

		traceParentsMap, errTraceParentsMap := cmotel.UnmarshalToSpanMap(r.Header.Get(cmotel.EnvTraceParentsMapHeaderName))
		if errTraceParentsMap == nil {
			for key, val := range traceParentsMap {
				if errSet := cmOtel.SetSpanFromTraceparent(key, val); errSet != nil {
					fmt.Printf("could not set span %s from traceparent because %s", key, errSet.Error())
				}
			}
		}

		ctx := context.WithValue(r.Context(), cmotel.ContextKey, cmOtel)
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}
