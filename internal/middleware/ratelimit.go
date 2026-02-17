package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

func RateLimiter(requestsPerMinute int) func(http.Handler) http.Handler {
	return httprate.Limit(
		requestsPerMinute,
		time.Minute,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"success":false,"error":{"code":"RATE_LIMITED","message":"Too many requests. Please try again later."}}`))
		}),
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	)
}
