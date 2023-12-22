package logger

import (
	"net/http"
)

type logWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *logWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

func (w *logWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *logWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *logWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *logWriter) Length() int {
	return w.length
}
