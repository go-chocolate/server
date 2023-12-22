package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout
// TODO
func Timeout(duration time.Duration) Middleware {
	if duration == 0 {
		return nopMiddleware
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			timeout, cancel := context.WithTimeout(ctx, duration)
			defer cancel()
			request = request.WithContext(timeout)
			ch := make(chan struct{})
			go run(next, writer, request, ch)
			select {
			case <-ctx.Done():
				http.Error(writer, "timeout", http.StatusGatewayTimeout)
			case <-ch:
				return
			}
		})
	}
}

func run(next http.Handler, w http.ResponseWriter, r *http.Request, ch chan struct{}) {
	defer func() {
		ch <- struct{}{}
	}()
	next.ServeHTTP(w, r)
}
