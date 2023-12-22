package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if recoveredError := recover(); recoveredError != nil {
					logger := logrus.WithContext(request.Context())
					stack := debug.Stack()
					traceId := writer.Header().Get(headerTraceId)
					logger.WithField("panic_trace_id", traceId).Errorf(string(stack))
				}
			}()
			next.ServeHTTP(writer, request)
		})
	}
}
