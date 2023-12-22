package middleware

import "net/http"

type Middleware func(next http.Handler) http.Handler

var nopMiddleware = Middleware(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(writer, request)
	})
})
