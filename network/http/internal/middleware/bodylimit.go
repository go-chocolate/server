package middleware

import (
	"fmt"
	"io"
	"net/http"
)

type limitReaderCloser struct {
	rc    io.Closer
	limit io.Reader
}

func (l *limitReaderCloser) Read(p []byte) (int, error) {
	return l.limit.Read(p)
}

func (l *limitReaderCloser) Close() error {
	return l.rc.Close()
}

func BodyLimit(limit int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if limit == 0 {
				next.ServeHTTP(writer, request)
				return
			}
			if request.ContentLength > limit {
				err := fmt.Sprintf("request entity too large, limit is %d, but got %d", limit, request.ContentLength)
				http.Error(writer, err, http.StatusRequestEntityTooLarge)
				return
			}

			request.Body = &limitReaderCloser{
				rc:    request.Body,
				limit: io.LimitReader(request.Body, limit),
			}
			next.ServeHTTP(writer, request)
		})
	}
}
