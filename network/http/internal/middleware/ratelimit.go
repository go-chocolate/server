package middleware

import (
	"net/http"
	"time"

	"github.com/juju/ratelimit"
)

func RateLimit(rateLimit int) Middleware {
	bucket := ratelimit.NewBucketWithRate(float64(rateLimit), 1<<16)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			select {
			case <-time.After(bucket.Take(1)):
				next.ServeHTTP(writer, request)
			case <-request.Context().Done():
				http.Error(writer, "timeout", http.StatusGatewayTimeout)
			}
		})
	}
}
