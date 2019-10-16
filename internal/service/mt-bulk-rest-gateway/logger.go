package mtbulkrestapi

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

// LogMiddleware logs requests and sets up job's unique id.
func (mtbulk *MTbulkRESTGateway) LogMiddleware(ctx context.Context) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")

			id := xid.New().String()

			rCtx, _ := context.WithCancel(r.Context())
			rCtx = context.WithValue(rCtx, "id", id)
			r = r.WithContext(rCtx)

			mtbulk.sugar.Infow("request", "method", r.Method, "url", r.URL, "remote addr", r.RemoteAddr, "id", id)
			next.ServeHTTP(w, r)
		})
	}
}
